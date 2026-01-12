package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware verifies JWT token before allowing access to protected routes
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("🔐 Middleware - Checking authentication")

		// STEP 1: Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("❌ No Authorization header found")
			errorResponseAuth(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		// STEP 2: Extract token (remove "Bearer " prefix)
		// Header format: "Bearer eyJhbGci...xyz"
		// We want just: "eyJhbGci...xyz"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			// TrimPrefix didn't remove anything = no "Bearer " prefix
			log.Println("❌ Invalid Authorization header format")
			errorResponseAuth(w, http.StatusUnauthorized, "Invalid authorization format. Use: Bearer <token>")
			return
		}

		// STEP 3: Verify token is valid and extract claims
		claims, err := validateToken(tokenString)
		if err != nil {
			log.Printf("❌ Token validation failed: %v", err)
			errorResponseAuth(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// STEP 4: Extract user_id from claims
		userID, ok := claims["user_id"].(float64) // JWT numbers come as float64
		if !ok {
			log.Println("❌ user_id not found in token claims")
			errorResponseAuth(w, http.StatusUnauthorized, "Invalid token claims")
			return
		}

		// STEP 5: Put user_id in request context (like a backpack for passing data)
		ctx := context.WithValue(r.Context(), "user_id", int64(userID))

		// STEP 6: Call next handler with updated context
		log.Printf("✅ User %d authenticated successfully", int64(userID))
		next(w, r.WithContext(ctx))
	}
}

// validateToken verifies JWT token signature and returns claims
func validateToken(tokenString string) (jwt.MapClaims, error) {
	// Get secret key from environment
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET not configured")
	}

	// Parse token (opposite of creating it!)
	// jwt.Parse: Takes token string, returns token object
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// This function is called by jwt.Parse to get the secret key
		// Check if signing method matches what we used (HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("Token has expired %w", err)
		}
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, fmt.Errorf("Invalid token signature %w", err)
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, fmt.Errorf("Malformed token %w", err)
		}
	}

	// Extract claims from token
	// token.Valid checks if token is not expired
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}
