package figure

import (
	"errors"
	"fmt"
	"strings"

	"github.com/storskegg/go-geomancer/internal/names"
)

// ErrInvalid is returned when input doesn't describe a figure.
var ErrInvalid = errors.New("figure: invalid figure")

// Parse reads a figure from any of these forms:
//
//   - a dot pattern, top to bottom: "1212"
//   - comma-separated parities: "odd,even,odd,even"
//   - a name, ignoring case and spacing: "Fortuna Major", "fortunamajor"
//     (common alternative spellings such as "Fortuna Maior" are accepted)
func Parse(s string) (Figure, error) {
	s = strings.TrimSpace(s)
	if f, ok := parsePattern(s); ok {
		return f, nil
	}
	if f, ok := parseParities(s); ok {
		return f, nil
	}
	if f, ok := parseName(s); ok {
		return f, nil
	}
	return 0, fmt.Errorf("%w: %q", ErrInvalid, s)
}

func parsePattern(s string) (Figure, bool) {
	if len(s) != 4 {
		return 0, false
	}
	var p [4]Parity
	for i := range 4 {
		switch s[i] {
		case '1':
			p[i] = Odd
		case '2':
			p[i] = Even
		default:
			return 0, false
		}
	}
	return FromSegments(p), true
}

func parseParities(s string) (Figure, bool) {
	parts := strings.Split(s, ",")
	if len(parts) != 4 {
		return 0, false
	}
	var p [4]Parity
	for i, part := range parts {
		switch strings.ToLower(strings.TrimSpace(part)) {
		case "odd":
			p[i] = Odd
		case "even":
			p[i] = Even
		default:
			return 0, false
		}
	}
	return FromSegments(p), true
}

func parseName(s string) (Figure, bool) {
	n := names.Normalize(s)
	if n == "" {
		return 0, false
	}
	for f, fi := range figures {
		if names.Normalize(fi.name) == n {
			return Figure(f), true
		}
		for _, a := range fi.aliases {
			if names.Normalize(a) == n {
				return Figure(f), true
			}
		}
	}
	return 0, false
}
