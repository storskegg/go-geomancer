package figure

import "testing"

func TestSegmentString(t *testing.T) {
	want := []string{"Head", "Neck", "Body", "Feet"}
	for i, s := range Segments() {
		if int(s) != i {
			t.Errorf("Segments()[%d] = %d", i, s)
		}
		if got := s.String(); got != want[i] {
			t.Errorf("Segment(%d).String() = %q, want %q", i, got, want[i])
		}
	}
	if got := Segment(4).String(); got != "Segment(4)" {
		t.Errorf("Segment(4).String() = %q", got)
	}
}

func TestParity(t *testing.T) {
	tests := []struct {
		p      Parity
		dots   int
		active bool
		str    string
	}{
		{Odd, 1, true, "odd"},
		{Even, 2, false, "even"},
	}
	for _, tt := range tests {
		if got := tt.p.Dots(); got != tt.dots {
			t.Errorf("%v.Dots() = %d, want %d", tt.p, got, tt.dots)
		}
		if got := tt.p.Active(); got != tt.active {
			t.Errorf("%v.Active() = %v, want %v", tt.p, got, tt.active)
		}
		if got := tt.p.String(); got != tt.str {
			t.Errorf("Parity(%d).String() = %q, want %q", tt.p, got, tt.str)
		}
	}
}
