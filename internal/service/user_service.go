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

func NewUserService(
	userRepo *repository.UserRepository,
	jwtSecret string,
) *UserService {
	return &UserService{
		UserRepo:  userRepo,
		JWTSecret: []byte(jwtSecret),
	}
}

// Tao JWT theo loai token.
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

// Login kiem tra tai khoan va tao Access Token, Refresh Token.
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

// RefreshAccessToken kiem tra lai tai khoan trong DB truoc khi cap token moi.
func (s *UserService) RefreshAccessToken(
	ctx context.Context,
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
		jwt.WithValidMethods([]string{"HS256"}),
	)
	if err != nil || parsedToken == nil || !parsedToken.Valid {
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

	userIDValue, ok := claims["user_id"].(float64)
	if !ok || userIDValue <= 0 || userIDValue != float64(int64(userIDValue)) {
		return "", ErrInvalidRefreshToken
	}

	// Kiem tra tai khoan con ton tai trong database.
	user, err := s.UserRepo.GetByID(ctx, int64(userIDValue))
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrInvalidRefreshToken
		}
		return "", err
	}

	// Tao Access Token voi username va role moi nhat trong database.
	return s.generateToken(
		user,
		"access",
		AccessTokenDuration,
	)
}
