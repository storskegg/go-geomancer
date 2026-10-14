// Package cli implements the geomancer command-line interface.
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/colorprofile"
	"github.com/spf13/cobra"

	"github.com/storskegg/go-geomancer/internal/cli/input"
	"github.com/storskegg/go-geomancer/system"
)

const defaultSystem = "traditional"

type options struct {
	system string
	color  string
	json   bool
}

// NewCommand returns the root geomancer command. Systems must be registered
// (by importing their packages) before it runs.
func NewCommand() *cobra.Command {
	var opts options
	root := &cobra.Command{
		Use:           "geomancer",
		Short:         "Cast and interpret geomantic charts",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	f := root.PersistentFlags()
	f.StringVarP(&opts.system, "system", "s", defaultSystem, "system of correspondences (see 'geomancer systems')")
	f.StringVar(&opts.color, "color", "auto", "colour output: auto, always or never")
	f.BoolVar(&opts.json, "json", false, "print JSON instead of text")

	root.AddCommand(
		newCastCommand(&opts),
		newFiguresCommand(&opts),
		newFigureCommand(&opts),
		newSystemsCommand(&opts),
	)
	return root
}

// Execute runs the CLI and returns the process exit code.
func Execute() int {
	cmd := NewCommand()
	if err := cmd.Execute(); err != nil {
		if errors.Is(err, input.ErrAborted) {
			return 130 // conventional exit code for an interrupted command
		}
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Error:", err)
		return 1
	}
	return 0
}

func (o *options) lookupSystem() (system.System, error) {
	if s, ok := system.Get(o.system); ok {
		return s, nil
	}
	ids := make([]string, 0)
	for _, s := range system.List() {
		ids = append(ids, s.ID())
	}
	return nil, fmt.Errorf("unknown system %q (available: %s)", o.system, strings.Join(ids, ", "))
}

// output returns a writer for styled text that strips or downsamples
// styling to suit the destination and the --color flag.
func (o *options) output(cmd *cobra.Command) (io.Writer, error) {
	w := colorprofile.NewWriter(cmd.OutOrStdout(), os.Environ())
	switch o.color {
	case "auto":
	case "always":
		if w.Profile < colorprofile.ANSI {
			w.Profile = colorprofile.ANSI
		}
	case "never":
		w.Profile = colorprofile.NoTTY
	default:
		return nil, fmt.Errorf("invalid --color %q: want auto, always or never", o.color)
	}
	return w, nil
}

// print writes v as JSON with --json, and otherwise renders it with text.
func (o *options) print(cmd *cobra.Command, v any, text func() string) error {
	if o.json {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}
	w, err := o.output(cmd)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, text())
	return err
}
