package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"personal-analytics-backend/internal/db"
	"personal-analytics-backend/internal/handlers"
	"personal-analytics-backend/internal/logger"
	"personal-analytics-backend/internal/redis"
	"personal-analytics-backend/internal/worker"
	"time"

	"github.com/joho/godotenv"
)

// Request comes in
//       ↓
// Handler validates (entries.go)
//       ↓
// Struct defines shape (CreateEntryRequest)
//       ↓
// DB function saves it (db.go - InsertEntry)
//       ↓
// Table schema stores it (db.go - createTables)

/*
	ResponseWrite is interface defines 2 function Write() and WriteHeader()
	w satisfies interfaces, you can write to it.(the output this is your connection
	back to user anything your write goes to their browser/client)
	r* http.Request is a pointer to struct containing all info about incoming request url,headers

	func healthHandler(w http.ResponseWriter, r *http.Request) {
		// F stands for file format allows you to print to any destination
		fmt.Fprintln(w, "ok")
	}
*/
func main() {
	// ========================================
	// STEP 0: Initialize structured logging FIRST
	// ========================================
	logger.InitLogger()

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		slog.Warn("No .env file found")
	}

	// Initialize database
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data.db"
	}
	err = db.InitDB(dbPath)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	err = redis.InitRedis("localhost:6379")
	if err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer redis.CloseRedis()

	// Start background worker pool (3 workers)
	worker.StartWorkerPool(3)

	// The "Defer" Magic: defer is a Go keyword that says: "Wait until this entire function (main) is finished, then immediately run this command."
	defer db.CloseDB()

	// As functions are values(First class citizens) you can pass a function just
	// like another function just like you pass int or string
	// we aren't calling healthHandler we are handling it to the http package
	// and saying execute this whenever someone visits /health

	// The Router htt.handleFunc tells go's default router "ServeMUX" the router
	// if a request comes in for the path run the function

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", handlers.RequestIDMiddleware(handlers.MetricsMiddleware(handlers.RateLimitMiddleware(handlers.LoggingMiddleware(handlers.HealthHandler)))))
	http.HandleFunc("/ping", handlers.RequestIDMiddleware(handlers.MetricsMiddleware(handlers.RateLimitMiddleware(handlers.LoggingMiddleware(handlers.PingHandler)))))

	// Auth endpoints (no protection needed)
	http.HandleFunc("/register", handlers.RequestIDMiddleware(handlers.MetricsMiddleware(handlers.RateLimitMiddleware(handlers.LoggingMiddleware(handlers.Register)))))
	http.HandleFunc("/login", handlers.RequestIDMiddleware(handlers.MetricsMiddleware(handlers.RateLimitMiddleware(handlers.LoggingMiddleware(handlers.Login)))))

	// Entries endpoints (PROTECTED - requires authentication)
	http.HandleFunc("/entries", handlers.RequestIDMiddleware(
		handlers.MetricsMiddleware(handlers.RateLimitMiddleware(handlers.LoggingMiddleware(
			handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					handlers.CreateEntry(w, r)
				} else if r.Method == http.MethodGet {
					handlers.GetEntries(w, r)
				} else if r.Method == http.MethodPatch {
					// PATCH /entries?id=5 - Update an entry
					handlers.UpdateEntry(w, r)
				} else if r.Method == http.MethodDelete {
					// Delete /entries?id=5 - Delete an entry
					handlers.DeleteEntry(w, r)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			}))))))

	http.HandleFunc("/metrics", handlers.RequestIDMiddleware(
		handlers.MetricsMiddleware(handlers.RateLimitMiddleware(
			handlers.LoggingMiddleware(handlers.GetMetrics)))))

	// ========================================
	// GRACEFUL SHUTDOWN IMPLEMENTATION
	// ========================================

	// STEP 1: Create HTTP server (instead of just ListenAndServe)
	// Why? So we can call server.Shutdown() later
	server := &http.Server{
		Addr: ":" + port,
	}

	// STEP 2: Start server in a GOROUTINE (background)
	// Why? So main() doesn't block here and can listen for Ctrl+C
	go func() {
		slog.Info("Server starting", "port", port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// STEP 3: Create a channel to receive OS signals
	// Channel = a pipe for communication between goroutines
	// Size 1 = can hold one signal before blocking
	quit := make(chan os.Signal, 1)

	// STEP 4: Tell Go: "When Ctrl+C (SIGINT) or kill (SIGTERM) happens, send it to 'quit' channel"
	signal.Notify(quit, os.Interrupt) // os.Interrupt = Ctrl+C

	// STEP 5: BLOCK HERE until a signal is received
	// This line waits forever until Ctrl+C is pressed
	<-quit
	slog.Warn("Shutdown signal received")

	// STEP 6: Create a timeout context (max 5 seconds to finish)
	// If requests take longer than 5 seconds, force close
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// STEP 7: Gracefully shutdown the server
	// - Stops accepting NEW requests
	// - Waits for current requests to finish (up to 5 seconds)
	slog.Info("Shutting down server", "timeout_seconds", 5)
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown error", "error", err)
	}

	// STEP 8: Close connections (defer will handle this, but log it)
	slog.Info("Closing connections", "redis", "closing", "database", "closing")
	slog.Info("Server stopped gracefully")
}

/*

The "Pointer vs. Interface" Secret
1. Why is r a Pointer? (*http.Request)
You said it's so we can change the value. While true, we rarely want to change the incoming request.
The real reason: An http.Request is a large struct. It contains headers, URL data, cookies, and more. If we didn't use a pointer, Go would have to copy that entire mountain of data every time a function is called. Using a pointer is fast and efficient.

2. Why is w NOT a Pointer? (http.ResponseWriter)
This is the trick question!

The secret: http.ResponseWriter is an Interface, not a struct.

In Go, we almost never pass a pointer to an interface. An interface is already a small "header" that internally points to the data.

When you see a type in Go that doesn't have a * but you can still call methods on it to change things (like writing a response), it's almost always an Interface.

*/
