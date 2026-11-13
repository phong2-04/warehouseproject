package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"warehouseproject/internal/middleware"
	"warehouseproject/internal/model"
	"warehouseproject/internal/repository"
	"warehouseproject/internal/service"

	"github.com/jackc/pgx/v5/pgconn"
)

type fakeUserManagementService struct {
	users     []model.User
	user      *model.User
	getErr    error
	createErr error
	updateErr error
	deleteErr error

	getCalled    bool
	createCalled bool
	updateCalled bool
	deleteCalled bool
}

func (f *fakeUserManagementService) GetAll(ctx context.Context) ([]model.User, error) {
	f.getCalled = true
	return f.users, f.getErr
}

func (f *fakeUserManagementService) Create(ctx context.Context, username, password, fullName, role string) (*model.User, error) {
	f.createCalled = true
	if f.createErr != nil {
		return nil, f.createErr
	}
	return f.user, nil
}

func (f *fakeUserManagementService) Update(ctx context.Context, id int64, username, fullName, role string) (*model.User, error) {
	f.updateCalled = true
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return f.user, nil
}

func (f *fakeUserManagementService) Delete(ctx context.Context, id int64) error {
	f.deleteCalled = true
	return f.deleteErr
}

func newUserHandlerTestRequest(method, path, body string) (*httptest.ResponseRecorder, *http.Request) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	return httptest.NewRecorder(), req
}

func TestUserHandlerUsersGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fake := &fakeUserManagementService{
			users: []model.User{{ID: 1, Username: "admin", FullName: "Admin", Role: "admin"}},
		}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodGet, "/api/users", "")

		h.Users(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !fake.getCalled {
			t.Fatal("expected GetAll to be called")
		}
	})

	t.Run("service error", func(t *testing.T) {
		fake := &fakeUserManagementService{getErr: errors.New("database error")}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodGet, "/api/users", "")

		h.Users(w, r)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func TestUserHandlerUsersPost(t *testing.T) {
	validBody := `{"username":"staff1","password":"password123","full_name":"Staff One","role":"staff"}`

	t.Run("success", func(t *testing.T) {
		fake := &fakeUserManagementService{
			user: &model.User{ID: 2, Username: "staff1", FullName: "Staff One", Role: "staff"},
		}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodPost, "/api/users", validBody)

		h.Users(w, r)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		if !fake.createCalled {
			t.Fatal("expected Create to be called")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		fake := &fakeUserManagementService{}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodPost, "/api/users", `{"username":`)

		h.Users(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
		if fake.createCalled {
			t.Fatal("Create should not be called")
		}
	})

	t.Run("invalid user data", func(t *testing.T) {
		fake := &fakeUserManagementService{createErr: service.ErrInvalidUserData}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodPost, "/api/users", validBody)

		h.Users(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("duplicate username", func(t *testing.T) {
		fake := &fakeUserManagementService{
			createErr: &pgconn.PgError{Code: "23505"},
		}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodPost, "/api/users", validBody)

		h.Users(w, r)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		fake := &fakeUserManagementService{createErr: errors.New("database error")}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodPost, "/api/users", validBody)

		h.Users(w, r)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		fake := &fakeUserManagementService{}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodDelete, "/api/users", "")

		h.Users(w, r)

		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405, got %d", w.Code)
		}
		if w.Header().Get("Allow") != "GET, POST" {
			t.Fatalf("unexpected Allow header: %s", w.Header().Get("Allow"))
		}
	})
}

