package handlers

import (
	"Bank-Management-System/config"
	"html/template"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

// LoanPage renders the loan application form
func LoanPage(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/loan.html"))
	tmpl.Execute(w, nil)
}

// ApplyLoan allows users to request a loan
func ApplyLoan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/loan", http.StatusSeeOther)
		return
	}

	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil || amount <= 0 {
		http.Error(w, "Invalid loan amount", http.StatusBadRequest)
		return
	}

	interestRate, err := strconv.ParseFloat(r.FormValue("interest_rate"), 64)
	if err != nil || interestRate < 0 {
		http.Error(w, "Invalid interest rate", http.StatusBadRequest)
		return
	}

	repaymentPeriod, err := strconv.Atoi(r.FormValue("repayment_period"))
	if err != nil || repaymentPeriod <= 0 {
		http.Error(w, "Invalid repayment period", http.StatusBadRequest)
		return
	}

	loanID := uuid.New().String()

	stmt, err := config.DB.Prepare("INSERT INTO loans (user_id, loan_id, amount, interest_rate, repayment_period, status) VALUES (?, ?, ?, ?, ?, 'pending')")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, loanID, amount, interestRate, repaymentPeriod)
	if err != nil {
		http.Error(w, "Failed to apply for loan", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// ViewLoans fetches and displays the user's loans
func ViewLoans(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	rows, err := config.DB.Query("SELECT loan_id, amount, interest_rate, repayment_period, status, created_at FROM loans WHERE user_id=?", userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var loans []map[string]interface{}
	for rows.Next() {
		var loanID, status, createdAt string
		var amount int
		var interestRate float64
		var repaymentPeriod int

		err := rows.Scan(&loanID, &amount, &interestRate, &repaymentPeriod, &status, &createdAt)
		if err != nil {
			http.Error(w, "Error retrieving loans", http.StatusInternalServerError)
			return
		}

		loans = append(loans, map[string]interface{}{
			"LoanID":          loanID,
			"Amount":          amount,
			"InterestRate":    interestRate,
			"RepaymentPeriod": repaymentPeriod,
			"Status":          status,
			"CreatedAt":       createdAt,
		})
	}

	tmpl := template.Must(template.ParseFiles("templates/view_loans.html"))
	tmpl.Execute(w, loans)
}
