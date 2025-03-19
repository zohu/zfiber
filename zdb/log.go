package zdb

import (
	"context"
	"errors"
	"fmt"
	"github.com/zohu/zfiber/zlog"
	"gorm.io/gorm"
	"time"

	gl "gorm.io/gorm/logger"
)

type GormLoggerOption struct {
	LogSlow                 time.Duration
	LogSkipCaller           bool
	LogIgnoreRecordNotFound bool
}

type GormLogger struct {
	ZLogger  *zlog.Logger
	LogLevel gl.LogLevel
	GormLoggerOption
}

func NewGormLogger(ops *GormLoggerOption) GormLogger {
	log := GormLogger{
		ZLogger:  zlog.ZLogger(),
		LogLevel: gl.Warn,
	}
	log.GormLoggerOption = *ops
	gl.Default = log
	return log
}

func (l GormLogger) LogMode(level gl.LogLevel) gl.Interface {
	return GormLogger{
		ZLogger:  l.ZLogger,
		LogLevel: level,
		GormLoggerOption: GormLoggerOption{
			LogSlow:                 l.LogSlow,
			LogSkipCaller:           l.LogSkipCaller,
			LogIgnoreRecordNotFound: l.LogIgnoreRecordNotFound,
		},
	}
}
func (l GormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	if l.LogLevel >= gl.Info {
		zlog.Infof(s, i...)
	}
}
func (l GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	if l.LogLevel >= gl.Warn {
		zlog.Warnf(s, i...)
	}
}
func (l GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	if l.LogLevel >= gl.Error {
		zlog.Errorf(s, i...)
	}
}
func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gl.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case l.LogLevel >= gl.Error && err != nil && (!l.LogIgnoreRecordNotFound || !errors.Is(err, gorm.ErrRecordNotFound)):
		sql, rows := fc()
		zlog.Errorf("err=%s rows=%d elapsed=%s sql=%s", err.Error(), rows, elapsed.String(), sql)
	case l.LogLevel >= gl.Warn && l.LogSlow != 0 && elapsed > l.LogSlow:
		sql, rows := fc()
		var e string
		if err != nil {
			e = fmt.Sprintf("err=%s ", err.Error())
		}
		zlog.Warnf("%selapsed=%s rows=%d sql=%s", e, elapsed.String(), rows, sql)
	case l.LogLevel >= gl.Info:
		sql, rows := fc()
		var e string
		if err != nil {
			e = fmt.Sprintf("err=%s ", err.Error())
		}
		zlog.Debugf("%selapsed=%s rows=%d sql=%s", e, elapsed.String(), rows, sql)
	}
}
