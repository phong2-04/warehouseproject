package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"warehouseproject/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v5"
)

func newCategoryHandlerMock(t *testing.T) (*CategoryHandler, pgxmock.PgxPoolIface) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(mock.Close)

	repo := repository.NewCategoryRepositoryWithDB(mock)
	return &CategoryHandler{repo: repo}, mock
}

func TestCategoryHandlerList(t *testing.T) {
	t.Run("thanh cong", func(t *testing.T) {
		h, mock := newCategoryHandlerMock(t)

		mock.ExpectQuery("SELECT id, name, description, created_at").
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "name", "description", "created_at",
			}).AddRow(int64(1), "Van phong pham", nil, time.Now()))

		req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("loi database", func(t *testing.T) {
		h, mock := newCategoryHandlerMock(t)

		mock.ExpectQuery("SELECT id, name, description, created_at").
			WillReturnError(errors.New("database error"))

		req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestCategoryHandlerGetByID(t *testing.T) {
	t.Run("thanh cong", func(t *testing.T) {
		h, mock := newCategoryHandlerMock(t)

		mock.ExpectQuery("SELECT id, name, description, created_at").
			WithArgs(int64(1)).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "name", "description", "created_at",
			}).AddRow(int64(1), "Van phong pham", nil, time.Now()))

		req := httptest.NewRequest(http.MethodGet, "/api/categories/1", nil)
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("khong tim thay", func(t *testing.T) {
		h, mock := newCategoryHandlerMock(t)

		mock.ExpectQuery("SELECT id, name, description, created_at").
			WithArgs(int64(99)).
			WillReturnError(repository.ErrCategoryNotFound)

		req := httptest.NewRequest(http.MethodGet, "/api/categories/99", nil)
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("loi database", func(t *testing.T) {
		h, mock := newCategoryHandlerMock(t)

		mock.ExpectQuery("SELECT id, name, description, created_at").
			WithArgs(int64(1)).
			WillReturnError(errors.New("database error"))

		req := httptest.NewRequest(http.MethodGet, "/api/categories/1", nil)
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ID khong hop le", func(t *testing.T) {
		h, _ := newCategoryHandlerMock(t)

		req := httptest.NewRequest(http.MethodGet, "/api/categories/abc", nil)
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestCategoryHandlerCreate(t *testing.T) {
	t.Run("thanh cong", func(t *testing.T) {
		h, mock := newCategoryHandlerMock(t)

		mock.ExpectQuery("(?s).*INSERT INTO categories.*").
			WithArgs("Danh muc moi", pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "created_at",
			}).AddRow(int64(5), time.Now()))

		req := httptest.NewRequest(http.MethodPost, "/api/categories",
			strings.NewReader(`{"name":"  Danh muc moi  ","description":"   "}`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusCreated, rec.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("JSON khong hop le", func(t *testing.T) {
		h, _ := newCategoryHandlerMock(t)

		req := httptest.NewRequest(http.MethodPost, "/api/categories",
			strings.NewReader(`{"name":`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("ten khong hop le", func(t *testing.T) {
		h, _ := newCategoryHandlerMock(t)

		req := httptest.NewRequest(http.MethodPost, "/api/categories",
			strings.NewReader(`{"name":"   "}`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("ten bi trung", func(t *testing.T) {
		h, mock := newCategoryHandlerMock(t)

		mock.ExpectQuery("(?s).*INSERT INTO categories.*").
			WithArgs("Danh muc trung", pgxmock.AnyArg()).
			WillReturnError(&pgconn.PgError{Code: "23505"})

		req := httptest.NewRequest(http.MethodPost, "/api/categories",
			strings.NewReader(`{"name":"Danh muc trung"}`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("loi database", func(t *testing.T) {
		h, mock := newCategoryHandlerMock(t)

		mock.ExpectQuery("(?s).*INSERT INTO categories.*").
			WithArgs("Danh muc loi", pgxmock.AnyArg()).
			WillReturnError(errors.New("database error"))

		req := httptest.NewRequest(http.MethodPost, "/api/categories",
			strings.NewReader(`{"name":"Danh muc loi"}`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestCategoryHandlerUpdate(t *testing.T) {
	t.Run("thanh cong", func(t *testing.T) {
		h, mock := newCategoryHandlerMock(t)

		mock.ExpectExec("(?s).*UPDATE categories.*").
			WithArgs("Danh muc sua", pgxmock.AnyArg(), int64(1)).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		req := httptest.NewRequest(http.MethodPut, "/api/categories/1",
			strings.NewReader(`{"name":"Danh muc sua"}`))
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("khong tim thay", func(t *testing.T) {
		h, mock := newCategoryHandlerMock(t)

		mock.ExpectExec("(?s).*UPDATE categories.*").
			WithArgs("Danh muc sua", pgxmock.AnyArg(), int64(99)).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))

		req := httptest.NewRequest(http.MethodPut, "/api/categories/99",
			strings.NewReader(`{"name":"Danh muc sua"}`))
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ten bi trung", func(t *testing.T) {
		h, mock := newCategoryHandlerMock(t)

		mock.ExpectExec("(?s).*UPDATE categories.*").
			WithArgs("Danh muc trung", pgxmock.AnyArg(), int64(1)).
			WillReturnError(&pgconn.PgError{Code: "23505"})

		req := httptest.NewRequest(http.MethodPut, "/api/categories/1",
			strings.NewReader(`{"name":"Danh muc trung"}`))
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("JSON khong hop le", func(t *testing.T) {
		h, _ := newCategoryHandlerMock(t)

		req := httptest.NewRequest(http.MethodPut, "/api/categories/1",
			strings.NewReader(`{"name":`))
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestCategoryHandlerDelete(t *testing.T) {
	t.Run("thanh cong", func(t *testing.T) {
		h, mock := newCategoryHandlerMock(t)

		mock.ExpectExec("DELETE FROM categories").
			WithArgs(int64(1)).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		req := httptest.NewRequest(http.MethodDelete, "/api/categories/1", nil)
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("khong tim thay", func(t *testing.T) {
		h, mock := newCategoryHandlerMock(t)

		mock.ExpectExec("DELETE FROM categories").
			WithArgs(int64(99)).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		req := httptest.NewRequest(http.MethodDelete, "/api/categories/99", nil)
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("category dang duoc su dung", func(t *testing.T) {
		h, mock := newCategoryHandlerMock(t)

		mock.ExpectExec("DELETE FROM categories").
			WithArgs(int64(1)).
			WillReturnError(&pgconn.PgError{Code: "23503"})

		req := httptest.NewRequest(http.MethodDelete, "/api/categories/1", nil)
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("loi database", func(t *testing.T) {
		h, mock := newCategoryHandlerMock(t)

		mock.ExpectExec("DELETE FROM categories").
			WithArgs(int64(1)).
			WillReturnError(errors.New("database error"))

		req := httptest.NewRequest(http.MethodDelete, "/api/categories/1", nil)
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}
