package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

var Conn *pgx.Conn

func ConnectDB() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),     // Your DB username
		os.Getenv("DB_PASSWORD"), // Your DB password
		os.Getenv("DB_HOST"),     // Typically 'localhost'
		os.Getenv("DB_PORT"),     // Typically '5432'
		os.Getenv("DB_NAME"),     // Your database name
	)

	var err error
	Conn, err = pgx.Connect(context.Background(), dsn)
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
		os.Exit(1)
	}

	fmt.Println("Connected to PostgreSQL!")
}
