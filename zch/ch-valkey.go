package zch

import (
	"github.com/valkey-io/valkey-go"
	"github.com/zohu/zfiber/zlog"
)

func NewValkey(conf valkey.ClientOption) valkey.Client {
	client, err := valkey.NewClient(conf)
	if err != nil {
		zlog.Fatalf("valkey client failed: %v", err)
	}
	return client
}
