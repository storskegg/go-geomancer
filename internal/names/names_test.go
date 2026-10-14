package names

import "testing"

func TestNormalize(t *testing.T) {
	tests := map[string]string{
		"Fortuna Major":   "fortunamajor",
		"fortuna_major":   "fortunamajor",
		"FORTUNA-MAJOR":   "fortunamajor",
		"  Via\t":         "via",
		"Caput Draconis'": "caputdraconis",
		"North Node":      "northnode",
		"":                "",
		" _-' ":           "",
		"Ünïcode":         "ünïcode",
	}
	for in, want := range tests {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIndex(t *testing.T) {
	labels := []string{"Sun", "Moon", "North Node"}
	tests := []struct {
		name string
		want int
		ok   bool
	}{
		{"Sun", 0, true},
		{"moon", 1, true},
		{"north_node", 2, true},
		{"NorthNode", 2, true},
		{"Mars", 0, false},
		{"", 0, false},
		{"   ", 0, false},
	}
	for _, tt := range tests {
		got, ok := Index(labels, tt.name)
		if got != tt.want || ok != tt.ok {
			t.Errorf("Index(%q) = %d, %v, want %d, %v", tt.name, got, ok, tt.want, tt.ok)
		}
	}
}

// An empty label must not match empty-after-normalization input.
func TestIndexEmptyLabel(t *testing.T) {
	if _, ok := Index([]string{""}, "-"); ok {
		t.Error(`Index([""], "-") matched`)
	}
}
