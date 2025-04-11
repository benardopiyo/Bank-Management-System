package handlers

import (
	"Bank-Management-System/config"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

// AdminLoanDashboard displays pending loan applications
func AdminLoanDashboard(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) || !isAdmin(r) {
		ErrorPage(w, r, http.StatusForbidden, "Access denied. Admins only.")
		return
	}

	rows, err := config.DB.Query(`
		SELECT l.loan_id, l.user_id, l.loan_type, l.amount, l.interest_rate, l.repayment_period, l.id_path, l.loan_form_path, u.user_name
		FROM loans l
		JOIN users u ON l.user_id = u.user_id
		WHERE l.status = 'pending'`)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, fmt.Sprintf("Database query error: %v", err))
		return
	}
	defer rows.Close()

	type PendingLoan struct {
		LoanID          string
		UserID          string
		LoanType        string
		Amount          int
		InterestRate    float64
		RepaymentPeriod int
		IDPath          string
		LoanFormPath    string
		Username        string // Moved to match query order
	}

	var pendingLoans []PendingLoan
	for rows.Next() {
		var loan PendingLoan
		err := rows.Scan(&loan.LoanID, &loan.UserID, &loan.LoanType, &loan.Amount, &loan.InterestRate, &loan.RepaymentPeriod, &loan.IDPath, &loan.LoanFormPath, &loan.Username)
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, fmt.Sprintf("Error scanning row: %v", err))
			return
		}
		pendingLoans = append(pendingLoans, loan)
	}

	if err = rows.Err(); err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, fmt.Sprintf("Row iteration error: %v", err))
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/admin_loan_dashboard.html"))
	err = tmpl.Execute(w, pendingLoans)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, fmt.Sprintf("Template rendering error: %v", err))
		return
	}
}

// ApproveLoan handles loan approval or rejection
func ApproveLoan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin-loans", http.StatusSeeOther)
		return
	}

	if !isAuthenticated(r) || !isAdmin(r) {
		ErrorPage(w, r, http.StatusForbidden, "Access denied. Admins only.")
		return
	}

	loanID := r.FormValue("loan_id")
	action := r.FormValue("action")
	adminID, err := getUserIDFromSession(r)
	if err != nil {
		ErrorPage(w, r, http.StatusUnauthorized, "Invalid session")
		return
	}

	var status string
	if action == "approve" {
		status = "approved"
	} else if action == "reject" {
		status = "rejected"
	} else {
		ErrorPage(w, r, http.StatusBadRequest, "Invalid action")
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to start transaction: %v", err))
		return
	}

	_, err = tx.Exec("UPDATE loans SET status = ?, approved_by = ?, approved_at = ? WHERE loan_id = ?", status, adminID, time.Now(), loanID)
	if err != nil {
		tx.Rollback()
		ErrorPage(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to update loan status: %v", err))
		return
	}

	if status == "approved" {
		// Fetch loan details
		var userID string
		var amount int
		var repaymentPeriod int
		var interestRate float64
		err = tx.QueryRow("SELECT user_id, amount, repayment_period, interest_rate FROM loans WHERE loan_id = ?", loanID).
			Scan(&userID, &amount, &repaymentPeriod, &interestRate)
		if err != nil {
			tx.Rollback()
			ErrorPage(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to fetch loan details: %v", err))
			return
		}

		// Disburse cash to user's account
		_, err = tx.Exec("INSERT INTO transactions (user_id, type, amount) VALUES (?, 'loan', ?)", userID, amount)
		if err != nil {
			tx.Rollback()
			ErrorPage(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to disburse loan: %v", err))
			return
		}

		// Update status to disbursed
		_, err = tx.Exec("UPDATE loans SET status = 'disbursed' WHERE loan_id = ?", loanID)
		if err != nil {
			tx.Rollback()
			ErrorPage(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to update loan status: %v", err))
			return
		}

		// Simple amortization: equal monthly payments
		totalAmount := float64(amount) * (1 + interestRate/100)
		monthlyPayment := int(totalAmount / float64(repaymentPeriod))
		for i := 1; i <= repaymentPeriod; i++ {
			dueDate := time.Now().AddDate(0, i, 0)
			_, err = tx.Exec("INSERT INTO loan_repayments (loan_id, payment_amount, due_date) VALUES (?, ?, ?)", loanID, monthlyPayment, dueDate)
			if err != nil {
				tx.Rollback()
				ErrorPage(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to create repayment schedule: %v", err))
				return
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, fmt.Sprintf("Transaction commit failed: %v", err))
		return
	}

	http.Redirect(w, r, "/admin-loans", http.StatusSeeOther)
}