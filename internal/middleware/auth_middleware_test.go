package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func createTestToken(t *testing.T, secret string, method jwt.SigningMethod, claims jwt.MapClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(method, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Khong tao duoc token: %v", err)
	}

	return tokenString
}

func TestAuthMiddlewareRejectsInvalidTokens(t *testing.T) {
	secret := "test-secret"

	tests := []struct {
		name        string
		authHeader  string
		expectedMsg string
	}{
		{
			name:        "Thieu Authorization",
			authHeader:  "",
			expectedMsg: "Thieu token xac thuc",
		},
		{
			name:        "Sai dinh dang Bearer",
			authHeader:  "Basic abc123",
			expectedMsg: "Thieu token xac thuc",
		},
		{
			name:        "Bearer rong",
			authHeader:  "Bearer ",
			expectedMsg: "Token khong hop le",
		},
		{
			name:        "Token khong hop le",
			authHeader:  "Bearer abc123",
			expectedMsg: "Token khong hop le hoac da het han",
		},
		{
			name: "Refresh Token khong duoc chap nhan",
			authHeader: "Bearer " + createTestToken(t, secret,
				jwt.SigningMethodHS256,
				jwt.MapClaims{
					"token_type": "refresh",
					"user_id":    1,
				}),
			expectedMsg: "Can su dung Access Token",
		},
		{
			name: "Token thieu user_id",
			authHeader: "Bearer " + createTestToken(t, secret,
				jwt.SigningMethodHS256,
				jwt.MapClaims{
					"token_type": "access",
				}),
			expectedMsg: "Thong tin nguoi dung khong hop le",
		},
		{
			name: "user_id bang 0",
			authHeader: "Bearer " + createTestToken(t, secret,
				jwt.SigningMethodHS256,
				jwt.MapClaims{
					"token_type": "access",
					"user_id":    0,
				}),
			expectedMsg: "Thong tin nguoi dung khong hop le",
		},
		{
			name: "user_id am",
			authHeader: "Bearer " + createTestToken(t, secret,
				jwt.SigningMethodHS256,
				jwt.MapClaims{
					"token_type": "access",
					"user_id":    -1,
				}),
			expectedMsg: "Thong tin nguoi dung khong hop le",
		},
		{
			name: "Sai chu ky token",
			authHeader: "Bearer " + createTestToken(t, "wrong-secret",
				jwt.SigningMethodHS256,
				jwt.MapClaims{
					"token_type": "access",
					"user_id":    1,
				}),
			expectedMsg: "Token khong hop le hoac da het han",
		},
		{
			name: "Khong chap nhan HS384",
			authHeader: "Bearer " + createTestToken(t, secret,
				jwt.SigningMethodHS384,
				jwt.MapClaims{
					"token_type": "access",
					"user_id":    1,
				}),
			expectedMsg: "Token khong hop le hoac da het han",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			req.Header.Set("Authorization", tt.authHeader)

			nextCalled := false

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
			})

			handler := AuthMiddleware(secret, nil, next)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusUnauthorized {
				t.Errorf("Status = %d, muon %d",
					recorder.Code, http.StatusUnauthorized)
			}

			if nextCalled {
				t.Error("Handler tiep theo khong duoc phep chay")
			}

			if !strings.Contains(recorder.Body.String(), tt.expectedMsg) {
				t.Errorf("Response khong chua thong bao mong doi: %s",
					tt.expectedMsg)
			}
		})
	}
}
