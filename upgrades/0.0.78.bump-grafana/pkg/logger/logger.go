package logger

import "fmt"

type Level int

const (
	// Debug means the system will output more technical messages
	Debug Level = iota

	// Info means the system will output default messages
	Info
)

type Logger struct {
	level Level
}

func (l Logger) Debug(args ...interface{}) {
	if l.levelIsEnabled(Debug) {
		out := make([]interface{}, 0)
		out = append(out, "[DEBUG]")
		out = append(out, args...)

		fmt.Println(out...)
	}
}

func (l Logger) Debugf(format string, args ...interface{}) {
	if l.levelIsEnabled(Debug) {
		fmt.Printf(format, args...)
	}
}

func (l Logger) Info(args ...interface{}) {
	if l.levelIsEnabled(Info) {
		fmt.Println(args...)
	}
}

func (l Logger) Infof(format string, args ...interface{}) {
	if l.levelIsEnabled(Info) {
		fmt.Printf(format, args...)
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
