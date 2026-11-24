package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"warehouseproject/internal/middleware"
	"warehouseproject/internal/service"
)

type fakeAuthUserService struct {
	loginResult   *service.LoginResponse
	loginErr      error
	refreshResult *service.RefreshResponse
	refreshErr    error
	logoutErr     error

	changePasswordErr    error
	changePasswordCalls  int
	changePasswordUserID int64
	currentPassword      string
	newPassword          string
	loginCalls           int
	refreshCalls         int
	logoutCalls          int
	username             string
	password             string
	refreshToken         string
}

func (f *fakeAuthUserService) Login(ctx context.Context, username, password string) (*service.LoginResponse, error) {
	f.loginCalls++
	f.username = username
	f.password = password
	return f.loginResult, f.loginErr
}

func (f *fakeAuthUserService) RefreshAccessToken(ctx context.Context, token string) (*service.RefreshResponse, error) {
	f.refreshCalls++
	f.refreshToken = token
	return f.refreshResult, f.refreshErr
}

func (f *fakeAuthUserService) Logout(ctx context.Context, token string) error {
	f.logoutCalls++
	f.refreshToken = token
	return f.logoutErr
}

func (f *fakeAuthUserService) ChangePassword(
	ctx context.Context,
	userID int64,
	currentPassword, newPassword string,
) error {
	f.changePasswordCalls++
	f.changePasswordUserID = userID
	f.currentPassword = currentPassword
	f.newPassword = newPassword
	return f.changePasswordErr
}
func newAuthTestRequest(method, body string) (*httptest.ResponseRecorder, *http.Request) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(method, "/api/test", strings.NewReader(body))
	req.RemoteAddr = "192.168.1.100:12345"
	return recorder, req
}

func TestAuthHandlerLogin(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		loginErr   error
		wantStatus int
		wantCalls  int
	}{
		{"method khong hop le", http.MethodGet, `{}`, nil, http.StatusMethodNotAllowed, 0},
		{"JSON khong hop le", http.MethodPost, `{`, nil, http.StatusBadRequest, 0},
		{"username rong", http.MethodPost, `{"username":" ","password":"123"}`, nil, http.StatusBadRequest, 0},
		{"password rong", http.MethodPost, `{"username":"test","password":""}`, nil, http.StatusBadRequest, 0},
		{"sai thong tin dang nhap", http.MethodPost, `{"username":"test","password":"wrong"}`, service.ErrInvalidCredentials, http.StatusUnauthorized, 1},
		{"thieu JWT secret", http.MethodPost, `{"username":"test","password":"123"}`, service.ErrJWTSecretRequired, http.StatusInternalServerError, 1},
		{"loi he thong", http.MethodPost, `{"username":"test","password":"123"}`, errors.New("database error"), http.StatusInternalServerError, 1},
		{"dang nhap thanh cong", http.MethodPost, `{"username":" test ","password":"123"}`, nil, http.StatusOK, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeAuthUserService{
				loginErr: tt.loginErr,
				loginResult: &service.LoginResponse{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token",
					TokenType:    "Bearer",
					ExpiresIn:    900,
				},
			}

			handler := NewAuthHandler(fake)
			recorder, req := newAuthTestRequest(tt.method, tt.body)
			handler.Login(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, muon %d; body: %s", recorder.Code, tt.wantStatus, recorder.Body.String())
			}

			if fake.loginCalls != tt.wantCalls {
				t.Errorf("so lan goi Login = %d, muon %d", fake.loginCalls, tt.wantCalls)
			}

			if tt.name == "dang nhap thanh cong" && fake.username != "test" {
				t.Errorf("username chua duoc trim: %q", fake.username)
			}
		})
	}
}

