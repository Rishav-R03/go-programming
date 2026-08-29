package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	user, err := simulateDBCall(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "Request timed out", http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			http.Error(w, "Client canceled  the request", http.StatusRequestTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"user":"%s}`, user)
}

func simulateDBCall(c context.Context) (string, error) {
	fmt.Println("[DB] Starting 3 sec query execution")
	select {
	case <-time.After(3 * time.Second):
		fmt.Println("[DB] Query completed successfully")
		return "John Doe", nil
	case <-c.Done():
		fmt.Printf("[DB] Query aborted early! Reason: %v\n", c.Err())
		return "", c.Err()
	}
}

func main() {
	http.HandleFunc("/users", getUserHandler)
	fmt.Println("Server starting on http://localhost:8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
