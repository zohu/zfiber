package zdb

import (
	"context"
	"fmt"
	"github.com/zohu/zfiber/zlog"
	"github.com/zohu/zfiber/zutil"
	"gorm.io/gorm"
	"sync"

	_ "github.com/lib/pq"
)

var conn sync.Map
var conf *Config

func New(c *Config, dst ...any) {
	conf = c
	if err := conf.Validate(); err != nil {
		zlog.Fatalf("validate db config failed: %v", err)
		return
	}
	db := DB(context.Background(), "")
	if db == nil {
		zlog.Fatalf("init db failed")
		return
	}
	// 初始化扩展
	for _, ext := range conf.Extension {
		if err := db.Exec(fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s;", ext)).Error; err != nil {
			zlog.Fatalf("create extension [%s] failed: %v", ext, err)
			return
		}
		zlog.Infof("create extension [%s] success", ext)
	}
	// 初始化库表
	if dst != nil && len(dst) > 0 {
		if err := db.AutoMigrate(dst...); err != nil {
			zlog.Fatalf("init db table failed: %v", err)
			return
		}
		zlog.Infof("init db table success: %d", len(dst))
	}
}

func DB(ctx context.Context, args ...string) *gorm.DB {
	if len(args) == 0 {
		args = append(args, "")
	}
	def := zutil.FirstTruth(args[0], conf.Db)
	if v, ok := conn.Load(def); ok {
		return v.(*gorm.DB).WithContext(ctx)
	}
	db, err := newDB(*conf, def)
	if err != nil {
		zlog.Fatalf("init db conn failed %v", err)
		return nil
	}
	conn.Store(def, db)
	return db.WithContext(ctx)
}
