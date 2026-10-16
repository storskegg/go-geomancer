package export

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"iter"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/parquet-go/parquet-go"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/system"
	"github.com/storskegg/go-geomancer/system/goldendawn"
	"github.com/storskegg/go-geomancer/system/traditional"
	"github.com/storskegg/go-geomancer/zodiac"
)

// unsigned is a valid system that gives no figure a sign.
var unsigned = &system.Mapping{
	Slug:            "unsigned",
	Title:           "Unsigned",
	SegmentElements: [4]zodiac.Element{zodiac.Fire, zodiac.Air, zodiac.Water, zodiac.Earth},
	Houses:          [12]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
}

// sample returns the first n records for each system.
func sample(t *testing.T, n int, systems ...system.System) []Record {
	t.Helper()
	var recs []Record
	for _, sys := range systems {
		for id, m := range chart.AllMothers() {
			if id == n {
				break
			}
			r, err := chart.NewReading(m, sys)
			if err != nil {
				t.Fatal(err)
			}
			recs = append(recs, Record{Chart: id, System: sys, Reading: r})
		}
	}
	return recs
}

func seq(recs []Record) iter.Seq2[Record, error] {
	return func(yield func(Record, error) bool) {
		for _, r := range recs {
			if !yield(r, nil) {
				return
			}
		}
	}
}

// example is chart 0xACF0 (Amissio, Fortuna Minor, Via, Populus).
func example(t *testing.T, sys system.System) Record {
	t.Helper()
	m := [4]figure.Figure{figure.Amissio, figure.FortunaMinor, figure.Via, figure.Populus}
	r, err := chart.NewReading(m, sys)
	if err != nil {
		t.Fatal(err)
	}
	return Record{Chart: chart.ChartID(m), System: sys, Reading: r}
}

func TestRecords(t *testing.T) {
	systems := []system.System{traditional.System, goldendawn.System}
	n := 0
	for rec, err := range Records(systems) {
		if err != nil {
			t.Fatal(err)
		}
		wantSys := systems[n/chart.ChartCount]
		if rec.System != wantSys || rec.Chart != n%chart.ChartCount || rec.Reading.System != wantSys.ID() {
			t.Fatalf("record %d = chart %d, system %s", n, rec.Chart, rec.System.ID())
		}
		if got := chart.ChartID(rec.Reading.Shield.Mothers); got != rec.Chart {
			t.Fatalf("record %d: chart %d holds mothers of chart %d", n, rec.Chart, got)
		}
		n++
	}
	if n != 2*chart.ChartCount {
		t.Errorf("%d records, want %d", n, 2*chart.ChartCount)
	}
}

func TestRecordsStopsEarly(t *testing.T) {
	n := 0
	for range Records([]system.System{traditional.System}) {
		if n++; n == 5 {
			break
		}
	}
	if n != 5 {
		t.Errorf("iterated %d times", n)
	}
}

func TestRecordsInvalidSystem(t *testing.T) {
	bad := *unsigned
	bad.Slug = ""
	n := 0
	var last error
	for _, err := range Records([]system.System{&bad}) {
		n++
		last = err
	}
	if n != 1 || !errors.Is(last, system.ErrInvalid) {
		t.Errorf("yielded %d records, last error %v; want 1 and ErrInvalid", n, last)
	}
}

func TestColumns(t *testing.T) {
	cols := Columns()
	if len(cols) != 42 {
		t.Fatalf("%d columns, want 42", len(cols))
	}
	if cols[0] != "system" || cols[1] != "chart" || cols[2] != "m1" || cols[17] != "jm" ||
		cols[18] != "h1_figure" || cols[29] != "h12_figure" || cols[30] != "h1_sign" || cols[41] != "h12_sign" {
		t.Errorf("columns = %v", cols)
	}
	if len(slices.Compact(slices.Sorted(slices.Values(cols)))) != 42 {
		t.Error("duplicate column names")
	}

	// The Parquet schema, taken from Row's tags, has the same columns in the
	// same order.
	var fields []string
	for _, f := range parquet.SchemaOf(Row{}).Fields() {
		fields = append(fields, f.Name())
	}
	if !slices.Equal(fields, cols) {
		t.Errorf("parquet columns = %v\nwant %v", fields, cols)
	}
}

