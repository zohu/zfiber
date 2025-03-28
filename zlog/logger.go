package zlog

import (
	"fmt"
	"log/slog"
)

type Logger struct {
	*slog.Logger
}

func ZLogger() *Logger {
	return &Logger{zlog}
}
func (l *Logger) Printf(format string, args ...any) {
	Infof(format, args...)
}
func (l *Logger) Debugf(format string, args ...any) {
	Debugf(format, args...)
}
func (l *Logger) Infof(format string, args ...any) {
	Infof(format, args...)
}
func (l *Logger) Warnf(format string, args ...any) {
	Warnf(format, args...)
}
func (l *Logger) Errorf(format string, args ...any) {
	Errorf(format, args...)
}
func (l *Logger) Fatalf(format string, args ...any) {
	Fatalf(format, args...)
}
func (l *Logger) Panicf(format string, args ...any) {
	Panicf(format, args...)
}

func (l *Logger) Error(err error, msg string, keysAndValues ...interface{}) {
	Error(fmt.Sprintf("%s %v", msg, err), keysAndValues)
}
