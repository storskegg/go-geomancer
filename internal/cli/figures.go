package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/storskegg/go-geomancer/content"
	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/internal/cli/render"
	"github.com/storskegg/go-geomancer/system"
	"github.com/storskegg/go-geomancer/zodiac"
)

// figureJSON is the JSON form of a figure under a system.
type figureJSON struct {
	Name     string          `json:"name"`
	Pattern  string          `json:"pattern"`
	Planet   zodiac.Planet   `json:"planet"`
	Sign     *zodiac.Sign    `json:"sign,omitempty"`
	Element  *zodiac.Element `json:"element,omitempty"`
	Dots     int             `json:"dots"`
	OddDots  int             `json:"oddDots"`
	EvenDots int             `json:"evenDots"`
	Content  *content.Entry  `json:"content,omitempty"`
}

func newFigureJSON(f figure.Figure, sys system.System) figureJSON {
	j := figureJSON{
		Name:     f.Name(),
		Pattern:  f.Pattern(),
		Planet:   f.Planet(),
		Dots:     f.Dots(),
		OddDots:  f.OddDots(),
		EvenDots: f.EvenDots(),
	}
	if s, ok := sys.Sign(f); ok {
		j.Sign = &s
	}
	if e, ok := sys.Element(f); ok {
		j.Element = &e
	}
	return j
}

func newFiguresCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "figures",
		Short: "List the sixteen figures and their correspondences",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			sys, err := opts.lookupSystem()
			if err != nil {
				return err
			}
			all := figure.All()
			figs := make([]figureJSON, 0, len(all))
			for _, f := range all {
				figs = append(figs, newFigureJSON(f, sys))
			}
			return opts.print(cmd, figs, func() string { return render.Figures(sys) })
		},
	}
}

func newFigureCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "figure NAME|PATTERN",
		Short: "Describe one figure, including its meaning in each house",
		Example: `  geomancer figure puer
  geomancer figure fortuna major
  geomancer figure 1212`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sys, err := opts.lookupSystem()
			if err != nil {
				return err
			}
			f, err := figure.Parse(strings.Join(args, " "))
			if err != nil {
				return err
			}
			entry := content.Default().For(f)
			j := newFigureJSON(f, sys)
			j.Content = &entry
			return opts.print(cmd, j, func() string { return render.Figure(f, sys, entry) })
		},
	}
}
