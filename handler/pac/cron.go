package pac

import (
	"fmt"

	"github.com/cntechpower/utils/log"
)

type cronLogger struct {
	h *log.Header
}

func newCronLogger() *cronLogger {
	return &cronLogger{h: log.NewHeader("cron")}
}

func (l *cronLogger) Info(msg string, keysAndValues ...interface{}) {
	log.Infof(l.h, msg, keysAndValues...)
}

func (l *cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Errorf(l.h, fmt.Sprintf("got error %v, msg: %v", err, msg), keysAndValues...)
}
