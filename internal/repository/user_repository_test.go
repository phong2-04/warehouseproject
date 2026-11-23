package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"warehouseproject/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
)

func newUserMock(t *testing.T) (pgxmock.PgxPoolIface, *UserRepository) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("khong tao duoc mock: %v", err)
	}

	return mock, NewUserRepositoryWithDB(mock)
}

func checkUserMock(t *testing.T, mock pgxmock.PgxPoolIface) {
	t.Helper()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		mock.ExpectQuery("SELECT id, username, password_hash, full_name, role").
			WithArgs("admin").
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "username", "password_hash", "full_name", "role",
			}).AddRow(int64(1), "admin", "hashed", "Admin User", "admin"))

		user, err := repo.GetByUsername(context.Background(), "admin")
		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}
		if user == nil || user.ID != 1 || user.Username != "admin" {
			t.Fatalf("du lieu user khong dung: %+v", user)
		}

		checkUserMock(t, mock)
	})

	t.Run("not found", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		mock.ExpectQuery("SELECT id, username, password_hash, full_name, role").
			WithArgs("unknown").
			WillReturnError(pgx.ErrNoRows)

		user, err := repo.GetByUsername(context.Background(), "unknown")
		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}
		if user != nil {
			t.Errorf("mong doi nil, nhan duoc: %+v", user)
		}

		checkUserMock(t, mock)
	})

	t.Run("database error", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		mock.ExpectQuery("SELECT id, username, password_hash, full_name, role").
			WithArgs("admin").
			WillReturnError(errors.New("database error"))

		_, err := repo.GetByUsername(context.Background(), "admin")
		if err == nil {
			t.Fatal("mong doi loi truy van")
		}

		checkUserMock(t, mock)
	})
}

func TestUserRepository_GetAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		mock.ExpectQuery("SELECT id, username, password_hash, full_name, role").
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "username", "password_hash", "full_name", "role",
			}).
				AddRow(int64(1), "admin", "hash1", "Admin", "admin").
				AddRow(int64(2), "staff", "hash2", "Staff", "staff"))

		users, err := repo.GetAll(context.Background())
		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}
		if len(users) != 2 {
			t.Fatalf("mong doi 2 user, nhan duoc %d", len(users))
		}

		checkUserMock(t, mock)
	})

	t.Run("query error", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		mock.ExpectQuery("SELECT id, username, password_hash, full_name, role").
			WillReturnError(errors.New("database error"))

		_, err := repo.GetAll(context.Background())
		if err == nil {
			t.Fatal("mong doi loi truy van")
		}

		checkUserMock(t, mock)
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		mock.ExpectQuery("SELECT id, username, password_hash, full_name, role").
			WithArgs(int64(1)).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "username", "password_hash", "full_name", "role",
			}).AddRow(int64(1), "admin", "hash", "Admin", "admin"))

		user, err := repo.GetByID(context.Background(), 1)
		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}
		if user.ID != 1 || user.Username != "admin" {
			t.Errorf("du lieu user khong dung: %+v", user)
		}

		checkUserMock(t, mock)
	})

	t.Run("not found", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		mock.ExpectQuery("SELECT id, username, password_hash, full_name, role").
			WithArgs(int64(99)).
			WillReturnError(pgx.ErrNoRows)

		_, err := repo.GetByID(context.Background(), 99)
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("mong doi ErrUserNotFound, nhan duoc: %v", err)
		}

		checkUserMock(t, mock)
	})

	t.Run("database error", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		mock.ExpectQuery("SELECT id, username, password_hash, full_name, role").
			WithArgs(int64(1)).
			WillReturnError(errors.New("database error"))

		_, err := repo.GetByID(context.Background(), 1)
		if err == nil {
			t.Fatal("mong doi loi truy van")
		}

		checkUserMock(t, mock)
	})
}

func TestUserRepository_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		user := &model.User{
			Username:     "newuser",
			PasswordHash: "hashed",
			FullName:     "New User",
			Role:         "staff",
		}

		mock.ExpectQuery("INSERT INTO users").
			WithArgs(user.Username, user.PasswordHash, user.FullName, user.Role).
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(5)))

		err := repo.Create(context.Background(), user)
		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}
		if user.ID != 5 {
			t.Errorf("ID mong doi 5, nhan duoc %d", user.ID)
		}

		checkUserMock(t, mock)
	})

	t.Run("database error", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		user := &model.User{Username: "newuser"}

		mock.ExpectQuery("INSERT INTO users").
			WithArgs(user.Username, user.PasswordHash, user.FullName, user.Role).
			WillReturnError(errors.New("database error"))

		err := repo.Create(context.Background(), user)
		if err == nil {
			t.Fatal("mong doi loi khi tao user")
		}

		checkUserMock(t, mock)
	})
}

