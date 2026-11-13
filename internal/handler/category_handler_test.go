package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"warehouseproject/internal/model"
)

func TestValidCategory(t *testing.T) {
	tests := []struct {
		name     string
		category *model.Category
		want     bool
	}{
		{
			name:     "Danh muc hop le",
			category: &model.Category{Name: "Thiet bi van phong"},
			want:     true,
		},
		{
			name:     "Danh muc nil",
			category: nil,
			want:     false,
		},
		{
			name:     "Ten rong",
			category: &model.Category{Name: ""},
			want:     false,
		},
		{
			name:     "Ten chi co khoang trang",
			category: &model.Category{Name: "   "},
			want:     false,
		},
		{
			name:     "Ten co 100 ky tu",
			category: &model.Category{Name: strings.Repeat("A", 100)},
			want:     true,
		},
		{
			name:     "Ten co 101 ky tu",
			category: &model.Category{Name: strings.Repeat("A", 101)},
			want:     false,
		},
		{
			name:     "Ten tieng Viet",
			category: &model.Category{Name: "Thiết bị văn phòng"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validCategory(tt.category)

			if got != tt.want {
				t.Errorf("validCategory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeCategory(t *testing.T) {
	tests := []struct {
		name     string
		category *model.Category
		wantName string
		wantDesc *string
	}{
		{
			name: "Xoa khoang trang",
			category: &model.Category{
				Name:        "  Thiet bi van phong  ",
				Description: stringPtr("  Mo ta danh muc  "),
			},
			wantName: "Thiet bi van phong",
			wantDesc: stringPtr("Mo ta danh muc"),
		},
		{
			name: "Mo ta rong thanh nil",
			category: &model.Category{
				Name:        "Danh muc test",
				Description: stringPtr("   "),
			},
			wantName: "Danh muc test",
			wantDesc: nil,
		},
		{
			name: "Mo ta nil",
			category: &model.Category{
				Name: "Danh muc test",
			},
			wantName: "Danh muc test",
			wantDesc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizeCategory(tt.category)

			if tt.category.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", tt.category.Name, tt.wantName)
			}

			if tt.wantDesc == nil {
				if tt.category.Description != nil {
					t.Errorf("Description = %v, want nil", *tt.category.Description)
				}
			} else if tt.category.Description == nil ||
				*tt.category.Description != *tt.wantDesc {
				t.Errorf("Description khong dung")
			}
		})
	}
}

func TestCategoryID(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantID     int64
		wantOK     bool
		wantStatus int
	}{
		{
			name:       "ID hop le",
			path:       "/api/categories/6",
			wantID:     6,
			wantOK:     true,
			wantStatus: 200,
		},
		{
			name:       "ID bang 0",
			path:       "/api/categories/0",
			wantOK:     false,
			wantStatus: 400,
		},
		{
			name:       "ID am",
			path:       "/api/categories/-1",
			wantOK:     false,
			wantStatus: 400,
		},
		{
			name:       "ID khong phai so",
			path:       "/api/categories/abc",
			wantOK:     false,
			wantStatus: 400,
		},
		{
			name:       "Sai duong dan",
			path:       "/api/products/6",
			wantOK:     false,
			wantStatus: 400,
		},
		{
			name:       "Thieu ID",
			path:       "/api/categories",
			wantOK:     false,
			wantStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()

			gotID, gotOK := categoryID(rec, req)

			if gotID != tt.wantID {
				t.Errorf("ID = %d, want %d", gotID, tt.wantID)
			}

			if gotOK != tt.wantOK {
				t.Errorf("OK = %v, want %v", gotOK, tt.wantOK)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("HTTP Status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
