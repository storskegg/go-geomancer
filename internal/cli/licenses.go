package cli

import (
	"github.com/spf13/cobra"

	"github.com/storskegg/go-geomancer/internal/cli/render"
	"github.com/storskegg/go-geomancer/internal/licenses"
)

func newLicenseCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "license",
		Short: "List the modules geomancer is built from and their licenses",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			mods := licenses.All()
			return opts.print(cmd, mods, func() string { return render.Licenses(mods) })
		},
	}
}
