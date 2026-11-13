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

func newCategoryMock(t *testing.T) (pgxmock.PgxPoolIface, *CategoryRepository) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("khong tao duoc mock: %v", err)
	}

	repo := NewCategoryRepositoryWithDB(mock)
	return mock, repo
}

func checkMockExpectations(t *testing.T, mock pgxmock.PgxPoolIface) {
	t.Helper()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}

func TestCategoryRepository_GetAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newCategoryMock(t)
		defer mock.Close()

		now := time.Now()
		description := "Thiet bi van phong"

		mock.ExpectQuery("SELECT id, name, description, created_at").
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "name", "description", "created_at"}).
					AddRow(int64(1), "Van phong", &description, now),
			)

		result, err := repo.GetAll(context.Background())

		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("mong doi 1 danh muc, nhan duoc %d", len(result))
		}
		if result[0].Name != "Van phong" {
			t.Errorf("ten danh muc khong dung: %s", result[0].Name)
		}

		checkMockExpectations(t, mock)
	})

	t.Run("query error", func(t *testing.T) {
		mock, repo := newCategoryMock(t)
		defer mock.Close()

		mock.ExpectQuery("SELECT id, name, description, created_at").
			WillReturnError(errors.New("database error"))

		_, err := repo.GetAll(context.Background())

		if err == nil {
			t.Fatal("mong doi loi truy van")
		}

		checkMockExpectations(t, mock)
	})
}

func TestCategoryRepository_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newCategoryMock(t)
		defer mock.Close()

		now := time.Now()
		description := "Mo ta danh muc"

		mock.ExpectQuery("SELECT id, name, description, created_at").
			WithArgs(int64(1)).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "name", "description", "created_at"}).
					AddRow(int64(1), "Van phong", &description, now),
			)

		result, err := repo.GetByID(context.Background(), 1)

		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}
		if result.ID != 1 || result.Name != "Van phong" {
			t.Errorf("du lieu danh muc khong dung: %+v", result)
		}

		checkMockExpectations(t, mock)
	})

	t.Run("not found", func(t *testing.T) {
		mock, repo := newCategoryMock(t)
		defer mock.Close()

		mock.ExpectQuery("SELECT id, name, description, created_at").
			WithArgs(int64(99)).
			WillReturnError(pgx.ErrNoRows)

		_, err := repo.GetByID(context.Background(), 99)

		if !errors.Is(err, ErrCategoryNotFound) {
			t.Errorf("mong doi ErrCategoryNotFound, nhan duoc: %v", err)
		}

		checkMockExpectations(t, mock)
	})

	t.Run("database error", func(t *testing.T) {
		mock, repo := newCategoryMock(t)
		defer mock.Close()

		mock.ExpectQuery("SELECT id, name, description, created_at").
			WithArgs(int64(1)).
			WillReturnError(errors.New("database error"))

		_, err := repo.GetByID(context.Background(), 1)

		if err == nil {
			t.Fatal("mong doi loi truy van")
		}

		checkMockExpectations(t, mock)
	})
}

func TestCategoryRepository_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newCategoryMock(t)
		defer mock.Close()

		now := time.Now()
		description := "Mo ta danh muc"
		category := &model.Category{
			Name:        "Van phong",
			Description: &description,
		}

		mock.ExpectQuery("INSERT INTO categories").
			WithArgs(category.Name, category.Description).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "created_at"}).
					AddRow(int64(1), now),
			)

		err := repo.Create(context.Background(), category)

		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}
		if category.ID != 1 {
			t.Errorf("ID mong doi 1, nhan duoc %d", category.ID)
		}
		if !category.CreatedAt.Equal(now) {
			t.Errorf("CreatedAt khong dung")
		}

		checkMockExpectations(t, mock)
	})

	t.Run("database error", func(t *testing.T) {
		mock, repo := newCategoryMock(t)
		defer mock.Close()

		description := "Mo ta"
		category := &model.Category{
			Name:        "Van phong",
			Description: &description,
		}

		mock.ExpectQuery("INSERT INTO categories").
			WithArgs(category.Name, category.Description).
			WillReturnError(errors.New("database error"))

		err := repo.Create(context.Background(), category)

		if err == nil {
			t.Fatal("mong doi loi khi tao danh muc")
		}

		checkMockExpectations(t, mock)
	})
}

func TestCategoryRepository_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newCategoryMock(t)
		defer mock.Close()

		description := "Mo ta moi"
		category := &model.Category{
			ID:          1,
			Name:        "Van phong moi",
			Description: &description,
		}

		mock.ExpectExec("UPDATE categories").
			WithArgs(category.Name, category.Description, category.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.Update(context.Background(), category)

		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}

		checkMockExpectations(t, mock)
	})

	t.Run("not found", func(t *testing.T) {
		mock, repo := newCategoryMock(t)
		defer mock.Close()

		description := "Mo ta"
		category := &model.Category{
			ID:          99,
			Name:        "Khong ton tai",
			Description: &description,
		}

		mock.ExpectExec("UPDATE categories").
			WithArgs(category.Name, category.Description, category.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))

		err := repo.Update(context.Background(), category)

		if !errors.Is(err, ErrCategoryNotFound) {
			t.Errorf("mong doi ErrCategoryNotFound, nhan duoc: %v", err)
		}

		checkMockExpectations(t, mock)
	})

	t.Run("database error", func(t *testing.T) {
		mock, repo := newCategoryMock(t)
		defer mock.Close()

		description := "Mo ta"
		category := &model.Category{
			ID:          1,
			Name:        "Van phong",
			Description: &description,
		}

		mock.ExpectExec("UPDATE categories").
			WithArgs(category.Name, category.Description, category.ID).
			WillReturnError(errors.New("database error"))

		err := repo.Update(context.Background(), category)

		if err == nil {
			t.Fatal("mong doi loi cap nhat")
		}

		checkMockExpectations(t, mock)
	})
}

func TestCategoryRepository_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newCategoryMock(t)
		defer mock.Close()

		mock.ExpectExec("DELETE FROM categories").
			WithArgs(int64(1)).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.Delete(context.Background(), 1)

		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}

		checkMockExpectations(t, mock)
	})

	t.Run("not found", func(t *testing.T) {
		mock, repo := newCategoryMock(t)
		defer mock.Close()

		mock.ExpectExec("DELETE FROM categories").
			WithArgs(int64(99)).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err := repo.Delete(context.Background(), 99)

		if !errors.Is(err, ErrCategoryNotFound) {
			t.Errorf("mong doi ErrCategoryNotFound, nhan duoc: %v", err)
		}

		checkMockExpectations(t, mock)
	})

	t.Run("database error", func(t *testing.T) {
		mock, repo := newCategoryMock(t)
		defer mock.Close()

		mock.ExpectExec("DELETE FROM categories").
			WithArgs(int64(1)).
			WillReturnError(errors.New("database error"))

		err := repo.Delete(context.Background(), 1)

		if err == nil {
			t.Fatal("mong doi loi xoa danh muc")
		}

		checkMockExpectations(t, mock)
	})
}
