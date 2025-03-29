package zdb

import (
	"context"
	"errors"
	"fmt"
	"github.com/zohu/zfiber/zlog"
	"gorm.io/gorm"
	"log/slog"
	"time"

	gl "gorm.io/gorm/logger"
)

type GormLoggerOption struct {
	LogSlow                 time.Duration
	LogIgnoreRecordNotFound bool
	Debug                   bool
}

type GormLogger struct {
	logger *zlog.Logger
	GormLoggerOption
}

func NewGormLogger(ops *GormLoggerOption) *GormLogger {
	zlogOptions := zlog.Options{}
	if ops.Debug {
		zlogOptions.Level = slog.LevelDebug
	}
	log := new(GormLogger)
	log.logger = zlog.NewZLogger(&zlogOptions)
	log.GormLoggerOption = *ops
	gl.Default = log
	return log
}

func (l GormLogger) LogMode(level gl.LogLevel) gl.Interface {
	return GormLogger{
		logger: l.logger,
		GormLoggerOption: GormLoggerOption{
			LogSlow:                 l.LogSlow,
			LogIgnoreRecordNotFound: l.LogIgnoreRecordNotFound,
		},
	}
}
func (l GormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	l.logger.Infof(s, i...)
}
func (l GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	l.logger.Warnf(s, i...)
}
func (l GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	l.logger.Errorf(s, i...)
}
func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	switch {
	case err != nil && (!l.LogIgnoreRecordNotFound || !errors.Is(err, gorm.ErrRecordNotFound)):
		sql, rows := fc()
		l.logger.Errorf("rows=%d elapsed=%6.3fs err=%s sql=%s", rows, elapsed.Seconds(), err.Error(), sql)
	case l.LogSlow != 0 && elapsed > l.LogSlow:
		sql, rows := fc()
		var e string
		if err != nil {
			e = fmt.Sprintf("err=%s ", err.Error())
		}
		l.logger.Warnf("rows=%d elapsed=%6.3fs %ssql=%s", rows, elapsed.Seconds(), e, sql)
	default:
		sql, rows := fc()
		var e string
		if err != nil {
			e = fmt.Sprintf("err=%s ", err.Error())
		}
		l.logger.Debugf("rows=%d elapsed=%6.3fs %ssql=%s", rows, elapsed.Seconds(), e, sql)
	}
}
