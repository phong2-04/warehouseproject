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

	maxUsernameFailures   = 10
	usernameFailureWindow = 15 * time.Minute
	usernameLockDuration  = 5 * time.Minute
	maxTrackedUsernames   = 10000
)

type loginAttempt struct {
	failures    int
	windowStart time.Time
	lockedUntil time.Time
	lastSeen    time.Time
}

type loginAttemptLimiter struct {
	mu          sync.Mutex
	attempts    map[string]*loginAttempt
	maxFailures int
	window      time.Duration
	lock        time.Duration
	maxTracked  int
}

func newLoginAttemptLimiter() *loginAttemptLimiter {
	return newConfiguredLoginAttemptLimiter(
		maxLoginFailures,
		loginFailureWindow,
		loginLockDuration,
		maxTrackedLoginIPs,
	)
}

func newUsernameAttemptLimiter() *loginAttemptLimiter {
	return newConfiguredLoginAttemptLimiter(
		maxUsernameFailures,
		usernameFailureWindow,
		usernameLockDuration,
		maxTrackedUsernames,
	)
}

func newConfiguredLoginAttemptLimiter(
	maxFailures int,
	window time.Duration,
	lock time.Duration,
	maxTracked int,
) *loginAttemptLimiter {
	return &loginAttemptLimiter{
		attempts:    make(map[string]*loginAttempt),
		maxFailures: maxFailures,
		window:      window,
		lock:        lock,
		maxTracked:  maxTracked,
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

// Tra ve so giay con lai neu key dang bi khoa.
func (l *loginAttemptLimiter) retryAfter(key string, now time.Time) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	attempt, exists := l.attempts[key]
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
		now.Sub(attempt.windowStart) >= l.window {
		delete(l.attempts, key)
	}

	return 0
}

// Ghi nhan mot lan dang nhap sai.
func (l *loginAttemptLimiter) recordFailure(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Don cac key da het thoi gian theo doi.
	for trackedKey, attempt := range l.attempts {
		if !attempt.lockedUntil.After(now) &&
			now.Sub(attempt.windowStart) >= l.window {
			delete(l.attempts, trackedKey)
		}
	}

	attempt, exists := l.attempts[key]

	if exists && attempt.lockedUntil.After(now) {
		return
	}

	if !exists {
		// Gioi han bo nho bang cach xoa key cu nhat.
		if len(l.attempts) >= l.maxTracked {
			var oldestKey string
			var oldestTime time.Time

			for trackedKey, item := range l.attempts {
				if oldestKey == "" || item.lastSeen.Before(oldestTime) {
					oldestKey = trackedKey
					oldestTime = item.lastSeen
				}
			}

			if oldestKey != "" {
				delete(l.attempts, oldestKey)
			}
		}

		attempt = &loginAttempt{
			windowStart: now,
			lastSeen:    now,
		}
		l.attempts[key] = attempt
	} else if now.Sub(attempt.windowStart) >= l.window {
		attempt = &loginAttempt{
			windowStart: now,
			lastSeen:    now,
		}
		l.attempts[key] = attempt
	}

	attempt.failures++
	attempt.lastSeen = now

	if attempt.failures >= l.maxFailures {
		attempt.lockedUntil = now.Add(l.lock)
	}
}

// Xoa bo dem khi dang nhap thanh cong.
func (l *loginAttemptLimiter) reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.attempts, key)
}
