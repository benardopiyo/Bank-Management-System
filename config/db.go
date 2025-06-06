package config

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
    var err error
    dbPath := os.Getenv("SQLITE_DB_PATH")
    if dbPath == "" {
        dbPath = "/app/data/bank.db" // Default for Render
    }
    DB, err = sql.Open("sqlite3", dbPath)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    fmt.Println("Database initialized successfully.")
    CreateTables()
    createAdminUser()
}

// Hash password
func hashPassword(pin string) string {
	hash := sha256.Sum256([]byte(pin))
	return hex.EncodeToString(hash[:])
}

func createAdminUser() {
	// Check if admin already exists
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM users WHERE user_name = 'admin'").Scan(&count)
	if err != nil {
		log.Fatal("Error checking admin existence:", err)
	}
	if count > 0 {
		fmt.Println("Admin user already exists.")
		return
	}

	// Create admin user
	adminID := uuid.New().String()

	adminPin := os.Getenv("ADMIN_PIN")
	if adminPin == "" {
		adminPin = "admin123" // Fallback for local dev, but enforce setting in production
	}
	adminPinHashed := hashPassword(adminPin)
	accountNumber := "00101000000" // Static admin account number (customize as needed)

	_, err = DB.Exec(`
		INSERT INTO users (user_id, name, user_name, user_pin, confirm_pin, account_number, branch, role, verification_status)
		VALUES (?, 'Admin User', 'admin', ?, ?, ?, 'Nairobi', 'admin', 'verified')`,
		adminID, adminPinHashed, adminPinHashed, accountNumber)
	if err != nil {
		log.Fatal("Error creating admin user:", err)
	}
	fmt.Println("Admin user created successfully: username=admin, pin=admin123")
}
