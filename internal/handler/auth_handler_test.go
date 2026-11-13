package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"warehouseproject/internal/service"
)

type fakeAuthUserService struct {
	loginResult   *service.LoginResponse
	loginErr      error
	refreshResult *service.RefreshResponse
	refreshErr    error
	logoutErr     error

	loginCalls   int
	refreshCalls int
	logoutCalls  int
	username     string
	password     string
	refreshToken string
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
