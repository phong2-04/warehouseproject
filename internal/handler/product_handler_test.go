package handler

import (
	"math"
	"strings"
	"testing"

	"warehouseproject/internal/model"
)

func TestValidProduct(t *testing.T) {
	newProduct := func() *model.Product {
		return &model.Product{
			Code:     "SP001",
			Name:     "Ban phim co",
			Price:    350000,
			Quantity: 10,
			MinStock: 5,
		}
	}

	int64Ptr := func(v int64) *int64 {
		return &v
	}

	tests := []struct {
		name    string
		product *model.Product
		want    bool
	}{
		{"San pham hop le", newProduct(), true},
		{"San pham nil", nil, false},
		{"Ma san pham rong", func() *model.Product {
			p := newProduct()
			p.Code = ""
			return p
		}(), false},
		{"Ma san pham qua dai", func() *model.Product {
			p := newProduct()
			p.Code = strings.Repeat("A", 51)
			return p
		}(), false},
		{"Ten san pham rong", func() *model.Product {
			p := newProduct()
			p.Name = ""
			return p
		}(), false},
		{"Ten san pham qua dai", func() *model.Product {
			p := newProduct()
			p.Name = strings.Repeat("A", 151)
			return p
		}(), false},
		{"Gia san pham am", func() *model.Product {
			p := newProduct()
			p.Price = -100
			return p
		}(), false},
		{"Gia vuot gioi han", func() *model.Product {
			p := newProduct()
			p.Price = 10000000000
			return p
		}(), false},
		{"Gia NaN", func() *model.Product {
			p := newProduct()
			p.Price = math.NaN()
			return p
		}(), false},
		{"Gia Infinity", func() *model.Product {
			p := newProduct()
			p.Price = math.Inf(1)
			return p
		}(), false},
		{"So luong am", func() *model.Product {
			p := newProduct()
			p.Quantity = -1
			return p
		}(), false},
		{"So luong vuot gioi han", func() *model.Product {
			p := newProduct()
			p.Quantity = 2147483648
			return p
		}(), false},
		{"MinStock am", func() *model.Product {
			p := newProduct()
			p.MinStock = -1
			return p
		}(), false},
		{"MinStock vuot gioi han", func() *model.Product {
			p := newProduct()
			p.MinStock = 2147483648
			return p
		}(), false},
		{"Category ID bang 0", func() *model.Product {
			p := newProduct()
			p.CategoryID = int64Ptr(0)
			return p
		}(), false},
		{"Category ID am", func() *model.Product {
			p := newProduct()
			p.CategoryID = int64Ptr(-1)
			return p
		}(), false},
		{"Category ID nil", newProduct(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validProduct(tt.product)
			if got != tt.want {
				t.Errorf("validProduct() = %v, want %v", got, tt.want)
			}
		})
	}
}
