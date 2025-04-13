package handlers

import (
	"Bank-Management-System/config"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/signintech/gopdf"
)

const (
	exchangeRateAPIKey = "c924d207c754d157d47665a7" // ExchangeRate-API key
	exchangeRateURL    = "https://v6.exchangerate-api.com/v6/" + exchangeRateAPIKey + "/latest/"
)

// ExchangeRateResponse defines the structure of the ExchangeRate-API response
type ExchangeRateResponse struct {
	Result          string             `json:"result"`
	BaseCode        string             `json:"base_code"`
	ConversionRates map[string]float64 `json:"conversion_rates"`
}

// fetchUserID retrieves user_id by username
func fetchUserID(username string) (string, error) {
	var userID string
	err := config.DB.QueryRow("SELECT user_id FROM users WHERE user_name = ?", username).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", nil // No user found
	}
	return userID, err
}

// getBalance calculates the user's balance in KES
func getBalance(userID string) (int, error) {
	var balance float64
	err := config.DB.QueryRow(`
		SELECT COALESCE(SUM(CASE 
			WHEN type='deposit' THEN amount * (1.0 / COALESCE(exchange_rate, 1.0))
			WHEN type='loan' THEN amount * (1.0 / COALESCE(exchange_rate, 1.0))
			WHEN type='receive' THEN amount * (1.0 / COALESCE(exchange_rate, 1.0))
			ELSE -amount * (1.0 / COALESCE(exchange_rate, 1.0))
		END), 0)
		FROM transactions WHERE user_id=?`, userID).Scan(&balance)
	return int(balance), err
}

