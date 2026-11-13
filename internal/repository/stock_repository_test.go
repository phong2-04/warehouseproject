package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	pgxmock "github.com/pashagolub/pgxmock/v5"
)

func TestStockRepositoryProcess(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name            string
		transactionType string
		quantity        int
		currentStock    int
		setup           func(pgxmock.PgxPoolIface)
		wantErr         error
	}{
		{
			name:            "Import success",
			transactionType: "import",
			quantity:        10,
			currentStock:    5,
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT quantity FROM products WHERE id = $1 FOR UPDATE`,
				)).
					WithArgs(int64(1)).
					WillReturnRows(pgxmock.NewRows([]string{"quantity"}).AddRow(5))

				mock.ExpectExec(regexp.QuoteMeta(
					`UPDATE products SET quantity = quantity + $1, updated_at = now() WHERE id = $2`,
				)).
					WithArgs(10, int64(1)).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO stock_transactions
            (product_id, user_id, transaction_type, quantity, note)
         VALUES ($1, $2, $3, $4, $5)
         RETURNING id, product_id, user_id, transaction_type, quantity,
                   COALESCE(note, ''), created_at`,
				)).
					WithArgs(int64(1), int64(2), "import", 10, "Test import").
					WillReturnRows(pgxmock.NewRows([]string{
						"id", "product_id", "user_id", "transaction_type",
						"quantity", "note", "created_at",
					}).AddRow(int64(1), int64(1), int64(2), "import", 10, "Test import", now))

				mock.ExpectCommit()
			},
		},
		{
			name:            "Export success",
			transactionType: "export",
			quantity:        3,
			currentStock:    10,
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT quantity FROM products WHERE id = $1 FOR UPDATE`,
				)).
					WithArgs(int64(1)).
					WillReturnRows(pgxmock.NewRows([]string{"quantity"}).AddRow(10))

				mock.ExpectExec(regexp.QuoteMeta(
					`UPDATE products SET quantity = quantity - $1, updated_at = now() WHERE id = $2`,
				)).
					WithArgs(3, int64(1)).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO stock_transactions
            (product_id, user_id, transaction_type, quantity, note)
         VALUES ($1, $2, $3, $4, $5)
         RETURNING id, product_id, user_id, transaction_type, quantity,
                   COALESCE(note, ''), created_at`,
				)).
					WithArgs(int64(1), int64(2), "export", 3, "Test export").
					WillReturnRows(pgxmock.NewRows([]string{
						"id", "product_id", "user_id", "transaction_type",
						"quantity", "note", "created_at",
					}).AddRow(int64(2), int64(1), int64(2), "export", 3, "Test export", now))

				mock.ExpectCommit()
			},
		},
		{
			name:            "Invalid quantity",
			transactionType: "import",
			quantity:        0,
			wantErr:         ErrInvalidStockQuantity,
			setup:           func(mock pgxmock.PgxPoolIface) {},
		},
		{
			name:            "Invalid transaction type",
			transactionType: "transfer",
			quantity:        5,
			wantErr:         errors.New("loai giao dich khong hop le"),
			setup:           func(mock pgxmock.PgxPoolIface) {},
		},
		{
			name:            "Begin transaction error",
			transactionType: "import",
			quantity:        5,
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin().WillReturnError(errors.New("begin error"))
			},
			wantErr: errors.New("begin error"),
		},
		{
			name:            "Product not found",
			transactionType: "import",
			quantity:        5,
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT quantity FROM products WHERE id = $1 FOR UPDATE`,
				)).
					WithArgs(int64(1)).
					WillReturnError(pgx.ErrNoRows)
				mock.ExpectRollback()
			},
			wantErr: ErrProductNotFound,
		},
		{
			name:            "Query product error",
			transactionType: "import",
			quantity:        5,
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT quantity FROM products WHERE id = $1 FOR UPDATE`,
				)).
					WithArgs(int64(1)).
					WillReturnError(errors.New("query error"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("query error"),
		},
		{
			name:            "Insufficient stock",
			transactionType: "export",
			quantity:        20,
			currentStock:    5,
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT quantity FROM products WHERE id = $1 FOR UPDATE`,
				)).
					WithArgs(int64(1)).
					WillReturnRows(pgxmock.NewRows([]string{"quantity"}).AddRow(5))
				mock.ExpectRollback()
			},
			wantErr: ErrInsufficientStock,
		},
		{
			name:            "Update product error",
			transactionType: "import",
			quantity:        5,
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT quantity FROM products WHERE id = $1 FOR UPDATE`,
				)).
					WithArgs(int64(1)).
					WillReturnRows(pgxmock.NewRows([]string{"quantity"}).AddRow(5))
				mock.ExpectExec(regexp.QuoteMeta(
					`UPDATE products SET quantity = quantity + $1, updated_at = now() WHERE id = $2`,
				)).
					WithArgs(5, int64(1)).
					WillReturnError(errors.New("update error"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("update error"),
		},
		{
			name:            "Insert transaction error",
			transactionType: "import",
			quantity:        5,
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT quantity FROM products WHERE id = $1 FOR UPDATE`,
				)).
					WithArgs(int64(1)).
					WillReturnRows(pgxmock.NewRows([]string{"quantity"}).AddRow(5))
				mock.ExpectExec(regexp.QuoteMeta(
					`UPDATE products SET quantity = quantity + $1, updated_at = now() WHERE id = $2`,
				)).
					WithArgs(5, int64(1)).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO stock_transactions
            (product_id, user_id, transaction_type, quantity, note)
         VALUES ($1, $2, $3, $4, $5)
         RETURNING id, product_id, user_id, transaction_type, quantity,
                   COALESCE(note, ''), created_at`,
				)).
					WithArgs(int64(1), int64(2), "import", 5, "Test import").
					WillReturnError(errors.New("insert error"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("insert error"),
		},
		{
			name:            "Commit error",
			transactionType: "import",
			quantity:        5,
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT quantity FROM products WHERE id = $1 FOR UPDATE`,
				)).
					WithArgs(int64(1)).
					WillReturnRows(pgxmock.NewRows([]string{"quantity"}).AddRow(5))
				mock.ExpectExec(regexp.QuoteMeta(
					`UPDATE products SET quantity = quantity + $1, updated_at = now() WHERE id = $2`,
				)).
					WithArgs(5, int64(1)).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO stock_transactions
            (product_id, user_id, transaction_type, quantity, note)
         VALUES ($1, $2, $3, $4, $5)
         RETURNING id, product_id, user_id, transaction_type, quantity,
                   COALESCE(note, ''), created_at`,
				)).
					WithArgs(int64(1), int64(2), "import", 5, "Test import").
					WillReturnRows(pgxmock.NewRows([]string{
						"id", "product_id", "user_id", "transaction_type",
						"quantity", "note", "created_at",
					}).AddRow(int64(3), int64(1), int64(2), "import", 5, "Test import", now))
				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			wantErr: errors.New("khong the luu giao dich: commit error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			if err != nil {
				t.Fatal(err)
			}
			defer mock.Close()

			tt.setup(mock)

			repo := NewStockRepositoryWithDB(mock)
			note := "Test import"
			if tt.transactionType == "export" {
				note = "Test export"
			}

			result, err := repo.Process(
				ctx,
				1,
				2,
				tt.transactionType,
				tt.quantity,
				note,
			)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("expected transaction record, got nil")
				}
				if result.ProductID != 1 || result.UserID != 2 {
					t.Errorf("unexpected record: %+v", result)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}