func TestNewRow(t *testing.T) {
	r := NewRow(example(t, goldendawn.System))
	if r.System != "goldendawn" || r.Chart != 0xACF0 {
		t.Errorf("system %q, chart %#x", r.System, r.Chart)
	}
	if r.M1 != "Amissio" || r.D1 != "Cauda Draconis" || r.N3 != "Laetitia" || r.J != "Carcer" || r.JM != "Fortuna Major" {
		t.Errorf("shield = %+v", r)
	}
	// Golden Dawn puts M2 (Fortuna Minor, Leo) in House 1 and M1 in House 10.
	if r.H1Figure != "Fortuna Minor" || r.H10Figure != "Amissio" {
		t.Errorf("houses 1 and 10 = %q, %q", r.H1Figure, r.H10Figure)
	}
	if r.H1Sign == nil || *r.H1Sign != "Leo" || r.H12Sign == nil || *r.H12Sign != "Cancer" {
		t.Errorf("signs = %v … %v", r.H1Sign, r.H12Sign)
	}

	vals := r.values()
	if len(vals) != len(Columns()) || vals[1] != strconv.Itoa(0xACF0) || vals[30] != "Leo" {
		t.Errorf("values = %v", vals)
	}
}

func TestNewRowNoSigns(t *testing.T) {
	r := NewRow(example(t, unsigned))
	for i, s := range r.houseSigns() {
		if *s != nil {
			t.Errorf("house %d sign = %q, want nil", i+1, **s)
		}
	}
	if vals := r.values(); vals[30] != "" || vals[41] != "" {
		t.Errorf("sign values = %q … %q, want empty", vals[30], vals[41])
	}
}

func TestFormats(t *testing.T) {
	for _, f := range Formats() {
		got, err := ParseFormat(strings.ToUpper(string(f)))
		if err != nil || got != f {
			t.Errorf("ParseFormat(%q) = %v, %v", strings.ToUpper(string(f)), got, err)
		}
		for _, ext := range f.Extensions() {
			if got, ok := FormatForPath("out/readings" + strings.ToUpper(ext)); !ok || got != f {
				t.Errorf("FormatForPath(%s) = %v, %v", ext, got, ok)
			}
		}
		if f.Streams() != (f != SQLite) {
			t.Errorf("%s.Streams() = %v", f, f.Streams())
		}
	}
	if Formats()[0] != CSV {
		t.Error("CSV is not the default format")
	}
	if _, err := ParseFormat("xml"); err == nil || !strings.Contains(err.Error(), "csv, jsonl, parquet, sqlite") {
		t.Errorf("ParseFormat(xml) error = %v", err)
	}
	if _, ok := FormatForPath("readings.txt"); ok {
		t.Error("FormatForPath(readings.txt) ok")
	}
}

func TestNewWriterErrors(t *testing.T) {
	if _, err := NewWriter(SQLite, &bytes.Buffer{}); err == nil {
		t.Error("NewWriter(sqlite) succeeded")
	}
	if _, err := NewWriter("xml", &bytes.Buffer{}); err == nil {
		t.Error("NewWriter(xml) succeeded")
	}
}

func TestCSV(t *testing.T) {
	recs := append(sample(t, 3, traditional.System), example(t, unsigned))
	var buf bytes.Buffer
	w, _ := NewWriter(CSV, &buf)
	if n, err := Run(w, seq(recs)); err != nil || n != len(recs) {
		t.Fatalf("Run = %d, %v", n, err)
	}

	rows, err := csv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(rows[0], Columns()) {
		t.Errorf("header = %v", rows[0])
	}
	if len(rows) != 1+len(recs) {
		t.Fatalf("%d rows, want %d", len(rows), 1+len(recs))
	}
	for i, rec := range recs {
		row := NewRow(rec)
		if !slices.Equal(rows[i+1], row.values()) {
			t.Errorf("row %d = %v\nwant %v", i+1, rows[i+1], row.values())
		}
	}
	// "Fortuna Minor" contains a space but needs no quoting; names never
	// contain commas.
	if strings.Contains(buf.String(), `"`) {
		t.Error("CSV contains quotes")
	}
}