// Deposit function
func Deposit(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		ErrorPage(w, r, http.StatusUnauthorized, "User not authenticated")
		return
	}

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil || amount <= 0 {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Invalid deposit amount")
		return
	}

	// Get user's currency
	var currency string
	err = config.DB.QueryRow("SELECT currency FROM users WHERE user_id = ?", userID).Scan(&currency)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	exchangeRate := 1.0
	if currency != "KES" {
		exchangeRate, err = getExchangeRate(currency, "KES")
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, "Failed to fetch exchange rate")
			return
		}
	}

	stmt, err := config.DB.Prepare("INSERT INTO transactions (user_id, type, amount, currency, exchange_rate, created_at) VALUES (?, 'deposit', ?, ?, ?, ?)")
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, amount, currency, exchangeRate, time.Now())
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to deposit")
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Withdraw function
func Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		ErrorPage(w, r, http.StatusUnauthorized, "User not authenticated")
		return
	}

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil || amount <= 0 {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Invalid withdrawal amount")
		return
	}

	balance, err := getBalance(userID)
	if err != nil {
		ErrorPageTrans(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	if balance < amount {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Insufficient funds")
		return
	}

	// Get user's currency
	var currency string
	err = config.DB.QueryRow("SELECT currency FROM users WHERE user_id = ?", userID).Scan(&currency)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	exchangeRate := 1.0
	if currency != "KES" {
		exchangeRate, err = getExchangeRate(currency, "KES")
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, "Failed to fetch exchange rate")
			return
		}
	}

	stmt, err := config.DB.Prepare("INSERT INTO transactions (user_id, type, amount, currency, exchange_rate, created_at) VALUES (?, 'withdraw', ?, ?, ?, ?)")
	if err != nil {
		ErrorPageTrans(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, amount, currency, exchangeRate, time.Now())
	if err != nil {
		ErrorPageTrans(w, r, http.StatusInternalServerError, "Failed to withdraw")
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// SendMoney function
func SendMoney(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		ErrorPage(w, r, http.StatusUnauthorized, "User not authenticated")
		return
	}

	recipientID, err := fetchUserID(r.FormValue("recipient"))
	if err != nil || recipientID == "" {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Invalid recipient")
		return
	}

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil || amount <= 0 {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Invalid amount")
		return
	}

	balance, err := getBalance(userID)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	if balance < amount {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Insufficient funds")
		return
	}

	// Get sender and recipient currencies
	var senderCurrency, recipientCurrency string
	err = config.DB.QueryRow("SELECT currency FROM users WHERE user_id = ?", userID).Scan(&senderCurrency)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	err = config.DB.QueryRow("SELECT currency FROM users WHERE user_id = ?", recipientID).Scan(&recipientCurrency)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	senderExchangeRate := 1.0
	if senderCurrency != "KES" {
		senderExchangeRate, err = getExchangeRate(senderCurrency, "KES")
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, "Failed to fetch sender exchange rate")
			return
		}
	}

	recipientExchangeRate := 1.0
	if recipientCurrency != "KES" {
		recipientExchangeRate, err = getExchangeRate(recipientCurrency, "KES")
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, "Failed to fetch recipient exchange rate")
			return
		}
	}

	// Amount in KES for balance check
	amountKES := float64(amount) * (1.0 / senderExchangeRate)

	// Amount in recipient's currency
	amountRecipient := int(float64(amountKES) * recipientExchangeRate)

	stmt, err := config.DB.Prepare("INSERT INTO transactions (user_id, type, amount, currency, exchange_rate, created_at) VALUES (?, 'send', ?, ?, ?, ?)")
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer stmt.Close()

	stmt2, err := config.DB.Prepare("INSERT INTO transactions (user_id, type, amount, currency, exchange_rate, created_at) VALUES (?, 'receive', ?, ?, ?, ?)")
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer stmt2.Close()

	_, err = stmt.Exec(userID, amount, senderCurrency, senderExchangeRate, time.Now())
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to send money")
		return
	}

	_, err = stmt2.Exec(recipientID, amountRecipient, recipientCurrency, recipientExchangeRate, time.Now())
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to receive money")
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// BuyAirtime function
func BuyAirtime(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		ErrorPage(w, r, http.StatusUnauthorized, "User not authenticated")
		return
	}

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil || amount <= 0 {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Invalid amount")
		return
	}

	balance, err := getBalance(userID)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	if balance < amount {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Insufficient funds")
		return
	}

	// Get user's currency
	var currency string
	err = config.DB.QueryRow("SELECT currency FROM users WHERE user_id = ?", userID).Scan(&currency)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	exchangeRate := 1.0
	if currency != "KES" {
		exchangeRate, err = getExchangeRate(currency, "KES")
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, "Failed to fetch exchange rate")
			return
		}
	}

	stmt, err := config.DB.Prepare("INSERT INTO transactions (user_id, type, amount, currency, exchange_rate, created_at) VALUES (?, 'buy_airtime', ?, ?, ?, ?)")
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, amount, currency, exchangeRate, time.Now())
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to buy airtime")
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Saving function
func Saving(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		ErrorPage(w, r, http.StatusUnauthorized, "User not authenticated")
		return
	}

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil || amount <= 0 {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Invalid amount")
		return
	}

	balance, err := getBalance(userID)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	if balance < amount {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Insufficient funds")
		return
	}

	// Get user's currency
	var currency string
	err = config.DB.QueryRow("SELECT currency FROM users WHERE user_id = ?", userID).Scan(&currency)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	exchangeRate := 1.0
	if currency != "KES" {
		exchangeRate, err = getExchangeRate(currency, "KES")
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, "Failed to fetch exchange rate")
			return
		}
	}

	stmt, err := config.DB.Prepare("INSERT INTO transactions (user_id, type, amount, currency, exchange_rate, created_at) VALUES (?, 'saving', ?, ?, ?, ?)")
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer stmt.Close()

	stmt2, err := config.DB.Prepare("INSERT INTO savings (user_id, savings_id, amount, created_at) VALUES (?, ?, ?, ?)")
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer stmt2.Close()

	_, err = stmt.Exec(userID, amount, currency, exchangeRate, time.Now())
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to save")
		return
	}

	_, err = stmt2.Exec(userID, uuid.New().String(), amount, time.Now())
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to save to savings")
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Balance function
func Balance(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		ErrorPageTrans(w, r, http.StatusUnauthorized, "User not authenticated")
		return
	}

	balance, err := getBalance(userID)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"balance": balance})
}