func TestAuthHandlerLoginRateLimit(t *testing.T) {
	fake := &fakeAuthUserService{
		loginErr: service.ErrInvalidCredentials,
	}
	handler := NewAuthHandler(fake)

	for i := 0; i < maxLoginFailures; i++ {
		recorder, req := newAuthTestRequest(
			http.MethodPost,
			`{"username":"test","password":"wrong"}`,
		)
		handler.Login(recorder, req)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("lan %d: status = %d, muon 401", i+1, recorder.Code)
		}
	}

	recorder, req := newAuthTestRequest(
		http.MethodPost,
		`{"username":"test","password":"wrong"}`,
	)
	handler.Login(recorder, req)

	if recorder.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, muon 429", recorder.Code)
	}

	if recorder.Header().Get("Retry-After") == "" {
		t.Error("thieu header Retry-After")
	}

	if fake.loginCalls != maxLoginFailures {
		t.Errorf("Login duoc goi %d lan, muon %d", fake.loginCalls, maxLoginFailures)
	}
}

func TestAuthHandlerLoginSuccessResetsLimiter(t *testing.T) {
	fake := &fakeAuthUserService{}
	handler := NewAuthHandler(fake)

	for i := 0; i < maxLoginFailures-1; i++ {
		fake.loginErr = service.ErrInvalidCredentials
		recorder, req := newAuthTestRequest(http.MethodPost, `{"username":"test","password":"wrong"}`)
		handler.Login(recorder, req)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("lan %d: status = %d, muon 401", i+1, recorder.Code)
		}
	}

	fake.loginErr = nil
	recorder, req := newAuthTestRequest(http.MethodPost, `{"username":"test","password":"correct"}`)
	handler.Login(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("dang nhap thanh cong: status = %d", recorder.Code)
	}

	fake.loginErr = service.ErrInvalidCredentials
	recorder, req = newAuthTestRequest(http.MethodPost, `{"username":"test","password":"wrong"}`)
	handler.Login(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("limiter khong duoc reset sau dang nhap thanh cong, status = %d", recorder.Code)
	}
}

func TestAuthHandlerRefreshToken(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		refreshErr error
		wantStatus int
		wantCalls  int
	}{
		{"method khong hop le", http.MethodGet, `{}`, nil, http.StatusMethodNotAllowed, 0},
		{"JSON khong hop le", http.MethodPost, `{`, nil, http.StatusBadRequest, 0},
		{"token rong", http.MethodPost, `{"refresh_token":" "}`, nil, http.StatusBadRequest, 0},
		{"token khong hop le", http.MethodPost, `{"refresh_token":"bad"}`, service.ErrInvalidRefreshToken, http.StatusUnauthorized, 1},
		{"thieu JWT secret", http.MethodPost, `{"refresh_token":"token"}`, service.ErrJWTSecretRequired, http.StatusInternalServerError, 1},
		{"loi he thong", http.MethodPost, `{"refresh_token":"token"}`, errors.New("database error"), http.StatusInternalServerError, 1},
		{"thanh cong", http.MethodPost, `{"refresh_token":" token "}`, nil, http.StatusOK, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeAuthUserService{
				refreshErr: tt.refreshErr,
				refreshResult: &service.RefreshResponse{
					AccessToken:  "new-access",
					RefreshToken: "new-refresh",
					TokenType:    "Bearer",
					ExpiresIn:    900,
				},
			}

			handler := NewAuthHandler(fake)
			recorder, req := newAuthTestRequest(tt.method, tt.body)
			handler.RefreshToken(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, muon %d; body: %s", recorder.Code, tt.wantStatus, recorder.Body.String())
			}

			if fake.refreshCalls != tt.wantCalls {
				t.Errorf("so lan goi RefreshAccessToken = %d, muon %d", fake.refreshCalls, tt.wantCalls)
			}

			if tt.name == "thanh cong" && fake.refreshToken != "token" {
				t.Errorf("refresh token chua duoc trim: %q", fake.refreshToken)
			}
		})
	}
}

