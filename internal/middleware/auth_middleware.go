package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"warehouseproject/internal/repository"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UsernameKey contextKey = "username"
	RoleKey     contextKey = "role"
)

// AuthMiddleware xac thuc Access Token va kiem tra tai khoan trong DB.
func AuthMiddleware(
	jwtSecret string,
	userRepo *repository.UserRepository,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeUnauthorized(w, "Thieu token xac thuc")
			return
		}

		tokenString := strings.TrimSpace(
			strings.TrimPrefix(authHeader, "Bearer "),
		)

		if tokenString == "" {
			writeUnauthorized(w, "Token khong hop le")
			return
		}

		token, err := jwt.Parse(
			tokenString,
			func(token *jwt.Token) (interface{}, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			},
			jwt.WithValidMethods([]string{"HS256"}),
		)

		if err != nil || token == nil || !token.Valid {
			writeUnauthorized(w, "Token khong hop le hoac da het han")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeUnauthorized(w, "Thong tin token khong hop le")
			return
		}

		tokenType, ok := claims["token_type"].(string)
		if !ok || tokenType != "access" {
			writeUnauthorized(w, "Can su dung Access Token")
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok || userID <= 0 || userID != float64(int64(userID)) {
			writeUnauthorized(w, "Thong tin nguoi dung khong hop le")
			return
		}

		// Kiem tra tai khoan va quyen hien tai trong database.
		user, err := userRepo.GetByID(r.Context(), int64(userID))
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				writeUnauthorized(w, "Tai khoan khong con ton tai")
				return
			}

			log.Printf("Loi kiem tra tai khoan trong AuthMiddleware: %v", err)
			writeInternalServerError(w)
			return
		}

		// Khong su dung username va role cu trong JWT.
		ctx := context.WithValue(r.Context(), UserIDKey, user.ID)
		ctx = context.WithValue(ctx, UsernameKey, user.Username)
		ctx = context.WithValue(ctx, RoleKey, user.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func writeInternalServerError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": "Loi kiem tra xac thuc",
	})
}
