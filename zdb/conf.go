package zdb

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/zohu/zfiber/zutil"
	"time"
)

type Config struct {
	Host                    string        `json:"host" yaml:"host" validate:"required" note:"数据库地址"`
	Port                    string        `json:"port" yaml:"port" validate:"required" note:"数据库端口"`
	User                    string        `json:"user" yaml:"user" validate:"required" note:"数据库用户"`
	Password                string        `json:"password" yaml:"password" validate:"required" note:"数据库密码"`
	Db                      string        `json:"db" yaml:"db" validate:"required" note:"数据库名"`
	Config                  string        `json:"config" yaml:"config" note:"数据库配置"`
	MaxIdle                 int           `json:"max_idle" yaml:"max_idle" note:"最大闲置连接数"`
	MaxAlive                int           `json:"max_alive" yaml:"max_alive" note:"最大存活连接数"`
	MaxAliveLife            time.Duration `json:"max_alive_life" yaml:"max_alive_life" note:"最大存活时间"`
	LogSlow                 int           `json:"log_slow" yaml:"log_slow" note:"慢阈值，秒"`
	LogSkipCaller           string        `json:"log_skip_caller" yaml:"log_skip_caller" note:"堆栈跳过层数,yes/no"`
	LogIgnoreRecordNotFound string        `json:"log_ignore_record_not_found" yaml:"log_ignore_record_not_found" note:"忽略无记录错误,yes/no"`
	Debug                   bool          `json:"debug" yaml:"debug" note:"是否开启debug日志"`
	Extension               []string      `json:"extension" yaml:"extension" note:"扩展配置"`
}

func (c *Config) Validate() error {
	c.Config = zutil.FirstTruth(c.Config, "sslmode=disable TimeZone=Asia/Shanghai")
	c.MaxIdle = zutil.FirstTruth(c.MaxIdle, 10)
	c.MaxAlive = zutil.FirstTruth(c.MaxAlive, 100)
	c.MaxAliveLife = zutil.FirstTruth(c.MaxAliveLife, time.Hour)
	c.LogSlow = zutil.FirstTruth(c.LogSlow, 5)
	c.LogSkipCaller = zutil.FirstTruth(c.LogSkipCaller, "yes")
	c.LogIgnoreRecordNotFound = zutil.FirstTruth(c.LogIgnoreRecordNotFound, "yes")
	return validator.New().Struct(c)
}
func (c *Config) Dsn(database string) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s %s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		database,
		c.Config,
	)
}
