package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"warehouseproject/internal/repository"

	"github.com/pashagolub/pgxmock/v5"
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
		{"Password vuot 72 byte", "user01", strings.Repeat("a", 73), "Nguyen Van A", "staff"},
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

// Username va full name Unicode vuot gioi han phai bi tu choi.
func TestUserManagementServiceCreateUnicodeTooLong(t *testing.T) {
	service := &UserManagementService{}

	_, err := service.Create(
		context.Background(),
		strings.Repeat("界", 51),
		"password123",
		"Nguyen Van A",
		"staff",
	)
	if !errors.Is(err, ErrInvalidUserData) {
		t.Errorf("Muon ErrInvalidUserData cho username, nhan duoc: %v", err)
	}

	_, err = service.Create(
		context.Background(),
		"user01",
		"password123",
		strings.Repeat("界", 101),
		"staff",
	)
	if !errors.Is(err, ErrInvalidUserData) {
		t.Errorf("Muon ErrInvalidUserData cho full name, nhan duoc: %v", err)
	}
}

func TestUserManagementServiceUpdateUnicodeTooLong(t *testing.T) {
	service := &UserManagementService{}

	_, err := service.Update(
		context.Background(),
		1,
		strings.Repeat("界", 51),
		"Nguyen Van A",
		"staff",
	)
	if !errors.Is(err, ErrInvalidUserData) {
		t.Errorf("Muon ErrInvalidUserData cho username, nhan duoc: %v", err)
	}

	_, err = service.Update(
		context.Background(),
		1,
		"user01",
		strings.Repeat("界", 101),
		"staff",
	)
	if !errors.Is(err, ErrInvalidUserData) {
		t.Errorf("Muon ErrInvalidUserData cho full name, nhan duoc: %v", err)
	}
}

// Create chap nhan username 50 va full name 100 ky tu Unicode.
func TestUserManagementServiceCreateUnicodeBoundary(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	service := NewUserManagementService(
		repository.NewUserRepositoryWithDB(mock),
	)

	username := strings.Repeat("界", 50)
	fullName := strings.Repeat("界", 100)

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(username, pgxmock.AnyArg(), fullName, "staff").
		WillReturnRows(
			pgxmock.NewRows([]string{"id"}).AddRow(int64(1)),
		)

	user, err := service.Create(
		context.Background(),
		username,
		"password123",
		fullName,
		"staff",
	)
	if err != nil {
		t.Fatalf("Create that bai voi Unicode dung gioi han: %v", err)
	}
	if user.ID != 1 {
		t.Errorf("ID = %d, muon 1", user.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations: %v", err)
	}
}

// Update chap nhan username 50 va full name 100 ky tu Unicode.
func TestUserManagementServiceUpdateUnicodeBoundary(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	service := NewUserManagementService(
		repository.NewUserRepositoryWithDB(mock),
	)

	username := strings.Repeat("界", 50)
	fullName := strings.Repeat("界", 100)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM users WHERE role = 'admin'").
		WillReturnRows(pgxmock.NewRows([]string{"id"}))
	mock.ExpectExec("UPDATE users").
		WithArgs(username, fullName, "staff", int64(1)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	user, err := service.Update(
		context.Background(),
		1,
		username,
		fullName,
		"staff",
	)
	if err != nil {
		t.Fatalf("Update that bai voi Unicode dung gioi han: %v", err)
	}
	if user.ID != 1 {
		t.Errorf("ID = %d, muon 1", user.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations: %v", err)
	}
}
