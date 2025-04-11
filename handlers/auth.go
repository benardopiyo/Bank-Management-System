package handlers

import (
	"Bank-Management-System/config"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

// User struct in JSON format
type User struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	Name       string `json:"name"`
	Username   string `json:"username"`
	UserPin    string `json:"user_pin"`
	ConfirmPin string `json:"confirm_pin"`
	CreatedAt  string `json:"created_at"`
}

// Predefined branch codes
var branchCodes = map[string]string{
	"Nairobi": "001",
	"Kisumu":  "002",
	"Mombasa": "003",
	"Nakuru":  "004",
}

// GenerateAccountNumber creates a unique account number
func GenerateAccountNumber(branch string) (string, error) {
	branchCode, ok := branchCodes[branch]
	if !ok {
		return "", fmt.Errorf("invalid branch: %s", branch)
	}

	// Get the latest account sequence for the branch
	var lastSeq int
	err := config.DB.QueryRow("SELECT COALESCE(MAX(CAST(SUBSTR(account_number, 6) AS INTEGER)), 0) FROM users WHERE branch = ?", branch).Scan(&lastSeq)
	if err != nil {
		return "", err
	}

	// Increment sequence and format (e.g., 00101 + 000001 = 00101000001)
	seq := lastSeq + 1
	accountNumber := fmt.Sprintf("%s01%06d", branchCode, seq) // 01 = Account Type (e.g., savings)
	return accountNumber, nil
}

// Hash password
func hashPassword(pin string) string {
	hash := sha256.Sum256([]byte(pin))
	return hex.EncodeToString(hash[:])
}

// Home Page
func HomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		ErrorPage(w, r, http.StatusNotFound, "Not Found")
		return
	}
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

	// Parse multipart form for file uploads
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		ErrorPage(w, r, http.StatusBadRequest, "Failed to parse form")
		return
	}

	name := r.FormValue("name")
	username := r.FormValue("username")
	pin := hashPassword(r.FormValue("pin"))
	confirmPin := hashPassword(r.FormValue("confirm-pin"))
	branch := r.FormValue("branch")

	if pin != confirmPin {
		ErrorPage(w, r, http.StatusBadRequest, "PINs do not match")
		return
	}

	var existingUser User
	err = config.DB.QueryRow("SELECT user_id, user_name FROM users WHERE user_name=?", username).Scan(&existingUser.ID, &existingUser.Name)
	if err == nil {
		ErrorPage(w, r, http.StatusBadRequest, "Username already exists")
		return
	}

	// Generate account number
	accountNumber, err := GenerateAccountNumber(branch)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to generate account number")
		return
	}

	// Handle photo and ID uploads
	photoFile, photoHeader, err := r.FormFile("photo")
	var photoPath string
	if err == nil {
		defer photoFile.Close()
		photoPath = fmt.Sprintf("uploads/photos/%s-%s", uuid.New().String(), photoHeader.Filename)
		err = saveFile(photoFile, photoPath)
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, "Failed to save photo")
			return
		}
	}

	idFile, idHeader, err := r.FormFile("id")
	var idPath string
	if err == nil {
		defer idFile.Close()
		idPath = fmt.Sprintf("uploads/ids/%s-%s", uuid.New().String(), idHeader.Filename)
		err = saveFile(idFile, idPath)
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, "Failed to save ID")
			return
		}
	}

	userID := uuid.New().String()
	stmt, err := config.DB.Prepare(`
		INSERT INTO users (user_id, name, user_name, user_pin, confirm_pin, account_number, branch, photo_path, id_path, verification_status, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending', CURRENT_TIMESTAMP)`)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	_, err = stmt.Exec(userID, name, username, pin, confirmPin, accountNumber, branch, photoPath, idPath)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to register")
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Helper function to save uploaded files
func saveFile(file multipart.File, path string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	return err
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
	var role string
	err := config.DB.QueryRow("SELECT user_id, user_name, role FROM users WHERE user_name=? AND user_pin=?", username, pin).Scan(&user.ID, &user.Name, &role)
	if err != nil {
		ErrorPage(w, r, http.StatusUnauthorized, "Invalid credentials")
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

	// Redirect based on role
	if role == "admin" {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
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

	userID, err := getUserIDFromSession(r)
	if err != nil {
		ErrorPage(w, r, http.StatusUnauthorized, "Invalid session")
		return
	}

	var verificationStatus string
	err = config.DB.QueryRow("SELECT verification_status FROM users WHERE user_id = ?", userID).Scan(&verificationStatus)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	if verificationStatus == "pending" {
		ErrorPage(w, r, http.StatusForbidden, "Your account is currently under review. Please allow some time for the verification process to be completed.")
		return
	} else if verificationStatus == "rejected" {
		ErrorPage(w, r, http.StatusForbidden, "Your account verification was not approved. Kindly review your submitted information and reapply.")
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

func AccountNumber(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromSession(r)
	if err != nil || userID == "" {
		ErrorPage(w, r, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var accountNumber string
	err = config.DB.QueryRow("SELECT account_number FROM users WHERE user_id = ?", userID).Scan(&accountNumber)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"accountNumber": accountNumber})
}

// Check if the user is an admin
func isAdmin(r *http.Request) bool {
	userID, err := getUserIDFromSession(r)
	if err != nil {
		return false
	}

	var role string
	err = config.DB.QueryRow("SELECT role FROM users WHERE user_id = ?", userID).Scan(&role)
	return err == nil && role == "admin"
}

// Middleware for admin-only routes
func AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if !isAdmin(r) {
			ErrorPage(w, r, http.StatusForbidden, "Access denied. Admins only.")
			return
		}
		next(w, r)
	}
}

// AdminDashboard displays pending users
func AdminDashboard(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query(`
		SELECT user_id, name, user_name, account_number, branch, photo_path, id_path 
		FROM users WHERE verification_status = 'pending'`)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	type PendingUser struct {
		UserID        string
		Name          string
		Username      string
		AccountNumber string
		Branch        string
		PhotoPath     string
		IDPath        string
	}

	var pendingUsers []PendingUser
	for rows.Next() {
		var user PendingUser
		err := rows.Scan(&user.UserID, &user.Name, &user.Username, &user.AccountNumber, &user.Branch, &user.PhotoPath, &user.IDPath)
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, "Database error")
			return
		}
		pendingUsers = append(pendingUsers, user)
	}

	tmpl := template.Must(template.ParseFiles("templates/admin_dashboard.html"))
	tmpl.Execute(w, pendingUsers)
}

// ApproveUser updates verification status
func ApproveUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	userID := r.FormValue("user_id")
	action := r.FormValue("action")

	var status string
	if action == "approve" {
		status = "verified"
	} else if action == "reject" {
		status = "rejected"
	} else {
		ErrorPage(w, r, http.StatusBadRequest, "Invalid action")
		return
	}

	_, err := config.DB.Exec("UPDATE users SET verification_status = ? WHERE user_id = ?", status, userID)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to update status")
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}