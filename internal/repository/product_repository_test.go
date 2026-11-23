package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"warehouseproject/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v5"
)

func newProductMock(t *testing.T) (pgxmock.PgxPoolIface, *ProductRepository) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("khong tao duoc mock: %v", err)
	}

	return mock, NewProductRepositoryWithDB(mock)
}

func checkProductMock(t *testing.T, mock pgxmock.PgxPoolIface) {
	t.Helper()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}

func productColumns() []string {
	return []string{
		"id", "code", "name", "category_id", "price",
		"quantity", "min_stock", "created_at", "updated_at", "version",
	}
}

func productRow() *model.Product {
	categoryID := int64(1)
	now := time.Now()

	return &model.Product{
		ID:         1,
		Code:       "SP001",
		Name:       "Ban phim co",
		CategoryID: &categoryID,
		Price:      350000,
		Quantity:   10,
		MinStock:   5,
		Version:    1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func addProductRow(rows *pgxmock.Rows, product *model.Product) *pgxmock.Rows {
	return rows.AddRow(
		product.ID,
		product.Code,
		product.Name,
		product.CategoryID,
		product.Price,
		product.Quantity,
		product.MinStock,
		product.CreatedAt,
		product.UpdatedAt,
		product.Version,
	)
}

func TestProductRepository_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newProductMock(t)
		defer mock.Close()

		now := time.Now()
		categoryID := int64(1)

		product := &model.Product{
			Code:       "SP001",
			Name:       "Ban phim co",
			CategoryID: &categoryID,
			Price:      350000,
			MinStock:   5,
		}

		mock.ExpectQuery("INSERT INTO products").
			WithArgs(product.Code, product.Name, product.CategoryID, product.Price, product.MinStock).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "version", "created_at", "updated_at"}).
					AddRow(int64(1), int64(1), now, now),
			)

		err := repo.Create(context.Background(), product)

		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}
		if product.ID != 1 || product.Version != 1 {
			t.Errorf("ID/version khong dung: ID=%d Version=%d", product.ID, product.Version)
		}
		if !product.CreatedAt.Equal(now) || !product.UpdatedAt.Equal(now) {
			t.Error("thoi gian tao hoac cap nhat khong dung")
		}

		checkProductMock(t, mock)
	})

	t.Run("database error", func(t *testing.T) {
		mock, repo := newProductMock(t)
		defer mock.Close()

		product := &model.Product{
			Code: "SP001",
			Name: "Ban phim co",
		}

		mock.ExpectQuery("INSERT INTO products").
			WithArgs(product.Code, product.Name, product.CategoryID, product.Price, product.MinStock).
			WillReturnError(errors.New("database error"))

		err := repo.Create(context.Background(), product)

		if err == nil {
			t.Fatal("mong doi loi khi tao san pham")
		}

		checkProductMock(t, mock)
	})
}

func TestProductRepository_GetAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newProductMock(t)
		defer mock.Close()

		product := productRow()
		rows := pgxmock.NewRows(productColumns())
		addProductRow(rows, product)

		mock.ExpectQuery("SELECT id, code, name, category_id").
			WillReturnRows(rows)

		result, err := repo.GetAll(context.Background())

		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("mong doi 1 san pham, nhan duoc %d", len(result))
		}
		if result[0].Code != "SP001" || result[0].Quantity != 10 || result[0].Version != 1 {
			t.Errorf("du lieu san pham khong dung: %+v", result[0])
		}

		checkProductMock(t, mock)
	})

	t.Run("query error", func(t *testing.T) {
		mock, repo := newProductMock(t)
		defer mock.Close()

		mock.ExpectQuery("SELECT id, code, name, category_id").
			WillReturnError(errors.New("database error"))

		_, err := repo.GetAll(context.Background())

		if err == nil {
			t.Fatal("mong doi loi truy van")
		}

		checkProductMock(t, mock)
	})
}

