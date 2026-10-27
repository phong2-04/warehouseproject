package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"warehouseproject/internal/middleware"
	"warehouseproject/internal/repository"
)

type StockHandler struct {
	repo *repository.StockRepository
}

func NewStockHandler(repo *repository.StockRepository) *StockHandler {
	return &StockHandler{repo: repo}
}

type stockRequest struct {
	ProductID int64  `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Note      string `json:"note"`
}

func (h *StockHandler) Import(w http.ResponseWriter, r *http.Request) {
	h.process(w, r, "import")
}

func (h *StockHandler) Export(w http.ResponseWriter, r *http.Request) {
	h.process(w, r, "export")
}

func (h *StockHandler) process(w http.ResponseWriter, r *http.Request, transactionType string) {
	if r.Method != http.MethodPost {
		writeStockJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "Method khong duoc ho tro",
		})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	var req stockRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeStockJSON(w, http.StatusBadRequest, map[string]string{
			"error": "JSON khong hop le",
		})
		return
	}

	if req.ProductID <= 0 || req.Quantity <= 0 {
		writeStockJSON(w, http.StatusBadRequest, map[string]string{
			"error": "product_id va quantity phai lon hon 0",
		})
		return
	}

	userIDValue := r.Context().Value(middleware.UserIDKey)
	userID, ok := stockUserID(userIDValue)
	if !ok || userID <= 0 {
		writeStockJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "Khong xac dinh duoc tai khoan dang nhap",
		})
		return
	}

	record, err := h.repo.Process(
		r.Context(),
		req.ProductID,
		userID,
		transactionType,
		req.Quantity,
		strings.TrimSpace(req.Note),
	)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInvalidStockQuantity):
			writeStockJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		case errors.Is(err, repository.ErrProductNotFound):
			writeStockJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case errors.Is(err, repository.ErrInsufficientStock):
			writeStockJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		default:
			log.Printf("Loi xu ly giao dich kho: %v", err)
			http.Error(w, `{"error":"Loi xu ly giao dich kho"}`, http.StatusInternalServerError)
		}
		return
	}

	writeStockJSON(w, http.StatusCreated, record)
}

func (h *StockHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeStockJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "Method khong duoc ho tro",
		})
		return
	}

	records, err := h.repo.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Khong the lay lich su giao dich"}`, http.StatusInternalServerError)
		return
	}

	writeStockJSON(w, http.StatusOK, records)
}

func stockUserID(value interface{}) (int64, bool) {
	switch id := value.(type) {
	case int:
		return int64(id), true
	case int64:
		return id, true
	case float64:
		return int64(id), true
	case json.Number:
		parsed, err := strconv.ParseInt(id.String(), 10, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func writeStockJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}


