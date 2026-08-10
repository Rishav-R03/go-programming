package servetoclient

import (
	"fmt"
	"quizapp/internal/model"
)

func ConductQuiz(probChan <-chan model.MathProblem) (string, string, int, error) {
	score := 0
	questionNum := 1
	fmt.Println("Enter your name")
	var name string
	fmt.Scanf("%s\n", &name)
	fmt.Println("Enter you rollno")
	var rollNo string
	fmt.Scanf("%s\n", &rollNo)

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
	return name, rollNo, score, nil
}

func VerifySolutionUpdateScore(solution int, input int) bool {
	return solution == input
}
