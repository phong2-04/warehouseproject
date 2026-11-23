package model

import "time"

type Product struct {
	ID         int64     `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	CategoryID *int64    `json:"category_id"`
	Price      float64   `json:"price"`
	Quantity   int       `json:"quantity"`
	MinStock   int       `json:"min_stock"`
	Version    int64     `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
