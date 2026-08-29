package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type User struct {
	ID        int
	Name      string
	Email     string
	CreatedAt string
}

func main() {
	connStr := "postgres://myuser:mypassword@0.0.0.0:5432/mydb?sslmode=disable"
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Error connecting database %v\n", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}
	fmt.Println("Successfully connected to PostgreSQL via pgx!")

	rows, err := db.Query("SELECT id, name, email FROM users")
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer rows.Close()

	fmt.Println("Registed users:= ")
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			log.Fatalf("Row scan failed: %v\n", err)
		}
		fmt.Printf("ID: %d | Name: %s | Email: %s\n", u.ID, u.Name, u.Email)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("Row iteration error: %v\n", err)
	}
}

// Helper function to read and execute any .sql file at runtime
func runSQLFile(db *sql.DB, filepath string) error {
	query, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("could not read file %s: %w", filepath, err)
	}

	_, err = db.Exec(string(query))
	if err != nil {
		return fmt.Errorf("failed executing query from file %s: %w", filepath, err)
	}

	fmt.Printf("Executed SQL file successfully: %s\n", filepath)
	return nil
}
