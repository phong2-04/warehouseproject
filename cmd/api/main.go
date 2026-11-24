package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	if len([]byte(jwtSecret)) < 32 {
		log.Fatal("JWT_SECRET phai co it nhat 32 byte")
	}

	// Khoi tao repository va service xac thuc.
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, jwtSecret)
	trustedProxies, err := handler.ParseTrustedProxies(os.Getenv("TRUSTED_PROXIES"))
	if err != nil {
		log.Fatal("Cau hinh TRUSTED_PROXIES khong hop le: ", err)
	}
	authHandler := handler.NewAuthHandlerWithTrustedProxies(userService, trustedProxies)

	mux := http.NewServeMux()

	// Health Check: kiem tra API va ket noi PostgreSQL.
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Allow", http.MethodGet)
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Method khong duoc ho tro",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")

		if err := db.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status":   "unhealthy",
				"database": "disconnected",
			})
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":   "healthy",
			"database": "connected",
		})
	})
	// Dang nhap va lam moi token.
	mux.HandleFunc("/api/login", authHandler.Login)
	mux.HandleFunc("/api/refresh-token", authHandler.RefreshToken)
	mux.HandleFunc("/api/logout", authHandler.Logout)
	mux.Handle(
		"/api/change-password",
		middleware.AuthMiddleware(jwtSecret, userRepo,
			http.HandlerFunc(authHandler.ChangePassword),
		),
	)

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
		middleware.AuthMiddleware(jwtSecret, userRepo, profileHandler),
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
		middleware.AuthMiddleware(jwtSecret, userRepo,
			middleware.RequireRole("admin")(adminHandler),
		),
	)

	// Khoi tao Repository va Handler san pham.
	productRepo := repository.NewProductRepository(db)
	productHandler := handler.NewProductHandler(productRepo)

	// API san pham yeu cau dang nhap.
	mux.Handle(
		"/api/products",
		middleware.AuthMiddleware(jwtSecret, userRepo, productHandler),
	)
	mux.Handle(
		"/api/products/",
		middleware.AuthMiddleware(jwtSecret, userRepo, productHandler),
	)

	// Khoi tao Service va Handler quan ly nguoi dung.
	userManagementService := service.NewUserManagementService(userRepo)
	userHandler := handler.NewUserHandler(userManagementService)

	// API quan ly nguoi dung chi danh cho Admin.
	mux.Handle(
		"/api/users",
		middleware.AuthMiddleware(jwtSecret, userRepo,
			middleware.RequireRole("admin")(
				http.HandlerFunc(userHandler.Users),
			),
		),
	)

	mux.Handle(
		"/api/users/",
		middleware.AuthMiddleware(jwtSecret, userRepo,
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
		middleware.AuthMiddleware(jwtSecret, userRepo,
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
		middleware.AuthMiddleware(jwtSecret, userRepo,
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
		middleware.AuthMiddleware(jwtSecret, userRepo,
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
		middleware.AuthMiddleware(jwtSecret, userRepo, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	server := &http.Server{
		Addr:              ":8080",
		Handler:           corsHandler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	shutdownContext, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	cleanupDone := make(chan struct{})

	go func() {
		defer close(cleanupDone)

		cleanup := func() {
			ctx, cancel := context.WithTimeout(
				shutdownContext,
				10*time.Second,
			)
			defer cancel()

			deleted, err := userRepo.DeleteExpiredRefreshTokens(ctx)
			if err != nil {
				log.Println("Cleanup Refresh Token that bai:", err)
				return
			}

			if deleted > 0 {
				log.Printf("Da xoa %d Refresh Token het han", deleted)
			}
		}

		// Chay mot lan khi khoi dong.
		cleanup()

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-shutdownContext.Done():
				return
			case <-ticker.C:
				cleanup()
			}
		}
	}()

	serverErrors := make(chan error, 1)

	go func() {
		log.Println("API server dang chay tai http://localhost:8080")
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Loi HTTP server: ", err)
		}

	case <-shutdownContext.Done():
		log.Println("Nhan tin hieu dung server, dang cho request hoan tat...")

		ctx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Println("Graceful shutdown that bai:", err)

			if closeErr := server.Close(); closeErr != nil {
				log.Println("Loi dong server:", closeErr)
			}
		}

		if err := <-serverErrors; err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Println("Loi khi server dung:", err)
		}

		log.Println("API server da dung")
	}

	stop()
	<-cleanupDone
}
