package handlers

import (
	"Bank-Management-System/config"
	"net/http"
	"strconv"
)

// RepayLoan allows a user to repay a loan using available deposit balance
func RepayLoan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	repaymentAmount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil || repaymentAmount <= 0 {
		http.Error(w, "Invalid repayment amount", http.StatusBadRequest)
		return
	}

	// Get user's deposit balance
	var depositBalance int
	err = config.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE user_id=? AND type='deposit'", userID).Scan(&depositBalance)
	if err != nil {
		http.Error(w, "Failed to fetch balance", http.StatusInternalServerError)
		return
	}

	// Get user's outstanding loan balance
	var loanBalance int
	err = config.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM loans WHERE user_id=? AND status='pending'", userID).Scan(&loanBalance)
	if err != nil {
		http.Error(w, "Failed to fetch loan balance", http.StatusInternalServerError)
		return
	}

	if loanBalance <= 0 {
		http.Error(w, "No outstanding loan balance", http.StatusBadRequest)
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		http.Error(w, "Transaction error", http.StatusInternalServerError)
		return
	}

	// Deduct from deposit if available
	if depositBalance >= repaymentAmount {
		_, err = tx.Exec("INSERT INTO transactions (user_id, type, amount) VALUES (?, 'repayment', ?)", userID, -repaymentAmount)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to process repayment", http.StatusInternalServerError)
			return
		}

		// Update deposit balance after repayment
		_, err = tx.Exec("UPDATE transactions SET amount = amount - ? WHERE user_id=? AND type='deposit'", repaymentAmount, userID)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to update deposit balance", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Insufficient deposit balance", http.StatusBadRequest)
		tx.Rollback()
		return
	}

	// Reduce the loan balance
	_, err = tx.Exec("UPDATE loans SET amount = amount - ? WHERE user_id=? AND status='pending'", repaymentAmount, userID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update loan balance", http.StatusInternalServerError)
		return
	}

	tx.Commit()
	http.Redirect(w, r, "/view_loans", http.StatusSeeOther)
}

// AutoDeductLoan automatically deducts from deposits when a user deposits money
func AutoDeductLoan(userID string) error {
	var depositBalance, outstandingDebt int

	// Check user's deposit balance
	err := config.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE user_id=? AND type='deposit'", userID).Scan(&depositBalance)
	if err != nil {
		return err
	}

	// Check if the user has an outstanding negative balance
	err = config.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE user_id=? AND type='debt'", userID).Scan(&outstandingDebt)
	if err != nil {
		return err
	}

	if outstandingDebt >= 0 {
		return nil // No debt to auto-deduct
	}

	tx, err := config.DB.Begin()
	if err != nil {
		return err
	}

	// Deduct outstanding debt from deposits if possible
	if depositBalance >= -outstandingDebt {
		_, err = tx.Exec("INSERT INTO transactions (user_id, type, amount) VALUES (?, 'debt_payment', ?)", userID, outstandingDebt)
		if err != nil {
			tx.Rollback()
			return err
		}

		_, err = tx.Exec("DELETE FROM transactions WHERE user_id=? AND type='debt'", userID)
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// Deduct whatever is available and update the remaining debt
		_, err = tx.Exec("INSERT INTO transactions (user_id, type, amount) VALUES (?, 'debt_payment', ?)", userID, -depositBalance)
		if err != nil {
			tx.Rollback()
			return err
		}

		_, err = tx.Exec("UPDATE transactions SET amount = amount + ? WHERE user_id=? AND type='debt'", depositBalance, userID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}

// Hook into deposit processing to trigger auto-loan deduction
func ProcessDeposit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	depositAmount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil || depositAmount <= 0 {
		http.Error(w, "Invalid deposit amount", http.StatusBadRequest)
		return
	}

	_, err = config.DB.Exec("INSERT INTO transactions (user_id, type, amount) VALUES (?, 'deposit', ?)", userID, depositAmount)
	if err != nil {
		http.Error(w, "Failed to process deposit", http.StatusInternalServerError)
		return
	}

	// Check if auto-deduction is needed
	err = AutoDeductLoan(userID)
	if err != nil {
		http.Error(w, "Auto deduction failed", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
