package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"personal-analytics-backend/internal/db"
	"personal-analytics-backend/internal/redis"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
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
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshRequest
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
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
	userID, err := db.CreateUser(r.Context(), req.Email, string(passwordHash))
	if err != nil {
		// Check if email already exists (SQLite UNIQUE constraint violation)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
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
	userID, passwordHash, err := db.GetUserByEmail(r.Context(), req.Email)
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
		"exp":     time.Now().Add(15 * time.Minute).Unix(), // Unix() converts time to NUMBER
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

	// Generating the refresh token
	random_UUID := uuid.New().String()
	refresh_token_key := "refresh_token:" + random_UUID
	refresh_token_value := fmt.Sprintf("%d", userID)

	if err := redis.Client.Set(r.Context(), refresh_token_key, refresh_token_value, 7*24*time.Hour).Err(); err != nil {
		slog.Error("Failed to store refresh token", "error", err)
		errorResponseAuth(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Success response with token
	slog.Info("User logged in", "user_id", userID)
	respondJSON(w, http.StatusOK, LoginResponse{
		Success:      true,
		Message:      "Login successful",
		AccessToken:  tokenString,
		RefreshToken: random_UUID,
	})
}

// Refresh the access token
func Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request body
	var req RefreshRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorResponseAuth(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		errorResponseAuth(w, http.StatusBadRequest, "Refresh token is required.")
		return
	}

	key := "refresh_token:" + req.RefreshToken
	val, err := redis.Client.Get(r.Context(), key).Result()

	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	userID, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Invalid session data")
		return
	}
	redis.Client.Del(r.Context(), key)

	// Generate JWT token
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		slog.Error("JWT_SECRET not set in environment")
		errorResponse(w, http.StatusInternalServerError, "Server configuration error")
		return
	}

	// data to put inside the token
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}

	// create token object
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// signing token and covert to string
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		slog.Error("Error generating JWT token", "error", err)
		errorResponse(w, http.StatusInternalServerError, "Failed to generate toten")
		return
	}

	// Generating the refresh token
	random_UUID := uuid.New().String()
	refresh_token_key := "refresh_token:" + random_UUID
	refresh_token_value := fmt.Sprintf("%d", userID)

	if err := redis.Client.Set(r.Context(), refresh_token_key, refresh_token_value, 7*24*time.Hour).Err(); err != nil {
		slog.Error("Failed to store refresh token", "error", err)
		errorResponseAuth(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Success response with token and refresh token
	respondJSON(w, http.StatusOK, LoginResponse{
		Success:      true,
		Message:      "Login successfully",
		AccessToken:  tokenString,
		RefreshToken: random_UUID,
	})
}

// Logout
func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RefreshRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		errorResponseAuth(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		errorResponseAuth(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	key := "refresh_token:" + req.RefreshToken
	val, err := redis.Client.Get(r.Context(), key).Result()

	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	if val == "" {
		errorResponse(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	redis.Client.Del(r.Context(), key)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
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

/*
=== INTERVIEW ANSWER: JWT AUTHENTICATION ===

WHAT:
JWT (JSON Web Token) is a self-contained identity token. After login, the server
issues a signed token. The client sends it with every request to prove identity.
No session stored on server — the token itself carries the proof.

AUTHENTICATION vs AUTHORIZATION:
- Authentication: proving who you are (login, token verification) ← what JWT does
- Authorization: what you're allowed to do after identity is confirmed

THE 3 PARTS OF A JWT (header.payload.signature):
- Header:    algorithm used (HS256). Not secret, just metadata.
- Payload:   claims — user_id, exp. Base64 encoded, NOT encrypted.
             Anyone can decode and read it. Never put passwords here.
- Signature: HMAC(header + payload + JWT_SECRET). This is the security.
             Tamper with payload → signature no longer matches → rejected.

FULL FLOW:
1. Register: validate → bcrypt hash password → save user to DB
2. Login: validate → fetch user by email → bcrypt.CompareHashAndPassword
   → create claims {user_id, exp} → jwt.NewWithClaims (unsigned Go object)
   → .SignedString([]byte(secret)) → signed token string returned to client
3. Protected request: client sends Authorization: Bearer <token>
4. AuthMiddleware: strip "Bearer " → ParseWithClaims (verifies signature + expiry)
   → extract user_id → context.WithValue(r.Context(), "user_id", id)
5. Handler: r.Context().Value("user_id") → query only that user's data
6. Token invalid/expired → 401 Unauthorized → return → handler never runs

WHAT IS context.Context:
A bag of request-scoped values that travels with the request through the entire
middleware chain into the handler. Standard Go way to pass data between middleware
and handlers without changing function signatures. Every http.Request carries one.

BCRYPT PASSWORD HASHING:
- One-way + salted. Same password hashed twice = different results.
- bcrypt.CompareHashAndPassword extracts salt from stored hash, re-hashes input
  with same salt, compares. You never store or see the plain password.
- Salt prevents: two users with same password having identical hashes in DB.

WHY JWT_SECRET IS CRITICAL:
It's not related to passwords. It's the stamp that proves tokens came from your
server. With the secret, an attacker forges valid tokens for ANY user_id with no
password, no DB, no login needed. Server cannot distinguish forged from real.
Store only in env var, never in code or logs. If leaked → rotate immediately.

TRADE-OFFS:
- JWT can't be invalidated before expiry (logout doesn't truly invalidate token)
- Production fix: token blocklist in Redis, or short expiry + refresh tokens
- Stateless = scales easily (no session DB needed), but revocation is hard
*/
