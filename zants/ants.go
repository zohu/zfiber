package zants

import (
	"github.com/panjf2000/ants/v2"
	"github.com/zohu/zfiber/zlog"
	"github.com/zohu/zfiber/zutil"
)

type Config struct {
	MultiSize int32 `json:"multi_size" yaml:"multi_size" note:"线程池数量"`
	PoolSize  int32 `json:"pool_size" yaml:"pool_size" note:"每池线程数量"`
}
type PoolStatus struct {
	Cap     int32 `json:"cap" note:"容量"`
	Running int32 `json:"running" note:"运行中"`
	Waiting int32 `json:"waiting" note:"等待中"`
	Idle    int32 `json:"idle" note:"空闲中"`
}

var multiPool *ants.MultiPool

func New(conf *Config) {
	size := zutil.FirstTruth(int(conf.MultiSize), 1)
	preSize := zutil.FirstTruth(int(conf.PoolSize), 10)
	p, err := ants.NewMultiPool(size, preSize, ants.LeastTasks, ants.WithLogger(zlog.NewZLogger(nil)))
	if err != nil {
		zlog.Fatalf("new multi pool error: %v", err)
	}
	multiPool = p
	zlog.Infof("init zants success, size=%dx%d", size, preSize)
}

// Submit
// @Description: 提交一个函数到池中执行
// @param fn
func Submit(fn func()) {
	if err := multiPool.Submit(fn); err != nil {
		zlog.Errorf("submit fn error: %v", err)
	}
}

// Status
// @Description: 获取池状态
// @return *PoolStatus
func Status() *PoolStatus {
	return &PoolStatus{
		Cap:     int32(multiPool.Cap()),
		Running: int32(multiPool.Running()),
		Waiting: int32(multiPool.Waiting()),
		Idle:    int32(multiPool.Free()),
	}
}

// Tune
// @Description: 调整每个池大小
// @param size
func Tune(size int) {
	multiPool.Tune(size)
}
