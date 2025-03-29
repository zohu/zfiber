package zlog

import "io"

type XLogger struct {
	logger *Logger
}

func NewXLogger(ops *Options, w ...io.Writer) *XLogger {
	return &XLogger{
		logger: NewZLogger(ops, w...),
	}
}

func (x *XLogger) Debug(args ...any) {
	x.logger.Debugf("%v", args...)
}
func (x *XLogger) Info(args ...any) {
	x.logger.Infof("%v", args...)
}
func (x *XLogger) Warn(args ...any) {
	x.logger.Warnf("%v", args...)
}
func (x *XLogger) Error(args ...any) {
	x.logger.Errorf("%v", args...)
}
func (x *XLogger) Debugf(fmt string, args ...any) {
	x.logger.Debugf(fmt, args...)
}
func (x *XLogger) Infof(fmt string, args ...any) {
	x.logger.Infof(fmt, args...)
}
func (x *XLogger) Warnf(fmt string, args ...any) {
	x.logger.Warnf(fmt, args...)
}
func (x *XLogger) Errorf(fmt string, args ...any) {
	x.logger.Errorf(fmt, args...)
}
