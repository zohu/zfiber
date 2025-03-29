package zfiber

import (
	"encoding/json"
	"errors"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/joho/godotenv"
	"github.com/zohu/zfiber/zants"
	"github.com/zohu/zfiber/zch"
	"github.com/zohu/zfiber/zdb"
	"github.com/zohu/zfiber/zid"
	"github.com/zohu/zfiber/zlog"
	"github.com/zohu/zfiber/zutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const LoggerFormat = "${respHeader:X-Request-ID} ${method} ${status} ${path} ${ip} ${queryParams} ${body} -> ${resBody} ${latency}\n"

type Options interface {
	ZantsOptions() *zants.Config
	ZchOptions() *zch.Config
	ZdbOptions() (*zdb.Config, []any)
	CorsOptions() cors.Config
	ServerOptions() *Config
}

// Config
// @Description: fiber服务配置
type Config struct {
	Addr       string      `json:"addr,omitempty" yaml:"addr,omitempty"`
	Domain     string      `json:"domain,omitempty" yaml:"domain,omitempty"`
	Middleware *Middleware `json:"middleware,omitempty" yaml:"middleware,omitempty"`
}
type Middleware struct {
	LoggerIgnore []string `json:"logger_ignore,omitempty" yaml:"logger_ignore,omitempty"`
}

type App struct {
	app       *fiber.App
	shutdowns []func()
	addr      string
}

func init() {
	// 加载环境变量
	_ = godotenv.Load()
}

func NewApp(ops Options, cfs ...fiber.Config) *App {
	if ops == nil {
		zlog.Fatalf("options is nil")
		return nil
	}
	// 初始化zants
	if ants := ops.ZantsOptions(); ants != nil {
		zants.New(ants)
	}
	// 初始化zch
	if ch := ops.ZchOptions(); ch != nil {
		zch.NewL2(ch)
		// 初始化ID生成器
		zid.AutoWorkerId(zch.V(), nil)
	}
	// 初始化zdb
	if db, dts := ops.ZdbOptions(); db != nil {
		zdb.New(db, dts...)
	}
	// 服务配置
	svrConf := ops.ServerOptions()
	if svrConf == nil {
		svrConf = new(Config)
	}

	// fiber
	conf := fiber.Config{}
	if len(cfs) > 0 {
		conf = cfs[0]
	}
	// 设置默认值
	conf.BodyLimit = zutil.FirstTruth(conf.BodyLimit, fiber.DefaultBodyLimit)
	conf.Concurrency = zutil.FirstTruth(conf.Concurrency, fiber.DefaultConcurrency)
	conf.JSONEncoder = sonic.Marshal
	conf.JSONDecoder = sonic.Unmarshal
	conf.StructValidator = NewFiberValidator()
	conf.ErrorHandler = errorHandler
	conf.StreamRequestBody = true

	app := fiber.New(conf)
	// 异常捕获
	app.Use(recoverer.New())
	// 跨域
	app.Use(cors.New(ops.CorsOptions()))
	// 请求ID
	app.Use(requestid.New(requestid.Config{Generator: zid.NextIdShort}))
	// 日志中间件
	app.Use(logger.New(loggerConfig(svrConf.Middleware)))

	return &App{app: app, addr: zutil.FirstTruth(svrConf.Addr, ":3000")}
}
func (s *App) Use(args ...any) *App {
	s.app.Use(args...)
	return s
}
func (s *App) Register(fn func(*fiber.App)) *App {
	fn(s.app)
	return s
}
func (s *App) Shutdown(shutdown func()) *App {
	s.shutdowns = append(s.shutdowns, shutdown)
	return s
}
func (s *App) Listen(config ...fiber.ListenConfig) {
	// 默认路由
	s.app.Get("health", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// 启动服务
	go func() {
		if len(config) == 0 {
			config = []fiber.ListenConfig{{DisableStartupMessage: true}}
		}
		_ = s.app.Listen(s.addr, config[0])
	}()
	zlog.Infof(">> listening on %s", s.addr)

	// 等待中断信号关闭服务器, 设置一个60秒的超时
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	zlog.Infof("serve shutdowning...")
	for _, shutdown := range s.shutdowns {
		shutdown()
	}
	if err := s.app.Shutdown(); err != nil {
		zlog.Fatalf("serve shutdown failed: %v", err)
	}
	zlog.Infof("serve shutdowned")
}
func errorHandler(c fiber.Ctx, err error) error {
	zlog.Warnf("%s %s %s: %v", requestid.FromContext(c), c.Method(), c.Path(), err)

	code := fiber.StatusInternalServerError
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
		switch code {
		case fiber.StatusUnauthorized:
			return AbortHttpCode(c, code, NewFlag(code, "登录态失效"))
		case fiber.StatusForbidden:
			return AbortHttpCode(c, code, NewFlag(code, "权限不足"))
		case fiber.StatusNotFound:
			return AbortHttpCode(c, code, NewFlag(code, "未知路径"))
		case fiber.StatusMethodNotAllowed:
			return AbortHttpCode(c, code, NewFlag(code, "拒绝访问"))
		default:
			return AbortHttpCode(c, e.Code, NewFlag(code, err.Error()))
		}
	}
	return AbortHttpCode(c, code, ErrNil)
}

func loggerConfig(conf *Middleware) logger.Config {
	if conf == nil {
		conf = new(Middleware)
	}
	return logger.Config{
		Format: LoggerFormat,
		Output: zlog.SafeWriter(),
		CustomTags: map[string]logger.LogFunc{
			logger.TagBody: func(output logger.Buffer, c fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				b := c.Body()
				if len(b) > 0 && b[0] == 123 && b[len(b)-1] == 125 {
					dst := zutil.NewByteBuff()
					defer zutil.ReleaseByteBuff(dst)
					_ = json.Compact(dst, b)
					b = dst.Bytes()
				}
				if len(b) > 2048 {
					_, _ = output.Write(b[:2048])
					return output.Write([]byte("..."))
				}
				return output.Write(b)
			},
		},
		Next: func(c fiber.Ctx) bool {
			for _, v := range conf.LoggerIgnore {
				if !strings.HasPrefix(v, "/") {
					v = "/" + v
				}
				if strings.HasPrefix(c.Path(), v) {
					return true
				}
			}
			return false
		},
	}
}
