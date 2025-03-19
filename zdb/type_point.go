package zdb

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type Point struct {
	X float64
	Y float64
}

func NewPoint(x, y float64) Point {
	return Point{X: x, Y: y}
}

func (p *Point) Equal(other Point) bool {
	return p.X == other.X && p.Y == other.Y
}
func (p *Point) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unsupported type for Point: %T", value)
	}
	str := string(bytes)
	str = strings.TrimPrefix(str, "POINT(")
	str = strings.TrimSuffix(str, ")")

	_, err := fmt.Sscanf(str, "%f %f", &p.X, &p.Y)
	return err
}
func (p Point) Value() (driver.Value, error) {
	return fmt.Sprintf("POINT(%f %f)", p.X, p.Y), nil
}
