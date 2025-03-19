package zmath

import "fmt"

type Chain struct {
	Num float64
}

func NewChain(num float64) *Chain {
	return &Chain{Num: num}
}

func (c *Chain) Add(args ...float64) *Chain {
	for _, v := range args {
		c.Num = Add(c.Num, v)
	}
	return c
}
func (c *Chain) Sub(args ...float64) *Chain {
	for _, v := range args {
		c.Num = Sub(c.Num, v)
	}
	return c
}
func (c *Chain) Mul(args ...float64) *Chain {
	for _, v := range args {
		c.Num = Mul(c.Num, v)
	}
	return c
}
func (c *Chain) Div(args ...float64) *Chain {
	for _, v := range args {
		c.Num = Div(c.Num, v)
	}
	return c
}
func (c *Chain) Round(round int32) float64 {
	return Round(c.Num, round)
}
func (c *Chain) Int64() int64 {
	return RoundInt64(c.Num)
}
func (c *Chain) String() string {
	return fmt.Sprintf("%f", c.Num)
}
