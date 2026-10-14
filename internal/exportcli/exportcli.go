// Package exportcli implements the geomancer-export command, which writes a
// reading for every possible chart under each system.
package exportcli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/internal/export"
	"github.com/storskegg/go-geomancer/system"
)

type options struct {
	format  string
	systems []string
	output  string
	force   bool

	// records produces the readings to export; tests replace it to export a
	// sample.
	records func([]system.System) iter.Seq2[export.Record, error]
}

// NewCommand returns the geomancer-export command. Systems must be
// registered (by importing their packages) before it runs.
func NewCommand() *cobra.Command { return newCommand(export.Records) }

func newCommand(records func([]system.System) iter.Seq2[export.Record, error]) *cobra.Command {
	opts := options{records: records}
	cmd := &cobra.Command{
		Use:   "geomancer-export",
		Short: "Export a reading for every possible chart under each system",
		Long: fmt.Sprintf(`Export a reading for every possible chart under each system.

Every set of four Mothers (%d in all) is cast under each system, and the
readings are written as CSV, JSON Lines, Parquet or SQLite.

The format defaults to CSV, or is taken from the output file's extension.
The output defaults to readings.<format>; use -o - for stdout (not SQLite).
Each chart has an ID from 0 to 65535 that packs its Mothers' 4-bit values,
M1 first.`, chart.ChartCount),
		Example: `  geomancer-export
  geomancer-export -o charts.parquet
  geomancer-export --format sqlite --system goldendawn
  geomancer-export --format jsonl -o - | head -1`,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          func(cmd *cobra.Command, _ []string) error { return opts.run(cmd) },
	}
	f := cmd.Flags()
	f.StringVarP(&opts.format, "format", "f", "", "output format: csv, jsonl, parquet or sqlite (default csv)")
	f.StringSliceVarP(&opts.systems, "system", "s", nil, "system to export; repeat for several (default all)")
	f.StringVarP(&opts.output, "output", "o", "", "output file, or - for stdout (default readings.<format>)")
	f.BoolVar(&opts.force, "force", false, "overwrite an existing output file")
	return cmd
}

// Execute runs the command with the process arguments and returns the exit
// code.
func Execute() int { return run(NewCommand(), os.Args[1:]) }

func run(cmd *cobra.Command, args []string) int {
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Error:", err)
		return 1
	}
	return 0
}

func (o *options) resolveFormat() (export.Format, error) {
	if o.format != "" {
		return export.ParseFormat(o.format)
	}
	if f, ok := export.FormatForPath(o.output); ok {
		return f, nil
	}
	return export.CSV, nil
}

func (o *options) resolveSystems() ([]system.System, error) {
	if len(o.systems) == 0 {
		return system.List(), nil
	}
	var systems []system.System
	seen := map[string]bool{}
	for _, id := range o.systems {
		s, ok := system.Get(id)
		if !ok {
			ids := make([]string, 0)
			for _, s := range system.List() {
				ids = append(ids, s.ID())
			}
			return nil, fmt.Errorf("unknown system %q (available: %s)", id, strings.Join(ids, ", "))
		}
		if !seen[s.ID()] {
			systems = append(systems, s)
			seen[s.ID()] = true
		}
	}
	return systems, nil
}

func (o *options) run(cmd *cobra.Command) error {
	format, err := o.resolveFormat()
	if err != nil {
		return err
	}
	systems, err := o.resolveSystems()
	if err != nil {
		return err
	}
	path := o.output
	if path == "" {
		path = "readings" + format.Extensions()[0]
	}

	start := time.Now()
	var n int
	if path == "-" {
		if !format.Streams() {
			return fmt.Errorf("%s output can't be written to stdout; give a file with -o", format)
		}
		n, err = o.exportTo(cmd.OutOrStdout(), format, systems)
		path = "stdout"
	} else {
		n, err = o.exportFile(cmd, path, format, systems)
	}
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(cmd.ErrOrStderr(), "Wrote %d readings for %d systems to %s in %s.\n",
		n, len(systems), path, time.Since(start).Round(time.Millisecond))
	return err
}

func (o *options) exportTo(w io.Writer, format export.Format, systems []system.System) (int, error) {
	ew, err := export.NewWriter(format, w)
	if err != nil {
		return 0, err
	}
	return export.Run(ew, o.records(systems))
}

// exportFile writes to a temporary file beside path and renames it into
// place once complete, so a failed export never leaves a partial file.
func (o *options) exportFile(cmd *cobra.Command, path string, format export.Format, systems []system.System) (int, error) {
	if _, err := os.Stat(path); err == nil && !o.force {
		return 0, fmt.Errorf("%s already exists; use --force to overwrite it", path)
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return 0, err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return 0, err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }() // no-op after a successful rename

	var n int
	if format.Streams() {
		n, err = o.exportTo(tmp, format, systems)
		if cerr := tmp.Close(); err == nil {
			err = cerr
		}
	} else {
		// SQLite opens the path itself and needs it to start empty.
		_ = tmp.Close()
		var w export.Writer
		if w, err = export.NewSQLiteWriter(cmd.Context(), tmpPath); err == nil {
			n, err = export.Run(w, o.records(systems))
		}
	}
	if err == nil {
		// CreateTemp makes the file private; give it ordinary permissions.
		err = os.Chmod(tmpPath, 0o644)
	}
	if err != nil {
		return 0, err
	}
	return n, os.Rename(tmpPath, path)
}
