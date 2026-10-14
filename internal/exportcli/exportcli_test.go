package exportcli

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"iter"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/parquet-go/parquet-go"
	"github.com/spf13/cobra"

	"github.com/storskegg/go-geomancer/internal/export"
	"github.com/storskegg/go-geomancer/system"
	_ "github.com/storskegg/go-geomancer/system/goldendawn"
	_ "github.com/storskegg/go-geomancer/system/traditional"
)

// sampleCharts is how many charts per system the tests export.
const sampleCharts = 10

// sampleRecords is export.Records limited to the first few charts of each
// system.
func sampleRecords(systems []system.System) iter.Seq2[export.Record, error] {
	return func(yield func(export.Record, error) bool) {
		for _, sys := range systems {
			for rec, err := range export.Records([]system.System{sys}) {
				if rec.Chart == sampleCharts {
					break
				}
				if !yield(rec, err) || err != nil {
					return
				}
			}
		}
	}
}

func newTestCommand() (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	cmd := newCommand(sampleRecords)
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	return cmd, &out, &errOut
}

func execute(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd, out, errOut := newTestCommand()
	cmd.SetArgs(args)
	err = cmd.Execute()
	return out.String(), errOut.String(), err
}

func readCSV(t *testing.T, path string) [][]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

func TestDefaultsToCSV(t *testing.T) {
	t.Chdir(t.TempDir())
	_, stderr, err := execute(t)
	if err != nil {
		t.Fatal(err)
	}
	rows := readCSV(t, "readings.csv")
	if len(rows) != 1+2*sampleCharts {
		t.Errorf("%d rows, want header + %d", len(rows), 2*sampleCharts)
	}
	// Systems in ID order.
	if rows[1][0] != "goldendawn" || rows[len(rows)-1][0] != "traditional" {
		t.Errorf("systems %s … %s", rows[1][0], rows[len(rows)-1][0])
	}
	if !strings.HasPrefix(stderr, "Wrote 20 readings for 2 systems to readings.csv in ") {
		t.Errorf("stderr = %q", stderr)
	}
}

func TestFormatFromExtension(t *testing.T) {
	dir := t.TempDir()
	tests := []struct {
		file  string
		check func(t *testing.T, path string)
	}{
		{"out.jsonl", func(t *testing.T, path string) {
			b, _ := os.ReadFile(path)
			if n := bytes.Count(b, []byte("\n")); n != 2*sampleCharts || !bytes.HasPrefix(b, []byte(`{"chart":0,`)) {
				t.Errorf("%d lines, starting %.20s", n, b)
			}
		}},
		{"out.parquet", func(t *testing.T, path string) {
			rows, err := parquet.ReadFile[export.Row](path)
			if err != nil || len(rows) != 2*sampleCharts {
				t.Errorf("read %d rows, %v", len(rows), err)
			}
		}},
		{"out.db", func(t *testing.T, path string) {
			db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = db.Close() }()
			var n int
			if err := db.QueryRow("SELECT count(*) FROM charts").Scan(&n); err != nil || n != sampleCharts {
				t.Errorf("charts = %d, %v", n, err)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			path := filepath.Join(dir, tt.file)
			if _, _, err := execute(t, "-o", path); err != nil {
				t.Fatal(err)
			}
			tt.check(t, path)
		})
	}
}

func TestFormatFlagOverridesExtension(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	if _, _, err := execute(t, "-f", "jsonl", "-o", path); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(path); !bytes.HasPrefix(b, []byte("{")) {
		t.Errorf("file starts %.20q, want JSON", b)
	}
}

func TestDefaultNameFollowsFormat(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, f := range []string{"parquet", "sqlite"} {
		if _, _, err := execute(t, "--format", f); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat("readings." + f); err != nil {
			t.Errorf("--format %s: %v", f, err)
		}
	}
}