func TestAuthHandlerLogout(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		logoutErr  error
		wantStatus int
		wantCalls  int
	}{
		{"method khong hop le", http.MethodGet, `{}`, nil, http.StatusMethodNotAllowed, 0},
		{"JSON khong hop le", http.MethodPost, `{`, nil, http.StatusBadRequest, 0},
		{"token rong", http.MethodPost, `{"refresh_token":""}`, nil, http.StatusBadRequest, 0},
		{"token khong hop le", http.MethodPost, `{"refresh_token":"bad"}`, service.ErrInvalidRefreshToken, http.StatusUnauthorized, 1},
		{"thieu JWT secret", http.MethodPost, `{"refresh_token":"token"}`, service.ErrJWTSecretRequired, http.StatusInternalServerError, 1},
		{"loi he thong", http.MethodPost, `{"refresh_token":"token"}`, errors.New("database error"), http.StatusInternalServerError, 1},
		{"thanh cong", http.MethodPost, `{"refresh_token":" token "}`, nil, http.StatusOK, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeAuthUserService{logoutErr: tt.logoutErr}
			handler := NewAuthHandler(fake)

			recorder, req := newAuthTestRequest(tt.method, tt.body)
			handler.Logout(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, muon %d; body: %s", recorder.Code, tt.wantStatus, recorder.Body.String())
			}

			if fake.logoutCalls != tt.wantCalls {
				t.Errorf("so lan goi Logout = %d, muon %d", fake.logoutCalls, tt.wantCalls)
			}

			if tt.name == "thanh cong" && fake.refreshToken != "token" {
				t.Errorf("refresh token chua duoc trim: %q", fake.refreshToken)
			}
		})
	}
}

func TestAuthHandlerUsernameRateLimitAcrossIPs(t *testing.T) {
	fake := &fakeAuthUserService{
		loginErr: service.ErrInvalidCredentials,
	}
	handler := NewAuthHandler(fake)

	// Cung mot username dang nhap sai tu 10 IP khac nhau.
	for i := 1; i <= maxUsernameFailures; i++ {
		recorder, req := newAuthTestRequest(
			http.MethodPost,
			`{"username":"target-user","password":"wrong"}`,
		)
		req.RemoteAddr = "192.168.1." + strconv.Itoa(i) + ":12345"

		handler.Login(recorder, req)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("lan %d: status = %d, muon 401", i, recorder.Code)
		}
	}

	if fake.loginCalls != maxUsernameFailures {
		t.Fatalf("Login duoc goi %d lan, muon %d", fake.loginCalls, maxUsernameFailures)
	}

	// IP moi van khong the thu dang nhap username da bi khoa.
	recorder, req := newAuthTestRequest(
		http.MethodPost,
		`{"username":"target-user","password":"wrong"}`,
	)
	req.RemoteAddr = "192.168.2.1:12345"
	handler.Login(recorder, req)

	if recorder.Code != http.StatusTooManyRequests {
		t.Errorf("username bi khoa: status = %d, muon 429", recorder.Code)
	}

	if recorder.Header().Get("Retry-After") == "" {
		t.Error("thieu Retry-After khi username bi khoa")
	}

	if fake.loginCalls != maxUsernameFailures {
		t.Errorf("Login van duoc goi khi username bi khoa: %d lan", fake.loginCalls)
	}

	// Username khac van dang nhap duoc, khong bi khoa theo username cu.
	recorder, req = newAuthTestRequest(
		http.MethodPost,
		`{"username":"other-user","password":"wrong"}`,
	)
	req.RemoteAddr = "192.168.2.2:12345"
	handler.Login(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("username khac: status = %d, muon 401", recorder.Code)
	}

	if fake.loginCalls != maxUsernameFailures+1 {
		t.Errorf("Login duoc goi %d lan, muon %d", fake.loginCalls, maxUsernameFailures+1)
	}
}

