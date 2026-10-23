package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"warehouseproject/internal/repository"
	"warehouseproject/internal/service"
)

type UserHandler struct {
	UserService *service.UserManagementService
}

func NewUserHandler(
	userService *service.UserManagementService,
) *UserHandler {
	return &UserHandler{UserService: userService}
}

func writeUserJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func decodeUserJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

// GET /api/users: danh sach; POST /api/users: tao tai khoan.
func (h *UserHandler) Users(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		users, err := h.UserService.GetAll(r.Context())
		if err != nil {
			writeUserJSON(w, http.StatusInternalServerError,
				map[string]string{"error": "Khong the lay danh sach"})
			return
		}
		writeUserJSON(w, http.StatusOK, users)

	case http.MethodPost:
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
			FullName string `json:"full_name"`
			Role     string `json:"role"`
		}

		if err := decodeUserJSON(r, &req); err != nil {
			writeUserJSON(w, http.StatusBadRequest,
				map[string]string{"error": "JSON khong hop le"})
			return
		}

		user, err := h.UserService.Create(
			r.Context(), req.Username, req.Password,
			req.FullName, req.Role,
		)
		if err != nil {
			if errors.Is(err, service.ErrInvalidUserData) {
				writeUserJSON(w, http.StatusBadRequest,
					map[string]string{"error": err.Error()})
				return
			}
			writeUserJSON(w, http.StatusConflict,
				map[string]string{"error": "Khong the tao tai khoan; kiem tra username co bi trung khong"})
			return
		}

		writeUserJSON(w, http.StatusCreated, user)

	default:
		w.Header().Set("Allow", "GET, POST")
		writeUserJSON(w, http.StatusMethodNotAllowed,
			map[string]string{"error": "Method khong duoc ho tro"})
	}
}

// PUT /api/users/{id}: sua; DELETE /api/users/{id}: xoa.
func (h *UserHandler) UserByID(w http.ResponseWriter, r *http.Request) {
	idText := strings.TrimPrefix(r.URL.Path, "/api/users/")
	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil || id <= 0 || strings.Contains(idText, "/") {
		writeUserJSON(w, http.StatusBadRequest,
			map[string]string{"error": "ID nguoi dung khong hop le"})
		return
	}

	switch r.Method {
	case http.MethodPut:
		var req struct {
			Username string `json:"username"`
			FullName string `json:"full_name"`
			Role     string `json:"role"`
		}

		if err := decodeUserJSON(r, &req); err != nil {
			writeUserJSON(w, http.StatusBadRequest,
				map[string]string{"error": "JSON khong hop le"})
			return
		}

		user, err := h.UserService.Update(
			r.Context(), id, req.Username, req.FullName, req.Role,
		)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidUserData):
				writeUserJSON(w, http.StatusBadRequest,
					map[string]string{"error": err.Error()})
			case errors.Is(err, repository.ErrUserNotFound):
				writeUserJSON(w, http.StatusNotFound,
					map[string]string{"error": "Khong tim thay nguoi dung"})
			default:
				writeUserJSON(w, http.StatusConflict,
					map[string]string{"error": "Khong the cap nhat; kiem tra username co bi trung khong"})
			}
			return
		}

		writeUserJSON(w, http.StatusOK, user)

	case http.MethodDelete:
		err := h.UserService.Delete(r.Context(), id)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidUserData):
				writeUserJSON(w, http.StatusBadRequest,
					map[string]string{"error": err.Error()})
			case errors.Is(err, repository.ErrUserNotFound):
				writeUserJSON(w, http.StatusNotFound,
					map[string]string{"error": "Khong tim thay nguoi dung"})
			default:
				writeUserJSON(w, http.StatusConflict,
					map[string]string{"error": "Khong the xoa tai khoan"})
			}
			return
		}

		writeUserJSON(w, http.StatusOK,
			map[string]string{"message": "Xoa tai khoan thanh cong"})

	default:
		w.Header().Set("Allow", "PUT, DELETE")
		writeUserJSON(w, http.StatusMethodNotAllowed,
			map[string]string{"error": "Method khong duoc ho tro"})
	}
}
