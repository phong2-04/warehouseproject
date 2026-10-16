
package main

import (
	"fmt"
	"log"

	"warehouseproject/internal/database"
)

func main() {
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal("Khong the ket noi database:", err)
	}
	defer db.Close()

	fmt.Println("Warehouse Management System")
	fmt.Println("Database da san sang!")
}
