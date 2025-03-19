package zsms

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/zohu/zfiber/zid"
	"github.com/zohu/zfiber/zlog"
	"github.com/zohu/zfiber/zutil"
	"slices"
	"sync"
)

type ProviderType int

const (
	ProviderTypeAliyun ProviderType = iota + 1
)

type Config struct {
	Provider     ProviderType `json:"provider" yaml:"provider" validate:"required" note:"短信服务商"`
	AccessKey    string       `json:"access_key" yaml:"access_key" validate:"required" note:"AccessKey"`
	AccessSecret string       `json:"access_secret" yaml:"access_secret" validate:"required" note:"AccessSecret"`
	Endpoint     string       `json:"endpoint" yaml:"endpoint" note:"接入点"`
	ProviderId   string       `json:"provider_id" yaml:"providerId" note:"短信服务商ID，如果为空则自动生成，生成的ID是临时的，重启后重置"`
}

func (c *Config) Validate() error {
	return validator.New().Struct(c)
}

/**
 * 短信服务
 */

type ReqSend struct {
	Phone         string            `json:"phone" validate:"required" note:"多个以逗号分隔，上限1000个"`
	SignName      string            `json:"sign_name" validate:"required" note:"签名"`
	Template      string            `json:"template" validate:"required" note:"模板ID"`
	TemplateParam map[string]string `json:"template_param" note:"模板参数"`
	ProviderId    string            `json:"provider_id" note:"可指定短信服务商ID"`
}
type ResSend struct {
	BizId string `json:"biz_id" note:"回执ID"`
}
type IService interface {
	Send(context.Context, *ReqSend) (*ResSend, error)
}

var services sync.Map
var providerIds []string

func AddProvider(conf *Config) error {
	if err := validator.New().Struct(conf); err != nil {
		return err
	}
	if conf.ProviderId == "" {
		conf.ProviderId = zid.NextIdShort()
	}
	switch conf.Provider {
	case ProviderTypeAliyun:
		service, err := newAliyunService(conf)
		if err != nil {
			return err
		}
		services.Store(conf.ProviderId, service)
		providerIds = append(providerIds, conf.ProviderId)
	default:
		return fmt.Errorf("unknown sms provider type: %d", conf.Provider)
	}
	zlog.Infof("add sms provider success: %s", conf.ProviderId)
	return nil
}
func AddProviders(cfs []*Config) {
	for _, conf := range cfs {
		if err := AddProvider(conf); err != nil {
			zlog.Errorf("add sms provider failed: %v", err)
		}
	}
}

func Send(ctx context.Context, h *ReqSend) (*ResSend, error) {
	if len(providerIds) == 0 {
		return nil, fmt.Errorf("no provider")
	}

	// 指定服务商或者默认第一个
	pid := zutil.FirstTruth(h.ProviderId, providerIds[0])
	resp, err := sendRetry(ctx, pid, h)
	if err == nil {
		return resp, nil
	}
	zlog.Warnf("send sms failed: %v", err)

	// 如果发送失败，用其他服务商重试
	excludes := []string{pid}
	for _, id := range providerIds {
		// 过滤已经用过的
		if slices.Contains(excludes, id) {
			continue
		}
		// 重新发送
		if resp, err = sendRetry(ctx, id, h); err == nil {
			return resp, nil
		}
		zlog.Warnf("send sms failed: %v", err)
		excludes = append(excludes, id)
	}
	return nil, fmt.Errorf("all provider send failed")
}

func sendRetry(ctx context.Context, pid string, h *ReqSend) (*ResSend, error) {
	svr, ok := services.Load(pid)
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", pid)
	}
	return svr.(IService).Send(ctx, h)
}
