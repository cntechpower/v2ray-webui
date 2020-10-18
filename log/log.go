package log

import (
	"fmt"
	_log "log"
	"os"
	"path"
	"runtime"
	"time"
)

var globalLogger *_log.Logger

func InitLogger(fileName string) {
	if globalLogger != nil {
		panic("Logger already init")
	}
	globalLogger = &_log.Logger{}
	if fileName != "" {
		// You could set this to any `io.Writer` such as a file
		file, err := os.OpenFile("anywhere.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			globalLogger.SetOutput(file)
		} else {
			globalLogger.SetOutput(os.Stderr)
			globalLogger.Printf("Failed to log to file, using default stderr: %v\n", err)
		}
	} else {
		globalLogger.SetOutput(os.Stderr)
		//globalLogger.Println("log to default stderr output")
	}
}

type Level string

const (
	info  Level = "INFO"
	warn  Level = "WARN"
	error Level = "ERROR"
	fatal Level = "FATAL"
)

func getCaller(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip)
	if ok {
		return path.Base(file), line
	}
	return "unknown.go", 0
}

func log(skip int, h *Header, level Level, format string, a ...interface{}) {
	file, line := getCaller(skip)

	globalLogger.Println(fmt.Sprintf("[%s] <%s> |%s| (%s:%v) %s", time.Now().Format("2006-01-02 15:04:05.000"), level, h, file, line, fmt.Sprintf(format, a...)))
}

type Header struct {
	name string
}

func NewHeader(n string) *Header {
	return &Header{name: n}
}

func (h *Header) String() string {
	return h.name
}

func Infof(h *Header, format string, a ...interface{}) {
	log(3, h, info, format, a...)
}

func Errorf(h *Header, format string, a ...interface{}) {
	log(3, h, error, format, a...)
}

func Warnf(h *Header, format string, a ...interface{}) {
	log(3, h, warn, format, a...)
}

func Fatalf(h *Header, format string, a ...interface{}) {
	log(3, h, error, format, a...)
}
