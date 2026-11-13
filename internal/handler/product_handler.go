package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"warehouseproject/internal/middleware"
	"warehouseproject/internal/model"
	"warehouseproject/internal/repository"
)

// ProductHandler xÃƒÂ¡Ã‚Â»Ã‚Â­ lÃƒÆ’Ã‚Â½ cÃƒÆ’Ã‚Â¡c API sÃƒÂ¡Ã‚ÂºÃ‚Â£n phÃƒÂ¡Ã‚ÂºÃ‚Â©m.
type ProductHandler struct {
	ProductRepo *repository.ProductRepository
}

func NewProductHandler(repo *repository.ProductRepository) *ProductHandler {
	return &ProductHandler{ProductRepo: repo}
}

func (h *ProductHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/products"), "/")

	// ChÃƒÂ¡Ã‚Â»Ã¢â‚¬Â° admin Ãƒâ€žÃ¢â‚¬ËœÃƒâ€ Ã‚Â°ÃƒÂ¡Ã‚Â»Ã‚Â£c thÃƒÆ’Ã‚Âªm, sÃƒÂ¡Ã‚Â»Ã‚Â­a hoÃƒÂ¡Ã‚ÂºÃ‚Â·c xÃƒÆ’Ã‚Â³a sÃƒÂ¡Ã‚ÂºÃ‚Â£n phÃƒÂ¡Ã‚ÂºÃ‚Â©m.
	role, _ := r.Context().Value(middleware.RoleKey).(string)

	switch r.Method {
	case http.MethodGet:
		if path == "" {
			h.GetAll(w, r)
		} else if path == "low-stock" {
			h.GetLowStock(w, r)
		} else {
			h.GetByID(w, r, path)
		}

	case http.MethodPost:
		if path != "" {
			http.NotFound(w, r)
			return
		}
		if !strings.EqualFold(role, "admin") {
			writeProductError(w, http.StatusForbidden, "Ban khong co quyen truy cap")
			return
		}
		h.Create(w, r)

	case http.MethodPut:
		if !strings.EqualFold(role, "admin") {
			writeProductError(w, http.StatusForbidden, "Ban khong co quyen truy cap")
			return
		}
		h.Update(w, r, path)

	case http.MethodDelete:
		if !strings.EqualFold(role, "admin") {
			writeProductError(w, http.StatusForbidden, "Ban khong co quyen truy cap")
			return
		}
		h.Delete(w, r, path)

	default:
		w.Header().Set("Allow", "GET, POST, PUT, DELETE")
		writeProductError(w, http.StatusMethodNotAllowed, "Method khong duoc ho tro")
	}
}

func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	products, err := h.ProductRepo.GetAll(r.Context())
	if err != nil {
		writeProductError(w, http.StatusInternalServerError, "Khong the lay danh sach san pham")
		return
	}

	writeProductJSON(w, http.StatusOK, products)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		writeProductError(w, http.StatusBadRequest, "ID san pham khong hop le")
		return
	}

	product, err := h.ProductRepo.GetByID(r.Context(), id)
	if err != nil {
		if repository.IsNotFound(err) {
			writeProductError(w, http.StatusNotFound, "Khong tim thay san pham")
			return
		}
		writeProductError(w, http.StatusInternalServerError, "Khong the lay san pham")
		return
	}

	writeProductJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var product model.Product
	if err := decodeJSONBody(w, r, &product, 1<<20, false); err != nil {
		if isRequestBodyTooLarge(err) {
			writeProductError(w, http.StatusRequestEntityTooLarge, "Request body qua lon")
			return
		}
		writeProductError(w, http.StatusBadRequest, "Du lieu JSON khong hop le")
		return
	}

	product.Code = strings.TrimSpace(product.Code)
	product.Name = strings.TrimSpace(product.Name)
	product.Quantity = 0

	if !validProduct(&product) {
		writeProductError(w, http.StatusBadRequest, "Thong tin san pham khong hop le")
		return
	}

	if err := h.ProductRepo.Create(r.Context(), &product); err != nil {
		if repository.IsUniqueViolation(err) {
			writeProductError(w, http.StatusConflict, "Ma san pham da ton tai")
			return
		}
		if repository.IsForeignKeyViolation(err) {
			writeProductError(w, http.StatusBadRequest, "Danh muc san pham khong ton tai")
			return
		}
		writeProductError(w, http.StatusInternalServerError, "Khong the them san pham")
		return
	}

	writeProductJSON(w, http.StatusCreated, product)
}

