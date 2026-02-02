package logger

import (
	"log/slog"
	"os"
)

// ========================================
// STRUCTURED LOGGING SETUP
// ========================================

// Log is our global structured logger
// All parts of the app will use this instead of log.Printf
var Log *slog.Logger

// InitLogger sets up structured logging for the application
// Call this once at startup in main.go
func InitLogger() {
	// Create a JSON handler that writes to stdout
	// JSONHandler = outputs logs as JSON (for production/log aggregation)
	// TextHandler = outputs human-readable logs (for local development)
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Show all log levels (Debug, Info, Warn, Error)
	})

	// Create the logger with the handler
	Log = slog.New(handler)

	// Also set as the default logger (so slog.Info() works globally)
	slog.SetDefault(Log)
}

// ========================================
// WHY STRUCTURED LOGGING?
// ========================================
//
// OLD WAY (log.Printf):
//   log.Printf("User %d created entry %d", userID, entryID)
//   Output: 2026/02/02 10:15:23 User 3 created entry 13
//
//   Problem: How do you search for "all logs for user 3"?
//   You'd have to regex through strings. Painful.
//
// NEW WAY (slog):
//   slog.Info("Entry created", "user_id", userID, "entry_id", entryID)
//   Output: {"time":"...","level":"INFO","msg":"Entry created","user_id":3,"entry_id":13}
//
//   Now you can:
//   - Filter: jq 'select(.user_id == 3)'
//   - Search: Datadog/ELK query: user_id:3
//   - Aggregate: Count entries per user
//
// ========================================
// LOG LEVELS
// ========================================
//
// slog.Debug("...")  - Detailed debugging info (usually off in production)
// slog.Info("...")   - Normal operations (request received, entry created)
// slog.Warn("...")   - Something unexpected but not broken
// slog.Error("...")  - Something broke, needs attention
//
// ========================================
// KEY-VALUE PAIRS
// ========================================
//
// Always pass pairs: "key", value, "key", value
//
// slog.Info("Request processed",
//     "method", "POST",
//     "path", "/entries",
//     "duration_ms", 45,
//     "user_id", 3,
// )
//
// Output:
// {"time":"...","level":"INFO","msg":"Request processed","method":"POST","path":"/entries","duration_ms":45,"user_id":3}
