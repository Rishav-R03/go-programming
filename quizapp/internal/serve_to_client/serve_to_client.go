package servetoclient

import (
	"fmt"
	"quizapp/internal/model"
	"time"
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
		fmt.Printf("\nQuestion %d: %s => ", questionNum, problem.Problem)
		answerChan := make(chan int)
		errorChan := make(chan error)

		timer := time.NewTimer(5 * time.Second)

		go func() {
			var input int
			_, err := fmt.Scanf("%d\n", &input)
			if err != nil {
				errorChan <- err
				return
			}
			answerChan <- input
		}()

		select {
		case <-timer.C:
			fmt.Println("\nTime's up! Question Skipped")
		case err := <-errorChan:
			timer.Stop()
			fmt.Printf("\nInvalid input! Question skipped. %v", err)
		case input := <-answerChan:
			timer.Stop()
			if VerifySolutionUpdateScore(problem.Solution, input) {
				fmt.Printf("Correct")
				score++
			} else {
				fmt.Printf("Incorrect! (correct answer: %d)\n", problem.Solution)
			}
		}
		questionNum++
	}
	return name, rollNo, score, nil
}

func VerifySolutionUpdateScore(solution int, input int) bool {
	return solution == input
}