func TestUserHandlerUserByID(t *testing.T) {
	validBody := `{"username":"staff1","full_name":"Staff One","role":"staff"}`

	t.Run("invalid ID", func(t *testing.T) {
		fake := &fakeUserManagementService{}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodPut, "/api/users/abc", validBody)

		h.UserByID(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid nested ID", func(t *testing.T) {
		fake := &fakeUserManagementService{}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodPut, "/api/users/2/extra", validBody)

		h.UserByID(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("update success", func(t *testing.T) {
		fake := &fakeUserManagementService{
			user: &model.User{ID: 2, Username: "staff1", FullName: "Staff One", Role: "staff"},
		}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodPut, "/api/users/2", validBody)

		h.UserByID(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !fake.updateCalled {
			t.Fatal("expected Update to be called")
		}
	})

	t.Run("update invalid JSON", func(t *testing.T) {
		fake := &fakeUserManagementService{}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodPut, "/api/users/2", `{"username":`)

		h.UserByID(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("update invalid data", func(t *testing.T) {
		fake := &fakeUserManagementService{updateErr: service.ErrInvalidUserData}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodPut, "/api/users/2", validBody)

		h.UserByID(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("update user not found", func(t *testing.T) {
		fake := &fakeUserManagementService{updateErr: repository.ErrUserNotFound}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodPut, "/api/users/2", validBody)

		h.UserByID(w, r)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("update last admin", func(t *testing.T) {
		fake := &fakeUserManagementService{updateErr: repository.ErrLastAdmin}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodPut, "/api/users/2", validBody)

		h.UserByID(w, r)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", w.Code)
		}
	})

	t.Run("update duplicate username", func(t *testing.T) {
		fake := &fakeUserManagementService{
			updateErr: &pgconn.PgError{Code: "23505"},
		}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodPut, "/api/users/2", validBody)

		h.UserByID(w, r)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", w.Code)
		}
	})

	t.Run("update internal error", func(t *testing.T) {
		fake := &fakeUserManagementService{updateErr: errors.New("database error")}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodPut, "/api/users/2", validBody)

		h.UserByID(w, r)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("delete without user context", func(t *testing.T) {
		fake := &fakeUserManagementService{}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodDelete, "/api/users/2", "")

		h.UserByID(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
		if fake.deleteCalled {
			t.Fatal("Delete should not be called")
		}
	})

	t.Run("delete self", func(t *testing.T) {
		fake := &fakeUserManagementService{}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodDelete, "/api/users/2", "")
		r = r.WithContext(context.WithValue(r.Context(), middleware.UserIDKey, int64(2)))

		h.UserByID(w, r)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
		if fake.deleteCalled {
			t.Fatal("Delete should not be called")
		}
	})

	t.Run("delete success", func(t *testing.T) {
		fake := &fakeUserManagementService{}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodDelete, "/api/users/2", "")
		r = r.WithContext(context.WithValue(r.Context(), middleware.UserIDKey, int64(1)))

		h.UserByID(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !fake.deleteCalled {
			t.Fatal("expected Delete to be called")
		}
	})

	t.Run("delete invalid data", func(t *testing.T) {
		fake := &fakeUserManagementService{deleteErr: service.ErrInvalidUserData}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodDelete, "/api/users/2", "")
		r = r.WithContext(context.WithValue(r.Context(), middleware.UserIDKey, int64(1)))

		h.UserByID(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("delete user not found", func(t *testing.T) {
		fake := &fakeUserManagementService{deleteErr: repository.ErrUserNotFound}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodDelete, "/api/users/2", "")
		r = r.WithContext(context.WithValue(r.Context(), middleware.UserIDKey, int64(1)))

		h.UserByID(w, r)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("delete last admin", func(t *testing.T) {
		fake := &fakeUserManagementService{deleteErr: repository.ErrLastAdmin}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodDelete, "/api/users/2", "")
		r = r.WithContext(context.WithValue(r.Context(), middleware.UserIDKey, int64(1)))

		h.UserByID(w, r)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", w.Code)
		}
	})

	t.Run("delete foreign key conflict", func(t *testing.T) {
		fake := &fakeUserManagementService{
			deleteErr: &pgconn.PgError{Code: "23503"},
		}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodDelete, "/api/users/2", "")
		r = r.WithContext(context.WithValue(r.Context(), middleware.UserIDKey, int64(1)))

		h.UserByID(w, r)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", w.Code)
		}
	})

	t.Run("delete internal error", func(t *testing.T) {
		fake := &fakeUserManagementService{deleteErr: errors.New("database error")}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodDelete, "/api/users/2", "")
		r = r.WithContext(context.WithValue(r.Context(), middleware.UserIDKey, int64(1)))

		h.UserByID(w, r)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		fake := &fakeUserManagementService{}
		h := NewUserHandler(fake)
		w, r := newUserHandlerTestRequest(http.MethodGet, "/api/users/2", "")

		h.UserByID(w, r)

		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405, got %d", w.Code)
		}
		if w.Header().Get("Allow") != "PUT, DELETE" {
			t.Fatalf("unexpected Allow header: %s", w.Header().Get("Allow"))
		}
	})
}
