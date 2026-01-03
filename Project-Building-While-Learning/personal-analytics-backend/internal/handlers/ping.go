package handlers

import (
	"fmt"
	"net/http"
	"os"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	msg := os.Getenv("PING_MESSAGE")
	if msg == "" {
		msg = "pong"
	}
	fmt.Printf("PingHandler called, msg: %s\n", msg)
	fmt.Fprintf(w, "Ping response: %s", msg)
}
