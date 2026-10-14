package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/x/term"

	"github.com/spf13/cobra"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/internal/cli/input"
	"github.com/storskegg/go-geomancer/internal/cli/render"
)

func newCastCommand(opts *options) *cobra.Command {
	var nephews bool
	cmd := &cobra.Command{
		Use:   "cast [M1 M2 M3 M4]",
		Short: "Cast the Shield and Astrological Square from four Mothers",
		Long: `Cast the Shield and Astrological Square from four Mothers.

With no arguments, you are prompted to choose each Mother from a list;
type to filter it by name or dot pattern. Set ACCESSIBLE=1 for numbered,
line-based prompts instead.

The Mothers may also be given as arguments, each as a dot pattern read
head to feet ("1212"), as parities ("odd,even,odd,even"), or by name
("Amissio", "fortuna-major"). Quote names that contain spaces.`,
		Example: `  geomancer cast
  geomancer cast 1212 1122 1111 2222
  geomancer cast --system goldendawn amissio "fortuna minor" via populus`,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 0 && len(args) != 4 {
				return fmt.Errorf("expected 4 mothers or none to choose them interactively, got %d", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			sys, err := opts.lookupSystem()
			if err != nil {
				return err
			}
			src, err := mothersSource(cmd, args)
			if err != nil {
				return err
			}
			mothers, err := src.Mothers()
			if err != nil {
				return err
			}
			reading, err := chart.NewReading(mothers, sys)
			if err != nil {
				return err
			}
			return opts.print(cmd, reading, func() string {
				return strings.Join([]string{
					render.Heading("The Shield"),
					render.Shield(reading.Shield, render.Options{Nephews: nephews}),
					render.Heading("Astrological Square · " + sys.Name()),
					render.Square(reading.Square),
				}, "\n\n")
			})
		},
	}
	cmd.Flags().BoolVar(&nephews, "nephews", false, "call N1–N4 Nephews rather than Nieces")
	return cmd
}

// mothersSource reads the Mothers from args if given, and otherwise prompts
// for them. The prompt draws on stderr so that stdout carries only the
// reading, e.g. for --json.
func mothersSource(cmd *cobra.Command, args []string) (input.MothersSource, error) {
	if len(args) > 0 {
		return input.Args(args), nil
	}
	in, out := cmd.InOrStdin(), cmd.ErrOrStderr()
	if os.Getenv("ACCESSIBLE") != "" {
		return input.Prompt{In: in, Out: out, Accessible: true}, nil
	}
	if !isTerminal(in) || !isTerminal(out) {
		return nil, errors.New("no mothers given and not running in a terminal; pass four mothers as arguments")
	}
	return input.Prompt{In: in, Out: out}, nil
}

func isTerminal(v any) bool {
	f, ok := v.(*os.File)
	return ok && term.IsTerminal(f.Fd())
}
