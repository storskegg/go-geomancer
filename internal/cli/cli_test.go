package cli

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/internal/cli/input"
	_ "github.com/storskegg/go-geomancer/system/goldendawn"
	_ "github.com/storskegg/go-geomancer/system/traditional"
)

var update = flag.Bool("update", false, "rewrite golden files")

func execute(t *testing.T, args ...string) (string, error) {
	t.Helper()
	stdout, _, err := runWithInput(t, "", args...)
	return stdout, err
}

func runWithInput(t *testing.T, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := NewCommand()
	var out, errOut bytes.Buffer
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return out.String(), errOut.String(), err
}

// trimLines drops trailing spaces, which Lip Gloss uses to pad blocks to a
// common width but which editors tend to strip from golden files.
func trimLines(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " ")
	}
	return strings.Join(lines, "\n")
}

func TestGolden(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"cast-traditional", []string{"cast", "-s", "traditional", "1212", "1122", "1111", "2222"}},
		{"cast-significator", []string{"cast", "--significator", "partnership", "1212", "1122", "1111", "2222"}},
		{"cast-question", []string{"cast", "--question", "Will the contract be signed?", "-q", "7", "1212", "1122", "1111", "2222"}},
		{"cast-occasion", []string{"cast", "--question", "Will the contract be signed?", "-q", "7", "--date", "2026-10-15", "--time", "3:30 pm", "--place", "Springfield, IL", "1212", "1122", "1111", "2222"}},
		{"cast-occasion-json", []string{"cast", "--json", "--date", "2026-10-15", "--time", "3:30 pm", "--place", "Springfield, IL", "1212", "1122", "1111", "2222"}},
		{"cast-goldendawn-nephews", []string{"cast", "-s", "goldendawn", "--nephews", "amissio", "fortuna minor", "via", "populus"}},
		{"cast-json", []string{"cast", "--json", "1212", "1122", "1111", "2222"}},
		{"cast-ascii", []string{"cast", "--symbols", "ascii", "1212", "1122", "1111", "2222"}},
		{"figures", []string{"figures"}},
		{"figure", []string{"figure", "fortuna", "major"}},
		{"figure-json", []string{"figure", "--json", "puer"}},
		{"systems", []string{"systems"}},
		{"systems-json", []string{"systems", "--json"}},
		{"suggest", []string{"suggest", "Will my husband get the job?"}},
		{"suggest-json", []string{"suggest", "--json", "Will", "the", "contract", "be", "signed?"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := execute(t, append(tt.args, "--color", "never")...)
			if err != nil {
				t.Fatal(err)
			}
			got = trimLines(got)

			path := filepath.Join("testdata", tt.name+".golden")
			if *update {
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("%v (run with -update to create it)", err)
			}
			if got != string(want) {
				t.Errorf("output differs from %s (run with -update to accept):\n%s", path, got)
			}
		})
	}
}

func TestCastJSONMatchesCore(t *testing.T) {
	out, err := execute(t, "cast", "--json", "1212", "1122", "1111", "2222")
	if err != nil {
		t.Fatal(err)
	}
	var r chart.Reading
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatal(err)
	}
	// Golden Dawn is the default system, and there's no significator unless
	// one is given.
	if r.System != "goldendawn" || r.Significator != 0 || r.Shield.Judge.Name() != "Carcer" {
		t.Errorf("reading = %+v", r)
	}
}

func TestSignificatorFlag(t *testing.T) {
	for _, q := range []string{"7", "Partnership", "partnership"} {
		out, err := execute(t, "cast", "--json", "-q", q, "1212", "1122", "1111", "2222")
		if err != nil {
			t.Fatal(err)
		}
		var r chart.Reading
		if err := json.Unmarshal([]byte(out), &r); err != nil {
			t.Fatal(err)
		}
		if r.Significator != 7 || len(r.Aspects) != 7 || r.Aspects[0].House != 1 {
			t.Errorf("-q %s: significator %d, aspects %v", q, r.Significator, r.Aspects)
		}
	}
	out, err := execute(t, "cast", "--color", "never", "-q", "10", "1212", "1122", "1111", "2222")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Significator: House 10 · Honour", "Aspects: ☍ Opposition 4 · □ Square 7 ↻, 1 ↺", "┏━━", "Heavy border: significator", "  Significator  "} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q", want)
		}
	}
}

