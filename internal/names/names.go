// Package names provides lenient matching of human-entered names against
// canonical labels, shared by the core enum parsers.
package names

import "strings"

// Normalize lowercases s and strips spaces, underscores, hyphens and
// apostrophes, so "Fortuna Major", "fortuna_major" and "FortunaMajor" compare
// equal.
func Normalize(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range strings.ToLower(s) {
		switch r {
		case ' ', '\t', '_', '-', '\'':
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Index returns the position of name in labels after normalizing both.
func Index(labels []string, name string) (int, bool) {
	n := Normalize(name)
	if n == "" {
		return 0, false
	}
	for i, l := range labels {
		if Normalize(l) == n {
			return i, true
		}
	}
	return 0, false
}
