
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
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal("Khong the ket noi database:", err)
	}
	defer db.Close()

	// Khoi tao cac thanh phan dang nhap.
	userRepo := repository.NewUserRepository(db)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("Hay cau hinh JWT_SECRET")
	}

	userService := service.NewUserService(userRepo, jwtSecret)
	authHandler := handler.NewAuthHandler(userService)

	// Dang ky cac API.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", authHandler.Login)

	// API thu nghiem yeu cau JWT hop le.
	profileHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey)
		username := r.Context().Value(middleware.UsernameKey)
		role := r.Context().Value(middleware.RoleKey)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":  userID,
			"username": username,
			"role":     role,
		})
	})

	mux.Handle(
		"/api/profile",
		middleware.AuthMiddleware(jwtSecret, profileHandler),
	)
	// API chi danh cho Admin.
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

	log.Println("API server dang chay tai http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
