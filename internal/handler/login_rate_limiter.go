package handler

import (
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	maxLoginFailures   = 5
	loginFailureWindow = 15 * time.Minute
	loginLockDuration  = 15 * time.Minute
	maxTrackedLoginIPs = 10000
)

type loginAttempt struct {
	failures    int
	windowStart time.Time
	lockedUntil time.Time
	lastSeen    time.Time
}

type loginAttemptLimiter struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt
}

func newLoginAttemptLimiter() *loginAttemptLimiter {
	return &loginAttemptLimiter{
		attempts: make(map[string]*loginAttempt),
	}
}

// Lay IP tu ket noi truc tiep, khong tin header X-Forwarded-For.
func getClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}

// Tra ve so giay con lai neu IP dang bi khoa.
func (l *loginAttemptLimiter) retryAfter(ip string, now time.Time) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	attempt, exists := l.attempts[ip]
	if !exists {
		return 0
	}

	if attempt.lockedUntil.After(now) {
		attempt.lastSeen = now
		remaining := attempt.lockedUntil.Sub(now)
		seconds := int((remaining + time.Second - 1) / time.Second)
		if seconds < 1 {
			return 1
		}
		return seconds
	}

	if !attempt.lockedUntil.IsZero() ||
		now.Sub(attempt.windowStart) >= loginFailureWindow {
		delete(l.attempts, ip)
	}

	return 0
}

// Ghi nhan mot lan dang nhap sai.
func (l *loginAttemptLimiter) recordFailure(ip string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Don cac IP da het thoi gian theo doi.
	for key, attempt := range l.attempts {
		if !attempt.lockedUntil.After(now) &&
			now.Sub(attempt.windowStart) >= loginFailureWindow {
			delete(l.attempts, key)
		}
	}

	attempt, exists := l.attempts[ip]

	if exists && attempt.lockedUntil.After(now) {
		return
	}

	if !exists || now.Sub(attempt.windowStart) >= loginFailureWindow {
		attempt = &loginAttempt{
			windowStart: now,
		}
		l.attempts[ip] = attempt
	}

	// Gioi han so IP duoc theo doi trong bo nho.
	if !exists && len(l.attempts) > maxTrackedLoginIPs {
		var oldestIP string
		var oldestTime time.Time

		for key, item := range l.attempts {
			if oldestIP == "" || item.lastSeen.Before(oldestTime) {
				oldestIP = key
				oldestTime = item.lastSeen
			}
		}

		if oldestIP != "" && oldestIP != ip {
			delete(l.attempts, oldestIP)
		}
	}

	attempt.failures++
	attempt.lastSeen = now

	if attempt.failures >= maxLoginFailures {
		attempt.lockedUntil = now.Add(loginLockDuration)
	}
}

// Xoa bo dem khi dang nhap thanh cong.
func (l *loginAttemptLimiter) reset(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.attempts, ip)
}
