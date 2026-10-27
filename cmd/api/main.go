package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"warehouseproject/internal/database"
	"warehouseproject/internal/handler"
	"warehouseproject/internal/middleware"
	"warehouseproject/internal/repository"
	"warehouseproject/internal/service"
)

func main() {
	// Ket noi PostgreSQL.
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal("Khong the ket noi database: ", err)
	}
	defer db.Close()

	// Cau hinh JWT.
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("Hay cau hinh JWT_SECRET")
	}

	// Khoi tao repository va service xac thuc.
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, jwtSecret)
	authHandler := handler.NewAuthHandler(userService)

	mux := http.NewServeMux()

	// Dang nhap va lam moi token.
	mux.HandleFunc("/api/login", authHandler.Login)
	mux.HandleFunc("/api/refresh-token", authHandler.RefreshToken)

	// Xem thong tin tai khoan dang dang nhap.
	profileHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":  r.Context().Value(middleware.UserIDKey),
			"username": r.Context().Value(middleware.UsernameKey),
			"role":     r.Context().Value(middleware.RoleKey),
		})
	})

	mux.Handle(
		"/api/profile",
		middleware.AuthMiddleware(jwtSecret, profileHandler),
	)

	// API kiem tra quyen Admin.
	adminHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Ban co quyen Admin",
		})
	})

	mux.Handle(
		"/api/admin",
		middleware.AuthMiddleware(
			jwtSecret,
			middleware.RequireRole("admin")(adminHandler),
		),
	)

	// Khoi tao Repository va Handler san pham.
	productRepo := repository.NewProductRepository(db)
	productHandler := handler.NewProductHandler(productRepo)

	// API san pham yeu cau dang nhap.
	mux.Handle(
		"/api/products",
		middleware.AuthMiddleware(jwtSecret, productHandler),
	)
	mux.Handle(
		"/api/products/",
		middleware.AuthMiddleware(jwtSecret, productHandler),
	)

	// Khoi tao Service va Handler quan ly nguoi dung.
	userManagementService := service.NewUserManagementService(userRepo)
	userHandler := handler.NewUserHandler(userManagementService)

	// API quan ly nguoi dung chi danh cho Admin.
	mux.Handle(
		"/api/users",
		middleware.AuthMiddleware(
			jwtSecret,
			middleware.RequireRole("admin")(
				http.HandlerFunc(userHandler.Users),
			),
		),
	)

	mux.Handle(
		"/api/users/",
		middleware.AuthMiddleware(
			jwtSecret,
			middleware.RequireRole("admin")(
				http.HandlerFunc(userHandler.UserByID),
			),
		),
	)

	// Khoi tao Handler quan ly danh muc.
	categoryHandler := handler.NewCategoryHandler(db)

	// Xem danh sach danh muc: can dang nhap.
	mux.Handle(
		"/api/categories",
		middleware.AuthMiddleware(jwtSecret,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					http.Error(w, `{"error":"Method khong duoc ho tro"}`, http.StatusMethodNotAllowed)
					return
				}
				categoryHandler.List(w, r)
			}),
		),
	)

	// Xem chi tiet, sua va xoa danh muc: chi Admin.
	mux.Handle(
		"/api/categories/",
		middleware.AuthMiddleware(jwtSecret,
			middleware.RequireRole("admin")(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.Method {
					case http.MethodGet:
						categoryHandler.GetByID(w, r)
					case http.MethodPut:
						categoryHandler.Update(w, r)
					case http.MethodDelete:
						categoryHandler.Delete(w, r)
					default:
						http.Error(w, `{"error":"Method khong duoc ho tro"}`, http.StatusMethodNotAllowed)
					}
				}),
			),
		),
	)

	// Them danh muc: chi Admin.
	mux.Handle(
		"/api/categories/create",
		middleware.AuthMiddleware(jwtSecret,
			middleware.RequireRole("admin")(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodPost {
						http.Error(w, `{"error":"Method khong duoc ho tro"}`, http.StatusMethodNotAllowed)
						return
					}
					categoryHandler.Create(w, r)
				}),
			),
		),
	)

	// Khoi tao Repository va Handler giao dich kho.
	stockRepo := repository.NewStockRepository(db)
	stockHandler := handler.NewStockHandler(stockRepo)

	// API nhap kho, xuat kho va xem lich su: yeu cau dang nhap.
	mux.Handle(
		"/api/stock/",
		middleware.AuthMiddleware(jwtSecret, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/stock/import":
				stockHandler.Import(w, r)
			case "/api/stock/export":
				stockHandler.Export(w, r)
			case "/api/stock/transactions":
				stockHandler.List(w, r)
			default:
				http.NotFound(w, r)
			}
		})),
	)
	// Cau hinh CORS cho cac cong frontend Vite.
	allowedOrigins := map[string]bool{
		"http://localhost:5173": true,
		"http://localhost:5174": true,
		"http://localhost:5175": true,
	}

	corsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		mux.ServeHTTP(w, r)
	})

	log.Println("API server dang chay tai http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}
