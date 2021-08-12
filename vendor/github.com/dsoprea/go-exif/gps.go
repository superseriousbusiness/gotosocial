package exif

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang/geo/s2"
)

var (
	ErrGpsCoordinatesNotValid = errors.New("GPS coordinates not valid")
)

type GpsDegrees struct {
	Orientation               byte
	Degrees, Minutes, Seconds float64
}

func (d GpsDegrees) String() string {
	return fmt.Sprintf("Degrees<O=[%s] D=(%g) M=(%g) S=(%g)>", string([]byte{d.Orientation}), d.Degrees, d.Minutes, d.Seconds)
}

func (d GpsDegrees) Decimal() float64 {
	decimal := float64(d.Degrees) + float64(d.Minutes)/60.0 + float64(d.Seconds)/3600.0

	if d.Orientation == 'S' || d.Orientation == 'W' {
		return -decimal
	} else {
		return decimal
	}
}

type GpsInfo struct {
	Latitude, Longitude GpsDegrees
	Altitude            int
	Timestamp           time.Time
}

func (gi *GpsInfo) String() string {
	return fmt.Sprintf("GpsInfo<LAT=(%.05f) LON=(%.05f) ALT=(%d) TIME=[%s]>", gi.Latitude.Decimal(), gi.Longitude.Decimal(), gi.Altitude, gi.Timestamp)
}

func (gi *GpsInfo) S2CellId() s2.CellID {
	latitude := gi.Latitude.Decimal()
	longitude := gi.Longitude.Decimal()

	ll := s2.LatLngFromDegrees(latitude, longitude)
	cellId := s2.CellIDFromLatLng(ll)

	if cellId.IsValid() == false {
		panic(ErrGpsCoordinatesNotValid)
	}

	return cellId
}