func TestProductRepository_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newProductMock(t)
		defer mock.Close()

		product := productRow()
		rows := pgxmock.NewRows(productColumns())
		addProductRow(rows, product)

		mock.ExpectQuery("SELECT id, code, name, category_id").
			WithArgs(int64(1)).
			WillReturnRows(rows)

		result, err := repo.GetByID(context.Background(), 1)

		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}
		if result.ID != 1 || result.Code != "SP001" || result.Version != 1 {
			t.Errorf("du lieu san pham khong dung: %+v", result)
		}

		checkProductMock(t, mock)
	})

	t.Run("not found", func(t *testing.T) {
		mock, repo := newProductMock(t)
		defer mock.Close()

		mock.ExpectQuery("SELECT id, code, name, category_id").
			WithArgs(int64(99)).
			WillReturnError(pgx.ErrNoRows)

		_, err := repo.GetByID(context.Background(), 99)

		if !IsNotFound(err) {
			t.Errorf("mong doi loi ErrNoRows, nhan duoc: %v", err)
		}

		checkProductMock(t, mock)
	})

	t.Run("database error", func(t *testing.T) {
		mock, repo := newProductMock(t)
		defer mock.Close()

		mock.ExpectQuery("SELECT id, code, name, category_id").
			WithArgs(int64(1)).
			WillReturnError(errors.New("database error"))

		_, err := repo.GetByID(context.Background(), 1)

		if err == nil {
			t.Fatal("mong doi loi truy van")
		}

		checkProductMock(t, mock)
	})
}

func TestProductRepository_Update(t *testing.T) {
	t.Run("success increments version", func(t *testing.T) {
		mock, repo := newProductMock(t)
		defer mock.Close()

		product := productRow()

		product.Name = "Ban phim moi"

		updatedProduct := *product

		updatedProduct.Version = 2

		rows := pgxmock.NewRows(productColumns())

		addProductRow(rows, &updatedProduct)

		mock.ExpectQuery("UPDATE products").
			WithArgs(
				product.Code,
				product.Name,
				product.CategoryID,
				product.Price,
				product.MinStock,
				product.ID,
				int64(1),
			).
			WillReturnRows(rows)

		err := repo.Update(context.Background(), product)

		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}
		if product.Name != "Ban phim moi" || product.Version != 2 {
			t.Errorf("du lieu sau cap nhat khong dung: %+v", product)
		}

		checkProductMock(t, mock)
	})

	t.Run("version conflict", func(t *testing.T) {
		mock, repo := newProductMock(t)
		defer mock.Close()

		product := productRow()

		mock.ExpectQuery("UPDATE products").
			WithArgs(
				product.Code,
				product.Name,
				product.CategoryID,
				product.Price,
				product.MinStock,
				product.ID,
				product.Version,
			).
			WillReturnError(pgx.ErrNoRows)

		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(product.ID).
			WillReturnRows(
				pgxmock.NewRows([]string{"exists"}).AddRow(true),
			)

		err := repo.Update(context.Background(), product)

		if !IsProductVersionConflict(err) {
			t.Errorf("mong doi loi version conflict, nhan duoc: %v", err)
		}

		checkProductMock(t, mock)
	})

	t.Run("not found", func(t *testing.T) {
		mock, repo := newProductMock(t)
		defer mock.Close()

		product := productRow()
		product.ID = 99

		mock.ExpectQuery("UPDATE products").
			WithArgs(
				product.Code,
				product.Name,
				product.CategoryID,
				product.Price,
				product.MinStock,
				product.ID,
				product.Version,
			).
			WillReturnError(pgx.ErrNoRows)

		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(product.ID).
			WillReturnRows(
				pgxmock.NewRows([]string{"exists"}).AddRow(false),
			)

		err := repo.Update(context.Background(), product)

		if !IsNotFound(err) {
			t.Errorf("mong doi loi ErrNoRows, nhan duoc: %v", err)
		}

		checkProductMock(t, mock)
	})

	t.Run("database error", func(t *testing.T) {
		mock, repo := newProductMock(t)
		defer mock.Close()

		product := productRow()

		mock.ExpectQuery("UPDATE products").
			WithArgs(
				product.Code,
				product.Name,
				product.CategoryID,
				product.Price,
				product.MinStock,
				product.ID,
				product.Version,
			).
			WillReturnError(errors.New("database error"))

		err := repo.Update(context.Background(), product)

		if err == nil {
			t.Fatal("mong doi loi cap nhat")
		}

		checkProductMock(t, mock)
	})
}