func TestCSVEmpty(t *testing.T) {
	var buf bytes.Buffer
	w, _ := NewWriter(CSV, &buf)
	if n, err := Run(w, seq(nil)); err != nil || n != 0 {
		t.Fatalf("Run = %d, %v", n, err)
	}
	if got := strings.TrimSpace(buf.String()); got != strings.Join(Columns(), ",") {
		t.Errorf("empty export = %q, want just the header", got)
	}
}

func TestJSONL(t *testing.T) {
	recs := sample(t, 3, traditional.System, goldendawn.System)
	var buf bytes.Buffer
	w, _ := NewWriter(JSONL, &buf)
	if _, err := Run(w, seq(recs)); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) != len(recs) {
		t.Fatalf("%d lines, want %d", len(lines), len(recs))
	}
	for i, line := range lines {
		if !strings.HasPrefix(line, `{"chart":`+strconv.Itoa(recs[i].Chart)+`,"system":"`+recs[i].System.ID()+`"`) {
			t.Errorf("line %d starts %.40s", i, line)
		}
		var got jsonlRecord
		if err := json.Unmarshal([]byte(line), &got); err != nil {
			t.Fatal(err)
		}
		if got.Chart != recs[i].Chart || got.Shield != recs[i].Reading.Shield || got.System != recs[i].Reading.System {
			t.Errorf("line %d = %+v", i, got)
		}
		for h := range got.Square {
			if got.Square[h].Figure != recs[i].Reading.Square[h].Figure {
				t.Errorf("line %d house %d differs", i, h+1)
			}
		}
	}
}

