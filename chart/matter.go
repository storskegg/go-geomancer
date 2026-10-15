package chart

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Matter is what a question asks about within the significator's house,
// when the question shows it: "Should I get a new pet?" asks about House
// 6's small animals rather than its sickness.
type Matter struct {
	// Topics are the parts of the house's signification the question
	// concerns, most relevant first, such as "small animals" in House 6.
	// For a derived house, they're in the subject's own house.
	Topics []string `json:"topics"`
	// Derived is set when the house is derived from a person's house.
	Derived *Derivation `json:"derived,omitempty"`
}

// Derivation explains a derived house: the Subject's house counted from the
// Person's. "Will my husband get the job?" asks about the 10th house (the
// job) from the 7th (the husband): House 4.
type Derivation struct {
	Person      int    `json:"person"`      // the person's house
	Subject     int    `json:"subject"`     // the matter's house, counted from the person's
	PersonWord  string `json:"personWord"`  // as written, such as "husband"
	SubjectWord string `json:"subjectWord"` // such as "job"
}

// maxTopics is the most topics Subject names.
const maxTopics = 2

// Subject names the matter, capitalized: its topics ("Small animals",
// "Sickness and small animals"), and for a derived house, whose they are
// ("Husband's career"). It's "" if there are no topics.
func (m *Matter) Subject() string {
	if m == nil || len(m.Topics) == 0 {
		return ""
	}
	s := strings.Join(m.Topics[:min(len(m.Topics), maxTopics)], " and ")
	if m.Derived != nil {
		s = possessive(m.Derived.PersonWord) + " " + s
	}
	return capitalize(s)
}

// Label is Subject with how a derived house is counted: "Husband's career
// (the 10th from the 7th)". It's "" if there are no topics.
func (m *Matter) Label() string {
	s := m.Subject()
	if s != "" && m.Derived != nil {
		s += fmt.Sprintf(" (the %s from the %s)", Ordinal(m.Derived.Subject), Ordinal(m.Derived.Person))
	}
	return s
}

// Ordinal returns 1st, 2nd, 3rd, 4th, ….
func Ordinal(n int) string {
	suffix := "th"
	switch n % 10 {
	case 1:
		suffix = "st"
	case 2:
		suffix = "nd"
	case 3:
		suffix = "rd"
	}
	if n%100 >= 11 && n%100 <= 13 {
		suffix = "th"
	}
	return fmt.Sprintf("%d%s", n, suffix)
}

// possessive returns "husband's", "boss's", or for a plural, "parents'".
func possessive(word string) string {
	if strings.HasSuffix(word, "s") && !strings.HasSuffix(word, "ss") {
		return word + "'"
	}
	return word + "'s"
}

func capitalize(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}
