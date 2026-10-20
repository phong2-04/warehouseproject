
package model

import "time"

// Product lưu thông tin sản phẩm trong kho.
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	SKU         string    `json:"sku"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
