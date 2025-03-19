package zutil

import (
	"github.com/bytedance/sonic"
	"reflect"
	"strconv"
)

func FirstTruth[T any](args ...T) T {
	for _, item := range args {
		if !reflect.ValueOf(item).IsZero() {
			return item
		}
	}
	return args[0]
}

// AnyToStruct
// @Description: any 转 struct
// @param src
// @param dst
// @return error
func AnyToStruct[T any](src any, dst *T) error {
	b, err := sonic.Marshal(src)
	if err != nil {
		return err
	}
	return sonic.Unmarshal(b, dst)
}

// Clean
// @Description: 清空对象
// @param v
func Clean[T any](v T) {
	p := reflect.ValueOf(v).Elem()
	p.Set(reflect.Zero(p.Type()))
}

func Ptr[T any](in T) *T {
	return &in
}
func Val[T any](in *T) T {
	if in == nil {
		return *new(T)
	}
	return *in
}

func MustInt64(in any) int64 {
	t := reflect.TypeOf(in)
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(in).Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(reflect.ValueOf(in).Uint())
	case reflect.Float32, reflect.Float64:
		return int64(reflect.ValueOf(in).Float())
	case reflect.String:
		i, _ := strconv.ParseInt(reflect.ValueOf(in).String(), 10, 64)
		return i
	default:
		return 0
	}
}

func When[T any](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}
