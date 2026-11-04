package repository

import (
	"context"
	"errors"

	"warehouseproject/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound = errors.New("khong tim thay nguoi dung")
	ErrLastAdmin    = errors.New("khong the xoa hoac ha quyen admin cuoi cung")
)

type UserRepository struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{DB: db}
}

// Tim tai khoan theo username.
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

// Lay danh sach nguoi dung.
func (r *UserRepository) GetAll(
	ctx context.Context,
) ([]model.User, error) {
	query := `
        SELECT id, username, password_hash, full_name, role
        FROM users
        ORDER BY id
    `

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]model.User, 0)

	for rows.Next() {
		var user model.User
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.FullName,
			&user.Role,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Tim nguoi dung theo ID.
func (r *UserRepository) GetByID(
	ctx context.Context,
	id int64,
) (*model.User, error) {
	query := `
        SELECT id, username, password_hash, full_name, role
        FROM users
        WHERE id = $1
    `

	var user model.User
	err := r.DB.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// Tao tai khoan moi.
func (r *UserRepository) Create(
	ctx context.Context,
	user *model.User,
) error {
	query := `
        INSERT INTO users (username, password_hash, full_name, role)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `

	return r.DB.QueryRow(
		ctx,
		query,
		user.Username,
		user.PasswordHash,
		user.FullName,
		user.Role,
	).Scan(&user.ID)
}

// Cap nhat thong tin nguoi dung.
func (r *UserRepository) Update(
	ctx context.Context,
	user *model.User,
) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Khoa cac tai khoan admin de tranh hai yeu cau
	// dong thoi cung ha quyen admin cuoi cung.
	rows, err := tx.Query(
		ctx,
		`SELECT id FROM users WHERE role = 'admin' ORDER BY id FOR UPDATE`,
	)
	if err != nil {
		return err
	}

	adminIDs := make(map[int64]bool)

	for rows.Next() {
		var adminID int64
		if err := rows.Scan(&adminID); err != nil {
			rows.Close()
			return err
		}
		adminIDs[adminID] = true
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	// Khong cho ha quyen admin cuoi cung.
	if adminIDs[user.ID] && user.Role != "admin" && len(adminIDs) <= 1 {
		return ErrLastAdmin
	}

	result, err := tx.Exec(
		ctx,
		`
            UPDATE users
            SET username = $1, full_name = $2, role = $3
            WHERE id = $4
        `,
		user.Username,
		user.FullName,
		user.Role,
		user.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return tx.Commit(ctx)
}

// Xoa tai khoan theo ID.
func (r *UserRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Khoa cac tai khoan admin truoc khi kiem tra.
	rows, err := tx.Query(
		ctx,
		`SELECT id FROM users WHERE role = 'admin' ORDER BY id FOR UPDATE`,
	)
	if err != nil {
		return err
	}

	adminIDs := make(map[int64]bool)

	for rows.Next() {
		var adminID int64
		if err := rows.Scan(&adminID); err != nil {
			rows.Close()
			return err
		}
		adminIDs[adminID] = true
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	// Khong cho xoa admin cuoi cung.
	if adminIDs[id] && len(adminIDs) <= 1 {
		return ErrLastAdmin
	}

	result, err := tx.Exec(
		ctx,
		`DELETE FROM users WHERE id = $1`,
		id,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return tx.Commit(ctx)
}
