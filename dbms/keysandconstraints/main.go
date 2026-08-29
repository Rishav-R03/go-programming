package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib" // Import the pgx driver for PostgreSQL
)

type DepartmentEntity struct {
	ID   int
	Name string
}

type EmployeeEntity struct {
	ID           int
	Name         string
	DepartmentID int
	Email        string
	Age          int
}

type DepartmentRequest struct {
	Name string `json:"name"`
}

type EmployeeRequest struct {
	Name         string `json:"name"`
	DepartmentID int    `json:"department_id"`
	Email        string `json:"email"`
	Age          int    `json:"age"`
}

type DepartmentResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type EmployeeResponse struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	DepartmentID int    `json:"department_id"`
	Email        string `json:"email"`
	Age          int    `json:"age"`
}

// Handler accepts *sql.DB by returning an http.HandlerFunc closure
func EmployeeHandlerLayerInsert(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var employeeRequest EmployeeRequest
		err := json.NewDecoder(r.Body).Decode(&employeeRequest)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		employeeEntity := EmployeeEntity{
			Name:         employeeRequest.Name,
			DepartmentID: employeeRequest.DepartmentID,
			Email:        employeeRequest.Email,
			Age:          employeeRequest.Age,
		}

		// FIX: Execute the database insert via the Service Layer
		response, err := EmployeeServiceLayerInsert(r.Context(), db, &employeeEntity)
		if err != nil {
			log.Printf("Error inserting employee: %v", err)
			http.Error(w, "Failed to create employee", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
		}
	}
}

func EmployeeServiceLayerInsert(ctx context.Context, db *sql.DB, employee *EmployeeEntity) (*EmployeeResponse, error) {
	eID, err := EmployeeRepositoryLayerInsert(ctx, db, *employee)
	if err != nil {
		return nil, err
	}

	return &EmployeeResponse{
		ID:           eID, // Now populated with the returned DB ID!
		Name:         employee.Name,
		DepartmentID: employee.DepartmentID,
		Email:        employee.Email,
		Age:          employee.Age,
	}, nil
}

func EmployeeRepositoryLayerInsert(ctx context.Context, db *sql.DB, employee EmployeeEntity) (int, error) {
	query := "INSERT INTO employees (name, department_id, email, age) VALUES ($1, $2, $3, $4) RETURNING id"
	var id int
	err := db.QueryRowContext(ctx, query, employee.Name, employee.DepartmentID, employee.Email, employee.Age).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func main() {
	connStr := "postgres://myuser:mypassword@127.0.0.1:5432/mydb?sslmode=disable"
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Error connecting to db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	mux := http.NewServeMux()
	// Pass db to the handler function generator
	mux.HandleFunc("/employees", EmployeeHandlerLayerInsert(db))

	log.Println("Server running on :8000...")
	log.Fatal(http.ListenAndServe(":8000", mux))
}
