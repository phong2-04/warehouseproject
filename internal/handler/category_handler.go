package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"warehouseproject/internal/model"
	"warehouseproject/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryHandler struct {
	repo *repository.CategoryRepository
}

func NewCategoryHandler(db *pgxpool.Pool) *CategoryHandler {
	return &CategoryHandler{
		repo: repository.NewCategoryRepository(db),
	}
}

func writeCategoryError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func validCategory(c *model.Category) bool {
	if c == nil {
		return false
	}

	name := strings.TrimSpace(c.Name)
	length := utf8.RuneCountInString(name)

	return length >= 1 && length <= 100
}

func normalizeCategory(c *model.Category) {
	c.Name = strings.TrimSpace(c.Name)

	if c.Description != nil {
		description := strings.TrimSpace(*c.Description)
		if description == "" {
			c.Description = nil
		} else {
			c.Description = &description
		}
	}
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.GetAll(r.Context())
	if err != nil {
		writeCategoryError(w, http.StatusInternalServerError, "Loi lay danh sach danh muc")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(categories)
}

func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := categoryID(w, r)
	if !ok {
		return
	}

	category, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			writeCategoryError(w, http.StatusNotFound, "Khong tim thay danh muc")
			return
		}

		writeCategoryError(w, http.StatusInternalServerError, "Loi lay danh muc")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(category)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var c model.Category

	if err := decodeJSONBody(w, r, &c, 256*1024, true); err != nil {
		if isRequestBodyTooLarge(err) {
			writeCategoryError(w, http.StatusRequestEntityTooLarge, "Request body qua lon")
			return
		}

		writeCategoryError(w, http.StatusBadRequest, "JSON khong hop le")
		return
	}

	normalizeCategory(&c)

	if !validCategory(&c) {
		writeCategoryError(w, http.StatusBadRequest, "Ten danh muc phai co tu 1 den 100 ky tu")
		return
	}

	if err := h.repo.Create(r.Context(), &c); err != nil {
		if repository.IsUniqueViolation(err) {
			writeCategoryError(w, http.StatusConflict, "Ten danh muc da ton tai")
			return
		}

		writeCategoryError(w, http.StatusInternalServerError, "Loi tao danh muc")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(c)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := categoryID(w, r)
	if !ok {
		return
	}

	var c model.Category

	if err := decodeJSONBody(w, r, &c, 256*1024, true); err != nil {
		if isRequestBodyTooLarge(err) {
			writeCategoryError(w, http.StatusRequestEntityTooLarge, "Request body qua lon")
			return
		}

		writeCategoryError(w, http.StatusBadRequest, "JSON khong hop le")
		return
	}

	normalizeCategory(&c)

	if !validCategory(&c) {
		writeCategoryError(w, http.StatusBadRequest, "Ten danh muc phai co tu 1 den 100 ky tu")
		return
	}

	c.ID = id

	if err := h.repo.Update(r.Context(), &c); err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			writeCategoryError(w, http.StatusNotFound, "Khong tim thay danh muc")
			return
		}

		if repository.IsUniqueViolation(err) {
			writeCategoryError(w, http.StatusConflict, "Ten danh muc da ton tai")
			return
		}

		writeCategoryError(w, http.StatusInternalServerError, "Loi cap nhat danh muc")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Cap nhat danh muc thanh cong",
	})
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := categoryID(w, r)
	if !ok {
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			writeCategoryError(w, http.StatusNotFound, "Khong tim thay danh muc")
			return
		}

		if repository.IsForeignKeyViolation(err) {
			writeCategoryError(w, http.StatusConflict, "Khong the xoa danh muc dang duoc san pham su dung")
			return
		}

		writeCategoryError(w, http.StatusInternalServerError, "Loi xoa danh muc")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Xoa danh muc thanh cong",
	})
}

func categoryID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(parts) != 3 || parts[0] != "api" || parts[1] != "categories" {
		writeCategoryError(w, http.StatusBadRequest, "Duong dan khong hop le")
		return 0, false
	}

	id, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || id <= 0 {
		writeCategoryError(w, http.StatusBadRequest, "ID khong hop le")
		return 0, false
	}

	return id, true
}
