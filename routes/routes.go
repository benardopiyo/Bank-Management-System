package routes

import (
	"Bank-Management-System/handlers"
	"net/http"

	"github.com/gorilla/mux"
)

func Routes() *mux.Router {
	mux := mux.NewRouter()

	mux.HandleFunc("/", handlers.HomePage).Methods("GET")
	mux.HandleFunc("/register", handlers.RegisterPage).Methods("GET")
	mux.HandleFunc("/register", handlers.Register).Methods("POST")
	mux.HandleFunc("/login", handlers.LoginPage).Methods("GET")
	mux.HandleFunc("/login", handlers.Login).Methods("POST")
	mux.HandleFunc("/logout", handlers.Logout).Methods("GET")

	// Protected Routes
	mux.HandleFunc("/dashboard", handlers.Dashboard).Methods("GET")
	mux.HandleFunc("/deposit", handlers.Deposit).Methods("POST")
	mux.HandleFunc("/withdraw", handlers.Withdraw).Methods("POST")
	mux.HandleFunc("/balance", handlers.Balance).Methods("GET")

	// Loan-related routes - FIXED
	mux.HandleFunc("/loan", handlers.LoanPage).Methods("GET")
	mux.HandleFunc("/apply-loan", handlers.ApplyLoan).Methods("POST")
	mux.HandleFunc("/view-loans", handlers.ViewLoans).Methods("GET")

	// Serve static files
	staticDir := "/static/"
	fs := http.StripPrefix(staticDir, http.FileServer(http.Dir("static")))
	mux.PathPrefix(staticDir).Handler(fs)

	return mux
}
