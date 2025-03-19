package zsms

import (
	"context"
	"fmt"
	"github.com/alibabacloud-go/tea/tea"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
)

type aliyunService struct {
	client *dysmsapi20170525.Client
}

func newAliyunService(conf *Config) (IService, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(conf.AccessKey),
		AccessKeySecret: tea.String(conf.AccessSecret),
		Endpoint:        tea.String(conf.Endpoint),
	}
	client, err := dysmsapi20170525.NewClient(config)
	if err != nil {
		return &aliyunService{}, fmt.Errorf("new aliyun client failed: %v", err)
	}
	return &aliyunService{client: client}, nil
}

func (s *aliyunService) Send(ctx context.Context, h *ReqSend) (*ResSend, error) {
	res, err := s.client.SendSms(&dysmsapi20170525.SendSmsRequest{
		PhoneNumbers:  tea.String(h.Phone),
		SignName:      tea.String(h.SignName),
		TemplateCode:  tea.String(h.Template),
		TemplateParam: util.ToJSONString(h.TemplateParam),
	})
	if err != nil {
		return nil, err
	}
	code := res.Body.Code
	if !tea.BoolValue(util.EqualString(code, tea.String("OK"))) {
		return nil, fmt.Errorf(tea.StringValue(res.Body.Message))
	}
	return &ResSend{BizId: tea.StringValue(res.Body.BizId)}, nil
}
