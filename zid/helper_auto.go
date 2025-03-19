package zid

import (
	"context"
	"github.com/valkey-io/valkey-go"
	"github.com/zohu/zfiber/zlog"
	"time"
)

func AutoWorkerId(c valkey.Client, ops *Options) {
	if ops == nil {
		ops = &Options{}
	}
	ops.Validate()
	ctx := context.TODO()
	ops.WorkerId = findIdx(ctx, c, ops, 0)

	singletonMutex.Lock()
	idGenerator = NewDefaultIdGenerator(ops)
	singletonMutex.Unlock()

	go alive(ctx, c, ops.prefix(ops.WorkerId))
	zlog.Infof("init zid success, workerid=%d", ops.WorkerId)
}

func findIdx(ctx context.Context, c valkey.Client, ops *Options, retry uint16) uint16 {
	if retry > ops.maxWorkerIdNumber() {
		zlog.Fatalf("all worker id [0-%d] are occupied, please extend WorkerIdBitLength", retry-1)
	}
	if ok, _ := c.Do(ctx, c.B().Set().Key(ops.prefix(retry)).Value("occupied").Nx().ExSeconds(60).Build()).AsBool(); ok {
		return retry
	}
	zlog.Warnf("worker id [%d] is occupied, try next", retry)
	return findIdx(ctx, c, ops, retry+1)
}
func alive(ctx context.Context, c valkey.Client, prefix string) {
	for range time.NewTicker(time.Second * 40).C {
		c.Do(ctx, c.B().Set().Key(prefix).Value("occupied").ExSeconds(60).Build())
	}
}
