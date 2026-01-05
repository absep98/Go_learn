package main

import (
	"log"
	"net/http"
	"os"
	"personal-analytics-backend/internal/db"
	"personal-analytics-backend/internal/handlers"

	"github.com/joho/godotenv"
)

// ResponseWrite is interface defines 2 function Write() and WriteHeader()
// w satisfies interfaces, you can write to it.(the output this is your connection
// back to user anything your write goes to their browser/client)
// r* http.Request is a pointer to struct containing all info about incoming request url,headers

/*
	func healthHandler(w http.ResponseWriter, r *http.Request) {
		// F stands for file format allows you to print to any destination
		fmt.Fprintln(w, "ok")
	}
*/
func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	// Initialize database
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data.db"
	}
	err = db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
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

	http.HandleFunc("/health", handlers.HealthHandler)
	http.HandleFunc("/ping", handlers.PingHandler)
	http.HandleFunc("/entries", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.CreateEntry(w, r)
		} else if r.Method == http.MethodGet {
			handlers.GetEntries(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Printf("Server starting on port %s", port)
	http.ListenAndServe(":"+port, nil)
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