type productUpdateRequest struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	CategoryID *int64  `json:"category_id"`
	Price      float64 `json:"price"`
	MinStock   int     `json:"min_stock"`
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		writeProductError(w, http.StatusBadRequest, "ID san pham khong hop le")
		return
	}

	var request productUpdateRequest
	if err := decodeJSONBody(w, r, &request, 1<<20, true); err != nil {
		if isRequestBodyTooLarge(err) {
			writeProductError(w, http.StatusRequestEntityTooLarge, "Request body qua lon")
			return
		}
		writeProductError(w, http.StatusBadRequest, "Du lieu JSON khong hop le")
		return
	}

	product, err := h.ProductRepo.GetByID(r.Context(), id)
	if err != nil {
		if repository.IsNotFound(err) {
			writeProductError(w, http.StatusNotFound, "Khong tim thay san pham")
			return
		}
		writeProductError(w, http.StatusInternalServerError, "Khong the lay san pham")
		return
	}

	product.Code = strings.TrimSpace(request.Code)
	product.Name = strings.TrimSpace(request.Name)
	product.CategoryID = request.CategoryID
	product.Price = request.Price
	product.MinStock = request.MinStock

	if !validProduct(product) {
		writeProductError(w, http.StatusBadRequest, "Thong tin san pham khong hop le")
		return
	}

	if err := h.ProductRepo.Update(r.Context(), product); err != nil {
		if repository.IsNotFound(err) {
			writeProductError(w, http.StatusNotFound, "Khong tim thay san pham")
			return
		}
		if repository.IsUniqueViolation(err) {
			writeProductError(w, http.StatusConflict, "Ma san pham da ton tai")
			return
		}
		if repository.IsForeignKeyViolation(err) {
			writeProductError(w, http.StatusBadRequest, "Danh muc san pham khong ton tai")
			return
		}
		writeProductError(w, http.StatusInternalServerError, "Khong the cap nhat san pham")
		return
	}

	writeProductJSON(w, http.StatusOK, product)
}
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		writeProductError(w, http.StatusBadRequest, "ID san pham khong hop le")
		return
	}

	if err := h.ProductRepo.Delete(r.Context(), id); err != nil {
		if err.Error() == "khong tim thay san pham" {
			writeProductError(w, http.StatusNotFound, "Khong tim thay san pham")
			return
		}
		if repository.IsForeignKeyViolation(err) {
			writeProductError(w, http.StatusConflict, "Khong the xoa san pham dang duoc su dung")
			return
		}
		writeProductError(w, http.StatusInternalServerError, "Khong the xoa san pham")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func validProduct(p *model.Product) bool {
	if p == nil {
		return false
	}

	code := utf8.RuneCountInString(p.Code)
	name := utf8.RuneCountInString(p.Name)

	if code == 0 || code > 50 {
		return false
	}

	if name == 0 || name > 150 {
		return false
	}

	if math.IsNaN(p.Price) || math.IsInf(p.Price, 0) ||
		p.Price < 0 || p.Price > 9999999999.99 {
		return false
	}

	if p.Quantity < 0 || p.Quantity > 2147483647 {
		return false
	}

	if p.MinStock < 0 || p.MinStock > 2147483647 {
		return false
	}

	if p.CategoryID != nil && *p.CategoryID <= 0 {
		return false
	}

	return true
}

func writeProductJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeProductError(w http.ResponseWriter, status int, message string) {
	writeProductJSON(w, status, map[string]string{"error": message})
}

// GetLowStock tra ve danh sach san pham ton kho thap.
func (h *ProductHandler) GetLowStock(w http.ResponseWriter, r *http.Request) {
	products, err := h.ProductRepo.GetLowStock(r.Context())
	if err != nil {
		writeProductError(w, http.StatusInternalServerError, "Khong the lay danh sach san pham ton kho thap")
		return
	}

	writeProductJSON(w, http.StatusOK, products)
}
