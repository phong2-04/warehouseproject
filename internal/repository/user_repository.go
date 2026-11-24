package repository

import (
	"context"
	"errors"
	"time"

	"warehouseproject/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound         = errors.New("khong tim thay nguoi dung")
	ErrLastAdmin            = errors.New("khong the xoa hoac ha quyen admin cuoi cung")
	ErrRefreshTokenNotFound = errors.New("refresh token khong ton tai hoac da het han")
	ErrRefreshTokenReuse    = errors.New("phat hien refresh token bi su dung lai")
)

type userDB interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type UserRepository struct {
	DB userDB
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{DB: db}
}

func NewUserRepositoryWithDB(db userDB) *UserRepository {
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

// Cap nhat mat khau va thu hoi cac Refresh Token dang hoat dong.
func (r *UserRepository) UpdatePasswordHash(
	ctx context.Context,
	userID int64,
	passwordHash string,
) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(
		ctx,
		`UPDATE users SET password_hash = $1 WHERE id = $2`,
		passwordHash,
		userID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE refresh_tokens
         SET revoked_at = NOW()
         WHERE user_id = $1 AND revoked_at IS NULL`,
		userID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
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

// Tao Refresh Token moi trong database.
func (r *UserRepository) CreateRefreshToken(
	ctx context.Context,
	userID int64,
	tokenHash string,
	expiresAt time.Time,
) error {
	_, err := r.DB.Exec(ctx, `
        INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
        VALUES ($1, $2, $3)
    `, userID, tokenHash, expiresAt)

	return err
}

// Thu hoi token cu va luu token moi trong cung mot transaction.
// Rotate Refresh Token và phát hiện token cũ bị sử dụng lại.
func (r *UserRepository) RotateRefreshToken(
	ctx context.Context,
	userID int64,
	oldTokenHash string,
	newTokenHash string,
	expiresAt time.Time,
) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var expired bool
	var revoked bool

	err = tx.QueryRow(ctx, `
        SELECT expires_at <= now(), revoked_at IS NOT NULL
        FROM refresh_tokens
        WHERE user_id = $1 AND token_hash = $2
        FOR UPDATE
    `, userID, oldTokenHash).Scan(&expired, &revoked)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrRefreshTokenNotFound
	}
	if err != nil {
		return err
	}

	if expired {
		return ErrRefreshTokenNotFound
	}

	if revoked {
		// Token đã bị thu hồi nhưng vẫn được sử dụng:
		// thu hồi tất cả Refresh Token còn hiệu lực của tài khoản.
		_, err = tx.Exec(ctx, `
                UPDATE refresh_tokens
                SET revoked_at = now()
                WHERE user_id = $1
                  AND revoked_at IS NULL
                  AND expires_at > now()
            `, userID)
		if err != nil {
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}

		return ErrRefreshTokenReuse
	}

	_, err = tx.Exec(ctx, `
        UPDATE refresh_tokens
        SET revoked_at = now()
        WHERE user_id = $1
          AND token_hash = $2
          AND revoked_at IS NULL
    `, userID, oldTokenHash)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
        VALUES ($1, $2, $3)
    `, userID, newTokenHash, expiresAt)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Thu hoi Refresh Token khi dang xuat.
func (r *UserRepository) RevokeRefreshToken(
	ctx context.Context,
	userID int64,
	tokenHash string,
) error {
	result, err := r.DB.Exec(ctx, `
        UPDATE refresh_tokens
        SET revoked_at = now()
        WHERE user_id = $1
          AND token_hash = $2
          AND revoked_at IS NULL
    `, userID, tokenHash)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrRefreshTokenNotFound
	}

	return nil
}

// Xoa cac Refresh Token da het han.
// Token bi thu hoi nhung chua het han van duoc giu lai de phat hien reuse.
func (r *UserRepository) DeleteExpiredRefreshTokens(ctx context.Context) (int64, error) {
	result, err := r.DB.Exec(ctx, `
        DELETE FROM refresh_tokens
        WHERE expires_at <= now()
    `)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}
