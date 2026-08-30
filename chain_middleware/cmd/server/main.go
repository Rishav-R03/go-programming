package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

type MiddleWare func(http.Handler) http.Handler

// Writing middleware
//1. CORS

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// 2. Recovery MiddleWare
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// 3. Logging
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		log.Printf("[%s] %s %d -%v", r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
	})
}

// Create helper

func Chain(h http.Handler, middlewares ...MiddleWare) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func WithMiddleWare(h http.HandlerFunc, middlewares ...MiddleWare) http.Handler {
	var handler http.Handler = h
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// Auth
/**
	RBAC requires a two-step process:

    Authentication: Authenticate the user and attach their \
	user ID/role to the request's context.Context.
    Authorization (RBAC): Read the role from context.
	Context and verify permissions.
*/

type ctxKey string

const UserRoleKey ctxKey = "user_role"

// Authenticates requests and stores roles in context
func AuthMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}

		userRole := "admin" // in production parse this from header or token
		ctx := context.WithValue(r.Context(), UserRoleKey, userRole)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RBAC takes required roles and returns a Middleware function
func RBAC(reqRole string) MiddleWare {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(UserRoleKey).(string)
			if !ok || role != reqRole {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

/*
Summary TableConceptWhat It IsWhy It Matterstype ctxKey stringCustom Go type aliasPrevents key collision
with string keys from other packages
UserRoleKeyConstant identifier of type ctxKeyActs as the unique lookup key inside the
context.Contextcontext.WithValueThread-safe context builderClones the request context to
safely store request-scoped dataval.(string)
Type assertionSafely converts Go's generic interface{} back to a string
*/

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello world")
	})

	mux.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("something went horribly wrong!")
	})

	mux.Handle("/dashboard", WithMiddleWare(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Welcome to User Dashboard")
		},
		AuthMiddleWare,
	))

	// applyed middlewares globally
	handler := Chain(
		mux,
		Logging,
		Recovery,
		CORS,
	)

	// per endpoint

	// mux.Handle("/admin", WithMiddleWare(
	// 	adminHandler,
	// 	Logging,
	// 	RBAC("admin"),
	// ))

	fmt.Println("Server running: http://localhost:8000")
	if err := http.ListenAndServe(":8000", handler); err != nil {
		log.Fatal(err)
	}
}
