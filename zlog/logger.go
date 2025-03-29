package zlog

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

type Logger struct {
	s *slog.Logger
}

func NewZLogger(ops *Options, w ...io.Writer) *Logger {
	return &Logger{slog.New(NewHandler(ops, w...))}
}
func (l *Logger) Debug(msg string, args ...any) {
	l.s.Debug(msg, args...)
}
func (l *Logger) Print(msg string, args ...any) {
	l.s.Info(msg, args...)
}
func (l *Logger) Info(msg string, args ...any) {
	l.s.Info(msg, args...)
}
func (l *Logger) Warn(msg string, args ...any) {
	l.s.Warn(msg, args...)
}
func (l *Logger) Error(msg string, args ...any) {
	l.s.Error(msg, args...)
}
func (l *Logger) Fatal(msg string, args ...any) {
	l.s.Error(msg, args...)
	os.Exit(1)
}
func (l *Logger) Panic(msg string, args ...any) {
	l.s.Error(msg, args...)
	panic(msg)
}

func (l *Logger) Debugf(format string, args ...any) {
	l.Debug(fmt.Sprintf(format, args...))
}
func (l *Logger) Printf(format string, args ...any) {
	l.Print(fmt.Sprintf(format, args...))
}
func (l *Logger) Infof(format string, args ...any) {
	l.Info(fmt.Sprintf(format, args...))
}
func (l *Logger) Warnf(format string, args ...any) {
	l.Warn(fmt.Sprintf(format, args...))
}
func (l *Logger) Errorf(format string, args ...any) {
	l.Error(fmt.Sprintf(format, args...))
}
func (l *Logger) Fatalf(format string, args ...any) {
	l.Fatal(fmt.Sprintf(format, args...))
}
func (l *Logger) Panicf(format string, args ...any) {
	l.Panic(fmt.Sprintf(format, args...))
}