// getExchangeRate fetches the exchange rate from source to target currency
func getExchangeRate(source, target string) (float64, error) {
	client := resty.New()
	resp, err := client.R().Get(exchangeRateURL + source)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch exchange rates: %v", err)
	}

	var result ExchangeRateResponse
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return 0, fmt.Errorf("failed to parse exchange rates: %v", err)
	}

	if result.Result != "success" {
		return 0, fmt.Errorf("exchange rate API error: %s", result.Result)
	}

	rate, ok := result.ConversionRates[target]
	if !ok {
		return 0, fmt.Errorf("target currency %s not supported", target)
	}

	return rate, nil
}

// ExchangePage renders the currency exchange page
func ExchangePage(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var verificationStatus string
	err = config.DB.QueryRow("SELECT verification_status FROM users WHERE user_id = ?", userID).Scan(&verificationStatus)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	if verificationStatus != "verified" {
		ErrorPage(w, r, http.StatusForbidden, "Account must be verified to access exchange services")
		return
	}

	// Fetch user's currency
	var currency string
	err = config.DB.QueryRow("SELECT currency FROM users WHERE user_id = ?", userID).Scan(&currency)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	data := struct {
		Currency string
	}{
		Currency: currency,
	}

	tmpl := template.Must(template.ParseFiles("templates/exchange.html"))
	err = tmpl.Execute(w, data)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Template rendering error")
		return
	}
}

