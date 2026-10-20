
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

var ErrInvalidCredentials = errors.New("username hoac password khong dung")
var ErrJWTSecretRequired = errors.New("JWT secret khong duoc de trong")

type UserService struct {
	UserRepo  *repository.UserRepository
	JWTSecret []byte
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
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

// Login kiểm tra tài khoản và tạo JWT nếu đăng nhập thành công.
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

	// So sánh mật khẩu nhập vào với mật khẩu đã mã hóa trong database.
	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Tạo JWT có thời hạn 24 giờ.
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.JWTSecret)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token: tokenString,
		User:  user,
	}, nil
}
