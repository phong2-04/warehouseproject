package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"warehouseproject/internal/middleware"
	"warehouseproject/internal/repository"
)

type fakeStockRepository struct {
	record       *repository.StockTransaction
	records      []repository.StockTransaction
	processErr   error
	listErr      error
	processCalls int
	listCalls    int
	productID    int64
	userID       int64
	transType    string
	quantity     int
	note         string
}

func (f *fakeStockRepository) Process(
	ctx context.Context,
	productID int64,
	userID int64,
	transactionType string,
	quantity int,
	note string,
) (*repository.StockTransaction, error) {
	f.processCalls++
	f.productID = productID
	f.userID = userID
	f.transType = transactionType
	f.quantity = quantity
	f.note = note
	return f.record, f.processErr
}

func (f *fakeStockRepository) List(ctx context.Context) ([]repository.StockTransaction, error) {
	f.listCalls++
	return f.records, f.listErr
}

func newStockTestRequest(method, body string, userID interface{}) (*httptest.ResponseRecorder, *http.Request) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(method, "/api/stock", strings.NewReader(body))

	if userID != nil {
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
	}

	return recorder, req
}

func TestStockHandlerImport(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		userID     interface{}
		repoErr    error
		wantStatus int
		wantCalls  int
	}{
		{"method khong hop le", http.MethodGet, `{}`, int64(1), nil, http.StatusMethodNotAllowed, 0},
		{"JSON sai", http.MethodPost, `{`, int64(1), nil, http.StatusBadRequest, 0},
		{"co truong khong hop le", http.MethodPost, `{"product_id":1,"quantity":2,"extra":true}`, int64(1), nil, http.StatusBadRequest, 0},
		{"product ID bang 0", http.MethodPost, `{"product_id":0,"quantity":2}`, int64(1), nil, http.StatusBadRequest, 0},
		{"quantity bang 0", http.MethodPost, `{"product_id":1,"quantity":0}`, int64(1), nil, http.StatusBadRequest, 0},
		{"khong co user ID", http.MethodPost, `{"product_id":1,"quantity":2}`, nil, nil, http.StatusUnauthorized, 0},
		{"user ID khong hop le", http.MethodPost, `{"product_id":1,"quantity":2}`, "abc", nil, http.StatusUnauthorized, 0},
		{"user ID khong duong", http.MethodPost, `{"product_id":1,"quantity":2}`, int64(0), nil, http.StatusUnauthorized, 0},
		{"so luong khong hop le", http.MethodPost, `{"product_id":1,"quantity":2}`, int64(1), repository.ErrInvalidStockQuantity, http.StatusBadRequest, 1},
		{"san pham khong ton tai", http.MethodPost, `{"product_id":1,"quantity":2}`, int64(1), repository.ErrProductNotFound, http.StatusNotFound, 1},
		{"loi he thong", http.MethodPost, `{"product_id":1,"quantity":2}`, int64(1), errors.New("database error"), http.StatusInternalServerError, 1},
		{"nhap kho thanh cong", http.MethodPost, `{"product_id":1,"quantity":5,"note":"  Nhap hang  "}`, int64(1), nil, http.StatusCreated, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeStockRepository{
				record: &repository.StockTransaction{
					ID:              1,
					ProductID:       1,
					UserID:          1,
					TransactionType: "import",
					Quantity:        5,
					Note:            "Nhap hang",
					CreatedAt:       time.Now(),
				},
				processErr: tt.repoErr,
			}

			handler := NewStockHandler(fake)
			recorder, req := newStockTestRequest(tt.method, tt.body, tt.userID)
			handler.Import(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, muon %d; body: %s", recorder.Code, tt.wantStatus, recorder.Body.String())
			}

			if fake.processCalls != tt.wantCalls {
				t.Errorf("so lan goi Process = %d, muon %d", fake.processCalls, tt.wantCalls)
			}

			if tt.name == "nhap kho thanh cong" {
				if fake.transType != "import" {
					t.Errorf("transaction type = %q, muon import", fake.transType)
				}
				if fake.note != "Nhap hang" {
					t.Errorf("note chua duoc trim: %q", fake.note)
				}
				if fake.userID != 1 || fake.productID != 1 || fake.quantity != 5 {
					t.Errorf("tham so Process khong dung")
				}
			}
		})
	}
}

func TestStockHandlerExport(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		repoErr    error
		wantStatus int
	}{
		{"xuat kho thanh cong", `{"product_id":2,"quantity":3}`, nil, http.StatusCreated},
		{"khong du ton kho", `{"product_id":2,"quantity":100}`, repository.ErrInsufficientStock, http.StatusConflict},
		{"san pham khong ton tai", `{"product_id":2,"quantity":3}`, repository.ErrProductNotFound, http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeStockRepository{
				record: &repository.StockTransaction{
					ID:              2,
					ProductID:       2,
					UserID:          1,
					TransactionType: "export",
					Quantity:        3,
				},
				processErr: tt.repoErr,
			}

			handler := NewStockHandler(fake)
			recorder, req := newStockTestRequest(http.MethodPost, tt.body, int64(1))
			handler.Export(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, muon %d; body: %s", recorder.Code, tt.wantStatus, recorder.Body.String())
			}

			if fake.processCalls != 1 {
				t.Errorf("Process duoc goi %d lan, muon 1", fake.processCalls)
			}

			if fake.transType != "export" {
				t.Errorf("transaction type = %q, muon export", fake.transType)
			}
		})
	}
}

func TestStockHandlerList(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		listErr    error
		records    []repository.StockTransaction
		wantStatus int
		wantCalls  int
	}{
		{"method khong hop le", http.MethodPost, nil, nil, http.StatusMethodNotAllowed, 0},
		{"loi repository", http.MethodGet, errors.New("database error"), nil, http.StatusInternalServerError, 1},
		{"danh sach rong", http.MethodGet, nil, []repository.StockTransaction{}, http.StatusOK, 1},
		{"co lich su giao dich", http.MethodGet, nil, []repository.StockTransaction{
			{ID: 1, ProductID: 2, UserID: 1, TransactionType: "import", Quantity: 5},
			{ID: 2, ProductID: 2, UserID: 1, TransactionType: "export", Quantity: 2},
		}, http.StatusOK, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeStockRepository{
				records: tt.records,
				listErr: tt.listErr,
			}

			handler := NewStockHandler(fake)
			recorder, req := newStockTestRequest(tt.method, "", nil)
			handler.List(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, muon %d; body: %s", recorder.Code, tt.wantStatus, recorder.Body.String())
			}

			if fake.listCalls != tt.wantCalls {
				t.Errorf("so lan goi List = %d, muon %d", fake.listCalls, tt.wantCalls)
			}
		})
	}
}

func TestStockUserID(t *testing.T) {
	tests := []struct {
		name   string
		value  interface{}
		wantID int64
		wantOK bool
	}{
		{"int", int(5), 5, true},
		{"int64", int64(6), 6, true},
		{"float64", float64(7), 7, true},
		{"json number", json.Number("8"), 8, true},
		{"json number sai", json.Number("abc"), 0, false},
		{"kieu khong ho tro", "10", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := stockUserID(tt.value)
			if gotID != tt.wantID || gotOK != tt.wantOK {
				t.Errorf("stockUserID(%v) = (%d, %v), muon (%d, %v)", tt.value, gotID, gotOK, tt.wantID, tt.wantOK)
			}
		})
	}
}
