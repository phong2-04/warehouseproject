
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"warehouseproject/internal/service"
)

type AuthHandler struct {
	UserService *service.UserService
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewAuthHandler(userService *service.UserService) *AuthHandler {
	return &AuthHandler{
		UserService: userService,
	}
}

// Login tiếp nhận yêu cầu đăng nhập và trả về JWT.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "Method khong duoc ho tro",
		})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Du lieu JSON khong hop le",
		})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Username va password khong duoc de trong",
		})
		return
	}

	result, err := h.UserService.Login(
		r.Context(),
		req.Username,
		req.Password,
	)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Username hoac password khong dung",
			})
			return
		}

		if errors.Is(err, service.ErrJWTSecretRequired) {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Server chua cau hinh JWT",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Loi xu ly dang nhap",
		})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
