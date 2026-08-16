package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

/*

Middleware in Go is simply a function that takes an http.Handler and returns a new http.Handler.

By wrapping handlers inside handlers, you form a execution chain where requests flow inwards through the middleware stack and responses flow back out.

*/

// Ex

/*

func MiddlewareName(next http.Handler) http.Handler {

return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 1. Code executed BEFORE the next handler (e.g., check auth, start timer)



next.ServeHTTP(w, r) // Pass execution to the next handler in the chain



// 2. Code executed AFTER the next handler (e.g., log status code, clean up)

})

}

*/

//---------------------------------------------------------------------------------
// Middleware Setup
//---------------------------------------------------------------------------------

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrappedWriter, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, wrappedWriter.statusCode, time.Since(start))
	})
}

type MiddleWare func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...MiddleWare) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// Handler
func finalHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, authenticated user"))
}

func main() {
	mux := http.NewServeMux()

	// 1. Basic Routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World")
	})

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to user's page")
	})

	// 2. Chained Route
	protectedHandler := Chain(
		http.HandlerFunc(finalHandler),
		Logging,
		// Auth,
		// Recovery,
	)
	mux.Handle("/dashboard", protectedHandler)

	// 3. Start Server (This blocks execution, so put it at the very end)
	fmt.Println("Server running on :8000...")
	if err := http.ListenAndServe(":8000", mux); err != nil {
		fmt.Println("server error:", err)
	}
}