func TestStdout(t *testing.T) {
	stdout, stderr, err := execute(t, "-o", "-", "-s", "traditional")
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.Count(stdout, "\n"); lines != 1+sampleCharts {
		t.Errorf("%d lines on stdout, want %d", lines, 1+sampleCharts)
	}
	if !strings.Contains(stderr, "to stdout") {
		t.Errorf("stderr = %q", stderr)
	}

	if _, _, err := execute(t, "-o", "-", "-f", "sqlite"); err == nil || !strings.Contains(err.Error(), "stdout") {
		t.Errorf("sqlite to stdout: error = %v", err)
	}
}

func TestSystems(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	if _, _, err := execute(t, "-o", path, "-s", "Traditional", "--system", "traditional"); err != nil {
		t.Fatal(err)
	}
	rows := readCSV(t, path)
	if len(rows) != 1+sampleCharts {
		t.Errorf("%d rows; duplicate --system should export once", len(rows))
	}
	for _, r := range rows[1:] {
		if r[0] != "traditional" {
			t.Fatalf("row for %s", r[0])
		}
	}

	path = filepath.Join(t.TempDir(), "out.csv")
	if _, _, err := execute(t, "-o", path, "-s", "traditional,goldendawn"); err != nil {
		t.Fatal(err)
	}
	if rows := readCSV(t, path); len(rows) != 1+2*sampleCharts || rows[1][0] != "traditional" {
		t.Errorf("comma-separated systems: %d rows, first %s", len(rows), rows[1][0])
	}
}

func TestOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.csv")
	if err := os.WriteFile(path, []byte("keep me"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := execute(t, "-o", path); err == nil || !strings.Contains(err.Error(), "--force") {
		t.Errorf("error = %v, want a hint about --force", err)
	}
	if b, _ := os.ReadFile(path); string(b) != "keep me" {
		t.Error("existing file was changed without --force")
	}

	if _, _, err := execute(t, "-o", path, "--force"); err != nil {
		t.Fatal(err)
	}
	if rows := readCSV(t, path); len(rows) != 1+2*sampleCharts {
		t.Errorf("%d rows after --force", len(rows))
	}

	// --force replaces a SQLite database rather than adding to it.
	db := filepath.Join(dir, "out.sqlite")
	for range 2 {
		if _, _, err := execute(t, "-o", db, "--force"); err != nil {
			t.Fatal(err)
		}
	}
}

func TestOutputFile(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"out.csv", "out.sqlite"} {
		path := filepath.Join(dir, name)
		if _, _, err := execute(t, "-o", path); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if perm := info.Mode().Perm(); perm&0o644 != 0o644 {
			t.Errorf("%s permissions = %v, want at least 0644", name, perm)
		}
	}
	// No temporary files are left behind.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("leftover %s", e.Name())
		}
	}
}

func TestFailedExportLeavesNoFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.csv")
	failing := func([]system.System) iter.Seq2[export.Record, error] {
		return func(yield func(export.Record, error) bool) {
			yield(export.Record{}, os.ErrInvalid)
		}
	}
	cmd := newCommand(failing)
	cmd.SetArgs([]string{"-o", path})
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err == nil {
		t.Fatal("export succeeded")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("left %d files behind, e.g. %s", len(entries), entries[0].Name())
	}
}

func TestErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing", "out.csv")
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"-f", "xml"}, `unknown format "xml"`},
		{[]string{"-s", "nope"}, `unknown system "nope" (available: goldendawn, traditional)`},
		{[]string{"extra"}, "unknown command"},
		{[]string{"-o", missing}, "no such file or directory"},
	}
	for _, tt := range tests {
		if _, _, err := execute(t, tt.args...); err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%v: error = %v, want it to contain %q", tt.args, err, tt.want)
		}
	}
}

func TestRunExitCodes(t *testing.T) {
	t.Chdir(t.TempDir())
	cmd, _, errOut := newTestCommand()
	if code := run(cmd, nil); code != 0 {
		t.Errorf("success: exit code %d, stderr %q", code, errOut)
	}
	cmd, _, errOut = newTestCommand()
	if code := run(cmd, []string{"-f", "xml"}); code != 1 || !strings.HasPrefix(errOut.String(), "Error: unknown format") {
		t.Errorf("failure: exit code %d, stderr %q", code, errOut)
	}
}
