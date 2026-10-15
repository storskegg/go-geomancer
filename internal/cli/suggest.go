package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/content"
	"github.com/storskegg/go-geomancer/internal/cli/render"
	"github.com/storskegg/go-geomancer/question"
)

// reason explains a suggestion: the words behind it, or for a derived
// house, how it's counted.
func reason(s question.Suggestion) string {
	if d := s.Derived; d != nil {
		return fmt.Sprintf("the %s from the %s (“%s”, “%s”)", chart.Ordinal(d.Subject), chart.Ordinal(d.Person), d.PersonWord, d.SubjectWord)
	}
	words := make([]string, len(s.Matches))
	for i, m := range s.Matches {
		words[i] = "“" + m.Word + "”"
	}
	return "from " + strings.Join(words, ", ")
}

// matter returns what question q asks about within house, or nil if q
// doesn't show it.
func matter(q string, house int) *chart.Matter { return question.Default().Matter(q, house) }

// title names a suggestion's house by what the question asks about there,
// such as "Small animals", or failing that by the house's title, such as
// "Illness".
func title(c *content.Set, s question.Suggestion) string {
	if t := s.Matter().Subject(); t != "" {
		return t
	}
	sig, _ := c.Signification(s.House)
	return sig.Title
}

// houseName names a suggestion's house for a hint: "⑥ Small animals", or
// in accessible prompts, which number houses themselves, "House 6 · Small
// animals".
func houseName(c *content.Set, s question.Suggestion, accessible bool) string {
	if accessible {
		return fmt.Sprintf("House %d · %s", s.House, title(c, s))
	}
	return string(rune('①'+s.House-1)) + " " + title(c, s)
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
		parts[i] = houseName(c, s, accessible) + ", " + reason(s)
	}
	return ss[0].House, "Suggested: " + strings.Join(parts, "; or ") + "."
}

type suggestionJSON struct {
	House   int                  `json:"house"`
	Title   string               `json:"title"`
	Topics  []string             `json:"topics,omitempty"`
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
				out[i] = suggestionJSON{House: s.House, Title: sig.Title, Topics: s.Topics, Score: s.Score, Matches: s.Matches, Derived: s.Derived}
			}
			return opts.print(cmd, out, func() string {
				if len(ss) == 0 {
					return "No suggestion: the significator prompt will start on House 1."
				}
				lines := []string{render.Heading("Suggestions for “" + q + "”")}
				for i, s := range ss {
					lines = append(lines, fmt.Sprintf("  %d. House %d · %s — %s", i+1, s.House, title(c, s), reason(s)))
				}
				return strings.Join(lines, "\n")
			})
		},
	}
}
