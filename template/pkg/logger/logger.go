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

func (l Logger) Debug(a ...interface{}) {
	if l.level <= Debug {
		fmt.Println(a...)
	}
}

func (l Logger) Debugf(format string, a ...interface{}) {
	if l.level <= Debug {
		fmt.Printf(format, a...)
	}
}

func (l Logger) Info(a ...interface{}) {
	if l.level <= Info {
		fmt.Println(a...)
	}
}

func (l Logger) Infof(format string, a ...interface{}) {
	if l.level <= Info {
		fmt.Printf(format, a...)
	}
}

func New(level Level) Logger {
	return Logger{
		level: level,
	}
}
