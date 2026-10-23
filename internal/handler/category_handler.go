package handler

import (
    "encoding/json"
    "errors"
    "net/http"
    "strconv"
    "strings"

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

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
    categories, err := h.repo.GetAll(r.Context())
    if err != nil {
        http.Error(w, `{"error":"Loi lay danh sach danh muc"}`, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(categories)
}

func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    id, ok := categoryID(w, r)
    if !ok {
        return
    }

    category, err := h.repo.GetByID(r.Context(), id)
    if err != nil {
        if errors.Is(err, repository.ErrCategoryNotFound) {
            http.Error(w, `{"error":"Khong tim thay danh muc"}`, http.StatusNotFound)
            return
        }
        http.Error(w, `{"error":"Loi lay danh muc"}`, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(category)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
    var c model.Category
    if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
        http.Error(w, `{"error":"JSON khong hop le"}`, http.StatusBadRequest)
        return
    }

    c.Name = strings.TrimSpace(c.Name)
    if c.Name == "" {
        http.Error(w, `{"error":"Ten danh muc khong duoc de trong"}`, http.StatusBadRequest)
        return
    }

    if err := h.repo.Create(r.Context(), &c); err != nil {
        http.Error(w, `{"error":"Loi tao danh muc"}`, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(c)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
    id, ok := categoryID(w, r)
    if !ok {
        return
    }

    var c model.Category
    if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
        http.Error(w, `{"error":"JSON khong hop le"}`, http.StatusBadRequest)
        return
    }

    c.Name = strings.TrimSpace(c.Name)
    if c.Name == "" {
        http.Error(w, `{"error":"Ten danh muc khong duoc de trong"}`, http.StatusBadRequest)
        return
    }
    c.ID = id

    if err := h.repo.Update(r.Context(), &c); err != nil {
        if errors.Is(err, repository.ErrCategoryNotFound) {
            http.Error(w, `{"error":"Khong tim thay danh muc"}`, http.StatusNotFound)
            return
        }
        http.Error(w, `{"error":"Loi cap nhat danh muc"}`, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Cap nhat danh muc thanh cong"})
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
    id, ok := categoryID(w, r)
    if !ok {
        return
    }

    if err := h.repo.Delete(r.Context(), id); err != nil {
        if errors.Is(err, repository.ErrCategoryNotFound) {
            http.Error(w, `{"error":"Khong tim thay danh muc"}`, http.StatusNotFound)
            return
        }
        http.Error(w, `{"error":"Loi xoa danh muc"}`, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Xoa danh muc thanh cong"})
}

func categoryID(w http.ResponseWriter, r *http.Request) (int64, bool) {
    parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
    if len(parts) != 3 || parts[0] != "api" || parts[1] != "categories" {
        http.Error(w, `{"error":"Duong dan khong hop le"}`, http.StatusBadRequest)
        return 0, false
    }

    id, err := strconv.ParseInt(parts[2], 10, 64)
    if err != nil || id <= 0 {
        http.Error(w, `{"error":"ID khong hop le"}`, http.StatusBadRequest)
        return 0, false
    }
    return id, true
}