// The prompts go to stderr, leaving stdout with only the reading.
func TestCastPromptAccessible(t *testing.T) {
	t.Setenv("ACCESSIBLE", "1")
	stdout, stderr, err := runWithInput(t, "6\n8\n1\n2\n", "cast", "--json", "--system", "traditional", "--significator", "7", "--question", "Will it?", "--date", "2026-10-15", "--time", "3:30 pm", "--place", "Springfield, IL")
	if err != nil {
		t.Fatal(err)
	}
	var r chart.Reading
	if err := json.Unmarshal([]byte(stdout), &r); err != nil {
		t.Fatalf("stdout is not a reading: %v\n%s", err, stdout)
	}
	if got := r.Shield.Mothers; got[0].Name() != "Amissio" || got[3].Name() != "Populus" {
		t.Errorf("mothers = %v", got)
	}
	if !strings.Contains(stderr, "Mother 1 (M1)") {
		t.Errorf("stderr lacks the prompt:\n%s", stderr)
	}
	for _, prompt := range []string{"System", "Significator", "Question", "Date", "Time", "Place"} {
		if strings.Contains(stderr, prompt) {
			t.Errorf("prompted for %s despite its flag:\n%s", prompt, stderr)
		}
	}
	if c := r.Circumstances; c == nil || c.Place.Label() != "Springfield, IL" || c.Zone != "CDT" {
		t.Errorf("circumstances = %+v", c)
	}
	if r.System != "traditional" || r.Significator != 7 {
		t.Errorf("system %s, significator %d", r.System, r.Significator)
	}
}

// Without --system and --significator, both are chosen before the Mothers.
func TestCastPromptSystem(t *testing.T) {
	t.Setenv("ACCESSIBLE", "1")
	// Systems are listed by ID, so 2 is traditional; then the date, time
	// and place; the question; houses are listed in order, so 9 is House 9.
	stdout, stderr, err := runWithInput(t, "2\n10/15/2026\n3:30 pm\nSpringfield, IL\nShall I travel?\n9\n6\n8\n1\n2\n", "cast", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var r chart.Reading
	if err := json.Unmarshal([]byte(stdout), &r); err != nil {
		t.Fatalf("stdout is not a reading: %v\n%s", err, stdout)
	}
	if r.System != "traditional" || r.Question != "Shall I travel?" || r.Significator != 9 || r.Shield.Mothers[0].Name() != "Amissio" || r.Shield.Mothers[3].Name() != "Populus" {
		t.Errorf("reading = %s, significator %d, mothers %v", r.System, r.Significator, r.Shield.Mothers)
	}
	if c := r.Circumstances; c == nil || c.Time.Format("2006-01-02 15:04 MST") != "2026-10-15 15:30 CDT" || c.Place.Label() != "Springfield, IL" {
		t.Errorf("circumstances = %+v", c)
	}
	for _, want := range []string{"System", "Date", "Time", "Place", "Question", "Significator", "9. Journeys — long journeys"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr lacks %q:\n%s", want, stderr)
		}
	}
}

var ansi = regexp.MustCompile("\x1b\\[")

func TestColor(t *testing.T) {
	always, err := execute(t, "systems", "--color", "always")
	if err != nil {
		t.Fatal(err)
	}
	if !ansi.MatchString(always) {
		t.Error("--color always produced no escape codes")
	}
	never, err := execute(t, "systems", "--color", "never")
	if err != nil {
		t.Fatal(err)
	}
	if ansi.MatchString(never) {
		t.Error("--color never produced escape codes")
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"cast", "1212"}, "expected 4 mothers or none to choose them interactively, got 1"},
		{[]string{"cast"}, "not running in a terminal"},
		{[]string{"cast", "1212", "1212", "1212", "9999"}, `mother M4: figure: invalid figure: "9999"`},
		{[]string{"cast", "-s", "nope", "1212", "1212", "1212", "1212"}, `unknown system "nope" (available: goldendawn, traditional)`},
		{[]string{"figure", "nobody"}, `invalid figure: "nobody"`},
		{[]string{"systems", "--color", "bogus"}, `invalid --color "bogus"`},
		{[]string{"cast", "--symbols", "emoji", "1212", "1212", "1212", "1212"}, `unknown symbols "emoji"`},
		{[]string{"cast", "-q", "13", "1212", "1212", "1212", "1212"}, `unknown significator "13"`},
		{[]string{"cast", "-q", "marriage", "1212", "1212", "1212", "1212"}, `unknown significator "marriage"`},
	}
	for _, tt := range tests {
		_, err := execute(t, tt.args...)
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%v: error = %v, want it to contain %q", tt.args, err, tt.want)
		}
	}
}

