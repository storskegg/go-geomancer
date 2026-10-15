package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/storskegg/go-geomancer/content"
	"github.com/storskegg/go-geomancer/internal/cli/render"
	"github.com/storskegg/go-geomancer/question"
)

// ordinal returns 1st, 2nd, 3rd, 4th, ….
func ordinal(n int) string {
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

// reason explains a suggestion: the words behind it, or for a derived
// house, how it's counted.
func reason(s question.Suggestion) string {
	if d := s.Derived; d != nil {
		return fmt.Sprintf("the %s from the %s (“%s”, “%s”)", ordinal(d.Subject), ordinal(d.Person), d.PersonWord, d.SubjectWord)
	}
	words := make([]string, len(s.Matches))
	for i, m := range s.Matches {
		words[i] = "“" + m.Word + "”"
	}
	return "from " + strings.Join(words, ", ")
}

// houseName names a house for a hint: "⑦ Partnership", or in accessible
// prompts, which number houses themselves, "House 7 · Partnership".
func houseName(c *content.Set, house int, accessible bool) string {
	sig, _ := c.Signification(house)
	if accessible {
		return fmt.Sprintf("House %d · %s", house, sig.Title)
	}
	return string(rune('①'+house-1)) + " " + sig.Title
}

// suggestHint suggests the significator for the prompt: the house to start
// on and a one-line hint, or 0 and "" if there's no suggestion.
func suggestHint(q string, accessible bool) (int, string) {
	ss := question.Default().Suggest(q)
	if len(ss) == 0 {
		return 0, ""
	}
	c := content.Default()
	parts := make([]string, len(ss))
	for i, s := range ss {
		parts[i] = houseName(c, s.House, accessible) + ", " + reason(s)
	}
	return ss[0].House, "Suggested: " + strings.Join(parts, "; or ") + "."
}

type suggestionJSON struct {
	House   int                  `json:"house"`
	Title   string               `json:"title"`
	Score   float64              `json:"score"`
	Matches []question.Match     `json:"matches"`
	Derived *question.Derivation `json:"derived,omitempty"`
}

func newSuggestCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "suggest QUESTION",
		Short: "Suggest the significator for a question",
		Long: `Suggest the significator, the house of what a question asks about, as
cast's significator prompt does. Up to two houses are suggested, best first.

A question about another person's affairs suggests the derived house: the
house of the matter, counted from the person's house.`,
		Example: `  geomancer suggest "Will the contract be signed?"
  geomancer suggest Will my husband get the job`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			q := strings.Join(args, " ")
			ss := question.Default().Suggest(q)
			c := content.Default()
			out := make([]suggestionJSON, len(ss))
			for i, s := range ss {
				sig, _ := c.Signification(s.House)
				out[i] = suggestionJSON{House: s.House, Title: sig.Title, Score: s.Score, Matches: s.Matches, Derived: s.Derived}
			}
			return opts.print(cmd, out, func() string {
				if len(ss) == 0 {
					return "No suggestion: the significator prompt will start on House 1."
				}
				lines := []string{render.Heading("Suggestions for “" + q + "”")}
				for i, s := range ss {
					sig, _ := c.Signification(s.House)
					lines = append(lines, fmt.Sprintf("  %d. House %d · %s — %s", i+1, s.House, sig.Title, reason(s)))
				}
				return strings.Join(lines, "\n")
			})
		},
	}
}
