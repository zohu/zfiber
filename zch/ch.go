package zch

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/valkey-io/valkey-go"
	"github.com/zohu/zfiber/zlog"
	"time"
)

type Config struct {
	Expiration    string   `json:"expiration" yaml:"expiration" note:"过期时间"`
	CleanInterval string   `json:"clean_interval" yaml:"clean_interval" note:"清理间隔"`
	Addrs         []string `json:"addrs" yaml:"addrs" validate:"required" note:"地址"`
	Db            int      `json:"db" yaml:"db" note:"数据库"`
	Password      string   `json:"password" yaml:"password" note:"密码"`
}

func (c *Config) Validate() error {
	return validator.New().Struct(c)
}

type L2 struct {
	m *Memory
	v valkey.Client
}

var l2 *L2

type Options struct {
	*Config
	ValkeyOptions valkey.ClientOption
}

// NewL2
// @Description: 创建l2缓存
// @param expiration
// @param cleanInterval
// @param ops
// @return *L2
func NewL2(conf *Config, opts ...valkey.ClientOption) *L2 {
	if err := conf.Validate(); err != nil {
		zlog.Fatalf("validate l2 config failed: %v", err)
		return nil
	}
	ops := &Options{
		Config: conf,
		ValkeyOptions: valkey.ClientOption{
			InitAddress: conf.Addrs,
			Password:    conf.Password,
			SelectDB:    conf.Db,
		},
	}
	expire, err := time.ParseDuration(ops.Expiration)
	if err != nil {
		zlog.Fatalf("parse expiration error: %v", err)
	}
	if expire == 0 {
		expire = time.Hour
	}
	internal, err := time.ParseDuration(ops.CleanInterval)
	if err != nil {
		zlog.Fatalf("parse clean interval error: %v", err)
	}
	if internal == 0 {
		internal = time.Minute * 5
	}
	if l2 == nil {
		l2 = &L2{
			m: NewMemory(expire, internal),
			v: NewValkey(ops.ValkeyOptions),
		}
	}
	zlog.Infof("init zch success")
	return l2
}

func L() *L2 {
	if l2 == nil {
		zlog.Fatalf("Please call NewL2 before using L")
	}
	return l2
}
func M() *Memory {
	if l2 == nil {
		zlog.Fatalf("Please call NewL2 before using M")
	}
	return l2.m
}
func V() valkey.Client {
	if l2 == nil {
		zlog.Fatalf("Please call NewL2 before using V")
	}
	return l2.v
}

func (l *L2) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	if err := l.v.Do(ctx, l.v.B().Set().Key(key).Value(value).Ex(expiration).Build()).Error(); err == nil {
		l.m.Set(key, value, l1(expiration))
		return nil
	} else {
		return err
	}
}

func (l *L2) Get(ctx context.Context, key string) (interface{}, error) {
	if v, ok := l.m.Get(key); ok {
		return v, nil
	} else {
		if v, err := l.v.Do(ctx, l.v.B().Get().Key(key).Build()).ToString(); err == nil {

			exp, err := l.v.Do(ctx, l.v.B().Ttl().Key(key).Build()).ToInt64()
			if err == nil && exp > 0 {
				l.m.Set(key, v, l1(time.Second*time.Duration(exp)))
			}
			return v, nil
		} else {
			return nil, err
		}
	}
}

// Del
// @Description: 删除缓存
// @receiver l
// @param ctx
// @param key
// @return error
func (l *L2) Del(ctx context.Context, key string) error {
	l.m.Delete(key)
	return l.v.Do(ctx, l.v.B().Del().Key(key).Build()).Error()
}

// Flush
// @Description: 释放二级缓存
// @receiver l
// @param ctx
// @return error
func (l *L2) Flush(ctx context.Context) error {
	l.m.Flush()
	return nil
}

// l1
// @Description: 计算l1缓存的过期时间, l1总是比l2短一些, 且最长是30min，减少内存占用且防止NX虚锁
// @param expiration
// @return time.Duration
func l1(expiration time.Duration) time.Duration {
	if expiration >= 35*time.Minute {
		return 30 * time.Minute
	} else if expiration >= 15*time.Minute {
		return 10 * time.Minute
	} else if expiration >= 10*time.Minute {
		return 5 * time.Minute
	} else if expiration >= 5*time.Minute {
		return time.Minute
	}
	return expiration
}
