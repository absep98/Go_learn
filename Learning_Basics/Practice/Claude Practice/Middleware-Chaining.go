package main

import (
	"fmt"
	"net/http"
)

func finalHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello!")
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("before")
		next.ServeHTTP(w, r)
		fmt.Println("after")
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "secret" {
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	})
}

func main() {
	final := http.HandlerFunc(finalHandler)
	chain := LoggingMiddleware(AuthMiddleware(final))

	http.Handle("/", chain)
	http.ListenAndServe(":8080", nil)
}

/*
Write LoggingMiddleware and AuthMiddleware (auth just checks if header Authorization equals
"secret", else responds 401), a dummy final handler that writes "Hello!", and wire them into
a chain in main using http.HandleFunc or http.ListenAndServe. Print something distinct in each
middleware's before/after so you can literally see the nesting order when you hit it.
*/
