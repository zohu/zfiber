package zdb

import (
	"database/sql/driver"
	"fmt"
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
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("unsupported type for Point: %T", value)
	}
	_, err := fmt.Sscanf(str, "SRID=4326;POINT(%f %f)", &p.X, &p.Y)
	return err
}
func (p Point) Value() (driver.Value, error) {
	return fmt.Sprintf("SRID=4326;POINT(%f %f)", p.X, p.Y), nil
}
