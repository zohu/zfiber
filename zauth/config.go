package zauth

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/zohu/zfiber/zutil"
	"strings"
	"time"
)

type Config struct {
	Prefix          string        `json:"prefix" yaml:"prefix" note:"前缀"`
	AuthAge         time.Duration `json:"auth_age" yaml:"auth_age" note:"过期时间"`
	MultipleCoexist bool          `json:"multiple_coexist" yaml:"multiple_coexist" note:"是否允许多个设备同时登录"`
	AllowIpChange   bool          `json:"allow_ip_change" yaml:"allow_ip_change" note:"是否允许ip变化"`
	AllowUaChange   bool          `json:"allow_ua_change" yaml:"allow_ua_change" note:"是否允许ua变化"`
	WhiteList       []string      `json:"white_list" yaml:"white_list" note:"白名单"`
}

func (c *Config) Validate() error {
	c.Prefix = zutil.FirstTruth(c.Prefix, "auth")
	c.AuthAge = zutil.FirstTruth(c.AuthAge, time.Hour*2)
	for i, p := range c.WhiteList {
		c.WhiteList[i] = strings.TrimSpace(strings.TrimPrefix(p, "/"))
	}
	return validator.New().Struct(c)
}
func (c *Config) IsWhite(path string) bool {
	for _, w := range c.WhiteList {
		if strings.HasPrefix(w, strings.TrimPrefix(path, "/")) {
			return true
		}
	}
	return false
}
func (c *Config) key(k string) string {
	return fmt.Sprintf("%s:%s", strings.TrimSuffix(c.Prefix, ":"), k)
}