func TestRunExitCodes(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		code   int
		stderr string
	}{
		{"success", []string{"systems", "--color", "never"}, 0, ""},
		{"error", []string{"figure", "nobody"}, 1, `Error: figure: invalid figure: "nobody"`},
		{"cancelled", []string{"cancel"}, 130, ""},
		{"unknown command", []string{"divine"}, 1, `Error: unknown command "divine"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCommand()
			cmd.AddCommand(&cobra.Command{
				Use:  "cancel",
				RunE: func(*cobra.Command, []string) error { return fmt.Errorf("prompt: %w", input.ErrAborted) },
			})
			var out, errOut bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&errOut)
			if got := run(cmd, tt.args); got != tt.code {
				t.Errorf("exit code = %d, want %d", got, tt.code)
			}
			if tt.stderr == "" && errOut.Len() > 0 {
				t.Errorf("unexpected stderr: %q", errOut.String())
			}
			if !strings.Contains(errOut.String(), tt.stderr) {
				t.Errorf("stderr = %q, want it to contain %q", errOut.String(), tt.stderr)
			}
		})
	}
}

func TestUnknownSystemEverywhere(t *testing.T) {
	for _, args := range [][]string{
		{"figures", "-s", "nope"},
		{"figure", "-s", "nope", "puer"},
		{"cast", "-s", "nope", "1212", "1212", "1212", "1212"},
	} {
		if _, err := execute(t, args...); err == nil || !strings.Contains(err.Error(), `unknown system "nope"`) {
			t.Errorf("%v: error = %v", args, err)
		}
	}
}

func TestSystemFlagCaseInsensitive(t *testing.T) {
	out, err := execute(t, "cast", "--json", "-s", "GoldenDawn", "1212", "1122", "1111", "2222")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"system": "goldendawn"`) {
		t.Errorf("output:\n%s", out)
	}
}

func TestFigureByPattern(t *testing.T) {
	byName, err := execute(t, "figure", "--json", "amissio")
	if err != nil {
		t.Fatal(err)
	}
	byPattern, err := execute(t, "figure", "--json", "1212")
	if err != nil {
		t.Fatal(err)
	}
	if byName != byPattern {
		t.Error("figure amissio and figure 1212 differ")
	}
}

func TestFiguresJSON(t *testing.T) {
	out, err := execute(t, "figures", "--json", "-s", "goldendawn")
	if err != nil {
		t.Fatal(err)
	}
	var figs []struct {
		Name    string  `json:"name"`
		Sign    *string `json:"sign"`
		Element *string `json:"element"`
	}
	if err := json.Unmarshal([]byte(out), &figs); err != nil {
		t.Fatal(err)
	}
	if len(figs) != 16 {
		t.Fatalf("%d figures, want 16", len(figs))
	}
	for _, f := range figs {
		node := f.Name == "Caput Draconis" || f.Name == "Cauda Draconis"
		if (f.Sign == nil) != node || f.Element == nil {
			t.Errorf("%s: sign %v, element %v", f.Name, f.Sign, f.Element)
		}
	}
}

func TestArgCounts(t *testing.T) {
	for _, args := range [][]string{
		{"figures", "extra"},
		{"systems", "extra"},
		{"figure"},
		{"cast", "1", "2", "3", "4", "5"},
	} {
		if _, err := execute(t, args...); err == nil {
			t.Errorf("%v succeeded", args)
		}
	}
}

