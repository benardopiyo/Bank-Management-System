package routes

import (
	"Bank-Management-System/handlers"
	"net/http"

	"github.com/gorilla/mux"
)

func Routes() *mux.Router {
	mux := mux.NewRouter()

	// Public routes
	mux.HandleFunc("/", handlers.HomePage).Methods("GET")
	mux.HandleFunc("/register", handlers.RegisterPage).Methods("GET")
	mux.HandleFunc("/register", handlers.Register).Methods("POST")
	mux.HandleFunc("/login", handlers.LoginPage).Methods("GET")
	mux.HandleFunc("/login", handlers.Login).Methods("POST")
	mux.HandleFunc("/logout", handlers.Logout).Methods("GET")

	// User routes
	mux.HandleFunc("/dashboard", handlers.Dashboard).Methods("GET")
	mux.HandleFunc("/account-number", handlers.AccountNumber).Methods("GET")
	mux.HandleFunc("/statement", handlers.DownloadStatement).Methods("GET")
	mux.HandleFunc("/exchange", handlers.ExchangePage).Methods("GET")
	mux.HandleFunc("/convert-currency", handlers.ConvertCurrency).Methods("POST")
	mux.HandleFunc("/international-transfer", handlers.InternationalTransfer).Methods("POST")

	// Admin routes
	mux.HandleFunc("/admin", handlers.AdminOnly(handlers.AdminDashboard)).Methods("GET")
	mux.HandleFunc("/approve-user", handlers.AdminOnly(handlers.ApproveUser)).Methods("POST")
	mux.HandleFunc("/admin-loans", handlers.AdminOnly(handlers.AdminLoanDashboard)).Methods("GET")
	mux.HandleFunc("/approve-loan", handlers.AdminOnly(handlers.ApproveLoan)).Methods("POST")

	// Transaction routes
	mux.HandleFunc("/deposit", handlers.Deposit).Methods("POST")
	mux.HandleFunc("/withdraw", handlers.Withdraw).Methods("POST")
	mux.HandleFunc("/balance", handlers.Balance).Methods("GET")
	mux.HandleFunc("/send-money", handlers.SendMoney).Methods("POST")
	mux.HandleFunc("/buy-airtime", handlers.BuyAirtime).Methods("POST")
	mux.HandleFunc("/saving", handlers.Saving).Methods("POST")

	// Loan-related routes
	mux.HandleFunc("/loan", handlers.LoanPage).Methods("GET")
	mux.HandleFunc("/apply-loan", handlers.ApplyLoan).Methods("POST")
	mux.HandleFunc("/repay-loan", handlers.RepayLoan).Methods("POST")
	mux.HandleFunc("/view-loans", handlers.ViewLoans).Methods("GET")

	// Serve static files
	mux.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	return mux
}
