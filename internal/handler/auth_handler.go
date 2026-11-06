package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"warehouseproject/internal/service"
)

type AuthHandler struct {
	UserService  *service.UserService
	loginLimiter *loginAttemptLimiter
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func NewAuthHandler(userService *service.UserService) *AuthHandler {
	return &AuthHandler{
		UserService:  userService,
		loginLimiter: newLoginAttemptLimiter(),
	}
}

// Login xu ly dang nhap va gioi han so lan dang nhap sai.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	clientIP := getClientIP(r)

	if retryAfter := h.loginLimiter.retryAfter(clientIP, time.Now()); retryAfter > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		writeJSON(w, http.StatusTooManyRequests, map[string]string{
			"error": "Ban da dang nhap sai qua nhieu lan. Vui long thu lai sau.",
		})
		return
	}

	result, err := h.UserService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			h.loginLimiter.recordFailure(clientIP, time.Now())
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": err.Error(),
			})
		case errors.Is(err, service.ErrJWTSecretRequired):
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Loi dang nhap",
			})
		}
		return
	}

	h.loginLimiter.reset(clientIP)
	writeJSON(w, http.StatusOK, result)
}

// RefreshToken thu hoi token cu va cap cap token moi.
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "Method khong duoc ho tro",
		})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Du lieu JSON khong hop le",
		})
		return
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Refresh token khong duoc de trong",
		})
		return
	}

	result, err := h.UserService.RefreshAccessToken(r.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidRefreshToken):
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": err.Error(),
			})
		case errors.Is(err, service.ErrJWTSecretRequired):
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Loi lam moi token",
			})
		}
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Logout thu hoi Refresh Token hien tai.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "Method khong duoc ho tro",
		})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Du lieu JSON khong hop le",
		})
		return
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Refresh token khong duoc de trong",
		})
		return
	}

	err := h.UserService.Logout(r.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidRefreshToken):
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": err.Error(),
			})
		case errors.Is(err, service.ErrJWTSecretRequired):
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Loi dang xuat",
			})
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Dang xuat thanh cong",
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
