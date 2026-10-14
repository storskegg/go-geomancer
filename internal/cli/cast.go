package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/internal/cli/input"
	"github.com/storskegg/go-geomancer/internal/cli/render"
)

func newCastCommand(opts *options) *cobra.Command {
	var nephews bool
	cmd := &cobra.Command{
		Use:   "cast M1 M2 M3 M4",
		Short: "Cast the Shield and Astrological Square from four Mothers",
		Long: `Cast the Shield and Astrological Square from four Mothers.

Each Mother may be given as a dot pattern read head to feet ("1212"),
as parities ("odd,even,odd,even"), or by name ("Amissio", "fortuna-major").
Quote names that contain spaces.`,
		Example: `  geomancer cast 1212 1122 1111 2222
  geomancer cast --system goldendawn amissio "fortuna minor" via populus`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			sys, err := opts.lookupSystem()
			if err != nil {
				return err
			}
			mothers, err := input.Args(args).Mothers()
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
