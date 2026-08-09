package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"quizapp/internal/model"
	"strconv"
	"strings"
)

func ReadMathProblemFromFile(path string) ([]model.MathProblem, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to open csv file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	if len(header) < 2 || strings.TrimSpace(header[0]) != "problem" || strings.TrimSpace(header[1]) != "solution" {
		log.Println("Warning: CSV header differ from expected ['problem', 'solution']")
	}

	var problems []model.MathProblem

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading csv row: %w", err)
		}
		if len(record) < 2 {
			continue
		}

		solutionVal, err := strconv.Atoi(strings.TrimSpace(record[1]))

		if err != nil {
			return nil, fmt.Errorf("invalid integer value for solution in row: %v: %w", record, err)
		}

		item := model.MathProblem{
			Problem:  strings.TrimSpace(record[0]),
			Solution: solutionVal,
		}

		problems = append(problems, item)
	}
	return problems, nil
}
