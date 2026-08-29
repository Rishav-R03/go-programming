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
	//Check if file exist
	_, osStatusErr := os.Stat(path)
	fileExists := !os.IsNotExist(osStatusErr)

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if !fileExists {
		header := Header{Name: "Name", RollNo: "RollNo", Score: "Score"}
		if err := writer.Write([]string{header.Name, header.RollNo, header.Score}); err != nil {
			return fmt.Errorf("cannot write header: %w", err)
		}
	}

	record := model.Result{Name: name, RollNO: rollno, Score: score}
	if err := writer.Write([]string{record.Name, record.RollNO, strconv.Itoa(record.Score)}); err != nil {
		return fmt.Errorf("cannot write record: %w", err)
	}
	return nil
}