func TestParquet(t *testing.T) {
	// More than one batch, plus a record without signs.
	recs := append(sample(t, parquetBatch+10, traditional.System), example(t, unsigned))
	var buf bytes.Buffer
	w, _ := NewWriter(Parquet, &buf)
	if n, err := Run(w, seq(recs)); err != nil || n != len(recs) {
		t.Fatalf("Run = %d, %v", n, err)
	}

	rows, err := parquet.Read[Row](bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != len(recs) {
		t.Fatalf("read %d rows, want %d", len(rows), len(recs))
	}
	for i, rec := range recs {
		want := NewRow(rec)
		if !slices.Equal(rows[i].values(), want.values()) {
			t.Errorf("row %d = %v\nwant %v", i, rows[i].values(), want.values())
		}
	}
	if last := rows[len(rows)-1]; last.H1Sign != nil {
		t.Errorf("unsigned row has sign %q", *last.H1Sign)
	}
}

func openSQLite(t *testing.T, records iter.Seq2[Record, error]) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "readings.db")
	w, err := NewSQLiteWriter(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Run(w, records); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func queryInt(t *testing.T, db *sql.DB, q string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(q).Scan(&n); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
	return n
}

func TestSQLite(t *testing.T) {
	recs := append(sample(t, 5, traditional.System, goldendawn.System), example(t, traditional.System), example(t, unsigned))
	db := openSQLite(t, seq(recs))

	counts := map[string]int{
		"figures":        16,
		"signs":          12,
		"positions":      16,
		"systems":        3,
		"system_figures": 3 * 16,
		"system_houses":  3 * 12,
		"charts":         6, // charts 0–4 and 0xACF0, shared across systems
		"chart_houses":   12 * len(recs),
	}
	for table, want := range counts {
		if got := queryInt(t, db, "SELECT count(*) FROM "+table); got != want {
			t.Errorf("%s has %d rows, want %d", table, got, want)
		}
	}

	// Lookup tables match the core.
	var name, pattern string
	if err := db.QueryRow("SELECT name, pattern FROM figures WHERE id = ?", int(figure.FortunaMajor)).Scan(&name, &pattern); err != nil || name != "Fortuna Major" || pattern != "2211" {
		t.Errorf("figure %d = %q %q, %v", figure.FortunaMajor, name, pattern, err)
	}
	if got := queryInt(t, db, "SELECT house FROM system_houses WHERE system_id = 'goldendawn' AND position = 'M1'"); got != 10 {
		t.Errorf("Golden Dawn M1 in house %d, want 10", got)
	}
	if got := queryInt(t, db, "SELECT count(*) FROM system_figures WHERE system_id = 'goldendawn' AND sign_id IS NULL"); got != 0 {
		t.Errorf("%d Golden Dawn figures without signs, want 0", got)
	}
	var node string
	if err := db.QueryRow("SELECT planet FROM system_figures WHERE system_id = 'goldendawn' AND figure_id = ?", int(figure.CaputDraconis)).Scan(&node); err != nil || node != "North Node" {
		t.Errorf("Caput Draconis planet = %q, %v", node, err)
	}
	if got := queryInt(t, db, "SELECT count(*) FROM system_figures WHERE system_id = 'unsigned' AND sign_id IS NULL AND element IS NULL AND planet IS NULL"); got != 16 {
		t.Errorf("%d unmapped figures for the unsigned system, want 16", got)
	}

	// The example chart's Shield and Square.
	var j, jm int
	if err := db.QueryRow("SELECT j, jm FROM charts WHERE id = ?", 0xACF0).Scan(&j, &jm); err != nil || j != int(figure.Carcer) || jm != int(figure.FortunaMajor) {
		t.Errorf("chart 0xACF0 judge %d, judgement %d, %v", j, jm, err)
	}
	// Traditional puts M1 (Amissio) in House 1, which takes Amissio's sign.
	var house1Sign string
	want, _ := traditional.System.Sign(figure.Amissio)
	if err := db.QueryRow(`SELECT s.name FROM chart_houses h JOIN signs s ON s.id = h.sign_id
		WHERE h.system_id = 'traditional' AND h.chart_id = ? AND h.house = 1`, 0xACF0).Scan(&house1Sign); err != nil || house1Sign != want.String() {
		t.Errorf("traditional house 1 sign = %q, %v; want %v", house1Sign, err, want)
	}
	if got := queryInt(t, db, "SELECT count(*) FROM chart_houses WHERE system_id = 'unsigned' AND sign_id IS NOT NULL"); got != 0 {
		t.Errorf("%d signed houses for the unsigned system", got)
	}

	// Every reference resolves, and the indexes exist.
	rows, err := db.Query("PRAGMA foreign_key_check")
	if err != nil {
		t.Fatal(err)
	}
	if rows.Next() {
		t.Error("foreign key violations")
	}
	_ = rows.Close()
	if got := queryInt(t, db, "SELECT count(*) FROM sqlite_master WHERE type = 'index' AND name IN ('chart_houses_by_figure', 'charts_by_judge')"); got != 2 {
		t.Errorf("%d of 2 indexes", got)
	}
}

func TestSQLiteRejectsNonEmptyDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "readings.db")
	w, err := NewSQLiteWriter(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := NewSQLiteWriter(context.Background(), path); err == nil || !strings.Contains(err.Error(), "not empty") {
		t.Errorf("second NewSQLiteWriter error = %v", err)
	}
}

func TestSQLiteBadPath(t *testing.T) {
	if _, err := NewSQLiteWriter(context.Background(), filepath.Join(t.TempDir(), "missing", "readings.db")); err == nil {
		t.Error("NewSQLiteWriter in a missing directory succeeded")
	}
}

// failWriter fails on the nth write and records whether it was closed.
type failWriter struct {
	n, writes int
	closed    bool
}

func (f *failWriter) Write(Record) error {
	if f.writes++; f.writes == f.n {
		return errors.New("disk full")
	}
	return nil
}

func (f *failWriter) Close() error { f.closed = true; return nil }

func TestRunStopsOnError(t *testing.T) {
	w := &failWriter{n: 3}
	n, err := Run(w, seq(sample(t, 10, traditional.System)))
	if err == nil || n != 2 || !w.closed {
		t.Errorf("Run = %d, %v; closed = %v", n, err, w.closed)
	}

	bad := *unsigned
	bad.Slug = ""
	w = &failWriter{}
	if _, err := Run(w, Records([]system.System{&bad})); !errors.Is(err, system.ErrInvalid) || !w.closed {
		t.Errorf("Run error = %v, closed = %v", err, w.closed)
	}
}

