package question

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/storskegg/go-geomancer/content"
)

func suggester(t *testing.T) *Suggester {
	t.Helper()
	s, err := New(content.Default().Significations())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

type labelled struct {
	q      string
	houses []int // acceptable, best first
}

func readLabelled(t *testing.T, file string) []labelled {
	t.Helper()
	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	var out []labelled
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		q, hs, ok := strings.Cut(line, "\t")
		if !ok {
			t.Fatalf("bad line %q", line)
		}
		l := labelled{q: q}
		for h := range strings.SplitSeq(hs, ",") {
			n, err := strconv.Atoi(h)
			if err != nil {
				t.Fatalf("bad houses in %q", line)
			}
			l.houses = append(l.houses, n)
		}
		out = append(out, l)
	}
	return out
}

func contains(hs []int, h int) bool {
	for _, x := range hs {
		if x == h {
			return true
		}
	}
	return false
}

// TestAccuracy scores the labelled questions the suggester was tuned on, and
// the held-out ones it wasn't, which measure how well it generalises. A
// question counts as right at top 1 if the first suggestion is an acceptable
// house, and at top 2 if either is.
func TestAccuracy(t *testing.T) {
	for _, set := range []struct {
		file       string
		top1, top2 float64
	}{
		{"testdata/questions.tsv", 0.95, 0.97},
		{"testdata/holdout.tsv", 0.80, 0.90},
		{"testdata/holdout2.tsv", 0.70, 0.80},
	} {
		t.Run(set.file, func(t *testing.T) { accuracy(t, set.file, set.top1, set.top2) })
	}
}

func accuracy(t *testing.T, file string, minTop1, minTop2 float64) {
	s := suggester(t)
	qs := readLabelled(t, file)
	top1, top2 := 0, 0
	var misses []string
	for _, l := range qs {
		got := s.Suggest(l.q)
		// With no suggestion the prompt opens on House 1, so that's the
		// answer.
		if len(got) == 0 {
			got = []Suggestion{{House: 1}}
		}
		ok1 := len(got) > 0 && contains(l.houses, got[0].House)
		ok2 := ok1 || (len(got) > 1 && contains(l.houses, got[1].House))
		if ok1 {
			top1++
		}
		if ok2 {
			top2++
		}
		if !ok1 {
			misses = append(misses, l.q+"  want "+intsToString(l.houses)+", got "+describe(got)+mark(ok2))
		}
	}
	n := float64(len(qs))
	t.Logf("%d questions: top 1 %.1f%%, top 2 %.1f%%", len(qs), 100*float64(top1)/n, 100*float64(top2)/n)
	for _, m := range misses {
		t.Log("  " + m)
	}
	if float64(top1)/n < minTop1 || float64(top2)/n < minTop2 {
		t.Errorf("accuracy below target: top 1 %.1f%% (want %.0f%%), top 2 %.1f%% (want %.0f%%)",
			100*float64(top1)/n, 100*minTop1, 100*float64(top2)/n, 100*minTop2)
	}
}

func intsToString(hs []int) string {
	parts := make([]string, len(hs))
	for i, h := range hs {
		parts[i] = strconv.Itoa(h)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func mark(ok2 bool) string {
	if ok2 {
		return " (top 2 ok)"
	}
	return ""
}

// describe lists suggestions with the words behind them, for diagnosing
// misses.
func describe(ss []Suggestion) string {
	parts := make([]string, len(ss))
	for i, s := range ss {
		var ms []string
		for _, m := range s.Matches {
			if m.Word == m.Term {
				ms = append(ms, m.Word)
			} else {
				ms = append(ms, m.Word+"≈"+m.Term)
			}
		}
		parts[i] = strconv.Itoa(s.House) + "(" + strings.Join(ms, " ") + ")"
	}
	return "[" + strings.Join(parts, " ") + "]"
}
