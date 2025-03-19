package zfile

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/zohu/zfiber/zdb"
	"github.com/zohu/zfiber/zlog"
	"github.com/zohu/zfiber/zutil"
	"io"
	"net/http"
	"path"
	"strings"
)

var config Config
var svr iService

func New(conf Config) {
	if err := validator.New().Struct(conf); err != nil {
		zlog.Fatalf("config error: %s", err)
		return
	}
	config = conf
	// 如果可能，同步数据库表
	if err := zdb.DB(context.TODO()).AutoMigrate(&ZfileRecord{}); err == nil {
		config.isPvMode = true
	}
	switch config.Provider {
	case ProviderTypeOss:
		svr = newOssService()
	default:
		zlog.Fatalf("unknown file provider type: %s", config.Provider)
	}
	zlog.Infof("init file provider success: %s", config.Provider)
}

type ReqUpload struct {
	Fid      string `json:"fid" validate:"required" note:"文件ID"`
	Path     string `json:"path" validate:"required" note:"文件存储路径"`
	Name     string `json:"name" validate:"required" note:"文件名"`
	IdleDays int64  `json:"idle_days" note:"最长闲置时间，0则永久"`
	Progress Progress
}

type RespUpload struct {
	Name string `json:"name" note:"文件名"`
	Url  string `json:"url" note:"文件地址"`
	Md5  string `json:"md5" note:"文件MD5"`
}

func Upload(ctx context.Context, h *ReqUpload, r io.Reader) (*RespUpload, error) {
	ext := path.Ext(h.Name)
	name := config.FullName(h.Path, h.Fid, ext)
	md5, err := svr.upload(ctx, r, name, h.Progress)
	if err != nil {
		return nil, err
	}

	if config.isPvMode {
		var exist ZfileRecord
		zdb.DB(ctx).Where("md5=?", md5).First(&exist)
		if exist.Name != "" {
			zlog.Debugf("file already exist in bucket, will cancel upload: %s", exist.Name)
			_ = svr.delete(ctx, name)
			return &RespUpload{
				Name: config.FullName(h.Path, exist.Fid, ext),
				Url:  config.HTTPDomain(exist.Name),
				Md5:  md5,
			}, nil
		}
		zdb.DB(ctx).Create(&ZfileRecord{
			Fid:    h.Fid,
			Md5:    md5,
			Bucket: config.Bucket,
			Name:   name,
			Expire: zutil.FirstTruth(h.IdleDays, config.IdleDays),
		})
	}

	return &RespUpload{
		Name: name,
		Url:  config.HTTPDomain(name),
		Md5:  md5,
	}, nil
}

func FiberForward(fc fiber.Ctx) error {
	arr := strings.Split(fc.Path(), "/")
	fid := strings.Split(arr[len(arr)-1], ".")[0]
	var ext ZfileRecord
	if err := zdb.DB(fc.Context()).Where("fid=?", fid).First(&ext).Error; err != nil {
		return fc.Status(404).SendString("404 Not Found")
	}
	ext.Pv += 1
	zdb.DB(fc.Context()).Updates(&ext)
	return fc.Redirect().Status(http.StatusFound).To(config.HTTPDomain(ext.Name))
}
