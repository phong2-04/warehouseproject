
package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"warehouseproject/internal/model"
)

// UserRepository xử lý truy vấn tài khoản trong database.
type UserRepository struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{DB: db}
}

// Tìm tài khoản theo username.
func (r *UserRepository) GetByUsername(
	ctx context.Context,
	username string,
) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, full_name, role
		FROM users
		WHERE username = $1
	`

	user := &model.User{}

	err := r.DB.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}