func TestCastPDF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "reading.pdf")
	stdout, stderr, err := runWithInput(t, "", "cast", "--color", "never", "--question", "Will it work?", "-q", "7", "--pdf", path, "1212", "1122", "1111", "2222")
	if err != nil {
		t.Fatal(err)
	}
	// The terminal output is unchanged; the report is extra.
	if !strings.Contains(stdout, "The Shield") || !strings.Contains(stdout, "Question  “Will it work?”") {
		t.Errorf("stdout lacks the reading:\n%.200s", stdout)
	}
	if stderr != "Wrote the report to "+path+".\n" {
		t.Errorf("stderr = %q", stderr)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(b, []byte("%PDF-")) || !bytes.Contains(b, []byte("/Count 3")) {
		t.Errorf("not a 3-page PDF: %.20q", b)
	}
	if info, _ := os.Stat(path); info.Mode().Perm()&0o644 != 0o644 {
		t.Errorf("permissions %v", info.Mode().Perm())
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Errorf("%d files in the directory, want just the report", len(entries))
	}

	// With --json, too.
	if _, _, err := runWithInput(t, "", "cast", "--json", "--pdf", path, "1212", "1122", "1111", "2222"); err != nil {
		t.Fatal(err)
	}
}

// A bad --pdf path fails before any prompt.
func TestCastPDFBadPath(t *testing.T) {
	t.Setenv("ACCESSIBLE", "1")
	for _, path := range []string{filepath.Join(t.TempDir(), "missing", "x.pdf"), "go.mod/x.pdf"} {
		_, stderr, err := runWithInput(t, "", "cast", "--pdf", path)
		if err == nil || !strings.Contains(err.Error(), "can't write "+path) {
			t.Errorf("%s: error = %v", path, err)
		}
		if stderr != "" {
			t.Errorf("%s: prompted before failing:\n%s", path, stderr)
		}
	}
}

func TestSuggestCommand(t *testing.T) {
	out, err := execute(t, "suggest", "--color", "never", "Will it?")
	if err != nil || !strings.Contains(out, "No suggestion") {
		t.Errorf("no suggestion: %q, %v", out, err)
	}
	if _, err := execute(t, "suggest"); err == nil {
		t.Error("suggest without a question succeeded")
	}
}

// With a question, the significator prompt starts on its suggestion.
func TestCastPromptSuggestion(t *testing.T) {
	t.Setenv("ACCESSIBLE", "1")
	// The question, Enter to accept the suggestion, then the Mothers.
	stdout, stderr, err := runWithInput(t, "Will my husband get the job?\n\n6\n8\n1\n2\n", "cast", "--json", "-s", "goldendawn", "--date", "2026-10-15", "--time", "3:30 pm", "--place", "Springfield, IL")
	if err != nil {
		t.Fatal(err)
	}
	var r chart.Reading
	if err := json.Unmarshal([]byte(stdout), &r); err != nil {
		t.Fatal(err)
	}
	if r.Significator != 4 {
		t.Errorf("significator = %d, want the suggested 4", r.Significator)
	}
	if !strings.Contains(stderr, "Suggested: House 4 · Home, the 10th from the 7th (“husband”, “job”)") {
		t.Errorf("stderr lacks the suggestion:\n%s", stderr)
	}

	// --question feeds the suggestion too.
	stdout, _, err = runWithInput(t, "\n6\n8\n1\n2\n", "cast", "--json", "-s", "goldendawn", "--question", "Will we divorce?", "--date", "2026-10-15", "--time", "3:30 pm", "--place", "Springfield, IL")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(stdout), &r); err != nil || r.Significator != 7 {
		t.Errorf("significator = %d, %v; want 7", r.Significator, err)
	}
}

func TestOrdinal(t *testing.T) {
	for n, want := range map[int]string{1: "1st", 2: "2nd", 3: "3rd", 4: "4th", 10: "10th", 11: "11th", 12: "12th", 21: "21st", 22: "22nd", 113: "113th"} {
		if got := ordinal(n); got != want {
			t.Errorf("ordinal(%d) = %q, want %q", n, got, want)
		}
	}
}
