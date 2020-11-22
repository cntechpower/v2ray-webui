package handler

import (
	"bufio"
	"os"
	"strings"
)

type FileHandler struct {
}

func NewFileHandler() *FileHandler {
	return &FileHandler{}
}

func (c *FileHandler) ReadFile(fileName string, from, to int) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()
	res := strings.Builder{}
	scanner := bufio.NewScanner(f)
	currentLine := 1
	for scanner.Scan() {
		if from <= currentLine && currentLine <= to {
			res.Write(scanner.Bytes())
			res.WriteString("\n")
		}
		if currentLine >= to {
			break
		}
		currentLine++
	}
	return res.String(), nil

}
