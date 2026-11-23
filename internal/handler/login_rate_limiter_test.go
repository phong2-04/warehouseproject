package handler

import (
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestLoginAttemptLimiter_RetryAfter(t *testing.T) {
	limiter := newLoginAttemptLimiter()
	now := time.Now()
	ip := "192.168.1.10"

	if got := limiter.retryAfter(ip, now); got != 0 {
		t.Errorf("IP moi khong nen bi khoa, nhan duoc %d", got)
	}

	for i := 0; i < maxLoginFailures; i++ {
		limiter.recordFailure(ip, now)
	}

	got := limiter.retryAfter(ip, now)
	if got != int(loginLockDuration.Seconds()) {
		t.Errorf("thoi gian khoa khong dung, nhan duoc %d", got)
	}
}

func TestLoginAttemptLimiter_LockAfterFiveFailures(t *testing.T) {
	limiter := newLoginAttemptLimiter()
	now := time.Now()
	ip := "192.168.1.11"

	for i := 0; i < maxLoginFailures-1; i++ {
		limiter.recordFailure(ip, now)
	}

	if got := limiter.retryAfter(ip, now); got != 0 {
		t.Errorf("chua du 5 lan sai thi khong duoc khoa, nhan %d", got)
	}

	limiter.recordFailure(ip, now)

	if got := limiter.retryAfter(ip, now); got <= 0 {
		t.Error("du 5 lan sai phai bi khoa")
	}
}

func TestLoginAttemptLimiter_UnlockAfterDuration(t *testing.T) {
	limiter := newLoginAttemptLimiter()
	now := time.Now()
	ip := "192.168.1.12"

	for i := 0; i < maxLoginFailures; i++ {
		limiter.recordFailure(ip, now)
	}

	later := now.Add(loginLockDuration + time.Second)

	if got := limiter.retryAfter(ip, later); got != 0 {
		t.Errorf("het thoi gian khoa phai cho dang nhap lai, nhan %d", got)
	}
}

func TestLoginAttemptLimiter_Reset(t *testing.T) {
	limiter := newLoginAttemptLimiter()
	now := time.Now()
	ip := "192.168.1.13"

	for i := 0; i < maxLoginFailures; i++ {
		limiter.recordFailure(ip, now)
	}

	limiter.reset(ip)

	if got := limiter.retryAfter(ip, now); got != 0 {
		t.Errorf("reset phai xoa trang thai khoa, nhan %d", got)
	}
}

func TestLoginAttemptLimiter_FailureWindowExpires(t *testing.T) {
	limiter := newLoginAttemptLimiter()
	now := time.Now()
	ip := "192.168.1.14"

	for i := 0; i < maxLoginFailures-1; i++ {
		limiter.recordFailure(ip, now)
	}

	later := now.Add(loginFailureWindow + time.Second)

	if got := limiter.retryAfter(ip, later); got != 0 {
		t.Errorf("het cua so 15 phut phai duoc thu lai, nhan %d", got)
	}
}

func TestGetClientIP(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/login", nil)
	req.RemoteAddr = "192.168.1.20:12345"
	req.Header.Set("X-Forwarded-For", "10.0.0.1")

	got := getClientIP(req)

	if got != "192.168.1.20" {
		t.Errorf("IP khong dung: nhan %s", got)
	}
}

func TestGetClientIP_WithoutPort(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/login", nil)
	req.RemoteAddr = "192.168.1.21"

	got := getClientIP(req)

	if got != "192.168.1.21" {
		t.Errorf("IP khong dung: nhan %s", got)
	}
}

func TestLoginAttemptLimiter_MaxTrackedIPs(t *testing.T) {
	limiter := newLoginAttemptLimiter()
	now := time.Now()

	oldestIP := "oldest-ip"

	limiter.attempts[oldestIP] = &loginAttempt{
		windowStart: now,
		lastSeen:    now.Add(-24 * time.Hour),
	}

	for i := 1; i < maxTrackedLoginIPs; i++ {
		ip := "ip-" + strconv.Itoa(i)
		limiter.attempts[ip] = &loginAttempt{
			windowStart: now,
			lastSeen:    now.Add(-time.Duration(i) * time.Second),
		}
	}

	newIP := "new-ip"
	limiter.recordFailure(newIP, now)

	if len(limiter.attempts) > maxTrackedLoginIPs {
		t.Fatalf(
			"So IP vuot gioi han: %d, toi da %d",
			len(limiter.attempts),
			maxTrackedLoginIPs,
		)
	}

	if _, exists := limiter.attempts[oldestIP]; exists {
		t.Error("IP cu nhat phai bi xoa khi vuot gioi han")
	}

	if _, exists := limiter.attempts[newIP]; !exists {
		t.Error("IP moi phai duoc theo doi")
	}
}
