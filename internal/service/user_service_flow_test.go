package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"warehouseproject/internal/model"
	"warehouseproject/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"golang.org/x/crypto/bcrypt"
)

func newUserServiceMock(t *testing.T) (pgxmock.PgxPoolIface, *UserService) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Khong tao duoc mock: %v", err)
	}

	repo := repository.NewUserRepositoryWithDB(mock)
	service := NewUserService(repo, "test-secret")

	return mock, service
}

func checkUserServiceMock(t *testing.T, mock pgxmock.PgxPoolIface) {
	t.Helper()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations: %v", err)
	}
}

func createTestUser(t *testing.T) (*model.User, string) {
	t.Helper()

	password := "password123"

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		t.Fatalf("Khong hash duoc password: %v", err)
	}

	return &model.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: string(hash),
		FullName:     "Test User",
		Role:         "admin",
	}, password
}

func createTestRefreshToken(t *testing.T, service *UserService) string {
	t.Helper()

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Role:     "admin",
	}

	token, err := service.generateToken(
		user,
		"refresh",
		RefreshTokenDuration,
	)
	if err != nil {
		t.Fatalf("Khong tao duoc refresh token: %v", err)
	}

	return token
}

func TestUserService_LoginFlow(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, service := newUserServiceMock(t)
		defer mock.Close()

		user, password := createTestUser(t)

		mock.ExpectQuery("SELECT id, username, password_hash, full_name, role").
			WithArgs("testuser").
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "username", "password_hash", "full_name", "role",
			}).AddRow(
				user.ID,
				user.Username,
				user.PasswordHash,
				user.FullName,
				user.Role,
			))

		mock.ExpectExec("INSERT INTO refresh_tokens").
			WithArgs(user.ID, pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		result, err := service.Login(
			context.Background(),
			"testuser",
			password,
		)

		if err != nil {
			t.Fatalf("Login that bai: %v", err)
		}

		if result == nil {
			t.Fatal("Login response khong duoc nil")
		}

		if result.AccessToken == "" || result.RefreshToken == "" {
			t.Error("Token khong duoc de trong")
		}

		if result.TokenType != "Bearer" {
			t.Errorf("TokenType = %s, muon Bearer", result.TokenType)
		}

		if result.ExpiresIn != int64(AccessTokenDuration.Seconds()) {
			t.Errorf("ExpiresIn khong dung: %d", result.ExpiresIn)
		}

		if result.User.ID != user.ID {
			t.Errorf("User ID = %d, muon %d", result.User.ID, user.ID)
		}

		checkUserServiceMock(t, mock)
	})

	t.Run("user not found", func(t *testing.T) {
		mock, service := newUserServiceMock(t)
		defer mock.Close()

		mock.ExpectQuery("SELECT id, username, password_hash, full_name, role").
			WithArgs("unknown").
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "username", "password_hash", "full_name", "role",
			}))

		result, err := service.Login(
			context.Background(),
			"unknown",
			"password123",
		)

		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("Muon ErrInvalidCredentials, nhan duoc: %v", err)
		}

		if result != nil {
			t.Error("Response phai nil")
		}

		checkUserServiceMock(t, mock)
	})

	t.Run("wrong password", func(t *testing.T) {
		mock, service := newUserServiceMock(t)
		defer mock.Close()

		user, _ := createTestUser(t)

		mock.ExpectQuery("SELECT id, username, password_hash, full_name, role").
			WithArgs("testuser").
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "username", "password_hash", "full_name", "role",
			}).AddRow(
				user.ID,
				user.Username,
				user.PasswordHash,
				user.FullName,
				user.Role,
			))

		result, err := service.Login(
			context.Background(),
			"testuser",
			"wrong-password",
		)

		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("Muon ErrInvalidCredentials, nhan duoc: %v", err)
		}

		if result != nil {
			t.Error("Response phai nil")
		}

		checkUserServiceMock(t, mock)
	})

	t.Run("database error", func(t *testing.T) {
		mock, service := newUserServiceMock(t)
		defer mock.Close()

		mock.ExpectQuery("SELECT id, username, password_hash, full_name, role").
			WithArgs("testuser").
			WillReturnError(errors.New("database error"))

		_, err := service.Login(
			context.Background(),
			"testuser",
			"password123",
		)

		if err == nil {
			t.Fatal("Muon nhan duoc loi database")
		}

		checkUserServiceMock(t, mock)
	})

	t.Run("save refresh token error", func(t *testing.T) {
		mock, service := newUserServiceMock(t)
		defer mock.Close()

		user, password := createTestUser(t)

		mock.ExpectQuery("SELECT id, username, password_hash, full_name, role").
			WithArgs("testuser").
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "username", "password_hash", "full_name", "role",
			}).AddRow(
				user.ID,
				user.Username,
				user.PasswordHash,
				user.FullName,
				user.Role,
			))

		mock.ExpectExec("INSERT INTO refresh_tokens").
			WithArgs(user.ID, pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errors.New("insert error"))

		_, err := service.Login(
			context.Background(),
			"testuser",
			password,
		)

		if err == nil {
			t.Fatal("Muon nhan duoc loi khi luu refresh token")
		}

		checkUserServiceMock(t, mock)
	})
}

