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
		ErrorPage(w, r, http.StatusUnauthorized, "You must be logged in to apply for a loan")
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
		ErrorPage(w, r, http.StatusUnauthorized, "You must be logged in to apply for a loan")
		return
	}

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil || amount <= 1000 {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Minimum loan amount is Ksh.1000")
		return
	}

	interestRate, err := strconv.ParseFloat(r.FormValue("interest_rate"), 64)
	if err != nil || interestRate < 0 {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Invalid interest rate")
		return
	}

	repaymentPeriod, err := strconv.Atoi(r.FormValue("repayment_period"))
	if err != nil || repaymentPeriod <= 0 {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Invalid repayment period")
		return
	}

	loanID := uuid.New().String()

	stmt, err := config.DB.Prepare("INSERT INTO loans (user_id, loan_id, amount, interest_rate, repayment_period, status) VALUES (?, ?, ?, ?, ?, 'disbursed')")
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer stmt.Close()

	transLoanStmt, err := config.DB.Prepare("INSERT INTO transactions (user_id, type, amount) VALUES (?, 'loan', ?)")
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer transLoanStmt.Close()

	_, err = stmt.Exec(userID, loanID, amount, interestRate, repaymentPeriod)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to apply for loan")
		return
	}

	_, err = transLoanStmt.Exec(userID, amount)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to process deposit")
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
		ErrorPage(w, r, http.StatusUnauthorized, "You must be logged in to view your loans")
		return
	}

	rows, err := config.DB.Query("SELECT loan_id, amount, interest_rate, repayment_period, status, created_at FROM loans WHERE user_id=?", userID)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
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
			ErrorPage(w, r, http.StatusInternalServerError, "Database error")
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
