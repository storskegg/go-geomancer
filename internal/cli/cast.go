package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/x/term"

	"github.com/spf13/cobra"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/content"
	"github.com/storskegg/go-geomancer/internal/cli/input"
	"github.com/storskegg/go-geomancer/internal/cli/render"
	"github.com/storskegg/go-geomancer/place"
	"github.com/storskegg/go-geomancer/report"
	"github.com/storskegg/go-geomancer/system"
)

func newCastCommand(opts *options) *cobra.Command {
	var (
		nephews      bool
		significator string
		question     string
		pdfPath      string
		occasion     occasionFlags
	)
	cmd := &cobra.Command{
		Use:   "cast [M1 M2 M3 M4]",
		Short: "Cast the Shield and Astrological Square from four Mothers",
		Long: `Cast the Shield and Astrological Square from four Mothers.

With no arguments, you are prompted to choose the system (unless --system
is given), for the date, local time and place of the reading (unless
--date, --time and --place are given), for your question (unless
--question is given; it's optional),
for the significator, the house of what you're asking about (unless
--significator is given), and for each Mother. The significator prompt
starts on the house your question suggests; see 'geomancer suggest'. Type to filter the
houses by subject and the figures by name or dot pattern. Set ACCESSIBLE=1
for numbered, line-based prompts instead.

The Mothers may also be given as arguments, each as a dot pattern read
head to feet ("1212"), as parities ("odd,even,odd,even"), or by name
("Amissio", "fortuna-major"). Quote names that contain spaces.

With --pdf, a printable report is also written to the given file.

Given the Mothers as arguments, the date, time and place come only from
their flags, all three or none. A time that happened twice, when clocks
went back, needs its zone: --time "1:30 am CST".`,
		Example: `  geomancer cast
  geomancer cast 1212 1122 1111 2222
  geomancer cast --significator partnership 1212 1122 1111 2222
  geomancer cast --question "Will the contract be signed?" -q 7 --pdf reading.pdf 1212 1122 1111 2222
  geomancer cast --date 2026-10-15 --time "3:30 pm" --place "Springfield, IL" 1212 1122 1111 2222
  geomancer cast --system traditional amissio "fortuna minor" via populus`,
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
			symbols, err := render.ParseSymbols(opts.symbols)
			if err != nil {
				return err
			}
			ropts := render.Options{Nephews: nephews, Symbols: symbols}
			if pdfPath != "" {
				// Fail before the prompts rather than after.
				if err := checkDir(pdfPath); err != nil {
					return err
				}
			}
			sig := 0
			if significator != "" {
				s, ok := content.Default().FindSignification(significator)
				if !ok {
					return fmt.Errorf("unknown significator %q: want a house number 1–12 or a title such as %q", significator, "partnership")
				}
				sig = s.House
			}
			interactive := len(args) == 0
			if interactive || occasion.place != "" {
				place.Preload()
			}
			occ, need, err := readOccasion(occasion, interactive, time.Now(), cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			src, err := mothersSource(cmd, args, input.Prompt{
				System: &sys, Significator: &sig, Question: &question,
				Occasion: occ, AskDate: need.date, AskTime: need.time, AskPlace: need.place,
				Places: places,
			})
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
			if err := reading.SetSignificator(sig); err != nil {
				return err
			}
			reading.Question = strings.TrimSpace(question)
			if reading.Question != "" && reading.Significator != 0 {
				reading.Matter = matter(reading.Question, reading.Significator)
			}
			if occ != nil {
				c := occ.Place.Place.Circumstances(occ.Time())
				reading.Circumstances = &c
			}

			err = opts.print(cmd, reading, func() string {
				square := render.Heading("Astrological Square · " + sys.Name())
				if s, ok := content.Default().Signification(reading.Significator); ok {
					square += "\n" + render.Significator(s, reading.Matter) + "\n" + render.Aspects(reading.Aspects, ropts)
				}
				var sections []string
				if head := render.Header(reading.Question, reading.Circumstances); head != "" {
					sections = append(sections, head)
				}
				return strings.Join(append(sections,
					render.Heading("The Shield"),
					render.Shield(reading.Shield, ropts),
					square,
					render.Square(reading.Square, reading.Significator, ropts),
					render.Heading("Reading"),
					render.Meanings(reading, content.Default(), ropts),
				), "\n\n")
			})
			if err != nil || pdfPath == "" {
				return err
			}
			if err := writePDF(pdfPath, reading, report.Options{Nephews: nephews}); err != nil {
				return fmt.Errorf("writing %s: %w", pdfPath, err)
			}
			_, err = fmt.Fprintf(cmd.ErrOrStderr(), "Wrote the report to %s.\n", pdfPath)
			return err
		},
	}
	cmd.Flags().BoolVar(&nephews, "nephews", false, "call N1–N4 Nephews rather than Nieces")
	cmd.Flags().StringVarP(&significator, "significator", "q", "", "house of the question: a number 1–12 or a title, such as partnership")
	cmd.Flags().StringVar(&question, "question", "", "the question, in your own words")
	cmd.Flags().StringVar(&occasion.date, "date", "", "the date of the reading, such as 2026-10-15 or \"October 15, 2026\"")
	cmd.Flags().StringVar(&occasion.time, "time", "", "the local time of the reading, such as \"3:30 pm\"")
	cmd.Flags().StringVar(&occasion.place, "place", "", "where the querent was: \"City, ST\", such as \"Springfield, IL\"")
	cmd.Flags().StringVar(&pdfPath, "pdf", "", "also write a printable report to `FILE`")
	return cmd
}

// mothersSource reads the Mothers from args if given, and otherwise prompts
// for them, and for the system, the occasion, the question and the
// significator unless their flags were given. p holds where the answers go.
// The prompt draws on stderr so that stdout carries only the reading, e.g.
// for --json.
func mothersSource(cmd *cobra.Command, args []string, p input.Prompt) (input.MothersSource, error) {
	if len(args) > 0 {
		return input.Args(args), nil
	}
	p.In, p.Out, p.Accessible = cmd.InOrStdin(), cmd.ErrOrStderr(), os.Getenv("ACCESSIBLE") != ""
	if !p.Accessible && (!isTerminal(p.In) || !isTerminal(p.Out)) {
		return nil, errors.New("no mothers given and not running in a terminal; pass four mothers as arguments")
	}
	if !cmd.Flags().Changed("system") {
		p.Systems = system.List()
	}
	p.AskQuestion = !cmd.Flags().Changed("question")
	if !cmd.Flags().Changed("significator") {
		sigs := content.Default().Significations()
		p.Significations = sigs[:]
		p.Suggest = func(q string) (int, string) { return suggestHint(q, p.Accessible) }
	}
	return p, nil
}

// checkDir reports whether path's directory exists, so a file can be
// written there.
func checkDir(path string) error {
	dir := filepath.Dir(path)
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("can't write %s: %w", path, errors.Unwrap(err))
	}
	if !info.IsDir() {
		return fmt.Errorf("can't write %s: %s is not a directory", path, dir)
	}
	return nil
}

// writePDF writes the reading's report to path through a temporary file
// beside it, so a failure never leaves a partial file.
func writePDF(path string, r chart.Reading, opts report.Options) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return errors.Unwrap(err) // the error names the temporary file
	}
	defer func() { _ = os.Remove(tmp.Name()) }() // no-op after a successful rename
	err = report.WritePDF(tmp, r, opts)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(tmp.Name(), 0o644)
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func isTerminal(v any) bool {
	f, ok := v.(*os.File)
	return ok && term.IsTerminal(f.Fd())
}
