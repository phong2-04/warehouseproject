package repository

import (
	"context"
	"errors"

	"warehouseproject/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct {
	DB *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{DB: db}
}

func (r *ProductRepository) Create(
	ctx context.Context,
	product *model.Product,
) error {
	query := `
        INSERT INTO products
            (code, name, category_id, price, quantity, min_stock)
        VALUES ($1, $2, $3, $4, 0, $5)
        RETURNING id, created_at, updated_at
    `

	return r.DB.QueryRow(
		ctx,
		query,
		product.Code,
		product.Name,
		product.CategoryID,
		product.Price,
		product.MinStock,
	).Scan(
		&product.ID,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
}

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

// Update chỉ cập nhật thông tin sản phẩm, không cập nhật số lượng tồn kho.
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
            min_stock = $5,
            updated_at = NOW()
        WHERE id = $6
        RETURNING id, code, name, category_id, price,
                  quantity, min_stock, created_at, updated_at
    `

	return r.DB.QueryRow(
		ctx,
		query,
		product.Code,
		product.Name,
		product.CategoryID,
		product.Price,
		product.MinStock,
		product.ID,
	).Scan(
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
}

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

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func (r *ProductRepository) GetLowStock(
	ctx context.Context,
) ([]model.Product, error) {
	query := `
        SELECT id, code, name, category_id, price,
               quantity, min_stock, created_at, updated_at
        FROM products
        WHERE quantity <= min_stock
        ORDER BY quantity ASC, id ASC
    `

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]model.Product, 0)

	for rows.Next() {
		var product model.Product

		if err := rows.Scan(
			&product.ID,
			&product.Code,
			&product.Name,
			&product.CategoryID,
			&product.Price,
			&product.Quantity,
			&product.MinStock,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// IsUniqueViolation kiểm tra lỗi trùng dữ liệu PostgreSQL.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// IsForeignKeyViolation kiểm tra lỗi vi phạm khóa ngoại PostgreSQL.
func IsForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
