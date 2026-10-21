
package service

import (
	"context"
	"errors"
	"time"

	"warehouseproject/internal/model"
	"warehouseproject/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials  = errors.New("username hoac password khong dung")
	ErrJWTSecretRequired   = errors.New("JWT secret khong duoc de trong")
	ErrInvalidRefreshToken = errors.New("refresh token khong hop le")
)

const (
	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 7 * 24 * time.Hour
)

type UserService struct {
	UserRepo  *repository.UserRepository
	JWTSecret []byte
}

type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	TokenType    string      `json:"token_type"`
	ExpiresIn    int64       `json:"expires_in"`
	User         *model.User `json:"user"`
}

// NewUserService khởi tạo service xử lý đăng nhập.
func NewUserService(
	userRepo *repository.UserRepository,
	jwtSecret string,
) *UserService {
	return &UserService{
		UserRepo:  userRepo,
		JWTSecret: []byte(jwtSecret),
	}
}

// Tạo JWT theo loại token và thời hạn tương ứng.
func (s *UserService) generateToken(
	user *model.User,
	tokenType string,
	duration time.Duration,
) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"user_id":    user.ID,
		"username":   user.Username,
		"role":       user.Role,
		"token_type": tokenType,
		"iat":        now.Unix(),
		"exp":        now.Add(duration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.JWTSecret)
}

// Login kiểm tra tài khoản và tạo Access Token, Refresh Token.
func (s *UserService) Login(
	ctx context.Context,
	username string,
	password string,
) (*LoginResponse, error) {
	if len(s.JWTSecret) == 0 {
		return nil, ErrJWTSecretRequired
	}

	user, err := s.UserRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.generateToken(
		user,
		"access",
		AccessTokenDuration,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken(
		user,
		"refresh",
		RefreshTokenDuration,
	)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(AccessTokenDuration.Seconds()),
		User:         user,
	}, nil
}

// RefreshAccessToken xác minh Refresh Token và cấp Access Token mới.
func (s *UserService) RefreshAccessToken(
	refreshToken string,
) (string, error) {
	if len(s.JWTSecret) == 0 {
		return "", ErrJWTSecretRequired
	}

	parsedToken, err := jwt.Parse(
		refreshToken,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidRefreshToken
			}
			return s.JWTSecret, nil
		},
	)
	if err != nil || !parsedToken.Valid {
		return "", ErrInvalidRefreshToken
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidRefreshToken
	}

	tokenType, ok := claims["token_type"].(string)
	if !ok || tokenType != "refresh" {
		return "", ErrInvalidRefreshToken
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return "", ErrInvalidRefreshToken
	}

	username, ok := claims["username"].(string)
	if !ok || username == "" {
		return "", ErrInvalidRefreshToken
	}

	role, ok := claims["role"].(string)
	if !ok || role == "" {
		return "", ErrInvalidRefreshToken
	}

	now := time.Now()
	newClaims := jwt.MapClaims{
		"user_id":    int64(userID),
		"username":   username,
		"role":       role,
		"token_type": "access",
		"iat":        now.Unix(),
		"exp":        now.Add(AccessTokenDuration).Unix(),
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	return newToken.SignedString(s.JWTSecret)
}
