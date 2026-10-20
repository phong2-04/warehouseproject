
package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

// RequireRole chi cho phep cac role duoc chi dinh truy cap API.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(RoleKey).(string)
			if !ok || role == "" {
				writeForbidden(w, "Khong tim thay thong tin quyen")
				return
			}

			for _, allowedRole := range allowedRoles {
				if strings.EqualFold(role, allowedRole) {
					next.ServeHTTP(w, r)
					return
				}
			}

			writeForbidden(w, "Ban khong co quyen truy cap")
		})
	}
}

func writeForbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
