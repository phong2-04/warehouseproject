package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"warehouseproject/internal/service"
)

type authUserService interface {
	Login(ctx context.Context, username, password string) (*service.LoginResponse, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*service.RefreshResponse, error)
	Logout(ctx context.Context, refreshToken string) error
}

type AuthHandler struct {
	UserService     authUserService
	loginLimiter    *loginAttemptLimiter
	usernameLimiter *loginAttemptLimiter
	trustedProxies  []netip.Prefix
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func NewAuthHandler(userService authUserService) *AuthHandler {
	return &AuthHandler{
		UserService:     userService,
		loginLimiter:    newLoginAttemptLimiter(),
		usernameLimiter: newUsernameAttemptLimiter(),
	}
}

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

	clientIP := getClientIPWithTrustedProxies(r, h.trustedProxies)

	if retryAfter := h.loginLimiter.retryAfter(clientIP, time.Now()); retryAfter > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		writeJSON(w, http.StatusTooManyRequests, map[string]string{
			"error": "Ban da dang nhap sai qua nhieu lan. Vui long thu lai sau.",
		})
		return
	}

	if retryAfter := h.usernameLimiter.retryAfter(req.Username, time.Now()); retryAfter > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		writeJSON(w, http.StatusTooManyRequests, map[string]string{
			"error": "Tai khoan tam thoi bi gioi han dang nhap. Vui long thu lai sau.",
		})
		return
	}

	result, err := h.UserService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			h.loginLimiter.recordFailure(clientIP, time.Now())
			h.usernameLimiter.recordFailure(req.Username, time.Now())
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
	h.usernameLimiter.reset(req.Username)
	writeJSON(w, http.StatusOK, result)
}

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
