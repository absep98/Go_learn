package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"personal-analytics-backend/internal/db"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse represents the registration response
type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	UserID  int64  `json:"user_id,omitempty"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token"`
}

// Register handles POST /register
func Register(w http.ResponseWriter, r *http.Request) {
	slog.Info("Request received", "method", "POST", "path", "/register")

	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request body
	var req RegisterRequest
	// json.NewDecoder: Use this for Network Requests (r.Body). Since the data is still "traveling" through the internet wires, the Decoder is more efficient. It processes the data as it arrives, piece by piece, rather than waiting for the whole thing to finish downloading.

	// json.NewDecoder: Prepares a "Streaming Machine" to read data from a pipe.

	// r.Body: The "Pipe" where the user's data is flowing in.

	// Decode: The "Action" of translating JSON text into Go variables.

	// &: The "GPS Coordinates" so the machine knows where to save the data.
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorResponseAuth(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate email
	if req.Email == "" {
		errorResponseAuth(w, http.StatusBadRequest, "email is required")
		return
	}

	// Basic email validation (contains @)
	if !strings.Contains(req.Email, "@") {
		errorResponseAuth(w, http.StatusBadRequest, "invalid email format")
		return
	}

	// Validate password
	if req.Password == "" {
		errorResponseAuth(w, http.StatusBadRequest, "password is required")
		return
	}

	// Password must be at least 6 characters
	if len(req.Password) < 6 {
		errorResponseAuth(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	// Hash the password using bcrypt
	// bcrypt.DefaultCost = 10 (good balance between security and speed)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Error hashing password", "error", err)
		errorResponseAuth(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Save user to database
	userID, err := db.CreateUser(req.Email, string(passwordHash))
	if err != nil {
		// Check if email already exists (SQLite UNIQUE constraint violation)
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			errorResponseAuth(w, http.StatusConflict, "Email already registered")
			return
		}

		slog.Error("Error creating user", "error", err)
		errorResponseAuth(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Success response
	slog.Info("User registered", "user_id", userID)
	respondJSON(w, http.StatusCreated, RegisterResponse{
		Success: true,
		Message: "User registered successfully",
		UserID:  userID,
	})
}

// Login handles POST /login
func Login(w http.ResponseWriter, r *http.Request) {
	slog.Info("Request received", "method", "POST", "path", "/login")

	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request body
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorResponseAuth(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate email and password are not empty
	if req.Email == "" {
		errorResponseAuth(w, http.StatusBadRequest, "email is required")
		return
	}

	if req.Password == "" {
		errorResponseAuth(w, http.StatusBadRequest, "password is required")
		return
	}

	// Get user from database
	userID, passwordHash, err := db.GetUserByEmail(req.Email)
	if err != nil {
		// Don't reveal if user exists or not (security best practice)
		slog.Warn("Login attempt for non-existent user", "email", req.Email)
		errorResponseAuth(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Compare password with stored hash
	// bcrypt.CompareHashAndPassword: Checks if plain password matches the encrypted hash
	// []byte(passwordHash): Convert hash string to bytes (bcrypt library requirement)
	// []byte(req.Password): Convert user's password to bytes
	// Returns: nil if match, error if mismatch
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		// Password doesn't match
		slog.Warn("Invalid password attempt", "user_id", userID)
		errorResponseAuth(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Generate JWT token
	// JWT token = "ticket" proving user logged in (contains user_id)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		slog.Error("JWT_SECRET not set in environment")
		errorResponseAuth(w, http.StatusInternalServerError, "Server configuration error")
		return
	}

	// STEP 1: Create claims (data to put inside token)
	// jwt.MapClaims: Just a map[string]interface{} - you already know this!
	// "user_id": userID - So we know WHO this token belongs to
	// "exp": Expiration time - Token becomes invalid after 24 hours
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // Unix() converts time to NUMBER
		// Why Unix()? JWT needs simple numbers, not complex Go time objects
		// Example: time.Now().Unix() = 1736359530 (just a number)
		// Add 24 hours = add 86400 seconds (24 * 60 * 60 = 86400)
	}

	// STEP 2: Create token object (NOT a string yet, just a Go object)
	// jwt.SigningMethodHS256: Encryption method (standard, secure, don't worry about it)
	// claims: The data map we just created above
	// Think: Creating an unsigned document with your data inside
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// STEP 3: Sign token and convert to string (this makes it official!)
	// This is like putting a wax seal on a letter - proves it's authentic
	// []byte(secret): Your secret key (converts string to bytes - library requirement)
	// Returns: "eyJhbGci...xyz" (the actual JWT string user receives)
	// Format: header.payload.signature (3 parts separated by dots)
	// "The secret key is a string, but encryption libraries need bytes (raw data format), so we convert it with []byte(secret). This lets the library encrypt and sign the token."

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		slog.Error("Error generating JWT token", "error", err)
		errorResponseAuth(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Success response with token
	slog.Info("User logged in", "user_id", userID)
	respondJSON(w, http.StatusOK, LoginResponse{
		Success: true,
		Message: "Login successful",
		Token:   tokenString,
	})
}

// errorResponseAuth sends error response for auth endpoints
func errorResponseAuth(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, RegisterResponse{
		Success: false,
		Message: message,
	})
}

// respondJSON is defined in entries.go and shared across all handlers in this package
