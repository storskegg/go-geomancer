// Package question suggests the significator, the house of what a question
// asks about, from the question's words. It runs entirely offline from data
// built in: a lexicon of each house's terms (from package content) and
// embedded word vectors for words the lexicon doesn't list.
//
// A question about another person's affairs suggests the derived house,
// counted from that person's house: "Will my husband get the job?" asks
// about the 10th house from the 7th, House 4.
package question

import (
	"cmp"
	"slices"
	"strings"
	"sync"
	"unicode"

	"github.com/kljensen/snowball/english"
	"golang.org/x/text/unicode/norm"

	"github.com/storskegg/go-geomancer/content"
)

// Suggestion is a house the question may be about.
type Suggestion struct {
	House int     `json:"house"`
	Score float64 `json:"score"`
	// Words are the question's words that point to the house: a word may
	// match a term directly, or by meaning (see Match).
	Matches []Match `json:"matches"`
	// Derived is set when the house is derived from a person's house.
	Derived *Derivation `json:"derived,omitempty"`
}

// Match is a word in the question that points to a house.
type Match struct {
	Word string `json:"word"` // as written in the question
	Term string `json:"term"` // the lexicon term it matched
	// ByMeaning is set when Word isn't the term but is similar in meaning,
	// by the word vectors.
	ByMeaning bool `json:"byMeaning,omitempty"`
}

// Derivation explains a derived house: the Subject's house counted from the
// Person's.
type Derivation struct {
	Person      int    `json:"person"`      // the person's house
	Subject     int    `json:"subject"`     // the matter's house, counted from the person's
	PersonWord  string `json:"personWord"`  // as written, such as "husband"
	SubjectWord string `json:"subjectWord"` // such as "job"
}

// Suggester scores questions against a lexicon.
type Suggester struct {
	terms     []term
	exact     map[string]bool // single-word terms as written
	people    map[string]int  // stem → house
	qualities map[string]bool // stems
	vectors   *vectors
}

// term is a lexicon entry, stemmed for matching.
type term struct {
	house  int
	text   string   // as written in the lexicon
	texts  []string // every spelling merged into it ("buy", "buying")
	stems  []string // one per word of a phrase
	weight float64
	vector []int8 // nil for phrases and words without one
}

// Tuning, chosen against the labelled questions in testdata.
const (
	// minSimilarity is how similar in meaning a word must be to a term to
	// count, and similarityWeight scales its contribution. Only terms of at
	// least minMeaningWeight are matched by meaning: weak terms are too
	// general.
	minSimilarity    = 0.62
	similarityWeight = 0.8
	minMeaningWeight = 2
	// minMeaningRank skips the most common words when matching by meaning:
	// general words ("making", "find", "feel") sit near too many terms to be
	// a sign of the subject. The vectors are stored most frequent first.
	minMeaningRank = 2000
	// minScore is the least score worth suggesting.
	minScore = 1.0
	// secondRatio is how close, as a fraction of the first, the second
	// suggestion's score must be to be offered.
	secondRatio = 0.3
)

// New builds a suggester from the significations' terms and people.
func New(sigs [12]content.Signification) (*Suggester, error) {
	v, err := loadVectors()
	if err != nil {
		return nil, err
	}
	s := &Suggester{exact: map[string]bool{}, people: map[string]int{}, qualities: map[string]bool{}, vectors: v}
	for _, sig := range sigs {
		// Terms with the same stems ("buy", "buying") are one term, with the
		// greater weight.
		byStems := map[string]int{} // joined stems → index in s.terms
		for _, text := range sortedKeys(sig.Terms) {
			w := sig.Terms[text]
			words := tokens(text)
			t := term{house: sig.House, text: text, texts: []string{text}, weight: w}
			for _, wd := range words {
				t.stems = append(t.stems, stem(wd))
			}
			if len(words) == 1 && w >= minMeaningWeight {
				t.vector = v.vector(words[0])
			}
			if len(words) == 1 {
				s.exact[words[0]] = true
			}
			key := strings.Join(t.stems, " ")
			if i, ok := byStems[key]; ok {
				s.terms[i].texts = append(s.terms[i].texts, text)
				if w > s.terms[i].weight {
					s.terms[i].weight = w
				}
				if s.terms[i].vector == nil {
					s.terms[i].vector = t.vector
				}
				continue
			}
			byStems[key] = len(s.terms)
			s.terms = append(s.terms, t)
		}
		for _, p := range sig.People {
			s.people[stem(p)] = sig.House
		}
		for _, q := range sig.Qualities {
			s.qualities[stem(q)] = true
		}
	}
	// Phrases first, so they take their words before single-word terms do;
	// then a stable order, so ties break the same way every time.
	slices.SortFunc(s.terms, func(a, b term) int {
		return cmp.Or(cmp.Compare(len(b.stems), len(a.stems)), cmp.Compare(a.house, b.house), strings.Compare(a.text, b.text))
	})
	return s, nil
}

