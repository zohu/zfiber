package zauth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/valkey-io/valkey-go"
	"github.com/zohu/zfiber"
	"github.com/zohu/zfiber/zcpt"
	"github.com/zohu/zfiber/zid"
	"github.com/zohu/zfiber/zlog"
	"github.com/zohu/zfiber/zutil"
	"strings"
	"time"
)

type Authorization[T any] struct {
	Session string `json:"session"`
	Value   T      `json:"value"`
}

const (
	UserAgent = "User-Agent"
	AESKey    = "315c2wd6vpc7q4hx"
	TokenKey  = "tk:"

	LocalsUserKey    = "user"
	LocalsSessionKey = "session"
)

var conf = &Config{}
var vk valkey.Client

func New[T any](client valkey.Client, ops *Config) fiber.Handler {
	if client == nil {
		zlog.Fatalf("valkey is nil")
		return nil
	}
	vk = client
	if ops != nil {
		conf = ops
	}
	if err := conf.Validate(); err != nil {
		zlog.Fatalf("validate auth config failed: %v", err)
		return nil
	}
	return func(c fiber.Ctx) error {
		defer func() {
			if err := recover(); err != nil {
				zlog.Errorf("auth panic: %v", err)
				_ = zfiber.Abort(c, zfiber.ErrInvalidToken)
			}
		}()
		// 过滤白名单
		if conf.IsWhite(c.Path()) {
			return c.Next()
		}
		// 校验登录态
		token := zutil.FirstTruth(c.Cookies("auth"), c.Get("Authorization"), c.Query("auth"))
		if strings.TrimSpace(token) == "" {
			return zfiber.Abort(c, zfiber.ErrInvalidToken)
		}
		d, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			return zfiber.Abort(c, zfiber.ErrInvalidToken)
		}
		d, err = zcpt.AesDecryptCBC(d, []byte(AESKey))
		if err != nil {
			return zfiber.Abort(c, zfiber.ErrInvalidToken)
		}
		tks := strings.Split(string(d), "##")
		if len(tks) != 5 {
			return zfiber.Abort(c, zfiber.ErrInvalidToken)
		}
		// 校验ua
		if !conf.AllowUaChange && tks[1] != zcpt.Md5(c.Get(UserAgent)) {
			return zfiber.Abort(c, zfiber.ErrInvalidToken)
		}
		// 校验ip
		if !conf.AllowIpChange && tks[2] != c.IP() {
			return zfiber.Abort(c, zfiber.ErrInvalidToken)
		}

		// 提取用户数据
		uid := tks[3]
		vKey := conf.key(TokenKey + uid)
		userStr, err := vk.Do(c.Context(), vk.B().Get().Key(vKey).Build()).ToString()
		if err != nil {
			return zfiber.Abort(c, zfiber.ErrInvalidToken)
		}
		var auth Authorization[T]
		if err = sonic.UnmarshalString(userStr, &auth); err != nil {
			return zfiber.Abort(c, zfiber.ErrInvalidToken)
		}
		// 是否允许多地同时登陆
		if !conf.MultipleCoexist && auth.Session != zcpt.Md5(token) {
			return zfiber.Abort(c, zfiber.ErrInvalidSession)
		}

		// 存储用户数据
		c.Locals(LocalsUserKey, zutil.Ptr(auth.Value))
		c.Locals(LocalsSessionKey, auth.Session)

		// 刷新Token有效期
		c.Cookie(&fiber.Cookie{
			Expires: time.Now().Add(conf.AuthAge),
			MaxAge:  int(conf.AuthAge.Seconds()),
			Name:    "auth",
			Value:   token,
		})
		vk.Do(c.Context(), vk.B().Expire().Key(vKey).Seconds(int64(conf.AuthAge.Seconds())).Build())
		return c.Next()
	}
}

func Login[T any](c fiber.Ctx, uid string, value T) zfiber.RespBean {
	vKey := conf.key(TokenKey + uid)
	// 判断是否可以多处登录
	if !conf.MultipleCoexist {
		vk.Do(c.Context(), vk.B().Del().Key(vKey).Build())
	}
	// 生成登录态
	tk := fmt.Sprintf("%s##%s##%s##%s##%d", zid.NextIdShort(), zcpt.Md5(c.Get(UserAgent)), c.IP(), uid, time.Now().Unix())
	d, _ := zcpt.AesEncryptCBC([]byte(tk), []byte(AESKey))
	token := base64.StdEncoding.EncodeToString(d)
	c.Cookie(&fiber.Cookie{
		Expires: time.Now().Add(conf.AuthAge),
		MaxAge:  int(conf.AuthAge.Seconds()),
		Name:    "auth",
		Value:   token,
	})
	userStr, _ := sonic.MarshalString(&Authorization[T]{Session: zcpt.Md5(token), Value: value})
	vk.Do(c.Context(), vk.B().Set().Key(vKey).Value(userStr).Ex(conf.AuthAge).Build())
	return zfiber.NewData(map[string]string{
		"token":  token,
		"expire": time.Now().Add(conf.AuthAge).Format(time.RFC3339),
	})
}

func UpdateAuth[T any](c fiber.Ctx, uid string, value T) {
	session := c.Locals(LocalsSessionKey).(string)
	userStr, _ := sonic.MarshalString(&Authorization[T]{Session: session, Value: value})
	vKey := conf.key(TokenKey + uid)
	vk.Do(c.Context(), vk.B().Set().Key(vKey).Value(userStr).Ex(conf.AuthAge).Build())
}

func Auth[T any](c fiber.Ctx) (*T, error) {
	if u := c.Locals(LocalsUserKey); u != nil {
		return u.(*T), nil
	}
	return nil, errors.New("auth info is nil")
}
