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

	createTables()
}

func createTables() {
	usersTable := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		user_name TEXT NOT NULL UNIQUE,
		user_pin TEXT NOT NULL,
		confirm_pin TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	transactionsTable := `CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		type TEXT NOT NULL,
		amount INTEGER NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(user_id)
	);`

	sessionsTable := `CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_token TEXT NOT NULL UNIQUE,
		user_id TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(user_id)
	);`

	loansTable := `CREATE TABLE IF NOT EXISTS loans (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		loan_id TEXT NOT NULL UNIQUE,
		amount INTEGER NOT NULL,
		interest_rate FLOAT NOT NULL,
		repayment_period INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(user_id)
	);`

	_, err := DB.Exec(usersTable)
	if err != nil {
		log.Fatal("Error creating users table:", err)
	}

	_, err = DB.Exec(transactionsTable)
	if err != nil {
		log.Fatal("Error creating transactions table:", err)
	}

	_, err = DB.Exec(sessionsTable)
	if err != nil {
		log.Fatal("Error creating sessions table:", err)
	}

	_, err = DB.Exec(loansTable)
	if err != nil {
		log.Fatal("Error creating loans table:", err)
	}

	fmt.Println("Database initialized successfully.")
}
