// Package logger defines a simple API for logging
package logger

import (
	"fmt"
	"os"
)

// Level defines the amount of logging to output
type Level int

const (
	// Debug means the system will output more technical messages
	Debug Level = iota

	// Info means the system will output default messages
	Info

	// Error means the system will output error messages
	Error
)

// Logger exposes logging functions
type Logger struct {
	level Level
}

// Debug logs content if the log level is more or equal to Debug
func (l Logger) Debug(args ...interface{}) {
	if l.levelIsEnabled(Debug) {
		out := make([]interface{}, 0)
		out = append(out, "[DEBUG]")
		out = append(out, args...)

		_, _ = fmt.Println(out...)
	}
}

// Debugf logs content if the log level is more or equal to Debug and allows for formatting
func (l Logger) Debugf(format string, args ...interface{}) {
	if l.levelIsEnabled(Debug) {
		_, _ = fmt.Printf("[DEBUG] "+format, args...)
	}
}

// Info logs content if the log level is more or equal to Info
func (l Logger) Info(args ...interface{}) {
	if l.levelIsEnabled(Info) {
		_, _ = fmt.Println(args...)
	}
}

// Infof logs content if the log level is more or equal to Info and allows for formatting
func (l Logger) Infof(format string, args ...interface{}) {
	if l.levelIsEnabled(Info) {
		_, _ = fmt.Printf(format, args...)
	}
}

// Error logs content if the log level is more or equal to Error
func (l Logger) Error(args ...interface{}) {
	if l.levelIsEnabled(Error) {
		_, _ = fmt.Fprintln(os.Stderr, args...)
	}
}

// Errorf logs content if the log level is more or equal to Error and allows for formatting
func (l Logger) Errorf(format string, args ...interface{}) {
	if l.levelIsEnabled(Error) {
		_, _ = fmt.Fprintf(os.Stderr, format, args...)
	}
}

func (l Logger) levelIsEnabled(level Level) bool {
	return l.level <= level
}

// New returns an initialized Logger
func New(level Level) Logger {
	return Logger{
		level: level,
	}
}
