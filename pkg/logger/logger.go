// pkg/logger/logger.go
package logger

import (
	"log"
	"os"
)

type Logger struct {
	info  *log.Logger
	warn  *log.Logger
	err   *log.Logger
	debug *log.Logger
}

func New(env string) *Logger {
	return &Logger{
		info:  log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime|log.Lshortfile),
		warn:  log.New(os.Stdout, "WARN ", log.Ldate|log.Ltime|log.Lshortfile),
		err:   log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile),
		debug: log.New(os.Stdout, "DEBUG ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *Logger) Info(msg string, keysAndValues ...any) {
	l.info.Printf(msg, keysAndValues...)
}

func (l *Logger) Warn(msg string, keysAndValues ...any) {
	l.warn.Printf(msg, keysAndValues...)
}

func (l *Logger) Error(msg string, keysAndValues ...any) {
	l.err.Printf(msg, keysAndValues...)
}

func (l *Logger) Debug(msg string, keysAndValues ...any) {
	l.debug.Printf(msg, keysAndValues...)
}

// NewTest: Logger dùng cho test (có thể disable output)
// func NewTest() *Logger {
// 	return New("test")
// }
