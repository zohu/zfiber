package zfiber

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/zohu/zfiber/zutil"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

type Empty struct{}

type RespBean struct {
	Code    int               `json:"code,omitempty" xml:"code,omitempty"`
	Data    any               `json:"data,omitempty" xml:"data,omitempty"`
	Message string            `json:"message,omitempty" xml:"message,omitempty"`
	Notes   map[string]string `json:"notes,omitempty" xml:"notes,omitempty"`
}
type RespListBean[T any] struct {
	Page  int   `json:"page" xml:"page" note:"页码"`
	Size  int   `json:"size" xml:"size" note:"每页数量"`
	List  []T   `json:"list" xml:"list"`
	Total int64 `json:"total" xml:"total"`
}

func NewResp(code int, message string, data any, notes map[string]string) RespBean {
	return RespBean{
		Code:    zutil.FirstTruth(code, 1),
		Data:    data,
		Message: zutil.FirstTruth(message, "success"),
		Notes:   notes,
	}
}
func NewData(data any) RespBean {
	return NewResp(1, "success", data, nil)
}
func NewDataList[T any](data *RespListBean[T]) RespBean {
	return NewData(data)
}
func Abort(c fiber.Ctx, resp RespBean) error {
	return AbortHttpCode(c, fiber.StatusOK, resp)
}
func AbortString(c fiber.Ctx, str string) error {
	return c.Status(fiber.StatusOK).SendString(str)
}

func AbortHttpCode(c fiber.Ctx, code int, resp RespBean) error {
	if strings.Contains(c.Get("Content-Type"), "application/xml") {
		return c.Status(code).XML(resp)
	}
	return c.Status(code).JSON(resp)
}

// translateErrors
// @Description: 翻译错误信息
// @param h
// @param errs
// @return map[string]string
func translateErrors(h any, errs validator.ValidationErrors) map[string]string {
	ets := make(map[string]string)
	elem := reflect.TypeOf(h)
	if elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}
	for _, e := range errs {
		field, _ := elem.FieldByName(e.StructField())
		key := strings.Split(field.Tag.Get("json"), ",")[0]
		if msg := field.Tag.Get("message"); msg != "" {
			ets[key] = msg
		} else {
			ets[key] = strings.ReplaceAll(e.Translate(Trans()), e.StructField(), fieldNameZh(field))
		}
	}
	return ets
}

// fieldNameZh
// @Description: 查找字段中文名
// @param field
// @return v
func fieldNameZh(field reflect.StructField) (v string) {
	if v = field.Tag.Get("note"); v != "" {
		return v
	}
	gormTag := field.Tag.Get("gorm")
	if gormTag != "" {
		arr := strings.Split(gormTag, ";")
		for _, tag := range arr {
			if strings.HasPrefix(tag, "comment:") {
				v = strings.TrimPrefix(tag, "comment:")
				v = strings.ReplaceAll(v, " ", "")
				return v
			}
		}
	}
	return field.Name
}

type Pages struct {
	Page int `json:"page" xml:"page" note:"页码"`
	Size int `json:"size" xml:"size" note:"每页数量"`
}

func (p *Pages) PageSizes() (int, int) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Size <= 0 {
		p.Size = 50
	}
	if p.Size > 1000 {
		p.Size = 1000
	}
	return p.Page, zutil.FirstTruth(p.Size, 50)
}
func (p *Pages) ScopePage(db *gorm.DB) *gorm.DB {
	page, size := p.PageSizes()
	return db.Offset((page - 1) * size).Limit(size)
}