func TestUserRepository_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		user := &model.User{
			ID:       2,
			Username: "staff2",
			FullName: "Staff Two",
			Role:     "staff",
		}

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT id FROM users WHERE role = 'admin'").
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(1)))
		mock.ExpectExec("UPDATE users").
			WithArgs(user.Username, user.FullName, user.Role, user.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		mock.ExpectCommit()

		err := repo.Update(context.Background(), user)
		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}

		checkUserMock(t, mock)
	})

	t.Run("last admin protection", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		user := &model.User{
			ID:       1,
			Username: "admin",
			FullName: "Admin",
			Role:     "staff",
		}

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT id FROM users WHERE role = 'admin'").
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(1)))
		mock.ExpectRollback()

		err := repo.Update(context.Background(), user)
		if !errors.Is(err, ErrLastAdmin) {
			t.Errorf("mong doi ErrLastAdmin, nhan duoc: %v", err)
		}

		checkUserMock(t, mock)
	})

	t.Run("user not found", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		user := &model.User{
			ID:       99,
			Username: "missing",
			FullName: "Missing",
			Role:     "staff",
		}

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT id FROM users WHERE role = 'admin'").
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(1)))
		mock.ExpectExec("UPDATE users").
			WithArgs(user.Username, user.FullName, user.Role, user.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))
		mock.ExpectRollback()

		err := repo.Update(context.Background(), user)
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("mong doi ErrUserNotFound, nhan duoc: %v", err)
		}

		checkUserMock(t, mock)
	})

	t.Run("begin error", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		mock.ExpectBegin().WillReturnError(errors.New("begin error"))

		err := repo.Update(context.Background(), &model.User{ID: 1})
		if err == nil {
			t.Fatal("mong doi loi Begin")
		}

		checkUserMock(t, mock)
	})
}

func TestUserRepository_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT id FROM users WHERE role = 'admin'").
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(1)))
		mock.ExpectExec("DELETE FROM users").
			WithArgs(int64(2)).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		mock.ExpectCommit()

		err := repo.Delete(context.Background(), 2)
		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}

		checkUserMock(t, mock)
	})

	t.Run("last admin protection", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT id FROM users WHERE role = 'admin'").
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(1)))
		mock.ExpectRollback()

		err := repo.Delete(context.Background(), 1)
		if !errors.Is(err, ErrLastAdmin) {
			t.Errorf("mong doi ErrLastAdmin, nhan duoc: %v", err)
		}

		checkUserMock(t, mock)
	})

	t.Run("user not found", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT id FROM users WHERE role = 'admin'").
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(1)))
		mock.ExpectExec("DELETE FROM users").
			WithArgs(int64(99)).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))
		mock.ExpectRollback()

		err := repo.Delete(context.Background(), 99)
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("mong doi ErrUserNotFound, nhan duoc: %v", err)
		}

		checkUserMock(t, mock)
	})
}

