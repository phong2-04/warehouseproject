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
	// Káº¿t ná»‘i PostgreSQL.
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal("Khong the ket noi database: ", err)
	}
	defer db.Close()

	// Khá»Ÿi táº¡o dá»‹ch vá»¥ xÃ¡c thá»±c.
	userRepo := repository.NewUserRepository(db)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("Hay cau hinh JWT_SECRET")
	}

	userService := service.NewUserService(userRepo, jwtSecret)
	authHandler := handler.NewAuthHandler(userService)

	mux := http.NewServeMux()

	// API Ä‘Äƒng nháº­p vÃ  lÃ m má»›i token.
	mux.HandleFunc("/api/login", authHandler.Login)
	mux.HandleFunc("/api/refresh-token", authHandler.RefreshToken)

	// API xem thÃ´ng tin tÃ i khoáº£n hiá»‡n táº¡i.
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

	// API chá»‰ dÃ nh cho admin.
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

	// Khá»Ÿi táº¡o Repository sáº£n pháº©m Ä‘á»ƒ sá»­ dá»¥ng á»Ÿ bÆ°á»›c CRUD.
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

	log.Println("API server dang chay tai http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
