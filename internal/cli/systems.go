package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/internal/cli/render"
	"github.com/storskegg/go-geomancer/system"
	"github.com/storskegg/go-geomancer/zodiac"
)

type systemJSON struct {
	ID       string                    `json:"id"`
	Name     string                    `json:"name"`
	Segments map[string]zodiac.Element `json:"segments"`
}

func newSystemsCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "systems",
		Short: "List the available systems",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			list := system.List()
			systems := make([]systemJSON, 0, len(list))
			for _, s := range list {
				j := systemJSON{ID: s.ID(), Name: s.Name(), Segments: map[string]zodiac.Element{}}
				for _, seg := range figure.Segments() {
					j.Segments[strings.ToLower(seg.String())] = s.SegmentElement(seg)
				}
				systems = append(systems, j)
			}
			return opts.print(cmd, systems, func() string { return render.Systems(list) })
		},
	}
}
