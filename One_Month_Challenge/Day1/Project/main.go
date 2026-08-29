package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const requestIDKey contextKey = "requestID"

// Custom response writer to catch status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// middle wares

type MiddleWare func(http.Handler) http.Handler

// generates unique requestid without external dependencies
func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

//RequestID middleware injects x-request-id into context and response header

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = generateID()
		}
		w.Header().Set("X-Request-ID", reqID)

		ctx := context.WithValue(r.Context(), requestIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get Request ID from context
		reqID, _ := r.Context().Value(requestIDKey).(string)

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		log.Printf(
			"[req_id=%s] %s %s %d %s",
			reqID,
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			time.Since(start),
		)
	})
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized: Missing or invalid token", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token != "secret-token-123" {
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Chain(h http.Handler, middlewares ...MiddleWare) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func publicHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Public Endpoint: Access Granted\n"))
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	reqID, _ := r.Context().Value(requestIDKey).(string)
	fmt.Fprintf(w, "Dashboard: Welcome Authenticated User! (Request ID: %s)\n", reqID)
}

func main() {
	mux := http.NewServeMux()
	publicChain := Chain(
		http.HandlerFunc(publicHandler),
		RequestID,
		Logging,
	)
	protectedChain := Chain(
		http.HandlerFunc(dashboardHandler),
		RequestID,
		Logging,
		Auth,
	)

	mux.Handle("/health", publicChain)
	mux.Handle("/dashboard", protectedChain)
	log.Println("Server listening: http://localhost:8000")
	if err := http.ListenAndServe(":8000", mux); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
