package config

import (
	"fmt"
	"log"
)

func CreateTables() {
	usersTable := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		user_name TEXT NOT NULL UNIQUE,
		user_pin TEXT NOT NULL,
		confirm_pin TEXT NOT NULL,
		account_number TEXT NOT NULL UNIQUE,
		branch TEXT NOT NULL,
		photo_path TEXT,
		id_path TEXT,
		verification_status TEXT DEFAULT 'pending',
		auto_verification_status TEXT DEFAULT 'pending',
		role TEXT DEFAULT 'user',
		currency TEXT NOT NULL DEFAULT 'KES', -- New column for currency
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	transactionsTable := `CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		type TEXT NOT NULL,
		amount INTEGER NOT NULL,
		currency TEXT NOT NULL DEFAULT 'KES', -- New column for transaction currency
		exchange_rate FLOAT, -- New column for exchange rate (relative to KES)
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- Added for transaction timing
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
		loan_type TEXT NOT NULL,
		amount INTEGER NOT NULL,
		interest_rate FLOAT NOT NULL,
		repayment_period INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		id_path TEXT,
		loan_form_path TEXT,
		approved_by TEXT,
		approved_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(user_id)
	);`

	loanRepaymentsTable := `CREATE TABLE IF NOT EXISTS loan_repayments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		loan_id TEXT NOT NULL,
		payment_amount INTEGER NOT NULL,
		due_date DATETIME NOT NULL,
		paid_amount INTEGER DEFAULT 0,
		paid_at DATETIME,
		status TEXT NOT NULL DEFAULT 'pending',
		FOREIGN KEY(loan_id) REFERENCES loans(loan_id)
	);`

	savingsTable := `CREATE TABLE IF NOT EXISTS savings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		savings_id TEXT NOT NULL UNIQUE,
		amount INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(user_id)
	);`

	tables := []string{usersTable, transactionsTable, sessionsTable, loansTable, loanRepaymentsTable, savingsTable}
	for _, table := range tables {
		_, err := DB.Exec(table)
		if err != nil {
			log.Fatalf("Error creating table: %v", err)
		}
	}

	fmt.Println("Tables created successfully.")
}
