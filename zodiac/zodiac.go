// Package zodiac defines the astrological vocabulary used by geomantic
// correspondences: elements, planets and signs.
//
// Each type is a small enum that marshals to and from its name, so values
// read naturally in JSON and on the command line.
package zodiac

import (
	"errors"
	"fmt"

	"github.com/storskegg/go-geomancer/internal/names"
)

// ErrUnknown is returned when a name or value doesn't match any member of an
// enum.
var ErrUnknown = errors.New("zodiac: unknown value")

func label[T ~uint8](labels []string, kind string, v T) string {
	if int(v) < len(labels) {
		return labels[v]
	}
	return fmt.Sprintf("%s(%d)", kind, v)
}

func parse[T ~uint8](labels []string, kind, s string) (T, error) {
	i, ok := names.Index(labels, s)
	if !ok {
		return 0, fmt.Errorf("%w: %s %q", ErrUnknown, kind, s)
	}
	return T(i), nil
}

func marshal[T ~uint8](labels []string, kind string, v T) ([]byte, error) {
	if int(v) >= len(labels) {
		return nil, fmt.Errorf("%w: %s(%d)", ErrUnknown, kind, v)
	}
	return []byte(labels[v]), nil
}
