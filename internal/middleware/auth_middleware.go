
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UsernameKey contextKey = "username"
	RoleKey     contextKey = "role"
)

// AuthMiddleware kiem tra JWT truoc khi cho phep truy cap API.
func AuthMiddleware(jwtSecret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// Yeu cau token theo dinh dang Bearer.
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

		// Xac minh chu ky va thoi han cua JWT.
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

		userID, ok := claims["user_id"].(float64)
		username, usernameOK := claims["username"].(string)
		role, roleOK := claims["role"].(string)

		if !ok || userID <= 0 ||
			!usernameOK || username == "" ||
			!roleOK || role == "" {
			writeUnauthorized(w, "Thieu thong tin nguoi dung")
			return
		}

		// Luu thong tin user de handler tiep theo su dung.
		ctx := context.WithValue(r.Context(), UserIDKey, int64(userID))
		ctx = context.WithValue(ctx, UsernameKey, username)
		ctx = context.WithValue(ctx, RoleKey, role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
