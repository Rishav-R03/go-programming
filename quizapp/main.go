package main

import (
	"fmt"
	"log"
	"quizapp/internal/csv"
	serveproblems "quizapp/internal/serve_problems"
	servetoclient "quizapp/internal/serve_to_client"
)

func main() {
	fmt.Println("Welcome to our quiz application!")
	filePath := "internal/csv/math_problems.csv"

	problems, err := csv.ReadMathProblemFromFile(filePath)
	if err != nil {
		log.Fatalf("error reading CSV: %v", err)
	}

	probChan, err := serveproblems.ServeProblemsToChannel(problems)
	if err != nil {
		log.Fatalf("error serving problems: %v", err)
	}

	// Interactively serve problems and take user inputs
	score, err := servetoclient.ConductQuiz(probChan)
	if err != nil {
		log.Fatalf("error running quiz: %v", err)
	}

	fmt.Printf("\nQuiz completed! Final Score: %d / %d\n", score, len(problems))
}
