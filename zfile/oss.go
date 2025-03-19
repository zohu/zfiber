package zfile

import (
	"context"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"io"
)

type ossService struct {
	client *oss.Client
}

func newOssService() *ossService {
	cfg := oss.LoadDefaultConfig().
		WithRegion(config.Region).
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.AccessKey, config.AccessSecret))
	return &ossService{
		client: oss.NewClient(cfg),
	}
}

func (s *ossService) upload(ctx context.Context, r io.Reader, name string, progress Progress) (md5 string, err error) {
	req := oss.PutObjectRequest{
		Bucket: oss.Ptr(config.Bucket),
		Key:    oss.Ptr(name),
		Body:   r,
	}
	if progress != nil {
		req.ProgressFn = func(increment, transferred, total int64) {
			progress(increment, transferred, total)
		}
	}
	resp, err := s.client.PutObject(ctx, &req)
	if err != nil {
		return "", err
	}
	return oss.ToString(resp.ContentMD5), nil
}
func (s *ossService) delete(ctx context.Context, name string) (err error) {
	_, err = s.client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(config.Bucket),
		Key:    oss.Ptr(name),
	})
	return err
}
