package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidStockQuantity = errors.New("so luong phai lon hon 0")
	ErrProductNotFound      = errors.New("khong tim thay san pham")
	ErrInsufficientStock    = errors.New("so luong ton kho khong du")
)

type StockTransaction struct {
	ID              int64  `json:"id"`
	ProductID       int64  `json:"product_id"`
	UserID          int64  `json:"user_id"`
	TransactionType string `json:"transaction_type"`
	Quantity        int    `json:"quantity"`
	Note            string `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

type StockRepository struct {
	db *pgxpool.Pool
}

func NewStockRepository(db *pgxpool.Pool) *StockRepository {
	return &StockRepository{db: db}
}

func (repo *StockRepository) Process(
	ctx context.Context,
	productID int64,
	userID int64,
	transactionType string,
	quantity int,
	note string,
) (*StockTransaction, error) {
	if quantity <= 0 {
		return nil, ErrInvalidStockQuantity
	}

	if transactionType != "import" && transactionType != "export" {
		return nil, errors.New("loai giao dich khong hop le")
	}

	tx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var currentQuantity int
	err = tx.QueryRow(
		ctx,
		`SELECT quantity FROM products WHERE id = $1 FOR UPDATE`,
		productID,
	).Scan(&currentQuantity)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}

	if transactionType == "export" && currentQuantity < quantity {
		return nil, ErrInsufficientStock
	}

	if transactionType == "import" {
		_, err = tx.Exec(
			ctx,
			`UPDATE products SET quantity = quantity + $1, updated_at = now() WHERE id = $2`,
			quantity, productID,
		)
	} else {
		_, err = tx.Exec(
			ctx,
			`UPDATE products SET quantity = quantity - $1, updated_at = now() WHERE id = $2`,
			quantity, productID,
		)
	}
	if err != nil {
		return nil, err
	}

	var record StockTransaction
	err = tx.QueryRow(
		ctx,
		`INSERT INTO stock_transactions
            (product_id, user_id, transaction_type, quantity, note)
         VALUES ($1, $2, $3, $4, $5)
         RETURNING id, product_id, user_id, transaction_type, quantity,
                   COALESCE(note, ''), created_at`,
		productID, userID, transactionType, quantity, note,
	).Scan(
		&record.ID,
		&record.ProductID,
		&record.UserID,
		&record.TransactionType,
		&record.Quantity,
		&record.Note,
		&record.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("khong the luu giao dich: %w", err)
	}

	return &record, nil
}

func (repo *StockRepository) List(
	ctx context.Context,
) ([]StockTransaction, error) {
	rows, err := repo.db.Query(
		ctx,
		`SELECT id, product_id, user_id, transaction_type, quantity,
                COALESCE(note, ''), created_at
         FROM stock_transactions
         ORDER BY created_at DESC, id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]StockTransaction, 0)
	for rows.Next() {
		var record StockTransaction
		if err := rows.Scan(
			&record.ID,
			&record.ProductID,
			&record.UserID,
			&record.TransactionType,
			&record.Quantity,
			&record.Note,
			&record.CreatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}


