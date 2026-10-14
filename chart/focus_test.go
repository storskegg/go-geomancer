package chart

import (
	"testing"

	"github.com/storskegg/go-geomancer/aspect"
	"github.com/storskegg/go-geomancer/system/goldendawn"
)

func TestFocusWithoutSignificator(t *testing.T) {
	r, _ := NewReading(exampleMothers, goldendawn.System)
	f := r.Focus()
	if len(f) != 12 {
		t.Fatalf("%d houses, want 12", len(f))
	}
	for i, h := range f {
		if h.House != i+1 || h.Querent != (i == 0) || h.Significator || h.Aspect != nil {
			t.Errorf("focus[%d] = %+v", i, h)
		}
	}
}

func TestFocus(t *testing.T) {
	r, _ := NewReading(exampleMothers, goldendawn.System)
	if err := r.SetSignificator(10); err != nil {
		t.Fatal(err)
	}
	// House 1 is the querent and also the Sinister Square to House 10, so
	// it appears once, first.
	want := []struct {
		house        int
		querent, sig bool
		aspect       aspect.Aspect
		side         aspect.Side
		aspected     bool
	}{
		{1, true, false, aspect.Square, aspect.Sinister, true},
		{10, false, true, 0, 0, false},
		{4, false, false, aspect.Opposition, aspect.Neither, true},
		{7, false, false, aspect.Square, aspect.Dexter, true},
		{6, false, false, aspect.Trine, aspect.Dexter, true},
		{2, false, false, aspect.Trine, aspect.Sinister, true},
		{8, false, false, aspect.Sextile, aspect.Dexter, true},
		{12, false, false, aspect.Sextile, aspect.Sinister, true},
	}
	got := r.Focus()
	if len(got) != len(want) {
		t.Fatalf("Focus() = %+v", got)
	}
	for i, w := range want {
		g := got[i]
		if g.House != w.house || g.Querent != w.querent || g.Significator != w.sig || (g.Aspect != nil) != w.aspected {
			t.Errorf("focus[%d] = %+v, want house %d", i, g, w.house)
			continue
		}
		if g.Aspect != nil && (g.Aspect.Aspect != w.aspect || g.Aspect.Side != w.side || g.Aspect.House != w.house) {
			t.Errorf("focus[%d] aspect = %+v", i, *g.Aspect)
		}
	}
}

// When the question is about the querent, House 1 is both.
func TestFocusSignificatorIsQuerent(t *testing.T) {
	r, _ := NewReading(exampleMothers, goldendawn.System)
	_ = r.SetSignificator(1)
	f := r.Focus()
	if len(f) != 8 || f[0].House != 1 || !f[0].Querent || !f[0].Significator {
		t.Errorf("Focus() = %+v", f)
	}
	seen := map[int]bool{}
	for _, h := range f {
		if seen[h.House] {
			t.Errorf("house %d listed twice", h.House)
		}
		seen[h.House] = true
	}
}
