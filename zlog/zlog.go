package zlog

import (
	"fmt"
	"log/slog"
	"os"
)

var zlog = slog.New(NewHandler(os.Stdout, &Options{
	AddSource: true,
	Level:     slog.LevelDebug,
}))

func Debug(msg string, args ...any) {
	zlog.Debug(msg, args...)
}
func Info(msg string, args ...any) {
	zlog.Info(msg, args...)
}
func Warn(msg string, args ...any) {
	zlog.Warn(msg, args...)
}
func Error(msg string, args ...any) {
	zlog.Error(msg, args...)
}
func Fatal(msg string, args ...any) {
	zlog.Error(msg, args...)
	os.Exit(1)
}
func Panic(msg string, args ...any) {
	zlog.Error(msg, args...)
	panic(msg)
}
func Debugf(format string, args ...any) {
	zlog.Debug(fmt.Sprintf(format, args...))
}
func Infof(format string, args ...any) {
	zlog.Info(fmt.Sprintf(format, args...))
}
func Warnf(format string, args ...any) {
	zlog.Warn(fmt.Sprintf(format, args...))
}
func Errorf(format string, args ...any) {
	zlog.Error(fmt.Sprintf(format, args...))
}
func Fatalf(format string, args ...any) {
	zlog.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}
func Panicf(format string, args ...any) {
	zlog.Error(fmt.Sprintf(format, args...))
	panic(fmt.Sprintf(format, args...))
}

//func Stack(skip int) string {
//	pcs := make([]uintptr, 16)
//	n := runtime.Callers(skip, pcs)
//	frames := runtime.CallersFrames(pcs[:n])
//	var str strings.Builder
//	for {
//		frame, more := frames.Next()
//		str.WriteString(frame.Function)
//		str.WriteByte('\n')
//		str.WriteByte('\t')
//		str.WriteString(frame.File)
//		str.WriteByte(':')
//		str.WriteString(strconv.Itoa(frame.Line))
//		str.WriteByte('\n')
//		if !more {
//			break
//		}
//	}
//	return str.String()
//}
