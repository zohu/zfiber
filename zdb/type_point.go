package zdb

import (
	"context"
	"database/sql/driver"
	"fmt"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// Point
// @Description: ewkb：专为 PostGIS 设计，简化了 SRID 和扩展类型的处理，是数据库交互的首选。
type Point ewkb.Point

func NewPoint(x, y float64) Point {
	return Point{Point: geom.NewPointFlat(geom.XY, []float64{x, y})}
}

func (p *Point) Equal(other Point) bool {
	return p.X() == other.X() && p.Y() == other.Y() && p.Z() == other.Z()
}

func (p *Point) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "GEOMETRY(Point,4326)"
}

// GormValue INSERT INTO `users` (`name`,`point`) VALUES ("jinzhu",ST_PointFromText("POINT(100 100)"))
func (p *Point) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return clause.Expr{
		SQL:  "ST_PointFromText(?)",
		Vars: []interface{}{fmt.Sprintf("POINT(%f %f)", p.X(), p.Y())},
	}
}

func (p *Point) Scan(value interface{}) error {
	if value == nil {
		*p = Point{}
		return nil
	}
	b := ewkb.Point(*p)
	return b.Scan(value)
}
func (p Point) Value() (driver.Value, error) {
	b := ewkb.Point(p)
	return b.Value()
}
