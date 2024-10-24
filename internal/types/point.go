package types

import (
	"database/sql/driver"

	vd "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/thanishsid/go-postgis"

	"github.com/thanishsid/dingilink-server/internal/types/apperror"
)

type LatLng struct {
	Lat float64
	Lng float64
}

func (i LatLng) Validate() error {
	return vd.ValidateStruct(&i,
		vd.Field(
			&i.Lat,
			vd.Min(float64(-90)).Error(apperror.INPUT_TOO_LOW),
			vd.Max(float64(90)).Error(apperror.INPUT_TOO_HIGH),
		),
		vd.Field(
			&i.Lng,
			vd.Min(float64(-180)).Error(apperror.INPUT_TOO_LOW),
			vd.Max(float64(180)).Error(apperror.INPUT_TOO_HIGH),
		),
	)
}

func LatLngToPoint(l *LatLng) Point {
	if l == nil {
		return Point{
			Valid: false,
		}
	}

	return NewCoordinates(l.Lat, l.Lng)
}

type Point struct {
	postgis.PointS
	Valid bool
}

func (p Point) LatLng() *LatLng {
	if p.Valid {
		return &LatLng{
			Lat: p.Y,
			Lng: p.X,
		}
	}

	return nil
}

func (p Point) Value() (driver.Value, error) {
	if !p.Valid {
		return nil, nil
	}

	return p.PointS.Value()
}

func (p *Point) Scan(src any) error {
	if src == nil {
		p.PointS = postgis.PointS{}
		p.Valid = false
		return nil
	}

	if err := p.PointS.Scan(src); err != nil {
		return err
	}

	p.Valid = true

	return nil
}

func NewCoordinates(lat, lng float64) Point {
	return Point{
		PointS: postgis.PointS{
			SRID: 4326,
			X:    lng,
			Y:    lat,
		},
		Valid: true,
	}
}
