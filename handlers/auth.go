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
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
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

// CustomerProfile struct for displaying customer details
type CustomerProfile struct {
	Name          string
	Username      string
	AccountNumber string
	PhotoPath     string
}

// Predefined branch codes
var branchCodes = map[string]string{
	"Nairobi": "001",
	"Kisumu":  "002",
	"Mombasa": "003",
	"Nakuru":  "004",
}

// Face++ API credentials and endpoints
const (
	facePlusPlusAPIKey     = "dX54O0lbt6ViV3q17fYdmDUXFbVSuf_V"
	facePlusPlusAPISecret  = "OtqRgnOLTk8jX90oALcuFPSwVMus6BPt"
	facePlusPlusCompareURL = "https://api-us.faceplusplus.com/facepp/v3/compare"
	facePlusPlusOCRURL     = "https://api-us.faceplusplus.com/facepp/v3/ocr/idcard"
)

// GenerateAccountNumber creates a unique account number
func GenerateAccountNumber(branch string) (string, error) {
	branchCode, ok := branchCodes[branch]
	if !ok {
		return "", fmt.Errorf("invalid branch: %s", branch)
	}

	var lastSeq int
	err := config.DB.QueryRow("SELECT COALESCE(MAX(CAST(SUBSTR(account_number, 6) AS INTEGER)), 0) FROM users WHERE branch = ?", branch).Scan(&lastSeq)
	if err != nil {
		return "", err
	}

	seq := lastSeq + 1
	accountNumber := fmt.Sprintf("%s01%06d", branchCode, seq)
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

// Register User with Face++ Verification and ID Authenticity
func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
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

	accountNumber, err := GenerateAccountNumber(branch)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to generate account number")
		return
	}

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

	client := resty.New()

	respCompare, err := client.R().
		SetFormData(map[string]string{
			"api_key":    facePlusPlusAPIKey,
			"api_secret": facePlusPlusAPISecret,
		}).
		SetFiles(map[string]string{
			"image_file1": photoPath,
			"image_file2": idPath,
		}).
		Post(facePlusPlusCompareURL)

	faceMatch := false
	if err != nil {
		fmt.Println("Face++ Compare API error:", err)
	} else {
		var result struct {
			Confidence float64 `json:"confidence"`
			Thresholds struct {
				E5 float64 `json:"1e-5"`
			} `json:"thresholds"`
		}
		err = json.Unmarshal(respCompare.Body(), &result)
		if err == nil && result.Confidence > result.Thresholds.E5 {
			faceMatch = true
		}
	}

	respOCR, err := client.R().
		SetFormData(map[string]string{
			"api_key":    facePlusPlusAPIKey,
			"api_secret": facePlusPlusAPISecret,
		}).
		SetFiles(map[string]string{
			"image_file": idPath,
		}).
		Post(facePlusPlusOCRURL)

	var idName string
	nameMatch := false
	if err != nil {
		fmt.Println("Face++ OCR API error:", err)
	} else {
		var ocrResult struct {
			Cards []struct {
				Name string `json:"name"`
			} `json:"cards"`
		}
		err = json.Unmarshal(respOCR.Body(), &ocrResult)
		if err == nil && len(ocrResult.Cards) > 0 {
			idName = ocrResult.Cards[0].Name
			nameMatch = strings.TrimSpace(strings.ToLower(idName)) == strings.TrimSpace(strings.ToLower(name))
		}
	}

	autoStatus := "pending"
	if faceMatch && nameMatch {
		autoStatus = "verified"
	} else if !faceMatch || !nameMatch {
		autoStatus = "failed"
	}

	userID := uuid.New().String()
	stmt, err := config.DB.Prepare(`
        INSERT INTO users (user_id, name, user_name, user_pin, confirm_pin, account_number, branch, photo_path, id_path, verification_status, auto_verification_status, currency, created_at) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending', ?, 'KES', CURRENT_TIMESTAMP)`)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	_, err = stmt.Exec(userID, name, username, pin, confirmPin, accountNumber, branch, photoPath, idPath, autoStatus)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to register")
		return
	}

	fmt.Printf("User: %s, ID Name: %s, Face Match: %v, Name Match: %v, Auto Status: %s\n", name, idName, faceMatch, nameMatch, autoStatus)

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

	config.DB.Exec("INSERT INTO sessions (session_token, user_id, expires_at) VALUES (?, ?, ?)", sessionToken, user.ID, expiration)

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

// Protected Dashboard with Customer Profile
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

	// Fetch customer profile data
	var profile CustomerProfile
	err = config.DB.QueryRow("SELECT name, user_name, account_number, photo_path FROM users WHERE user_id = ?", userID).
		Scan(&profile.Name, &profile.Username, &profile.AccountNumber, &profile.PhotoPath)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Failed to fetch profile")
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/dashboard.html"))
	err = tmpl.Execute(w, profile)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Template rendering error")
		return
	}
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

// AdminDashboard displays pending users with auto verification status
func AdminDashboard(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query(`
		SELECT user_id, name, user_name, account_number, branch, photo_path, id_path, auto_verification_status 
		FROM users WHERE verification_status = 'pending'`)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	type PendingUser struct {
		UserID                 string
		Name                   string
		Username               string
		AccountNumber          string
		Branch                 string
		PhotoPath              string
		IDPath                 string
		AutoVerificationStatus string
	}

	var pendingUsers []PendingUser
	for rows.Next() {
		var user PendingUser
		err := rows.Scan(&user.UserID, &user.Name, &user.Username, &user.AccountNumber, &user.Branch, &user.PhotoPath, &user.IDPath, &user.AutoVerificationStatus)
		if err != nil {
			ErrorPage(w, r, http.StatusInternalServerError, "Database error")
			return
		}
		pendingUsers = append(pendingUsers, user)
	}

	fmt.Printf("Pending users: %+v\n", pendingUsers)

	tmpl := template.Must(template.ParseFiles("templates/admin_dashboard.html"))
	err = tmpl.Execute(w, pendingUsers)
	if err != nil {
		ErrorPage(w, r, http.StatusInternalServerError, "Template rendering error")
		return
	}
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
