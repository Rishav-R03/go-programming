package serveproblems

import (
	"fmt"
	"quizapp/internal/model"
)

func ServeProblemsToChannel(problems []model.MathProblem) (<-chan model.MathProblem, error) {
	if len(problems) == 0 {
		return nil, fmt.Errorf("problems slice is empty")
	}

	ch := make(chan model.MathProblem)

	go func() {
		defer close(ch)
		for _, p := range problems {
			ch <- p
		}
	}()

	return ch, nil
}
