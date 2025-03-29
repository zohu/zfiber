package zlog

import (
	"log/slog"
	"os"
)

/**
 * zlog 默认实例，可以直接使用zlog.XXX
 */

var zlog = NewZLogger(&Options{Level: slog.LevelDebug, SkipCallers: 1}, os.Stdout)

func WithOptions(ops *Options) {
	zlog = NewZLogger(ops, os.Stdout)
}

func Debugf(format string, args ...any) {
	zlog.Debugf(format, args...)
}
func Infof(format string, args ...any) {
	zlog.Infof(format, args...)
}
func Warnf(format string, args ...any) {
	zlog.Warnf(format, args...)
}
func Errorf(format string, args ...any) {
	zlog.Errorf(format, args...)
}
func Fatalf(format string, args ...any) {
	zlog.Fatalf(format, args...)
}
func Panicf(format string, args ...any) {
	zlog.Panicf(format, args...)
}