func TestUserService_RefreshAccessTokenFlow(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, service := newUserServiceMock(t)
		defer mock.Close()

		user := &model.User{
			ID:       1,
			Username: "testuser",
			Role:     "admin",
		}

		oldToken := createTestRefreshToken(t, service)

		mock.ExpectQuery("SELECT id, username, password_hash, full_name, role").
			WithArgs(int64(1)).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "username", "password_hash", "full_name", "role",
			}).AddRow(
				user.ID,
				user.Username,
				"hashed",
				"Test User",
				user.Role,
			))

		mock.ExpectBegin()

		mock.ExpectQuery("UPDATE refresh_tokens").
			WithArgs(int64(1), hashRefreshToken(oldToken)).
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(10)))

		mock.ExpectExec("INSERT INTO refresh_tokens").
			WithArgs(
				int64(1),
				pgxmock.AnyArg(),
				pgxmock.AnyArg(),
			).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectCommit()

		result, err := service.RefreshAccessToken(
			context.Background(),
			oldToken,
		)

		if err != nil {
			t.Fatalf("Refresh token that bai: %v", err)
		}

		if result == nil {
			t.Fatal("Refresh response khong duoc nil")
		}

		if result.AccessToken == "" || result.RefreshToken == "" {
			t.Error("Token moi khong duoc de trong")
		}

		if result.TokenType != "Bearer" {
			t.Errorf("TokenType khong dung: %s", result.TokenType)
		}

		checkUserServiceMock(t, mock)
	})

	t.Run("user not found", func(t *testing.T) {
		mock, service := newUserServiceMock(t)
		defer mock.Close()

		token := createTestRefreshToken(t, service)

		mock.ExpectQuery("SELECT id, username, password_hash, full_name, role").
			WithArgs(int64(1)).
			WillReturnError(pgx.ErrNoRows)

		result, err := service.RefreshAccessToken(
			context.Background(),
			token,
		)

		if !errors.Is(err, ErrInvalidRefreshToken) {
			t.Errorf("Muon ErrInvalidRefreshToken, nhan duoc: %v", err)
		}

		if result != nil {
			t.Error("Response phai nil")
		}

		checkUserServiceMock(t, mock)
	})

	t.Run("database error", func(t *testing.T) {
		mock, service := newUserServiceMock(t)
		defer mock.Close()

		token := createTestRefreshToken(t, service)

		mock.ExpectQuery("SELECT id, username, password_hash, full_name, role").
			WithArgs(int64(1)).
			WillReturnError(errors.New("database error"))

		_, err := service.RefreshAccessToken(
			context.Background(),
			token,
		)

		if err == nil {
			t.Fatal("Muon nhan duoc loi database")
		}

		checkUserServiceMock(t, mock)
	})

	t.Run("old refresh token not found", func(t *testing.T) {
		mock, service := newUserServiceMock(t)
		defer mock.Close()

		token := createTestRefreshToken(t, service)

		mock.ExpectQuery("SELECT id, username, password_hash, full_name, role").
			WithArgs(int64(1)).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "username", "password_hash", "full_name", "role",
			}).AddRow(
				int64(1),
				"testuser",
				"hashed",
				"Test User",
				"admin",
			))

		mock.ExpectBegin()

		mock.ExpectQuery("UPDATE refresh_tokens").
			WithArgs(int64(1), hashRefreshToken(token)).
			WillReturnError(pgx.ErrNoRows)

		mock.ExpectRollback()

		result, err := service.RefreshAccessToken(
			context.Background(),
			token,
		)

		if !errors.Is(err, ErrInvalidRefreshToken) {
			t.Errorf("Muon ErrInvalidRefreshToken, nhan duoc: %v", err)
		}

		if result != nil {
			t.Error("Response phai nil")
		}

		checkUserServiceMock(t, mock)
	})

	t.Run("invalid token", func(t *testing.T) {
		mock, service := newUserServiceMock(t)
		defer mock.Close()

		result, err := service.RefreshAccessToken(
			context.Background(),
			"invalid-token",
		)

		if !errors.Is(err, ErrInvalidRefreshToken) {
			t.Errorf("Muon ErrInvalidRefreshToken, nhan duoc: %v", err)
		}

		if result != nil {
			t.Error("Response phai nil")
		}

		checkUserServiceMock(t, mock)
	})
}

