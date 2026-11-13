package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"warehouseproject/internal/middleware"
	"warehouseproject/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v5"
)

func newProductHandlerMock(t *testing.T) (*ProductHandler, pgxmock.PgxPoolIface) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		mock.Close()
	})

	repo := repository.NewProductRepositoryWithDB(mock)
	return NewProductHandler(repo), mock
}

func productTestRows() *pgxmock.Rows {
	return pgxmock.NewRows([]string{
		"id", "code", "name", "category_id", "price",
		"quantity", "min_stock", "created_at", "updated_at",
	}).AddRow(
		int64(1), "SP001", "Ban phim co", nil, 350000.0,
		10, 5, time.Now(), time.Now(),
	)
}

func TestProductHandlerGetAll(t *testing.T) {
	t.Run("thanh cong", func(t *testing.T) {
		h, mock := newProductHandlerMock(t)
		mock.ExpectQuery("SELECT id, code, name").
			WillReturnRows(productTestRows())

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
		h.GetAll(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("loi database", func(t *testing.T) {
		h, mock := newProductHandlerMock(t)
		mock.ExpectQuery("SELECT id, code, name").
			WillReturnError(errors.New("database error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
		h.GetAll(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rec.Code)
		}
	})
}

func TestProductHandlerGetByID(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockError  error
		wantStatus int
	}{
		{"thanh cong", "1", nil, http.StatusOK},
		{"ID khong hop le", "abc", nil, http.StatusBadRequest},
		{"ID bang 0", "0", nil, http.StatusBadRequest},
		{"khong tim thay", "99", pgx.ErrNoRows, http.StatusNotFound},
		{"loi database", "1", errors.New("database error"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, mock := newProductHandlerMock(t)

			if tt.id == "1" {
				if tt.mockError != nil {
					mock.ExpectQuery("SELECT id, code, name").
						WithArgs(int64(1)).
						WillReturnError(tt.mockError)
				} else {
					mock.ExpectQuery("SELECT id, code, name").
						WithArgs(int64(1)).
						WillReturnRows(productTestRows())
				}
			} else if tt.id == "99" {
				mock.ExpectQuery("SELECT id, code, name").
					WithArgs(int64(99)).
					WillReturnError(tt.mockError)
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/products/"+tt.id, nil)
			h.GetByID(rec, req, tt.id)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s",
					rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestProductHandlerCreate(t *testing.T) {
	t.Run("thanh cong", func(t *testing.T) {
		h, mock := newProductHandlerMock(t)

		mock.ExpectQuery("INSERT INTO products").
			WithArgs("SP001", "Ban phim co", pgxmock.AnyArg(), 350000.0, 5).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "created_at", "updated_at",
			}).AddRow(int64(1), time.Now(), time.Now()))

		body := `{"code":" SP001 ","name":" Ban phim co ","price":350000,"quantity":99,"min_stock":5}`
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/products", strings.NewReader(body))
		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	tests := []struct {
		name       string
		body       string
		mockError  error
		wantStatus int
	}{
		{"JSON khong hop le", `{"code":`, nil, http.StatusBadRequest},
		{"du lieu khong hop le", `{"code":"","name":"Test","price":100,"min_stock":5}`, nil, http.StatusBadRequest},
		{"ma san pham bi trung", `{"code":"SP001","name":"Test","price":100,"min_stock":5}`, &pgconn.PgError{Code: "23505"}, http.StatusConflict},
		{"category khong ton tai", `{"code":"SP002","name":"Test","category_id":1,"price":100,"min_stock":5}`, &pgconn.PgError{Code: "23503"}, http.StatusBadRequest},
		{"loi database", `{"code":"SP003","name":"Test","price":100,"min_stock":5}`, errors.New("database error"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, mock := newProductHandlerMock(t)

			if tt.mockError != nil {
				switch tt.name {
				case "category khong ton tai":
					mock.ExpectQuery("INSERT INTO products").
						WithArgs("SP002", "Test", pgxmock.AnyArg(), 100.0, 5).
						WillReturnError(tt.mockError)
				case "ma san pham bi trung":
					mock.ExpectQuery("INSERT INTO products").
						WithArgs("SP001", "Test", pgxmock.AnyArg(), 100.0, 5).
						WillReturnError(tt.mockError)
				default:
					mock.ExpectQuery("INSERT INTO products").
						WithArgs("SP003", "Test", pgxmock.AnyArg(), 100.0, 5).
						WillReturnError(tt.mockError)
				}
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/products", strings.NewReader(tt.body))
			h.Create(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s",
					rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestProductHandlerUpdate(t *testing.T) {
	validBody := `{"code":"SP002","name":"Chuot khong day","price":200000,"min_stock":3}`

	t.Run("thanh cong", func(t *testing.T) {
		h, mock := newProductHandlerMock(t)

		mock.ExpectQuery("SELECT id, code, name").
			WithArgs(int64(1)).
			WillReturnRows(productTestRows())

		mock.ExpectQuery("UPDATE products").
			WithArgs("SP002", "Chuot khong day", pgxmock.AnyArg(), 200000.0, 3, int64(1)).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "code", "name", "category_id", "price",
				"quantity", "min_stock", "created_at", "updated_at",
			}).AddRow(int64(1), "SP002", "Chuot khong day", nil, 200000.0,
				10, 3, time.Now(), time.Now()))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/products/1", strings.NewReader(validBody))
		h.Update(rec, req, "1")

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("ID khong hop le", func(t *testing.T) {
		h, _ := newProductHandlerMock(t)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/products/abc", strings.NewReader(validBody))
		h.Update(rec, req, "abc")

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("JSON khong hop le", func(t *testing.T) {
		h, _ := newProductHandlerMock(t)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/products/1", strings.NewReader(`{"code":`))
		h.Update(rec, req, "1")

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("khong tim thay khi lay san pham", func(t *testing.T) {
		h, mock := newProductHandlerMock(t)
		mock.ExpectQuery("SELECT id, code, name").
			WithArgs(int64(99)).
			WillReturnError(pgx.ErrNoRows)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/products/99", strings.NewReader(validBody))
		h.Update(rec, req, "99")

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", rec.Code)
		}
	})

	t.Run("loi database khi lay san pham", func(t *testing.T) {
		h, mock := newProductHandlerMock(t)
		mock.ExpectQuery("SELECT id, code, name").
			WithArgs(int64(1)).
			WillReturnError(errors.New("database error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/products/1", strings.NewReader(validBody))
		h.Update(rec, req, "1")

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rec.Code)
		}
	})

	t.Run("du lieu cap nhat khong hop le", func(t *testing.T) {
		h, mock := newProductHandlerMock(t)
		mock.ExpectQuery("SELECT id, code, name").
			WithArgs(int64(1)).
			WillReturnRows(productTestRows())

		body := `{"code":"","name":"Test","price":100,"min_stock":5}`
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/products/1", strings.NewReader(body))
		h.Update(rec, req, "1")

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	updateErrors := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"khong tim thay khi cap nhat", pgx.ErrNoRows, http.StatusNotFound},
		{"ma san pham bi trung", &pgconn.PgError{Code: "23505"}, http.StatusConflict},
		{"category khong ton tai", &pgconn.PgError{Code: "23503"}, http.StatusBadRequest},
		{"loi database khi cap nhat", errors.New("database error"), http.StatusInternalServerError},
	}

	for _, tt := range updateErrors {
		t.Run(tt.name, func(t *testing.T) {
			h, mock := newProductHandlerMock(t)

			mock.ExpectQuery("SELECT id, code, name").
				WithArgs(int64(1)).
				WillReturnRows(productTestRows())

			mock.ExpectQuery("UPDATE products").
				WithArgs("SP002", "Chuot khong day", pgxmock.AnyArg(), 200000.0, 3, int64(1)).
				WillReturnError(tt.err)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/api/products/1", strings.NewReader(validBody))
			h.Update(rec, req, "1")

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s",
					rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestProductHandlerDelete(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockError  error
		rows       int64
		wantStatus int
	}{
		{"thanh cong", "1", nil, 1, http.StatusNoContent},
		{"ID khong hop le", "abc", nil, 0, http.StatusBadRequest},
		{"khong tim thay", "99", nil, 0, http.StatusNotFound},
		{"san pham dang duoc su dung", "1", &pgconn.PgError{Code: "23503"}, 0, http.StatusConflict},
		{"loi database", "1", errors.New("database error"), 0, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, mock := newProductHandlerMock(t)

			if tt.id != "abc" {
				if tt.mockError != nil {
					mock.ExpectExec("DELETE FROM products").
						WithArgs(int64(1)).
						WillReturnError(tt.mockError)
				} else {
					mock.ExpectExec("DELETE FROM products").
						WithArgs(func() int64 {
							if tt.id == "99" {
								return 99
							}
							return 1
						}()).
						WillReturnResult(pgxmock.NewResult("DELETE", tt.rows))
				}
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/api/products/"+tt.id, nil)
			h.Delete(rec, req, tt.id)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s",
					rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestProductHandlerGetLowStock(t *testing.T) {
	t.Run("thanh cong", func(t *testing.T) {
		h, mock := newProductHandlerMock(t)
		mock.ExpectQuery("WHERE quantity <= min_stock").
			WillReturnRows(productTestRows())

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/products/low-stock", nil)
		h.GetLowStock(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("loi database", func(t *testing.T) {
		h, mock := newProductHandlerMock(t)
		mock.ExpectQuery("WHERE quantity <= min_stock").
			WillReturnError(errors.New("database error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/products/low-stock", nil)
		h.GetLowStock(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rec.Code)
		}
	})
}

func TestProductHandlerServeHTTP(t *testing.T) {
	t.Run("staff khong duoc them san pham", func(t *testing.T) {
		h, _ := newProductHandlerMock(t)
		ctx := context.WithValue(context.Background(), middleware.RoleKey, "staff")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/products", strings.NewReader(`{}`)).WithContext(ctx)
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", rec.Code)
		}
	})

	t.Run("duong dan POST khong hop le", func(t *testing.T) {
		h, _ := newProductHandlerMock(t)
		ctx := context.WithValue(context.Background(), middleware.RoleKey, "admin")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/products/1", nil).WithContext(ctx)
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", rec.Code)
		}
	})

	t.Run("staff khong duoc sua", func(t *testing.T) {
		h, _ := newProductHandlerMock(t)
		ctx := context.WithValue(context.Background(), middleware.RoleKey, "staff")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/products/1", nil).WithContext(ctx)
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", rec.Code)
		}
	})

	t.Run("staff khong duoc xoa", func(t *testing.T) {
		h, _ := newProductHandlerMock(t)
		ctx := context.WithValue(context.Background(), middleware.RoleKey, "staff")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/api/products/1", nil).WithContext(ctx)
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", rec.Code)
		}
	})

	t.Run("method khong ho tro", func(t *testing.T) {
		h, _ := newProductHandlerMock(t)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/products", nil)
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want 405", rec.Code)
		}
	})

	t.Run("admin duoc them san pham", func(t *testing.T) {
		h, mock := newProductHandlerMock(t)
		mock.ExpectQuery("INSERT INTO products").
			WithArgs("SP001", "Ban phim co", pgxmock.AnyArg(), 350000.0, 5).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "created_at", "updated_at",
			}).AddRow(int64(1), time.Now(), time.Now()))

		ctx := context.WithValue(context.Background(), middleware.RoleKey, "admin")
		body := `{"code":"SP001","name":"Ban phim co","price":350000,"min_stock":5}`

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/products", strings.NewReader(body)).WithContext(ctx)
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body.String())
		}
	})
}
