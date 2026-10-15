package chart

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCircumstances(t *testing.T) {
	chicago, err := time.LoadLocation("America/Chicago")
	if err != nil {
		t.Fatal(err)
	}
	c := Circumstances{
		Time: time.Date(2026, 10, 15, 15, 30, 0, 0, chicago), TimeZone: "America/Chicago", Zone: "CDT",
		Place: Location{Name: "Springfield", State: "IL", County: "Sangamon County", Latitude: 39.80172, Longitude: -89.64371},
	}
	if got := c.UTCOffset(); got != "UTC−05:00" {
		t.Errorf("UTCOffset = %q", got)
	}
	if got := c.Place.Coordinates(); got != "39.8017° N, 89.6437° W" {
		t.Errorf("Coordinates = %q", got)
	}
	if got := (Location{Latitude: -33.9, Longitude: 151.2}).Coordinates(); got != "33.9000° S, 151.2000° E" {
		t.Errorf("southern Coordinates = %q", got)
	}
	if got := c.Place.Label(); got != "Springfield, IL" {
		t.Errorf("Label = %q", got)
	}

	// Kolkata is UTC+5:30, and the minutes count.
	kolkata, _ := time.LoadLocation("Asia/Kolkata")
	if got := (Circumstances{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, kolkata)}).UTCOffset(); got != "UTC+05:30" {
		t.Errorf("Kolkata UTCOffset = %q", got)
	}

	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"time":"2026-10-15T15:30:00-05:00","timeZone":"America/Chicago","zone":"CDT",` +
		`"place":{"name":"Springfield","state":"IL","county":"Sangamon County","latitude":39.80172,"longitude":-89.64371}}`
	if string(b) != want {
		t.Errorf("Marshal =\n%s\nwant\n%s", b, want)
	}
}
