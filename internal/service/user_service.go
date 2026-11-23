package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
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

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
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
	jti := make([]byte, 16)
	if _, err := rand.Read(jti); err != nil {
		return "", err
	}

	now := time.Now()

	claims := jwt.MapClaims{
		"jti":        hex.EncodeToString(jti),
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

// Tao SHA-256 hash cua Refresh Token de luu trong database.
func hashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// Login kiem tra tai khoan va luu Refresh Token vao database.
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

	err = s.UserRepo.CreateRefreshToken(
		ctx,
		user.ID,
		hashRefreshToken(refreshToken),
		time.Now().Add(RefreshTokenDuration),
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

// Kiem tra chu ky, loai va user ID cua Refresh Token.
func (s *UserService) parseRefreshToken(
	refreshToken string,
) (int64, error) {
	if len(s.JWTSecret) == 0 {
		return 0, ErrJWTSecretRequired
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
		return 0, ErrInvalidRefreshToken
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return 0, ErrInvalidRefreshToken
	}

	tokenType, ok := claims["token_type"].(string)
	if !ok || tokenType != "refresh" {
		return 0, ErrInvalidRefreshToken
	}

	userIDValue, ok := claims["user_id"].(float64)
	if !ok || userIDValue <= 0 || userIDValue != float64(int64(userIDValue)) {
		return 0, ErrInvalidRefreshToken
	}

	return int64(userIDValue), nil
}

// Refresh Token rotation: thu hoi token cu va cap cap token moi.
func (s *UserService) RefreshAccessToken(
	ctx context.Context,
	refreshToken string,
) (*RefreshResponse, error) {
	userID, err := s.parseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.UserRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}

	accessToken, err := s.generateToken(
		user,
		"access",
		AccessTokenDuration,
	)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.generateToken(
		user,
		"refresh",
		RefreshTokenDuration,
	)
	if err != nil {
		return nil, err
	}

	err = s.UserRepo.RotateRefreshToken(
		ctx,
		user.ID,
		hashRefreshToken(refreshToken),
		hashRefreshToken(newRefreshToken),
		time.Now().Add(RefreshTokenDuration),
	)
	if err != nil {
		if errors.Is(err, repository.ErrRefreshTokenNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}

	return &RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(AccessTokenDuration.Seconds()),
	}, nil
}

// Logout thu hoi Refresh Token hien tai.
func (s *UserService) Logout(
	ctx context.Context,
	refreshToken string,
) error {
	userID, err := s.parseRefreshToken(refreshToken)
	if err != nil {
		return err
	}

	err = s.UserRepo.RevokeRefreshToken(
		ctx,
		userID,
		hashRefreshToken(refreshToken),
	)
	if errors.Is(err, repository.ErrRefreshTokenNotFound) {
		return ErrInvalidRefreshToken
	}

	return err
}