func TestUsernameLimiterUnlocksAfterFiveMinutes(t *testing.T) {
	limiter := newUsernameAttemptLimiter()
	now := time.Now()

	for i := 0; i < maxUsernameFailures; i++ {
		limiter.recordFailure("target-user", now)
	}

	if got := limiter.retryAfter("target-user", now); got <= 0 {
		t.Fatal("username phai bi khoa sau 10 lan dang nhap sai")
	}

	unlockTime := now.Add(usernameLockDuration)
	if got := limiter.retryAfter("target-user", unlockTime); got != 0 {
		t.Errorf("username van bi khoa sau 5 phut, Retry-After = %d", got)
	}
}

func TestAuthHandlerChangePassword(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		withUser   bool
		serviceErr error
		wantStatus int
		wantCalls  int
	}{
		{
			name:       "method khong hop le",
			method:     http.MethodGet,
			body:       `{"current_password":"oldpass123","new_password":"newpass123"}`,
			withUser:   true,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "JSON khong hop le",
			method:     http.MethodPost,
			body:       `{`,
			withUser:   true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "thieu mat khau",
			method:     http.MethodPost,
			body:       `{"current_password":"","new_password":"newpass123"}`,
			withUser:   true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "chua xac thuc",
			method:     http.MethodPost,
			body:       `{"current_password":"oldpass123","new_password":"newpass123"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "sai mat khau hien tai",
			method:     http.MethodPost,
			body:       `{"current_password":"wrongpass","new_password":"newpass123"}`,
			withUser:   true,
			serviceErr: service.ErrInvalidCurrentPassword,
			wantStatus: http.StatusUnauthorized,
			wantCalls:  1,
		},
		{
			name:       "mat khau moi khong hop le",
			method:     http.MethodPost,
			body:       `{"current_password":"oldpass123","new_password":"123"}`,
			withUser:   true,
			serviceErr: service.ErrInvalidNewPassword,
			wantStatus: http.StatusBadRequest,
			wantCalls:  1,
		},
		{
			name:       "mat khau moi trung mat khau cu",
			method:     http.MethodPost,
			body:       `{"current_password":"oldpass123","new_password":"oldpass123"}`,
			withUser:   true,
			serviceErr: service.ErrPasswordUnchanged,
			wantStatus: http.StatusBadRequest,
			wantCalls:  1,
		},
		{
			name:       "loi he thong",
			method:     http.MethodPost,
			body:       `{"current_password":"oldpass123","new_password":"newpass123"}`,
			withUser:   true,
			serviceErr: errors.New("database error"),
			wantStatus: http.StatusInternalServerError,
			wantCalls:  1,
		},
		{
			name:       "doi mat khau thanh cong",
			method:     http.MethodPost,
			body:       `{"current_password":"oldpass123","new_password":"newpass123"}`,
			withUser:   true,
			wantStatus: http.StatusOK,
			wantCalls:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeAuthUserService{
				changePasswordErr: tt.serviceErr,
			}
			handler := NewAuthHandler(fake)

			recorder, req := newAuthTestRequest(tt.method, tt.body)

			if tt.withUser {
				ctx := context.WithValue(
					req.Context(),
					middleware.UserIDKey,
					int64(7),
				)
				req = req.WithContext(ctx)
			}

			handler.ChangePassword(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf(
					"status = %d, muon %d; body: %s",
					recorder.Code,
					tt.wantStatus,
					recorder.Body.String(),
				)
			}

			if fake.changePasswordCalls != tt.wantCalls {
				t.Errorf(
					"so lan goi ChangePassword = %d, muon %d",
					fake.changePasswordCalls,
					tt.wantCalls,
				)
			}

			if tt.name == "doi mat khau thanh cong" {
				if fake.changePasswordUserID != 7 {
					t.Errorf("userID = %d, muon 7", fake.changePasswordUserID)
				}
				if fake.currentPassword != "oldpass123" {
					t.Errorf("currentPassword = %q", fake.currentPassword)
				}
				if fake.newPassword != "newpass123" {
					t.Errorf("newPassword = %q", fake.newPassword)
				}
			}
		})
	}
}
