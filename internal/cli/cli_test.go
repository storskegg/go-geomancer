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
		{"cast-traditional", []string{"cast", "1212", "1122", "1111", "2222"}},
		{"cast-goldendawn-nephews", []string{"cast", "-s", "goldendawn", "--nephews", "amissio", "fortuna minor", "via", "populus"}},
		{"cast-json", []string{"cast", "--json", "1212", "1122", "1111", "2222"}},
		{"figures", []string{"figures"}},
		{"figure", []string{"figure", "fortuna", "major"}},
		{"figure-json", []string{"figure", "--json", "puer"}},
		{"systems", []string{"systems"}},
		{"systems-json", []string{"systems", "--json"}},
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
	if r.System != "traditional" || r.Shield.Judge.Name() != "Carcer" {
		t.Errorf("reading = %+v", r)
	}
}

// The prompts go to stderr, leaving stdout with only the reading.
func TestCastPromptAccessible(t *testing.T) {
	t.Setenv("ACCESSIBLE", "1")
	stdout, stderr, err := runWithInput(t, "6\n8\n1\n2\n", "cast", "--json", "--system", "traditional")
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
	if strings.Contains(stderr, "System") {
		t.Errorf("prompted for the system despite --system:\n%s", stderr)
	}
}

// Without --system, the system is chosen first.
func TestCastPromptSystem(t *testing.T) {
	t.Setenv("ACCESSIBLE", "1")
	// Systems are listed by ID: 1 is goldendawn.
	stdout, stderr, err := runWithInput(t, "1\n6\n8\n1\n2\n", "cast", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var r chart.Reading
	if err := json.Unmarshal([]byte(stdout), &r); err != nil {
		t.Fatalf("stdout is not a reading: %v\n%s", err, stdout)
	}
	if r.System != "goldendawn" || r.Shield.Mothers[3].Name() != "Populus" {
		t.Errorf("reading = %s, mothers %v", r.System, r.Shield.Mothers)
	}
	if !strings.Contains(stderr, "System") {
		t.Errorf("stderr lacks the system prompt:\n%s", stderr)
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
