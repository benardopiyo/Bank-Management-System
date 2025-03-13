package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "bank.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	fmt.Println("Database initialized successfully.")
	CreateTables()
}