func TestUserRepository_CreateRefreshToken(t *testing.T) {
	mock, repo := newUserMock(t)
	defer mock.Close()

	expiresAt := time.Now().Add(24 * time.Hour)

	mock.ExpectExec("INSERT INTO refresh_tokens").
		WithArgs(int64(1), "token-hash", expiresAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err := repo.CreateRefreshToken(context.Background(), 1, "token-hash", expiresAt)
	if err != nil {
		t.Fatalf("khong mong doi loi: %v", err)
	}

	checkUserMock(t, mock)
}

func TestUserRepository_RotateRefreshToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		expiresAt := time.Now().Add(24 * time.Hour)

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT expires_at <= now()").
			WithArgs(int64(1), "old-hash").
			WillReturnRows(pgxmock.NewRows([]string{
				"expired", "revoked",
			}).AddRow(false, false))
		mock.ExpectExec("UPDATE refresh_tokens").
			WithArgs(int64(1), "old-hash").
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		mock.ExpectExec("INSERT INTO refresh_tokens").
			WithArgs(int64(1), "new-hash", expiresAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
		mock.ExpectCommit()

		err := repo.RotateRefreshToken(
			context.Background(), 1, "old-hash", "new-hash", expiresAt,
		)
		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}

		checkUserMock(t, mock)
	})

	t.Run("revoked token reuse revokes active sessions", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		expiresAt := time.Now().Add(24 * time.Hour)

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT expires_at <= now()").
			WithArgs(int64(1), "old-hash").
			WillReturnRows(pgxmock.NewRows([]string{
				"expired", "revoked",
			}).AddRow(false, true))
		mock.ExpectExec("UPDATE refresh_tokens").
			WithArgs(int64(1)).
			WillReturnResult(pgxmock.NewResult("UPDATE", 2))
		mock.ExpectCommit()

		err := repo.RotateRefreshToken(
			context.Background(), 1, "old-hash", "new-hash", expiresAt,
		)
		if !errors.Is(err, ErrRefreshTokenReuse) {
			t.Fatalf("mong doi ErrRefreshTokenReuse, nhan duoc: %v", err)
		}

		checkUserMock(t, mock)
	})

	t.Run("expired token is not reuse", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		expiresAt := time.Now().Add(24 * time.Hour)

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT expires_at <= now()").
			WithArgs(int64(1), "old-hash").
			WillReturnRows(pgxmock.NewRows([]string{
				"expired", "revoked",
			}).AddRow(true, false))
		mock.ExpectRollback()

		err := repo.RotateRefreshToken(
			context.Background(), 1, "old-hash", "new-hash", expiresAt,
		)
		if !errors.Is(err, ErrRefreshTokenNotFound) {
			t.Fatalf("mong doi ErrRefreshTokenNotFound, nhan duoc: %v", err)
		}

		checkUserMock(t, mock)
	})

	t.Run("old token not found", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		expiresAt := time.Now().Add(24 * time.Hour)

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT expires_at <= now()").
			WithArgs(int64(1), "old-hash").
			WillReturnError(pgx.ErrNoRows)
		mock.ExpectRollback()

		err := repo.RotateRefreshToken(
			context.Background(), 1, "old-hash", "new-hash", expiresAt,
		)
		if !errors.Is(err, ErrRefreshTokenNotFound) {
			t.Errorf("mong doi ErrRefreshTokenNotFound, nhan duoc: %v", err)
		}

		checkUserMock(t, mock)
	})
}
func TestUserRepository_RevokeRefreshToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		mock.ExpectExec("UPDATE refresh_tokens").
			WithArgs(int64(1), "token-hash").
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.RevokeRefreshToken(context.Background(), 1, "token-hash")
		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}

		checkUserMock(t, mock)
	})

	t.Run("token not found", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		mock.ExpectExec("UPDATE refresh_tokens").
			WithArgs(int64(1), "missing-hash").
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))

		err := repo.RevokeRefreshToken(context.Background(), 1, "missing-hash")
		if !errors.Is(err, ErrRefreshTokenNotFound) {
			t.Errorf("mong doi ErrRefreshTokenNotFound, nhan duoc: %v", err)
		}

		checkUserMock(t, mock)
	})
}

// Kiem tra cleanup Refresh Token het han.
func TestUserRepository_DeleteExpiredRefreshTokens(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		mock.ExpectExec("DELETE FROM refresh_tokens").
			WillReturnResult(pgxmock.NewResult("DELETE", 3))

		deleted, err := repo.DeleteExpiredRefreshTokens(context.Background())
		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}
		if deleted != 3 {
			t.Errorf("mong doi xoa 3 token, nhan duoc %d", deleted)
		}

		checkUserMock(t, mock)
	})

	t.Run("database error", func(t *testing.T) {
		mock, repo := newUserMock(t)
		defer mock.Close()

		mock.ExpectExec("DELETE FROM refresh_tokens").
			WillReturnError(errors.New("database error"))

		deleted, err := repo.DeleteExpiredRefreshTokens(context.Background())
		if err == nil {
			t.Fatal("mong doi loi database")
		}
		if deleted != 0 {
			t.Errorf("so token xoa khi gap loi phai la 0, nhan duoc %d", deleted)
		}

		checkUserMock(t, mock)
	})
}
