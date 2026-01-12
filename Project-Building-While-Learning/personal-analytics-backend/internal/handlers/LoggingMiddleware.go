package handlers

import (
	"log"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current_time := time.Now().Format("2006-01-02 15:04:05")
		request_type := r.Method
		request_path := r.URL.Path

		log.Printf("%s %s %s", current_time, request_type, request_path)

		next(w, r)
	}
}
