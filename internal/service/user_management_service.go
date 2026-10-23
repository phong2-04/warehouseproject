package service

import (
	"context"
	"errors"
	"strings"

	"warehouseproject/internal/model"
	"warehouseproject/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidUserData = errors.New("du lieu nguoi dung khong hop le")

type UserManagementService struct {
	UserRepo *repository.UserRepository
}

func NewUserManagementService(
	userRepo *repository.UserRepository,
) *UserManagementService {
	return &UserManagementService{UserRepo: userRepo}
}

// Lay danh sach nguoi dung.
func (s *UserManagementService) GetAll(
	ctx context.Context,
) ([]model.User, error) {
	return s.UserRepo.GetAll(ctx)
}

// Tao tai khoan va ma hoa mat khau.
func (s *UserManagementService) Create(
	ctx context.Context,
	username, password, fullName, role string,
) (*model.User, error) {
	username = strings.TrimSpace(username)
	fullName = strings.TrimSpace(fullName)
	role = strings.TrimSpace(role)

	if username == "" || password == "" || fullName == "" {
		return nil, ErrInvalidUserData
	}
	if len(username) > 50 || len(password) < 8 {
		return nil, ErrInvalidUserData
	}
	if role != "admin" && role != "staff" {
		return nil, ErrInvalidUserData
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		PasswordHash: string(passwordHash),
		FullName:     fullName,
		Role:         role,
	}

	if err := s.UserRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Sua username, ho ten va quyen cua tai khoan.
func (s *UserManagementService) Update(
	ctx context.Context,
	id int64,
	username, fullName, role string,
) (*model.User, error) {
	username = strings.TrimSpace(username)
	fullName = strings.TrimSpace(fullName)
	role = strings.TrimSpace(role)

	if id <= 0 || username == "" || fullName == "" {
		return nil, ErrInvalidUserData
	}
	if len(username) > 50 {
		return nil, ErrInvalidUserData
	}
	if role != "admin" && role != "staff" {
		return nil, ErrInvalidUserData
	}

	user := &model.User{
		ID:       id,
		Username: username,
		FullName: fullName,
		Role:     role,
	}

	if err := s.UserRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Xoa tai khoan.
func (s *UserManagementService) Delete(
	ctx context.Context,
	id int64,
) error {
	if id <= 0 {
		return ErrInvalidUserData
	}
	return s.UserRepo.Delete(ctx, id)
}
