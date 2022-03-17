// Package logger contains functionality for logging to console
package logger

import (
	"fmt"
	"os"
)

// Level is the current log level (Debug, Info, etc)
type Level int

const (
	// Debug means the system will output more technical messages
	Debug Level = iota

	// Info means the system will output default messages
	Info

	// Error means the system will output error messages
	Error
)

// Logger provides logging functionality
type Logger struct {
	level Level
}

// Debug prints to console if debug level is enabled
func (l Logger) Debug(args ...interface{}) {
	if l.levelIsEnabled(Debug) {
		out := make([]interface{}, 0)
		out = append(out, "[DEBUG]")
		out = append(out, args...)

		_, _ = fmt.Println(out...)
	}
}

// Debugf prints to console if debug level is enabled
func (l Logger) Debugf(format string, args ...interface{}) {
	if l.levelIsEnabled(Debug) {
		_, _ = fmt.Printf("[DEBUG] "+format, args...)
	}
}

// Info prints to console if info level is enabled
func (l Logger) Info(args ...interface{}) {
	if l.levelIsEnabled(Info) {
		_, _ = fmt.Println(args...)
	}
}

// Infof prints to console if info level is enabled
func (l Logger) Infof(format string, args ...interface{}) {
	if l.levelIsEnabled(Info) {
		_, _ = fmt.Printf(format, args...)
	}
}

// Error prints to console if error level is enabled
func (l Logger) Error(args ...interface{}) {
	if l.levelIsEnabled(Error) {
		_, _ = fmt.Fprintln(os.Stderr, args...)
	}
}

// Errorf prints to console if error level is enabled
func (l Logger) Errorf(format string, args ...interface{}) {
	if l.levelIsEnabled(Error) {
		_, _ = fmt.Fprintf(os.Stderr, format, args...)
	}
}

func (l Logger) levelIsEnabled(level Level) bool {
	return l.level <= level
}

// New returns a Logger
func New(level Level) Logger {
	return Logger{
		level: level,
	}
}
