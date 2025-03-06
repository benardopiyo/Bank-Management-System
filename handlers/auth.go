package handlers

import (
	"Bank-Management-System/config"
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// User struct in JSON format
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	PIN      string `json:"-"`
}

// Hash password
func hashPassword(pin string) string {
	hash := sha256.Sum256([]byte(pin))
	return hex.EncodeToString(hash[:])
}

// Home Page
func HomePage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/home.html"))
	tmpl.Execute(w, nil)
}

// Register Page
func RegisterPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/register.html"))
	tmpl.Execute(w, nil)
}

// Register User
func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	username := r.FormValue("username")
	pin := hashPassword(r.FormValue("pin"))
	confirmPin := hashPassword(r.FormValue("confirm-pin"))

	if pin != confirmPin {
		http.Error(w, "PINs do not match", http.StatusBadRequest)
		return
	}

	var existingUser User
	err := config.DB.QueryRow("SELECT user_id, name FROM users WHERE user_name=?", username).Scan(&existingUser.ID, &existingUser.Name)
	if err == nil {
		http.Error(w, "Username already exists", http.StatusBadRequest)
		return
	}

	userID := uuid.New().String()
	stmt, err := config.DB.Prepare(`
		INSERT INTO users (user_id, name, user_name, user_pin, confirm_pin, created_at) 
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
	)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	_, err = stmt.Exec(userID, name, username, pin, confirmPin)
	if err != nil {
		http.Error(w, "Failed to register", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Login Page
func LoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/login.html"))
	tmpl.Execute(w, nil)
}

// Login User (Set Session Cookie)
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := r.FormValue("user-name")
	pin := hashPassword(r.FormValue("pin"))

	var user User
	err := config.DB.QueryRow("SELECT user_id, name FROM users WHERE user_name=? AND user_pin=?", username, pin).Scan(&user.ID, &user.Name)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	sessionToken := uuid.New().String()
	expiration := time.Now().Add(24 * time.Hour)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  expiration,
		HttpOnly: true,
		Path:     "/",
	})

	// Store session mapping to user UUID
	config.DB.Exec("INSERT INTO sessions (session_token, user_id, expires_at) VALUES (?, ?, ?)", sessionToken, user.ID, expiration)

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Logout User (Clear Cookie)
func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Middleware to Protect Routes
func isAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie("session_token")
	return err == nil && cookie.Value != ""
}

// Protected Dashboard
func Dashboard(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/dashboard.html"))
	tmpl.Execute(w, nil)
}

// Middleware: Get user ID from session
func getUserIDFromSession(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return "", err
	}

	var userID string
	err = config.DB.QueryRow("SELECT user_id FROM sessions WHERE session_token=?", cookie.Value).Scan(&userID)
	if err != nil {
		return "", err
	}
	return userID, nil
}
