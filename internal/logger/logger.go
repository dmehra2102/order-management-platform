package logger

import (
	"encoding/json"
	"log"
	"time"
)

type Level string

const (
	Debug Level = "DEBUG"
	Info  Level = "INFO"
	Warn  Level = "WARN"
	Error Level = "ERROR"
)

type Logger struct {
	level Level
}

type entry struct {
	Level   string `json:"level"`
	Time    string `json:"time"`
	Message string `json:"message"`
	Fields  any    `json:"fields,omitempty"`
}

func New(level string) *Logger {
	return &Logger{level: Level(level)}
}

func (l *Logger) Debug(msg string, fields ...any) {
	if l.level == Debug {
		l.log(Debug, msg, fields...)
	}
}

func (l *Logger) Info(msg string, fields ...any) {
	l.log(Info, msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...any) {
	l.log(Warn, msg, fields...)
}

func (l *Logger) Error(msg string, fields ...any) {
	l.log(Error, msg, fields...)
}

func (l *Logger) log(level Level, msg string, fields ...any) {
	e := entry{
		Level:   string(level),
		Time:    time.Now().UTC().Format(time.RFC3339),
		Message: msg,
	}

	if len(fields) > 0 {
		e.Fields = fields
	}

	b, _ := json.Marshal(e)
	log.Println(string(b))
}
