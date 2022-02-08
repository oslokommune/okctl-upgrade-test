package logger

import (
	"fmt"
	"os"
)

type Level int

const (
	// Debug means the system will output more technical messages
	Debug Level = iota

	// Info means the system will output default messages
	Info

	// Error means the system will output error messages
	Error
)

type Logger struct {
	level Level
}

func (l Logger) Debug(args ...interface{}) {
	if l.levelIsEnabled(Debug) {
		out := make([]interface{}, 0)
		out = append(out, "[DEBUG]")
		out = append(out, args...)

		_, _ = fmt.Println(out...)
	}
}

func (l Logger) Debugf(format string, args ...interface{}) {
	if l.levelIsEnabled(Debug) {
		_, _ = fmt.Printf("[DEBUG] "+format, args...)
	}
}

func (l Logger) Info(args ...interface{}) {
	if l.levelIsEnabled(Info) {
		_, _ = fmt.Println(args...)
	}
}

func (l Logger) Infof(format string, args ...interface{}) {
	if l.levelIsEnabled(Info) {
		_, _ = fmt.Printf(format, args...)
	}
}

func (l Logger) Error(args ...interface{}) {
	if l.levelIsEnabled(Error) {
		_, _ = fmt.Fprintln(os.Stderr, args...)
	}
}

func (l Logger) Errorf(format string, args ...interface{}) {
	if l.levelIsEnabled(Error) {
		_, _ = fmt.Fprintf(os.Stderr, format, args...)
	}
}

func (l Logger) levelIsEnabled(level Level) bool {
	return l.level <= level
}

func New(level Level) Logger {
	return Logger{
		level: level,
	}
}
