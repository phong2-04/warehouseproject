
package repository

import (
	"context"
	"errors"
	"time"

	"warehouseproject/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProductRepository xử lý truy vấn sản phẩm trong PostgreSQL.
type ProductRepository struct {
	DB *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{DB: db}
}

// Create thêm sản phẩm mới.
func (r *ProductRepository) Create(
	ctx context.Context,
	product *model.Product,
) error {
	query := `
		INSERT INTO products
			(code, name, category_id, price, quantity, min_stock)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	return r.DB.QueryRow(
		ctx,
		query,
		product.Code,
		product.Name,
		product.CategoryID,
		product.Price,
		product.Quantity,
		product.MinStock,
	).Scan(
		&product.ID,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
}

// GetAll lấy danh sách sản phẩm.
func (r *ProductRepository) GetAll(
	ctx context.Context,
) ([]model.Product, error) {
	query := `
		SELECT id, code, name, category_id, price,
		       quantity, min_stock, created_at, updated_at
		FROM products
		ORDER BY id
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]model.Product, 0)

	for rows.Next() {
		var product model.Product

		err := rows.Scan(
			&product.ID,
			&product.Code,
			&product.Name,
			&product.CategoryID,
			&product.Price,
			&product.Quantity,
			&product.MinStock,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// GetByID lấy một sản phẩm theo ID.
func (r *ProductRepository) GetByID(
	ctx context.Context,
	id int64,
) (*model.Product, error) {
	query := `
		SELECT id, code, name, category_id, price,
		       quantity, min_stock, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var product model.Product

	err := r.DB.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.Code,
		&product.Name,
		&product.CategoryID,
		&product.Price,
		&product.Quantity,
		&product.MinStock,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

// Update cập nhật thông tin sản phẩm.
func (r *ProductRepository) Update(
	ctx context.Context,
	product *model.Product,
) error {
	query := `
		UPDATE products
		SET code = $1,
		    name = $2,
		    category_id = $3,
		    price = $4,
		    quantity = $5,
		    min_stock = $6,
		    updated_at = $7
		WHERE id = $8
		RETURNING updated_at
	`

	return r.DB.QueryRow(
		ctx,
		query,
		product.Code,
		product.Name,
		product.CategoryID,
		product.Price,
		product.Quantity,
		product.MinStock,
		time.Now(),
		product.ID,
	).Scan(&product.UpdatedAt)
}

// Delete xóa sản phẩm theo ID.
func (r *ProductRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	tag, err := r.DB.Exec(
		ctx,
		`DELETE FROM products WHERE id = $1`,
		id,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return errors.New("khong tim thay san pham")
	}

	return nil
}

// IsNotFound kiểm tra lỗi không tìm thấy dữ liệu.
func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
