package servetoclient

import (
	"fmt"
	"quizapp/internal/model"
)

func ConductQuiz(probChan <-chan model.MathProblem) (int, error) {
	score := 0
	questionNum := 1

	for problem := range probChan {
		fmt.Printf("\nQuestion: %d: %s => ", questionNum, problem.Problem)
		var input int
		_, err := fmt.Scanf("%d\n", &input)
		if err != nil {
			fmt.Println("Invalid input. Skipped question")
			questionNum++
			continue
		}

		if VerifySolutionUpdateScore(problem.Solution, input) {
			fmt.Println("Correct.")
			score++
		} else {
			fmt.Printf("Incorrect! (correct answer: %d)", problem.Solution)
		}
		questionNum++
	}
	return score, nil
}

func VerifySolutionUpdateScore(solution int, input int) bool {
	if solution == input {
		return true
	}
	return false
}