func TestUserService_LogoutFlow(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, service := newUserServiceMock(t)
		defer mock.Close()

		token := createTestRefreshToken(t, service)

		mock.ExpectExec("UPDATE refresh_tokens").
			WithArgs(int64(1), hashRefreshToken(token)).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := service.Logout(context.Background(), token)
		if err != nil {
			t.Fatalf("Logout that bai: %v", err)
		}

		checkUserServiceMock(t, mock)
	})

	t.Run("refresh token not found", func(t *testing.T) {
		mock, service := newUserServiceMock(t)
		defer mock.Close()

		token := createTestRefreshToken(t, service)

		mock.ExpectExec("UPDATE refresh_tokens").
			WithArgs(int64(1), hashRefreshToken(token)).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))

		err := service.Logout(context.Background(), token)

		if !errors.Is(err, ErrInvalidRefreshToken) {
			t.Errorf("Muon ErrInvalidRefreshToken, nhan duoc: %v", err)
		}

		checkUserServiceMock(t, mock)
	})

	t.Run("database error", func(t *testing.T) {
		mock, service := newUserServiceMock(t)
		defer mock.Close()

		token := createTestRefreshToken(t, service)

		mock.ExpectExec("UPDATE refresh_tokens").
			WithArgs(int64(1), hashRefreshToken(token)).
			WillReturnError(errors.New("database error"))

		err := service.Logout(context.Background(), token)

		if err == nil {
			t.Fatal("Muon nhan duoc loi database")
		}

		checkUserServiceMock(t, mock)
	})

	t.Run("invalid token", func(t *testing.T) {
		mock, service := newUserServiceMock(t)
		defer mock.Close()

		err := service.Logout(
			context.Background(),
			"invalid-token",
		)

		if !errors.Is(err, ErrInvalidRefreshToken) {
			t.Errorf("Muon ErrInvalidRefreshToken, nhan duoc: %v", err)
		}

		checkUserServiceMock(t, mock)
	})
}

func TestUserService_RefreshTokenExpiry(t *testing.T) {
	service := &UserService{
		JWTSecret: []byte("test-secret"),
	}

	token := createServiceTestToken(
		t,
		"test-secret",
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"token_type": "refresh",
			"user_id":    1,
			"exp":        time.Now().Add(-time.Hour).Unix(),
		},
	)

	_, err := service.RefreshAccessToken(
		context.Background(),
		token,
	)

	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("Muon ErrInvalidRefreshToken, nhan duoc: %v", err)
	}
}

func TestDummyPasswordHashUsesDefaultCost(t *testing.T) {
	cost, err := bcrypt.Cost(dummyPasswordHash)
	if err != nil {
		t.Fatalf("Dummy hash khong hop le: %v", err)
	}

	if cost != bcrypt.DefaultCost {
		t.Errorf("Dummy hash cost = %d, muon %d",
			cost, bcrypt.DefaultCost)
	}

	err = bcrypt.CompareHashAndPassword(
		dummyPasswordHash,
		[]byte("password123"),
	)

	if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Errorf("Dummy hash khong hoat dong dung: %v", err)
	}
}
