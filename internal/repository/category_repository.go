package repository

import (
    "context"
    "errors"

    "warehouseproject/internal/model"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

var ErrCategoryNotFound = errors.New("khong tim thay danh muc")

type CategoryRepository struct {
    db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
    return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetAll(ctx context.Context) ([]model.Category, error) {
    rows, err := r.db.Query(ctx,
        `SELECT id, name, description, created_at
         FROM categories ORDER BY id`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    categories := make([]model.Category, 0)
    for rows.Next() {
        var c model.Category
        if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt); err != nil {
            return nil, err
        }
        categories = append(categories, c)
    }

    return categories, rows.Err()
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*model.Category, error) {
    var c model.Category
    err := r.db.QueryRow(ctx,
        `SELECT id, name, description, created_at
         FROM categories WHERE id = $1`, id,
    ).Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt)

    if errors.Is(err, pgx.ErrNoRows) {
        return nil, ErrCategoryNotFound
    }
    if err != nil {
        return nil, err
    }
    return &c, nil
}

func (r *CategoryRepository) Create(ctx context.Context, c *model.Category) error {
    return r.db.QueryRow(ctx,
        `INSERT INTO categories (name, description)
         VALUES ($1, $2)
         RETURNING id, created_at`,
        c.Name, c.Description,
    ).Scan(&c.ID, &c.CreatedAt)
}

func (r *CategoryRepository) Update(ctx context.Context, c *model.Category) error {
    tag, err := r.db.Exec(ctx,
        `UPDATE categories
         SET name = $1, description = $2
         WHERE id = $3`,
        c.Name, c.Description, c.ID,
    )
    if err != nil {
        return err
    }
    if tag.RowsAffected() == 0 {
        return ErrCategoryNotFound
    }
    return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
    tag, err := r.db.Exec(ctx,
        `DELETE FROM categories WHERE id = $1`, id,
    )
    if err != nil {
        return err
    }
    if tag.RowsAffected() == 0 {
        return ErrCategoryNotFound
    }
    return nil
}
