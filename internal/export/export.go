// Package export casts every possible chart under a set of systems and
// writes the readings as CSV, JSON Lines, Parquet or SQLite.
package export

import (
	"fmt"
	"io"
	"iter"
	"strings"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/position"
	"github.com/storskegg/go-geomancer/system"
)

// Record is one reading to export.
type Record struct {
	Chart   int // chart.ChartID of the Mothers
	System  system.System
	Reading chart.Reading
}

// Records yields a reading for every set of Mothers under each system, in
// the order given, and then in chart ID order.
func Records(systems []system.System) iter.Seq2[Record, error] {
	return func(yield func(Record, error) bool) {
		for _, sys := range systems {
			for id, mothers := range chart.AllMothers() {
				r, err := chart.NewReading(mothers, sys)
				if !yield(Record{Chart: id, System: sys, Reading: r}, err) || err != nil {
					return
				}
			}
		}
	}
}

// Writer writes records in one format. Close flushes buffered output and
// finishes the file, but doesn't close the underlying io.Writer.
type Writer interface {
	Write(Record) error
	Close() error
}

// Run writes records to w, closes w, and returns the number written. It
// stops at the first error.
func Run(w Writer, records iter.Seq2[Record, error]) (int, error) {
	n := 0
	for rec, err := range records {
		if err == nil {
			err = w.Write(rec)
		}
		if err != nil {
			_ = w.Close()
			return n, err
		}
		n++
	}
	return n, w.Close()
}

// Format is an output format.
type Format string

const (
	CSV     Format = "csv"
	JSONL   Format = "jsonl"
	Parquet Format = "parquet"
	SQLite  Format = "sqlite"
)

// Formats returns every format, the default first.
func Formats() []Format { return []Format{CSV, JSONL, Parquet, SQLite} }

// ParseFormat parses a format name, ignoring case.
func ParseFormat(s string) (Format, error) {
	for _, f := range Formats() {
		if strings.EqualFold(s, string(f)) {
			return f, nil
		}
	}
	return "", fmt.Errorf("unknown format %q (want %s)", s, formatList())
}

// FormatForPath returns the format implied by a file extension, if any.
func FormatForPath(path string) (Format, bool) {
	for _, f := range Formats() {
		for _, ext := range f.Extensions() {
			if strings.HasSuffix(strings.ToLower(path), ext) {
				return f, true
			}
		}
	}
	return "", false
}

// Extensions returns the file extensions for f, the preferred one first.
func (f Format) Extensions() []string {
	if f == SQLite {
		return []string{".sqlite", ".sqlite3", ".db"}
	}
	return []string{"." + string(f)}
}

// Streams reports whether f can be written to an io.Writer, such as stdout.
// SQLite needs a file path.
func (f Format) Streams() bool { return f != SQLite }

// NewWriter returns a writer for a streaming format.
func NewWriter(f Format, w io.Writer) (Writer, error) {
	switch f {
	case CSV:
		return newCSVWriter(w), nil
	case JSONL:
		return newJSONLWriter(w), nil
	case Parquet:
		return newParquetWriter(w), nil
	case SQLite:
		return nil, fmt.Errorf("%s output needs a file path", f)
	}
	return nil, fmt.Errorf("unknown format %q (want %s)", f, formatList())
}

func formatList() string {
	names := make([]string, 0, len(Formats()))
	for _, f := range Formats() {
		names = append(names, string(f))
	}
	return strings.Join(names, ", ")
}

// NewRow flattens a record.
func NewRow(rec Record) Row {
	r := Row{System: rec.System.ID(), Chart: int32(rec.Chart)}
	for i, f := range r.shield() {
		*f = rec.Reading.Shield.At(position.Position(i)).Name()
	}
	figs, signs := r.houseFigures(), r.houseSigns()
	for i, h := range rec.Reading.Square {
		*figs[i] = h.Figure.Name()
		if h.Sign != nil {
			name := h.Sign.String()
			*signs[i] = &name
		}
	}
	return r
}
