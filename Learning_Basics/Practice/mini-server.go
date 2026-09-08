package main

import (
	"net/http"
	"encoding/json"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// 1. check method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2 set header
	w.Header().Set("Content-Type", "application/json")

	// 3 write response
	json.NewEncoder(w).Encode(map[string]string{"message": "hello world"})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status" : "ok"})
}


func main() {


	http.HandleFunc("/hello", helloHandler)

	http.HandleFunc("/health", healthHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	})

	http.ListenAndServe(":8081", nil)

}