func TestStockRepositoryList(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name    string
		setup   func(pgxmock.PgxPoolIface)
		wantLen int
		wantErr bool
	}{
		{
			name: "List success",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, product_id, user_id, transaction_type, quantity,
                COALESCE(note, ''), created_at
         FROM stock_transactions
         ORDER BY created_at DESC, id DESC`,
				)).
					WillReturnRows(pgxmock.NewRows([]string{
						"id", "product_id", "user_id", "transaction_type",
						"quantity", "note", "created_at",
					}).
						AddRow(int64(1), int64(1), int64(2), "import", 10, "Import test", now).
						AddRow(int64(2), int64(1), int64(2), "export", 3, "Export test", now))
			},
			wantLen: 2,
		},
		{
			name: "List empty",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, product_id, user_id, transaction_type, quantity,
                COALESCE(note, ''), created_at
         FROM stock_transactions
         ORDER BY created_at DESC, id DESC`,
				)).
					WillReturnRows(pgxmock.NewRows([]string{
						"id", "product_id", "user_id", "transaction_type",
						"quantity", "note", "created_at",
					}))
			},
			wantLen: 0,
		},
		{
			name: "Query error",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, product_id, user_id, transaction_type, quantity,
                COALESCE(note, ''), created_at
         FROM stock_transactions
         ORDER BY created_at DESC, id DESC`,
				)).
					WillReturnError(errors.New("query error"))
			},
			wantErr: true,
		},
		{
			name: "Scan error",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, product_id, user_id, transaction_type, quantity,
                COALESCE(note, ''), created_at
         FROM stock_transactions
         ORDER BY created_at DESC, id DESC`,
				)).
					WillReturnRows(pgxmock.NewRows([]string{
						"id", "product_id", "user_id", "transaction_type",
						"quantity", "note", "created_at",
					}).AddRow("invalid-id", int64(1), int64(2), "import", 10, "Test", now))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			if err != nil {
				t.Fatal(err)
			}
			defer mock.Close()

			tt.setup(mock)

			repo := NewStockRepositoryWithDB(mock)
			records, err := repo.List(ctx)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(records) != tt.wantLen {
					t.Fatalf("expected %d records, got %d", tt.wantLen, len(records))
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}