// ConvertCurrency handles currency conversion
func ConvertCurrency(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		ErrorPage(w, r, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/exchange", http.StatusSeeOther)
		return
	}

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil || amount <= 0 {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Invalid amount")
		return
	}

	fromCurrency := r.FormValue("from_currency")
	toCurrency := r.FormValue("to_currency")

	if fromCurrency == toCurrency {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Source and target currencies must be different")
		return
	}

	// Get user's currency
	var userCurrency string
	err = config.DB.QueryRow("SELECT currency FROM users WHERE user_id = ?", userID).Scan(&userCurrency)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	// Check balance
	balance, err := getBalance(userID)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	// Convert amount to KES for balance check
	exchangeRateToKES := 1.0
	if fromCurrency != "KES" {
		exchangeRateToKES, err = getExchangeRate(fromCurrency, "KES")
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, "Failed to fetch exchange rate")
			return
		}
	}
	amountKES := float64(amount) * (1.0 / exchangeRateToKES)

	if balance < int(amountKES) {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Insufficient funds")
		return
	}

	// Calculate amount in target currency
	exchangeRate, err := getExchangeRate(fromCurrency, toCurrency)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to fetch exchange rate")
		return
	}
	amountConverted := int(float64(amount) * exchangeRate)

	// Record withdrawal in source currency
	stmt, err := config.DB.Prepare("INSERT INTO transactions (user_id, type, amount, currency, exchange_rate, created_at) VALUES (?, 'exchange_out', ?, ?, ?, ?)")
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, amount, fromCurrency, exchangeRateToKES, time.Now())
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to record exchange")
		return
	}

	// Record deposit in target currency
	stmt2, err := config.DB.Prepare("INSERT INTO transactions (user_id, type, amount, currency, exchange_rate, created_at) VALUES (?, 'exchange_in', ?, ?, ?, ?)")
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer stmt2.Close()

	exchangeRateToKES2 := 1.0
	if toCurrency != "KES" {
		exchangeRateToKES2, err = getExchangeRate(toCurrency, "KES")
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, "Failed to fetch exchange rate")
			return
		}
	}

	_, err = stmt2.Exec(userID, amountConverted, toCurrency, exchangeRateToKES2, time.Now())
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to record exchange")
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// InternationalTransfer handles transfers to users in different currencies
func InternationalTransfer(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		ErrorPage(w, r, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/exchange", http.StatusSeeOther)
		return
	}

	recipientID, err := fetchUserID(r.FormValue("recipient"))
	if err != nil || recipientID == "" {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Invalid recipient")
		return
	}

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil || amount <= 0 {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Invalid amount")
		return
	}

	// Get sender and recipient currencies
	var senderCurrency, recipientCurrency string
	err = config.DB.QueryRow("SELECT currency FROM users WHERE user_id = ?", userID).Scan(&senderCurrency)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	err = config.DB.QueryRow("SELECT currency FROM users WHERE user_id = ?", recipientID).Scan(&recipientCurrency)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	// Check balance
	balance, err := getBalance(userID)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	// Convert amount to KES for balance check
	senderExchangeRate := 1.0
	if senderCurrency != "KES" {
		senderExchangeRate, err = getExchangeRate(senderCurrency, "KES")
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, "Failed to fetch sender exchange rate")
			return
		}
	}
	amountKES := float64(amount) * (1.0 / senderExchangeRate)

	if balance < int(amountKES) {
		ErrorPageTrans(w, r, http.StatusBadRequest, "Insufficient funds")
		return
	}

	// Calculate amount in recipient's currency
	recipientExchangeRate := 1.0
	if recipientCurrency != "KES" {
		recipientExchangeRate, err = getExchangeRate(recipientCurrency, "KES")
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, "Failed to fetch recipient exchange rate")
			return
		}
	}
	amountRecipient := int(amountKES * recipientExchangeRate)

	// Record sender's transaction
	stmt, err := config.DB.Prepare("INSERT INTO transactions (user_id, type, amount, currency, exchange_rate, created_at) VALUES (?, 'international_send', ?, ?, ?, ?)")
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, amount, senderCurrency, senderExchangeRate, time.Now())
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to send money")
		return
	}

	// Record recipient's transaction
	stmt2, err := config.DB.Prepare("INSERT INTO transactions (user_id, type, amount, currency, exchange_rate, created_at) VALUES (?, 'international_receive', ?, ?, ?, ?)")
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer stmt2.Close()

	_, err = stmt2.Exec(recipientID, amountRecipient, recipientCurrency, recipientExchangeRate, time.Now())
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to receive money")
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// DownloadStatement generates and serves a bank statement PDF
func DownloadStatement(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		ErrorPage(w, r, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var verificationStatus string
	err = config.DB.QueryRow("SELECT verification_status FROM users WHERE user_id = ?", userID).Scan(&verificationStatus)
	if err != nil {
		fmt.Println("Error fetching verification status:", err)
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	if verificationStatus != "verified" {
		ErrorPage(w, r, http.StatusForbidden, "Account must be verified to download statements")
		return
	}

	var name, accountNumber, currency string
	err = config.DB.QueryRow("SELECT name, account_number, currency FROM users WHERE user_id = ?", userID).Scan(&name, &accountNumber, &currency)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to fetch user details")
		return
	}

	rows, err := config.DB.Query(`
		SELECT id, type, amount, currency, created_at 
		FROM transactions 
		WHERE user_id = ? 
		ORDER BY created_at DESC`, userID)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	type Transaction struct {
		ID        int
		Type      string
		Amount    int
		Currency  string
		CreatedAt time.Time
	}

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		err := rows.Scan(&t.ID, &t.Type, &t.Amount, &t.Currency, &t.CreatedAt)
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, "Database error")
			return
		}
		transactions = append(transactions, t)
	}

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	err = pdf.AddTTFFont("roboto", "static/fonts/Roboto-Regular.ttf")
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to load font")
		return
	}
	err = pdf.AddTTFFont("roboto-bold", "static/fonts/Roboto-Bold.ttf")
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to load bold font")
		return
	}

	pdf.AddPage()

	pdf.SetFont("roboto-bold", "", 16)
	pdf.SetXY(50, 30)
	pdf.Text("Insight Bank Statement")

	pdf.SetFont("roboto", "", 12)
	pdf.SetXY(50, 50)
	pdf.Text(fmt.Sprintf("Account Holder: %s", name))
	pdf.SetXY(50, 65)
	pdf.Text(fmt.Sprintf("Account Number: %s", accountNumber))
	pdf.SetXY(50, 80)
	pdf.Text(fmt.Sprintf("Currency: %s", currency))
	pdf.SetXY(50, 95)
	pdf.Text(fmt.Sprintf("Date: %s", time.Now().Format("02 January 2006")))

	pdf.SetFont("roboto", "", 40)
	pdf.SetTextColor(200, 200, 200)
	pdf.Rotate(45, 300, 400)
	pdf.SetXY(200, 350)
	pdf.Text("OFFICIAL STATEMENT")
	pdf.RotateReset()
	pdf.SetTextColor(0, 0, 0)

	pdf.SetFont("roboto-bold", "", 10)
	pdf.SetXY(50, 120)
	pdf.CellWithOption(&gopdf.Rect{W: 50, H: 10}, "Trans. ID", gopdf.CellOption{Align: gopdf.Left})
	pdf.SetXY(100, 120)
	pdf.CellWithOption(&gopdf.Rect{W: 100, H: 10}, "Type", gopdf.CellOption{Align: gopdf.Left})
	pdf.SetXY(200, 120)
	pdf.CellWithOption(&gopdf.Rect{W: 80, H: 10}, "Amount", gopdf.CellOption{Align: gopdf.Right})
	pdf.SetXY(280, 120)
	pdf.CellWithOption(&gopdf.Rect{W: 50, H: 10}, "Currency", gopdf.CellOption{Align: gopdf.Left})
	pdf.SetXY(330, 120)
	pdf.CellWithOption(&gopdf.Rect{W: 150, H: 10}, "Date", gopdf.CellOption{Align: gopdf.Left})

	pdf.SetLineWidth(0.5)
	pdf.Line(50, 130, 500, 130)

	pdf.SetFont("roboto", "", 10)
	y := 140.0
	for _, t := range transactions {
		pdf.SetXY(50, y)
		pdf.CellWithOption(&gopdf.Rect{W: 50, H: 10}, fmt.Sprintf("%d", t.ID), gopdf.CellOption{Align: gopdf.Left})
		pdf.SetXY(100, y)
		pdf.CellWithOption(&gopdf.Rect{W: 100, H: 10}, t.Type, gopdf.CellOption{Align: gopdf.Left})
		pdf.SetXY(200, y)
		pdf.CellWithOption(&gopdf.Rect{W: 80, H: 10}, fmt.Sprintf("%d", t.Amount), gopdf.CellOption{Align: gopdf.Right})
		pdf.SetXY(280, y)
		pdf.CellWithOption(&gopdf.Rect{W: 50, H: 10}, t.Currency, gopdf.CellOption{Align: gopdf.Left})
		pdf.SetXY(330, y)
		pdf.CellWithOption(&gopdf.Rect{W: 150, H: 10}, t.CreatedAt.Format("02 Jan 2006 15:04"), gopdf.CellOption{Align: gopdf.Left})
		y += 15
	}

	balance, err := getBalance(userID)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to calculate balance")
		return
	}
	pdf.SetFont("roboto-bold", "", 12)
	pdf.SetXY(50, y+20)
	pdf.Text(fmt.Sprintf("Current Balance: KES %d", balance))

	tempDir := "uploads/statements"
	err = os.MkdirAll(tempDir, 0755)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to create temp directory")
		return
	}
	tempFile := filepath.Join(tempDir, fmt.Sprintf("statement_%s.pdf", uuid.New().String()))
	err = pdf.WritePdf(tempFile)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to generate PDF")
		return
	}
	defer os.Remove(tempFile)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"InsightBank_Statement_%s.pdf\"", time.Now().Format("20060102")))
	http.ServeFile(w, r, tempFile)
}
