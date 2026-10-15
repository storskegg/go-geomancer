package question

import (
	"bytes"
	"encoding/binary"
	"slices"
	"testing"

	"github.com/storskegg/go-geomancer/content"
)

func TestTokens(t *testing.T) {
	tests := map[string][]string{
		"Will my husband's job last?": {"will", "my", "husband", "s", "job", "last"},
		"Fiancée—and café!":           {"fiancee", "and", "cafe"},
		"  ":                          nil,
		"I'd say 42 things":           {"i", "d", "say", "things"},
	}
	for in, want := range tests {
		if got := tokens(in); !slices.Equal(got, want) {
			t.Errorf("tokens(%q) = %q, want %q", in, got, want)
		}
	}
}

func houses(ss []Suggestion) []int {
	out := make([]int, len(ss))
	for i, s := range ss {
		out[i] = s.House
	}
	return out
}

func TestSuggestNothing(t *testing.T) {
	s := suggester(t)
	for _, q := range []string{"", "   ", "Will it?", "What will happen to me?"} {
		if got := s.Suggest(q); len(got) != 0 {
			t.Errorf("Suggest(%q) = %v, want none", q, houses(got))
		}
	}
}

func TestSuggestMatches(t *testing.T) {
	s := suggester(t)
	got := s.Suggest("Will we get married?")
	if len(got) == 0 || got[0].House != 7 {
		t.Fatalf("Suggest = %v", houses(got))
	}
	if m := got[0].Matches; len(m) != 1 || m[0] != (Match{Word: "married", Term: "married"}) {
		t.Errorf("matches = %+v", m)
	}

	// A phrase takes its words: "short trip" is House 3, not "trip", House 9.
	if got := s.Suggest("Will the short trip be pleasant?"); len(got) == 0 || got[0].House != 3 {
		t.Errorf("short trip: %v", houses(got))
	}

	// Terms sharing a stem count once: "buy" and "buying" are one term.
	buy := s.Suggest("Should I buy it?")
	buying := s.Suggest("Should I buy it, buying it?")
	if len(buy) == 0 || len(buying) == 0 || buy[0].Score != buying[0].Score {
		t.Errorf("buy scored %v, buy and buying %v", buy, buying)
	}

	// A word that is itself a term isn't claimed by another sharing its
	// stem: "faithful" is House 7, not "faith", House 9.
	for _, sg := range s.Suggest("Will he be faithful?") {
		if sg.House == 9 {
			t.Errorf("faithful matched House 9: %+v", sg)
		}
	}
}

func TestSuggestByMeaning(t *testing.T) {
	s := suggester(t)
	// "chemotherapy" isn't in the lexicon, but is close to its illness terms.
	got := s.Suggest("Will the chemotherapy succeed?")
	if len(got) == 0 || got[0].House != 6 || !got[0].Matches[0].ByMeaning {
		t.Fatalf("Suggest = %+v", got)
	}
	// Common words aren't matched by meaning, however close.
	for _, sg := range s.Suggest("Am I making progress?") {
		for _, m := range sg.Matches {
			if m.Word == "making" {
				t.Errorf("matched the common word %q by meaning to %q", m.Word, m.Term)
			}
		}
	}
}

func TestSuggestDerived(t *testing.T) {
	s := suggester(t)
	tests := []struct {
		q               string
		house           int
		person, subject int // 0: not derived
	}{
		{"Will my husband get the job?", 4, 7, 10},
		{"Will my boss be fired?", 7, 10, 10},    // in the person's own house
		{"Is my daughter pregnant?", 9, 5, 5},    // likewise
		{"Will I marry my boss?", 7, 0, 0},       // the querent's own matter
		{"Will my boss like my work?", 10, 0, 0}, // the querent's work
		{"Is my boyfriend faithful?", 7, 0, 0},   // a quality, not an affair
		{"How is my sister doing?", 3, 0, 0},     // no subject
		{"Will my wife forgive me?", 7, 0, 0},    // in the first person
		{"Will my business partner be honest?", 7, 0, 0},
	}
	for _, tt := range tests {
		got := s.Suggest(tt.q)
		if len(got) == 0 || got[0].House != tt.house {
			t.Errorf("%q: %v, want %d first", tt.q, houses(got), tt.house)
			continue
		}
		d := got[0].Derived
		switch {
		case tt.person == 0 && d != nil:
			t.Errorf("%q: derived %+v, want none", tt.q, *d)
		case tt.person != 0 && (d == nil || d.Person != tt.person || d.Subject != tt.subject):
			t.Errorf("%q: derived %+v, want %d from %d", tt.q, d, tt.subject, tt.person)
		}
	}

	// The subject's own house is the alternative.
	got := s.Suggest("Will my husband get the job?")
	if len(got) != 2 || got[1].House != 10 || got[1].Derived != nil {
		t.Errorf("alternative = %+v", got)
	}
}

func TestSuggestAtMostTwo(t *testing.T) {
	s := suggester(t)
	got := s.Suggest("money marriage career health travel friends prison")
	if len(got) > 2 {
		t.Errorf("%d suggestions, want at most 2", len(got))
	}
	// A distant second isn't offered.
	if got := s.Suggest("Will my husband and I divorce, and will the divorce be fair?"); len(got) != 1 {
		t.Errorf("Suggest = %+v, want just House 7", got)
	}
}

func TestVectors(t *testing.T) {
	v, err := loadVectors()
	if err != nil {
		t.Fatal(err)
	}
	if v.dims != 100 || len(v.index) != 50000 {
		t.Errorf("%d words × %d dimensions", len(v.index), v.dims)
	}
	if v.rank("the") != 0 || v.rank("xyzzyq") != -1 || v.vector("xyzzyq") != nil {
		t.Error("rank or vector of known and unknown words")
	}
	self := similarity(v.vector("king"), v.vector("king"))
	near := similarity(v.vector("king"), v.vector("queen"))
	far := similarity(v.vector("king"), v.vector("banana"))
	if self < 0.99 || near <= far {
		t.Errorf("similarities: self %.2f, king–queen %.2f, king–banana %.2f", self, near, far)
	}
}

func TestParseVectorsInvalid(t *testing.T) {
	good := func() []byte {
		var b bytes.Buffer
		b.WriteString("GMV1")
		_ = binary.Write(&b, binary.LittleEndian, uint32(2))
		_ = binary.Write(&b, binary.LittleEndian, uint32(1))
		b.Write([]byte{2, 'h', 'i', 10, 0xF6})
		return b.Bytes()
	}
	if v, err := parseVectors(good()); err != nil || v.vector("hi")[1] != -10 {
		t.Fatalf("valid file: %v, %v", v, err)
	}
	for name, b := range map[string][]byte{
		"empty":     nil,
		"magic":     append([]byte("XXXX"), good()[4:]...),
		"truncated": good()[:14],
		"short":     good()[:len(good())-1],
		"long":      append(good(), 0),
	} {
		if _, err := parseVectors(b); err == nil {
			t.Errorf("%s: parsed", name)
		}
	}
}

func TestNewWithOtherContent(t *testing.T) {
	var sigs [12]content.Signification
	for i := range sigs {
		sigs[i].House = i + 1
	}
	sigs[6].Terms = map[string]float64{"zebra": 3}
	s, err := New(sigs)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Suggest("A zebra?"); len(got) != 1 || got[0].House != 7 {
		t.Errorf("Suggest = %+v", got)
	}
	if Default() == nil {
		t.Error("Default() is nil")
	}
}
