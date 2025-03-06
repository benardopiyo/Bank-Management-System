package handlers

import (
	"Bank-Management-System/config"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
)

// Fetch user's UUID from the database using their username/email
func getUserID(username string) (string, error) {
	var userID string
	err := config.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", nil // No user found
	}
	return userID, err
}

// Get user's balance using their UUID
func getBalance(userID string) (int, error) {
	var balance int
	err := config.DB.QueryRow(`
		SELECT COALESCE(SUM(CASE WHEN type='deposit' THEN amount ELSE -amount END), 0)
		FROM transactions WHERE user_id=?`, userID).Scan(&balance)
	return balance, err
}

// Deposit function
func Deposit(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil || amount <= 0 {
		http.Error(w, "Invalid deposit amount", http.StatusBadRequest)
		return
	}

	stmt, err := config.DB.Prepare("INSERT INTO transactions (user_id, type, amount) VALUES (?, 'deposit', ?)")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, amount)
	if err != nil {
		http.Error(w, "Failed to deposit", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Withdraw function
func Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil || amount <= 0 {
		http.Error(w, "Invalid withdrawal amount", http.StatusBadRequest)
		return
	}

	balance, err := getBalance(userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if balance < amount {
		http.Error(w, "Insufficient funds", http.StatusBadRequest)
		return
	}

	stmt, err := config.DB.Prepare("INSERT INTO transactions (user_id, type, amount) VALUES (?, 'withdraw', ?)")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, amount)
	if err != nil {
		http.Error(w, "Failed to withdraw", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Balance function
func Balance(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	balance, err := getBalance(userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"balance": balance})
}