func sortedKeys(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// Suggest returns up to two houses the question may be about, best first:
// the second only if it scores close to the first. It returns none for an
// empty question or one with nothing to go on.
func (s *Suggester) Suggest(q string) []Suggestion {
	words := tokens(q)
	if len(words) == 0 {
		return nil
	}
	firstPerson := false
	var content []string // words that aren't stop words
	var owned []bool     // whether each follows "my" or "our"
	for i, w := range words {
		if firstPersonWords[w] {
			firstPerson = true
		}
		if !stopWords[w] {
			content = append(content, w)
			owned = append(owned, i > 0 && (words[i-1] == "my" || words[i-1] == "our"))
		}
	}
	stems := make([]string, len(content))
	for i, w := range content {
		stems[i] = stem(w)
	}

	var scores [13]float64
	var matches [13][]Match
	matched := make([]bool, len(content))
	credit := func(h int, score float64, m Match) {
		scores[h] += score
		for _, prev := range matches[h] {
			if prev == m {
				return
			}
		}
		matches[h] = append(matches[h], m)
	}

	// Lexicon terms, including phrases, matched by stem. Each term counts
	// once, and a word already taken by a phrase isn't matched again. A word
	// that is itself a term isn't claimed by other terms that merely share
	// its stem: "faithful" is a term, so it doesn't also match "faith".
	credited := make([]bool, len(s.terms)) // by index in s.terms
	for k := range s.terms {
		t := s.terms[k]
		for i := 0; i+len(t.stems) <= len(stems); i++ {
			if len(t.stems) == 1 && s.exact[content[i]] && !slices.Contains(t.texts, content[i]) {
				continue
			}
			if slices.Equal(stems[i:i+len(t.stems)], t.stems) && !slices.Contains(matched[i:i+len(t.stems)], true) {
				// The term scores once, but takes every occurrence, so a
				// repeat isn't matched again by meaning.
				if !credited[k] {
					credit(t.house, t.weight, Match{Word: strings.Join(content[i:i+len(t.stems)], " "), Term: t.text})
					credited[k] = true
				}
				for j := i; j < i+len(t.stems); j++ {
					matched[j] = true
				}
			}
		}
	}

	// Other words, by meaning: each credits the house of its most similar
	// term, if similar enough.
	for i, w := range content {
		if matched[i] {
			continue
		}
		if s.vectors.rank(w) < minMeaningRank {
			continue
		}
		vw := s.vectors.vector(w)
		if vw == nil {
			continue
		}
		var best *term
		bestSim := minSimilarity
		for k := range s.terms {
			t := &s.terms[k]
			if t.vector == nil {
				continue
			}
			if sim := similarity(vw, t.vector); sim > bestSim {
				best, bestSim = t, sim
			}
		}
		if best != nil {
			credit(best.house, best.weight*bestSim*similarityWeight, Match{Word: w, Term: best.text, ByMeaning: true})
		}
	}

	var ranked []Suggestion
	for h := 1; h <= 12; h++ {
		if scores[h] >= minScore {
			ranked = append(ranked, Suggestion{House: h, Score: scores[h], Matches: matches[h]})
		}
	}
	slices.SortStableFunc(ranked, func(a, b Suggestion) int { return cmp.Compare(b.Score, a.Score) })

	// A derived house comes first, with the subject's own house as the
	// alternative.
	if d := s.derive(content, stems, owned, ranked, firstPerson); d != nil {
		out := []Suggestion{*d}
		for _, r := range ranked {
			if r.House == d.Derived.Subject && r.House != d.House {
				out = append(out, r)
				break
			}
		}
		return out
	}

	if len(ranked) > 2 {
		ranked = ranked[:2]
	}
	if len(ranked) == 2 && ranked[1].Score < ranked[0].Score*secondRatio {
		ranked = ranked[:1]
	}
	return ranked
}

// derive returns the derived-house suggestion for a question about another
// person's affairs, or nil. The question must name a person signified by a
// house other than the 1st, not be in the first person ("Will I marry my
// boss?" is the querent's own matter), and have a subject: a lexicon term,
// matched exactly, besides the person and not the querent's ("my work" in
// "Will my boss like my work?"). The subject may be in the person's own
// house: "Will my boss be fired?" asks about the 10th from the 10th.
func (s *Suggester) derive(words, stems []string, owned []bool, ranked []Suggestion, firstPerson bool) *Suggestion {
	if firstPerson {
		return nil
	}
	person, pi := 0, -1
	for i, st := range stems {
		if h, ok := s.people[st]; ok && h != 1 {
			person, pi = h, i
			break
		}
	}
	if person == 0 {
		return nil
	}
	// The subject: the best-scoring house with a match other than the
	// person.
	ownedWord := map[string]bool{}
	for i, w := range words {
		if owned[i] {
			ownedWord[w] = true
		}
	}
	for _, r := range ranked {
		for _, m := range r.Matches {
			if m.ByMeaning || ownedWord[m.Word] || s.qualities[stem(m.Word)] || slices.Contains(strings.Fields(m.Word), words[pi]) {
				continue
			}
			derived := (person-1+r.House-1)%12 + 1
			return &Suggestion{
				House:   derived,
				Score:   r.Score,
				Matches: []Match{{Word: words[pi], Term: words[pi]}, m},
				Derived: &Derivation{Person: person, Subject: r.House, PersonWord: words[pi], SubjectWord: m.Word},
			}
		}
	}
	return nil
}

// tokens splits text into lowercase words of letters, without accents.
// Possessives and contractions split off ("husband's" gives "husband", "s").
func tokens(text string) []string {
	var b strings.Builder
	for _, r := range norm.NFD.String(strings.ToLower(text)) {
		switch {
		case unicode.Is(unicode.Mn, r): // a combining accent
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		default:
			b.WriteByte(' ')
		}
	}
	return strings.Fields(b.String())
}

func stem(word string) string { return english.Stem(word, true) }

// firstPersonWords mark a question about the querent's own matters.
var firstPersonWords = set("i", "me", "myself", "mine")

// stopWords carry no subject.
var stopWords = set(
	"a", "about", "above", "after", "again", "against", "all", "also", "am", "an", "and", "any", "are",
	"as", "at", "be", "been", "before", "being", "best", "better", "both", "but", "by", "can", "could",
	"d", "did", "do", "does", "doing", "done", "down", "during", "each", "ever", "few", "for", "from",
	"further", "get", "gets", "getting", "go", "goes", "going", "gone", "good", "got", "had", "has",
	"have", "having", "he", "her", "here", "hers", "herself", "him", "himself", "his", "how", "i", "if",
	"in", "into", "is", "it", "its", "itself", "just", "know", "last", "let", "like", "ll", "m",
	"make", "many", "may", "me", "might", "mine", "more", "most", "much", "must", "my", "myself", "new",
	"next", "no", "nor", "not", "now", "of", "off", "on", "once", "one", "only", "or", "other", "our",
	"ours", "out", "over", "own", "re", "really", "right", "s", "same", "shall", "she", "should", "so",
	"some", "soon", "still", "such", "t", "take", "than", "that", "the", "their", "theirs", "them",
	"themselves", "then", "there", "these", "they", "thing", "things", "think", "this", "those",
	"through", "time", "to", "too", "under", "until", "up", "us", "ve", "very", "want", "was", "way",
	"we", "week", "well", "were", "what", "when", "where", "whether", "which", "while", "who", "whom",
	"why", "will", "with", "would", "year", "yes", "you", "your", "yours", "yourself",
	// Common verbs and fillers too general to point anywhere.
	"along", "anything", "change", "come", "comes", "coming", "enough", "everything", "face", "happen",
	"happens", "hold", "okay", "someone", "something", "stop", "worth",
)

func set(words ...string) map[string]bool {
	m := make(map[string]bool, len(words))
	for _, w := range words {
		m[w] = true
	}
	return m
}

// Default returns a suggester for the default content's significations.
func Default() *Suggester {
	s, err := defaultSuggester()
	if err != nil {
		panic(err) // the embedded data is checked by the package tests
	}
	return s
}

var defaultSuggester = sync.OnceValues(func() (*Suggester, error) {
	return New(content.Default().Significations())
})
