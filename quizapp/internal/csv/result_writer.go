package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"quizapp/internal/model"
	"strconv"
)

type Header struct {
	Name   string
	RollNo string
	Score  string
}

func WriteResult(path string, name string, rollno string, score int) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := Header{Name: "Name", RollNo: "RollNo", Score: "Score"}
	record := model.Result{Name: name, RollNO: rollno, Score: score}

	if err := writer.Write([]string{header.Name, header.RollNo, header.Score}); err != nil {
		return fmt.Errorf("cannot write: %w", err)
	}
	if err := writer.Write([]string{record.Name, record.RollNO, strconv.Itoa(record.Score)}); err != nil {
		return fmt.Errorf("cannot write: %w", err)
	}
	return nil
}
