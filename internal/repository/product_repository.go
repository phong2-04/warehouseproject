package repository

import (
	"context"
	"errors"

	"warehouseproject/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrProductVersionConflict = errors.New("product version conflict")

type productDB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type ProductRepository struct {
	DB productDB
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{DB: db}
}

func NewProductRepositoryWithDB(db productDB) *ProductRepository {
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
RETURNING id, version, created_at, updated_at
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
		&product.Version,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
}

func (r *ProductRepository) GetAll(
	ctx context.Context,
) ([]model.Product, error) {
	query := `
SELECT id, code, name, category_id, price,
       quantity, min_stock, created_at, updated_at, version
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
			&product.Version,
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

func (r *ProductRepository) GetByID(
	ctx context.Context,
	id int64,
) (*model.Product, error) {
	query := `
SELECT id, code, name, category_id, price,
       quantity, min_stock, created_at, updated_at, version
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
		&product.Version,
	)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

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
    version = version + 1,
    updated_at = NOW()
WHERE id = $6 AND version = $7
RETURNING id, code, name, category_id, price,
          quantity, min_stock, created_at, updated_at, version
`

	err := r.DB.QueryRow(
		ctx,
		query,
		product.Code,
		product.Name,
		product.CategoryID,
		product.Price,
		product.MinStock,
		product.ID,
		product.Version,
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
		&product.Version,
	)

	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`

	if checkErr := r.DB.QueryRow(ctx, checkQuery, product.ID).Scan(&exists); checkErr != nil {
		return checkErr
	}

	if exists {
		return ErrProductVersionConflict
	}

	return pgx.ErrNoRows
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

func IsProductVersionConflict(err error) bool {
	return errors.Is(err, ErrProductVersionConflict)
}

func (r *ProductRepository) GetLowStock(
	ctx context.Context,
) ([]model.Product, error) {
	query := `
SELECT id, code, name, category_id, price,
       quantity, min_stock, created_at, updated_at, version
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
			&product.Version,
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

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func IsForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
