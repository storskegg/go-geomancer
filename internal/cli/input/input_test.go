package input

import (
	"errors"
	"strings"
	"testing"

	"github.com/storskegg/go-geomancer/figure"
)

func TestArgs(t *testing.T) {
	got, err := Args{"1212", "Fortuna Minor", "odd,odd,odd,odd", "populus"}.Mothers()
	if err != nil {
		t.Fatal(err)
	}
	want := [4]figure.Figure{figure.Amissio, figure.FortunaMinor, figure.Via, figure.Populus}
	if got != want {
		t.Errorf("Mothers() = %v, want %v", got, want)
	}
}

func TestArgsErrors(t *testing.T) {
	if _, err := (Args{"1212"}).Mothers(); err == nil || !strings.Contains(err.Error(), "got 1") {
		t.Errorf("one arg: error = %v", err)
	}
	_, err := Args{"1212", "1212", "nope", "1212"}.Mothers()
	if !errors.Is(err, figure.ErrInvalid) || !strings.HasPrefix(err.Error(), "mother M3:") {
		t.Errorf("bad figure: error = %v", err)
	}
}
