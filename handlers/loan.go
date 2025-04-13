package handlers

import (
	"Bank-Management-System/config"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// LoanPage renders the loan application form with customer profile
func LoanPage(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		ErrorPage(w, r, http.StatusUnauthorized, "You must be logged in to apply for a loan")
		return
	}

	userID, err := getUserIDFromSession(r)
	if err != nil {
		ErrorPage(w, r, http.StatusUnauthorized, "Invalid session")
		return
	}

	var profile CustomerProfile
	err = config.DB.QueryRow("SELECT name, user_name, account_number, photo_path FROM users WHERE user_id = ?", userID).
		Scan(&profile.Name, &profile.Username, &profile.AccountNumber, &profile.PhotoPath)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to fetch profile")
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/loan.html"))
	err = tmpl.Execute(w, profile)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Template rendering error")
		return
	}
}

// ApplyLoan allows users to request a loan with document uploads
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

	err = r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		ErrorPage(w, r, http.StatusBadRequest, "Failed to parse form")
		return
	}

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil || amount <= 1000 {
		ErrorPage(w, r, http.StatusBadRequest, "Minimum loan amount is Ksh.1000")
		return
	}

	interestRate, err := strconv.ParseFloat(r.FormValue("interest_rate"), 64)
	if err != nil || interestRate < 0 {
		ErrorPage(w, r, http.StatusBadRequest, "Invalid interest rate")
		return
	}

	repaymentPeriod, err := strconv.Atoi(r.FormValue("repayment_period"))
	if err != nil || repaymentPeriod <= 0 {
		ErrorPage(w, r, http.StatusBadRequest, "Invalid repayment period")
		return
	}

	loanType := r.FormValue("loan_type")
	if loanType != "personal" && loanType != "mortgage" && loanType != "commercial" {
		ErrorPage(w, r, http.StatusBadRequest, "Invalid loan type")
		return
	}

	// Handle ID upload
	idFile, idHeader, err := r.FormFile("id")
	var idPath string
	if err == nil {
		defer idFile.Close()
		idPath = fmt.Sprintf("uploads/loan_docs/%s-%s", uuid.New().String(), idHeader.Filename)
		err = SaveFile(idFile, idPath)
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to save ID33: %v", err))
			return
		}
	} else {
		ErrorPage(w, r, http.StatusBadRequest, "ID document is required")
		return
	}

	// Handle loan form upload
	formFile, formHeader, err := r.FormFile("loan_form")
	var formPath string
	if err == nil {
		defer formFile.Close()
		formPath = fmt.Sprintf("uploads/loan_docs/%s-%s", uuid.New().String(), formHeader.Filename)
		err = SaveFile(formFile, formPath)
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to save loan form: %v", err))
			return
		}
	} else {
		ErrorPage(w, r, http.StatusBadRequest, "Loan form is required")
		return
	}

	loanID := uuid.New().String()
	stmt, err := config.DB.Prepare(`
		INSERT INTO loans (user_id, loan_id, loan_type, amount, interest_rate, repayment_period, status, id_path, loan_form_path) 
		VALUES (?, ?, ?, ?, ?, ?, 'pending', ?, ?)`)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, loanID, loanType, amount, interestRate, repaymentPeriod, idPath, formPath)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to apply for loan")
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Helper function to save uploaded files
func SaveFile(file multipart.File, path string) error {
	// Ensure the directory exists
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Create the file
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", path, err)
	}
	defer out.Close()

	// Copy the uploaded file content to the new file
	_, err = io.Copy(out, file)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", path, err)
	}

	return nil
}

// ViewLoans (unchanged for now, but can be updated later to show repayment schedules)
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

	var profile CustomerProfile
	err = config.DB.QueryRow("SELECT name, user_name, account_number, photo_path FROM users WHERE user_id = ?", userID).
		Scan(&profile.Name, &profile.Username, &profile.AccountNumber, &profile.PhotoPath)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to fetch profile")
		return
	}

	rows, err := config.DB.Query("SELECT loan_id, loan_type, amount, interest_rate, repayment_period, status, created_at FROM loans WHERE user_id=?", userID)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var loans []map[string]interface{}
	for rows.Next() {
		var loanID, loanType, status, createdAt string
		var amount int
		var interestRate float64
		var repaymentPeriod int

		err := rows.Scan(&loanID, &loanType, &amount, &interestRate, &repaymentPeriod, &status, &createdAt)
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, "Database error")
			return
		}

		loans = append(loans, map[string]interface{}{
			"LoanID":          loanID,
			"LoanType":        loanType,
			"Amount":          amount,
			"InterestRate":    interestRate,
			"RepaymentPeriod": repaymentPeriod,
			"Status":          status,
			"CreatedAt":       createdAt,
		})
	}

	data := struct {
		Profile CustomerProfile
		Loans   []map[string]interface{}
	}{
		Profile: profile,
		Loans:   loans,
	}

	tmpl := template.Must(template.ParseFiles("templates/view_loans.html"))
	err = tmpl.Execute(w, data)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Template rendering error")
		return
	}
}

// CheckDelinquency checks for overdue loan repayments
func CheckDelinquency() error {
	rows, err := config.DB.Query("SELECT loan_id, due_date FROM loan_repayments WHERE status = 'pending' AND due_date < CURRENT_TIMESTAMP")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var loanID string
		var dueDate time.Time
		err := rows.Scan(&loanID, &dueDate)
		if err != nil {
			return err
		}
		_, err = config.DB.Exec("UPDATE loan_repayments SET status = 'overdue' WHERE loan_id = ? AND due_date = ?", loanID, dueDate)
		if err != nil {
			return err
		}
	}
	return nil
}