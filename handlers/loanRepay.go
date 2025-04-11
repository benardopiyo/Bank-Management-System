package handlers

import (
	"Bank-Management-System/config"
	"net/http"
	"strconv"
)

// RepayLoan allows a user to repay a loan using available deposit balance
func RepayLoan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/repay-loan", http.StatusSeeOther)
		return
	}

	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		ErrorPage(w, r, http.StatusUnauthorized, "You must be logged in to repay a loan")
		return
	}

	repayAmt, err := strconv.Atoi(r.FormValue("repay_amount"))
	if err != nil || repayAmt <= 0 {
		ErrorPage(w, r, http.StatusBadRequest, "Invalid repayment amount")
		return
	}

	repayStmt, err := config.DB.Prepare("INSERT INTO transactions (user_id, type, amount) VALUES (?, 'repay', ?)")
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer repayStmt.Close()

	_, err = repayStmt.Exec(userID, repayAmt)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to process deposit")
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
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
		ErrorPage(w, r, http.StatusUnauthorized, "User not authenticated")
		return
	}

	depositAmount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil || depositAmount <= 0 {
		ErrorPage(w, r, http.StatusBadRequest, "Invalid deposit amount")
		return
	}

	_, err = config.DB.Exec("INSERT INTO transactions (user_id, type, amount) VALUES (?, 'deposit', ?)", userID, depositAmount)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to process deposit")
		return
	}

	// Check if auto-deduction is needed
	err = AutoDeductLoan(userID)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Auto deduction failed")
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
