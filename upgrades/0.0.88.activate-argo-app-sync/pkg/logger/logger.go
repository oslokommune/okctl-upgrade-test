package logger

import (
	"fmt"
	"io"
	"os"
)

type Level int

const (
	// Debug means the system will output debug messages and all levels below
	Debug Level = iota
	// Info means the system will output default messages and all levels below
	Info
	// Error means the system will output error and all levels below
	Error
)

type Logger struct {
	level Level
	Out   io.Writer
	Err   io.Writer
}

func (l Logger) Debug(args ...interface{}) {
	if !l.levelIsEnabled(Debug) {
		return
	}

	out := make([]interface{}, 0)
	out = append(out, "[DEBUG]")
	out = append(out, args...)

	_, _ = fmt.Fprintln(l.Out, out...)
}

func (l Logger) Debugf(format string, args ...interface{}) {
	if !l.levelIsEnabled(Debug) {
		return
	}

	_, _ = fmt.Fprintf(l.Out, "[DEBUG] "+format, args...)
}

func (l Logger) Info(args ...interface{}) {
	if !l.levelIsEnabled(Info) {
		return
	}

	_, _ = fmt.Fprintln(l.Out, args...)
}

func (l Logger) Infof(format string, args ...interface{}) {
	if !l.levelIsEnabled(Info) {
		return
	}

	_, _ = fmt.Fprintf(l.Out, format, args...)
}

func (l Logger) Error(args ...interface{}) {
	if !l.levelIsEnabled(Error) {
		return
	}

	out := make([]interface{}, 0)
	out = append(out, "[ERROR]")
	out = append(out, args...)

	_, _ = fmt.Fprintln(l.Err, out...)
}

func (l Logger) Errorf(format string, args ...interface{}) {
	if !l.levelIsEnabled(Error) {
		return
	}

	_, _ = fmt.Fprintf(l.Err, "[ERROR] "+format, args...)
}

func (l Logger) levelIsEnabled(level Level) bool {
	return l.level <= level
}

func New(level Level) Logger {
	return NewWithWriters(level, os.Stdout, os.Stderr)
}

func NewWithWriters(level Level, out io.Writer, err io.Writer) Logger {
	return Logger{
		level: level,
		Out:   out,
		Err:   err,
	}
}
