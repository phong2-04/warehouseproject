package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectDB tạo kết nối đến PostgreSQL.
func ConnectDB() (*pgxpool.Pool, error) {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "warehouse_db"
	}

	connString := "postgres://" + dbUser + ":" + dbPassword +
		"@" + dbHost + ":" + dbPort + "/" + dbName +
		"?sslmode=disable"

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 10
	config.MinConns = 1
	config.MaxConnLifetime = time.Hour

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(context.Background()); err != nil {
		db.Close()
		return nil, err
	}

	log.Println("Ket noi PostgreSQL thanh cong!")
	return db, nil
}