func TestProductRepository_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newProductMock(t)
		defer mock.Close()

		mock.ExpectExec("DELETE FROM products").
			WithArgs(int64(1)).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.Delete(context.Background(), 1)

		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}

		checkProductMock(t, mock)
	})

	t.Run("not found", func(t *testing.T) {
		mock, repo := newProductMock(t)
		defer mock.Close()

		mock.ExpectExec("DELETE FROM products").
			WithArgs(int64(99)).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err := repo.Delete(context.Background(), 99)

		if err == nil || err.Error() != "khong tim thay san pham" {
			t.Errorf("mong doi loi khong tim thay san pham, nhan duoc: %v", err)
		}

		checkProductMock(t, mock)
	})

	t.Run("database error", func(t *testing.T) {
		mock, repo := newProductMock(t)
		defer mock.Close()

		mock.ExpectExec("DELETE FROM products").
			WithArgs(int64(1)).
			WillReturnError(errors.New("database error"))

		err := repo.Delete(context.Background(), 1)

		if err == nil {
			t.Fatal("mong doi loi xoa san pham")
		}

		checkProductMock(t, mock)
	})
}

func TestProductRepository_GetLowStock(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, repo := newProductMock(t)
		defer mock.Close()

		product := productRow()
		product.Quantity = 2
		product.MinStock = 5

		rows := pgxmock.NewRows(productColumns())
		addProductRow(rows, product)

		mock.ExpectQuery("WHERE quantity <= min_stock").
			WillReturnRows(rows)

		result, err := repo.GetLowStock(context.Background())

		if err != nil {
			t.Fatalf("khong mong doi loi: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("mong doi 1 san pham sap het hang, nhan duoc %d", len(result))
		}
		if result[0].Quantity > result[0].MinStock || result[0].Version != 1 {
			t.Error("du lieu san pham sap het hang khong dung")
		}

		checkProductMock(t, mock)
	})

	t.Run("query error", func(t *testing.T) {
		mock, repo := newProductMock(t)
		defer mock.Close()

		mock.ExpectQuery("WHERE quantity <= min_stock").
			WillReturnError(errors.New("database error"))

		_, err := repo.GetLowStock(context.Background())

		if err == nil {
			t.Fatal("mong doi loi truy van")
		}

		checkProductMock(t, mock)
	})
}

func TestIsNotFound(t *testing.T) {
	if !IsNotFound(pgx.ErrNoRows) {
		t.Error("mong doi true voi pgx.ErrNoRows")
	}

	if IsNotFound(errors.New("database error")) {
		t.Error("mong doi false voi loi khac")
	}
}

func TestIsProductVersionConflict(t *testing.T) {
	if !IsProductVersionConflict(ErrProductVersionConflict) {
		t.Error("mong doi true voi version conflict")
	}

	if IsProductVersionConflict(errors.New("database error")) {
		t.Error("mong doi false voi loi khac")
	}
}

func TestIsUniqueViolation(t *testing.T) {
	duplicateErr := &pgconn.PgError{Code: "23505"}

	if !IsUniqueViolation(duplicateErr) {
		t.Error("mong doi true voi loi trung du lieu")
	}

	if IsUniqueViolation(errors.New("database error")) {
		t.Error("mong doi false voi loi thong thuong")
	}
}

func TestIsForeignKeyViolation(t *testing.T) {
	foreignKeyErr := &pgconn.PgError{Code: "23503"}

	if !IsForeignKeyViolation(foreignKeyErr) {
		t.Error("mong doi true voi loi khoa ngoai")
	}

	if IsForeignKeyViolation(errors.New("database error")) {
		t.Error("mong doi false voi loi thong thuong")
	}
}
