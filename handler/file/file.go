package file

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	v2rayAccessLog  = "v2ray_access.log"
	v2rayErrorLog   = "v2ray_error.log"
	generalStdLog   = "std.log"
	generalErrorLog = "error.log"
)

var fileNameSlice = [4]string{
	v2rayAccessLog,
	v2rayErrorLog,
	generalStdLog,
	generalErrorLog,
}

type Handler struct {
}

func New() *Handler {
	return &Handler{}
}

func (c *Handler) getFileNameByType(typ int) (string, error) {
	if typ <= 0 || typ > 3 {
		return "", fmt.Errorf("no such file type %v", typ)
	}
	return fileNameSlice[typ-1], nil
}

func (c *Handler) Read(typ, from, to int) (string, error) {
	fileName, err := c.getFileNameByType(typ)
	if err != nil {
		return "", err
	}
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
