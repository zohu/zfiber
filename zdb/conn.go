package zdb

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"os"
	"time"
)

const DriverName = "pgx/v5"

func newDB(config Config, database string) (*gorm.DB, error) {
	db, err := gorm.Open(
		postgres.New(postgres.Config{
			DriverName: DriverName,
			DSN:        config.Dsn(database),
		}),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // 关闭复数表名
			},
			Logger: NewGormLogger(&GormLoggerOption{
				LogSlow:                 time.Second * time.Duration(config.LogSlow),
				LogSkipCaller:           config.LogSkipCaller == "yes",
				LogIgnoreRecordNotFound: config.LogIgnoreRecordNotFound == "yes",
			}),
		})
	if err != nil {
		return nil, fmt.Errorf("链接数据库失败 %s", err.Error())
	}
	d, _ := db.DB()
	d.SetMaxIdleConns(config.MaxIdle)
	d.SetMaxOpenConns(config.MaxAlive)
	d.SetConnMaxLifetime(config.MaxAliveLife)
	if config.Debug || os.Getenv("DEBUG") == "true" {
		db = db.Debug()
	}
	return db, nil
}
