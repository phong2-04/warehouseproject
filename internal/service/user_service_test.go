package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"warehouseproject/internal/model"

	"github.com/golang-jwt/jwt/v5"
)

func createServiceTestToken(
	t *testing.T,
	secret string,
	method jwt.SigningMethod,
	claims jwt.MapClaims,
) string {
	t.Helper()

	token := jwt.NewWithClaims(method, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Khong tao duoc token: %v", err)
	}

	return tokenString
}

func TestHashRefreshToken(t *testing.T) {
	token := "test-refresh-token"

	hash := hashRefreshToken(token)

	expected := sha256.Sum256([]byte(token))
	expectedHash := hex.EncodeToString(expected[:])

	if hash != expectedHash {
		t.Errorf("Hash khong dung")
	}

	if len(hash) != 64 {
		t.Errorf("Do dai hash = %d, muon 64", len(hash))
	}

	if hashRefreshToken(token) != hash {
		t.Error("Hash cung mot token phai cho ket qua giong nhau")
	}
}

func TestGenerateToken(t *testing.T) {
	service := &UserService{
		JWTSecret: []byte("test-secret"),
	}

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Role:     "admin",
	}

	tests := []struct {
		name      string
		tokenType string
		duration  time.Duration
	}{
		{
			name:      "Tao Access Token",
			tokenType: "access",
			duration:  AccessTokenDuration,
		},
		{
			name:      "Tao Refresh Token",
			tokenType: "refresh",
			duration:  RefreshTokenDuration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, err := service.generateToken(
				user,
				tt.tokenType,
				tt.duration,
			)

			if err != nil {
				t.Fatalf("Tao token that bai: %v", err)
			}

			token, err := jwt.Parse(
				tokenString,
				func(token *jwt.Token) (interface{}, error) {
					return service.JWTSecret, nil
				},
			)

			if err != nil {
				t.Fatalf("Khong parse duoc token: %v", err)
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				t.Fatal("Claims khong hop le")
			}

			if claims["token_type"] != tt.tokenType {
				t.Errorf("Token type khong dung")
			}

			if claims["username"] != user.Username {
				t.Errorf("Username khong dung")
			}

			if claims["role"] != user.Role {
				t.Errorf("Role khong dung")
			}

			if claims["user_id"] != float64(user.ID) {
				t.Errorf("User ID khong dung")
			}
		})
	}
}

func TestParseRefreshToken(t *testing.T) {
	secret := "test-secret"

	service := &UserService{
		JWTSecret: []byte(secret),
	}

	tests := []struct {
		name        string
		token       string
		expectedID  int64
		expectedErr error
	}{
		{
			name: "Refresh Token hop le",
			token: createServiceTestToken(t, secret,
				jwt.SigningMethodHS256,
				jwt.MapClaims{
					"token_type": "refresh",
					"user_id":    10,
					"exp":        time.Now().Add(time.Hour).Unix(),
				}),
			expectedID: 10,
		},
		{
			name: "Access Token khong duoc chap nhan",
			token: createServiceTestToken(t, secret,
				jwt.SigningMethodHS256,
				jwt.MapClaims{
					"token_type": "access",
					"user_id":    10,
					"exp":        time.Now().Add(time.Hour).Unix(),
				}),
			expectedErr: ErrInvalidRefreshToken,
		},
		{
			name: "Sai chu ky",
			token: createServiceTestToken(t, "wrong-secret",
				jwt.SigningMethodHS256,
				jwt.MapClaims{
					"token_type": "refresh",
					"user_id":    10,
					"exp":        time.Now().Add(time.Hour).Unix(),
				}),
			expectedErr: ErrInvalidRefreshToken,
		},
		{
			name: "Token het han",
			token: createServiceTestToken(t, secret,
				jwt.SigningMethodHS256,
				jwt.MapClaims{
					"token_type": "refresh",
					"user_id":    10,
					"exp":        time.Now().Add(-time.Hour).Unix(),
				}),
			expectedErr: ErrInvalidRefreshToken,
		},
		{
			name: "User ID bang 0",
			token: createServiceTestToken(t, secret,
				jwt.SigningMethodHS256,
				jwt.MapClaims{
					"token_type": "refresh",
					"user_id":    0,
					"exp":        time.Now().Add(time.Hour).Unix(),
				}),
			expectedErr: ErrInvalidRefreshToken,
		},
		{
			name: "Thieu User ID",
			token: createServiceTestToken(t, secret,
				jwt.SigningMethodHS256,
				jwt.MapClaims{
					"token_type": "refresh",
					"exp":        time.Now().Add(time.Hour).Unix(),
				}),
			expectedErr: ErrInvalidRefreshToken,
		},
		{
			name:        "Sai dinh dang token",
			token:       "invalid-token",
			expectedErr: ErrInvalidRefreshToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := service.parseRefreshToken(tt.token)

			if tt.expectedErr != nil {
				if err != tt.expectedErr {
					t.Errorf("Error = %v, muon %v",
						err, tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Khong mong doi loi: %v", err)
			}

			if userID != tt.expectedID {
				t.Errorf("User ID = %d, muon %d",
					userID, tt.expectedID)
			}
		})
	}
}

func TestUserServiceRequiresJWTSecret(t *testing.T) {
	service := &UserService{}

	_, err := service.Login(
		context.Background(),
		"testuser",
		"password123",
	)

	if err != ErrJWTSecretRequired {
		t.Errorf("Login error = %v, muon %v",
			err, ErrJWTSecretRequired)
	}

	_, err = service.parseRefreshToken("some-token")

	if err != ErrJWTSecretRequired {
		t.Errorf("Parse error = %v, muon %v",
			err, ErrJWTSecretRequired)
	}
}

func TestGenerateTokenUniqueJTI(t *testing.T) {
	service := &UserService{
		JWTSecret: []byte("test-secret"),
	}

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Role:     "admin",
	}

	firstToken, err := service.generateToken(user, "refresh", RefreshTokenDuration)
	if err != nil {
		t.Fatalf("Tao token thu nhat that bai: %v", err)
	}

	secondToken, err := service.generateToken(user, "refresh", RefreshTokenDuration)
	if err != nil {
		t.Fatalf("Tao token thu hai that bai: %v", err)
	}

	if firstToken == secondToken {
		t.Fatal("Hai JWT tao lien tiep khong duoc trung nhau")
	}

	parseJTI := func(tokenString string) string {
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return service.JWTSecret, nil
		})
		if err != nil {
			t.Fatalf("Parse JWT that bai: %v", err)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			t.Fatal("Claims khong hop le")
		}

		jti, ok := claims["jti"].(string)
		if !ok || jti == "" {
			t.Fatal("JWT khong co JTI hop le")
		}

		return jti
	}

	firstJTI := parseJTI(firstToken)
	secondJTI := parseJTI(secondToken)

	if firstJTI == secondJTI {
		t.Fatal("JTI cua hai token bi trung nhau")
	}
}
