package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name           string
		role           interface{}
		allowedRoles   []string
		expectedStatus int
		nextCalled     bool
	}{
		{
			name:           "Admin duoc truy cap",
			role:           "admin",
			allowedRoles:   []string{"admin"},
			expectedStatus: http.StatusOK,
			nextCalled:     true,
		},
		{
			name:           "Staff bi chan",
			role:           "staff",
			allowedRoles:   []string{"admin"},
			expectedStatus: http.StatusForbidden,
			nextCalled:     false,
		},
		{
			name:           "Role khong phan biet chu hoa",
			role:           "ADMIN",
			allowedRoles:   []string{"admin"},
			expectedStatus: http.StatusOK,
			nextCalled:     true,
		},
		{
			name:           "Khong co thong tin role",
			role:           nil,
			allowedRoles:   []string{"admin"},
			expectedStatus: http.StatusForbidden,
			nextCalled:     false,
		},
		{
			name:           "Role rong",
			role:           "",
			allowedRoles:   []string{"admin"},
			expectedStatus: http.StatusForbidden,
			nextCalled:     false,
		},
		{
			name:           "Khong co role duoc phep",
			role:           "admin",
			allowedRoles:   []string{},
			expectedStatus: http.StatusForbidden,
			nextCalled:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)

			if tt.role != nil {
				ctx := context.WithValue(req.Context(), RoleKey, tt.role)
				req = req.WithContext(ctx)
			}

			nextCalled := false

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			handler := RequireRole(tt.allowedRoles...)(next)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			if recorder.Code != tt.expectedStatus {
				t.Errorf("Status = %d, muon %d",
					recorder.Code, tt.expectedStatus)
			}

			if nextCalled != tt.nextCalled {
				t.Errorf("nextCalled = %v, muon %v",
					nextCalled, tt.nextCalled)
			}
		})
	}
}
