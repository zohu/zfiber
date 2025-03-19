package zfiber

import (
	"testing"
)

func TestValidate(t *testing.T) {
	_ = Trans()
	_ = Validator()
	v := NewFiberValidator()
	type Args struct {
		F1 any `validate:"required" json:"f_1"`
		F2 any `validate:"required" json:"f_2" note:"NOTE"`
		F3 any `validate:"required" json:"f_3" gorm:"comment:GORM"`
		F4 any `validate:"required" json:"f_4" message:"MESSAGE"`
		F5 any `validate:"datetime" json:"f_5"`
		F6 any `validate:"datetime=RFC3339" json:"f_6"`
		F7 any `validate:"regular=.*" json:"f_7"`
	}
	args := &Args{
		F5: "2023-01-01 00:00:00",
		F6: "2023-01-01",
		F7: "123",
	}
	errs := ErrParameter.WithValidateErrs(args, v.Validate(args))
	t.Logf("%+v", errs)
}
