package service

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestUserManagementServiceCreateInvalidData(t *testing.T) {
	service := &UserManagementService{}

	tests := []struct {
		name     string
		username string
		password string
		fullName string
		role     string
	}{
		{"Username rong", "", "password123", "Nguyen Van A", "staff"},
		{"Password rong", "user01", "", "Nguyen Van A", "staff"},
		{"FullName rong", "user01", "password123", "", "staff"},
		{"Username qua dai", strings.Repeat("a", 51), "password123", "Nguyen Van A", "staff"},
		{"Password qua ngan", "user01", "1234567", "Nguyen Van A", "staff"},
		{"Role khong hop le", "user01", "password123", "Nguyen Van A", "manager"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Create(
				context.Background(),
				tt.username,
				tt.password,
				tt.fullName,
				tt.role,
			)

			if !errors.Is(err, ErrInvalidUserData) {
				t.Errorf("Muon ErrInvalidUserData, nhan duoc: %v", err)
			}
		})
	}
}

func TestUserManagementServiceUpdateInvalidData(t *testing.T) {
	service := &UserManagementService{}

	tests := []struct {
		name     string
		id       int64
		username string
		fullName string
		role     string
	}{
		{"ID bang 0", 0, "user01", "Nguyen Van A", "staff"},
		{"ID am", -1, "user01", "Nguyen Van A", "staff"},
		{"Username rong", 1, "", "Nguyen Van A", "staff"},
		{"FullName rong", 1, "user01", "", "staff"},
		{"Username qua dai", 1, strings.Repeat("a", 51), "Nguyen Van A", "staff"},
		{"Role khong hop le", 1, "user01", "Nguyen Van A", "manager"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Update(
				context.Background(),
				tt.id,
				tt.username,
				tt.fullName,
				tt.role,
			)

			if !errors.Is(err, ErrInvalidUserData) {
				t.Errorf("Muon ErrInvalidUserData, nhan duoc: %v", err)
			}
		})
	}
}

func TestUserManagementServiceDeleteInvalidID(t *testing.T) {
	service := &UserManagementService{}

	tests := []struct {
		name string
		id   int64
	}{
		{"ID bang 0", 0},
		{"ID am", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Delete(context.Background(), tt.id)

			if !errors.Is(err, ErrInvalidUserData) {
				t.Errorf("Muon ErrInvalidUserData, nhan duoc: %v", err)
			}
		})
	}
}
