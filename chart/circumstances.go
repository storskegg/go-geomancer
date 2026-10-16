package chart

import (
	"fmt"
	"math"
	"time"

	"github.com/storskegg/go-geomancer/sky"
)

// Circumstances are when and where a chart was cast.
type Circumstances struct {
	// Time is the local time, with the offset then in effect. In JSON it's
	// RFC 3339: "2026-10-15T15:30:00-05:00".
	Time time.Time `json:"time"`
	// TimeZone is the IANA time zone, such as "America/Chicago".
	TimeZone string `json:"timeZone"`
	// Zone is the zone in effect at Time, such as "CDT". Daylight and
	// standard time are separate zones: CDT is UTC−5, CST UTC−6.
	Zone  string   `json:"zone"`
	Place Location `json:"place"`
}

// Location is where a chart was cast.
type Location struct {
	Name      string  `json:"name"`
	State     string  `json:"state"`
	County    string  `json:"county,omitempty"`
	Latitude  float64 `json:"latitude"`  // degrees north
	Longitude float64 `json:"longitude"` // degrees east; negative is west
}

// Label names the location "City, ST".
func (l Location) Label() string { return l.Name + ", " + l.State }

// Coordinates writes the location as "39.8017° N, 89.6437° W".
func (l Location) Coordinates() string {
	ns, ew := "N", "E"
	if l.Latitude < 0 {
		ns = "S"
	}
	if l.Longitude < 0 {
		ew = "W"
	}
	return fmt.Sprintf("%.4f° %s, %.4f° %s", math.Abs(l.Latitude), ns, math.Abs(l.Longitude), ew)
}

// Sky works out the astrology of the moment and place.
func (c Circumstances) Sky() (*sky.Sky, error) {
	return sky.At(c.Time, c.Place.Latitude, c.Place.Longitude)
}

// UTCOffset writes the offset in effect, such as "UTC−05:00".
func (c Circumstances) UTCOffset() string {
	_, off := c.Time.Zone()
	sign := "+"
	if off < 0 {
		sign, off = "−", -off
	}
	return fmt.Sprintf("UTC%s%02d:%02d", sign, off/3600, off%3600/60)
}