// TestFullExport writes every chart for both systems in each format and
// checks the counts. It takes several seconds, so -short skips it.
func TestFullExport(t *testing.T) {
	if testing.Short() {
		t.Skip("full export is slow")
	}
	systems := []system.System{goldendawn.System, traditional.System}
	want := len(systems) * chart.ChartCount
	for _, f := range []Format{CSV, JSONL, Parquet} {
		t.Run(string(f), func(t *testing.T) {
			var buf bytes.Buffer
			w, _ := NewWriter(f, &buf)
			if n, err := Run(w, Records(systems)); err != nil || n != want {
				t.Fatalf("Run = %d, %v", n, err)
			}
			if f == Parquet {
				pf, err := parquet.OpenFile(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
				if err != nil || pf.NumRows() != int64(want) {
					t.Errorf("parquet rows = %d, %v", pf.NumRows(), err)
				}
				return
			}
			lines := bytes.Count(buf.Bytes(), []byte("\n"))
			if f == CSV {
				lines-- // header
			}
			if lines != want {
				t.Errorf("%d records, want %d", lines, want)
			}
		})
	}
	t.Run("sqlite", func(t *testing.T) {
		db := openSQLite(t, Records(systems))
		if got := queryInt(t, db, "SELECT count(*) FROM charts"); got != chart.ChartCount {
			t.Errorf("%d charts", got)
		}
		if got := queryInt(t, db, "SELECT count(*) FROM chart_houses"); got != 12*want {
			t.Errorf("%d chart houses", got)
		}
		// Only figures with an even number of dots can be the Judge.
		if got := queryInt(t, db, "SELECT count(*) FROM charts c JOIN figures f ON f.id = c.j WHERE f.dots % 2 = 1"); got != 0 {
			t.Errorf("%d charts with an odd Judge", got)
		}
	})
}

// brokenWriter fails every write.
type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, errors.New("broken pipe") }

func TestStreamWriteErrors(t *testing.T) {
	// Enough rows to overflow each writer's buffer.
	recs := sample(t, parquetBatch+1, traditional.System)
	for _, f := range []Format{CSV, JSONL, Parquet} {
		w, _ := NewWriter(f, brokenWriter{})
		if _, err := Run(w, seq(recs)); err == nil || !strings.Contains(err.Error(), "broken pipe") {
			t.Errorf("%s: Run error = %v, want broken pipe", f, err)
		}
	}
	// Close fails too when there's nothing but a header to write.
	w, _ := NewWriter(CSV, brokenWriter{})
	if err := w.Close(); err == nil {
		t.Error("CSV Close on a broken writer succeeded")
	}
}

func TestSQLiteDuplicateRecord(t *testing.T) {
	w, err := NewSQLiteWriter(context.Background(), filepath.Join(t.TempDir(), "readings.db"))
	if err != nil {
		t.Fatal(err)
	}
	rec := example(t, traditional.System)
	if _, err := Run(w, seq([]Record{rec, rec})); err == nil || !strings.Contains(err.Error(), "houses") {
		t.Errorf("Run error = %v, want a chart_houses conflict", err)
	}
}

func TestSQLiteCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := NewSQLiteWriter(ctx, filepath.Join(t.TempDir(), "readings.db")); err == nil {
		t.Error("NewSQLiteWriter with a cancelled context succeeded")
	}

	// Cancelling partway through fails the next insert: first the system,
	// then a new chart, then a known chart's houses.
	recs := []Record{example(t, traditional.System), example(t, goldendawn.System)}
	tests := []struct {
		name   string
		before int // records written before cancelling
		next   Record
		want   string
	}{
		{"system", 0, recs[0], "system traditional"},
		{"chart", 1, sample(t, 1, traditional.System)[0], "chart 0"},
		{"houses", 1, Record{Chart: recs[0].Chart, System: traditional.System, Reading: sample(t, 1, traditional.System)[0].Reading}, "houses"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			w, err := NewSQLiteWriter(ctx, filepath.Join(t.TempDir(), "readings.db"))
			if err != nil {
				t.Fatal(err)
			}
			for _, r := range recs[:tt.before] {
				if err := w.Write(r); err != nil {
					t.Fatal(err)
				}
			}
			cancel()
			if err := w.Write(tt.next); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Write error = %v, want it to mention %q", err, tt.want)
			}
			if err := w.Close(); err == nil {
				t.Error("Close after cancelling succeeded")
			}
		})
	}
}
