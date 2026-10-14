package cli

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/storskegg/go-geomancer/chart"
	_ "github.com/storskegg/go-geomancer/system/goldendawn"
	_ "github.com/storskegg/go-geomancer/system/traditional"
)

var update = flag.Bool("update", false, "rewrite golden files")

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := NewCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
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
			got, err := run(t, append(tt.args, "--color", "never")...)
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
	out, err := run(t, "cast", "--json", "1212", "1122", "1111", "2222")
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

var ansi = regexp.MustCompile("\x1b\\[")

func TestColor(t *testing.T) {
	always, err := run(t, "systems", "--color", "always")
	if err != nil {
		t.Fatal(err)
	}
	if !ansi.MatchString(always) {
		t.Error("--color always produced no escape codes")
	}
	never, err := run(t, "systems", "--color", "never")
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
		{[]string{"cast", "1212"}, "accepts 4 arg(s)"},
		{[]string{"cast", "1212", "1212", "1212", "9999"}, `mother M4: figure: invalid figure: "9999"`},
		{[]string{"cast", "-s", "nope", "1212", "1212", "1212", "1212"}, `unknown system "nope" (available: goldendawn, traditional)`},
		{[]string{"figure", "nobody"}, `invalid figure: "nobody"`},
		{[]string{"systems", "--color", "bogus"}, `invalid --color "bogus"`},
	}
	for _, tt := range tests {
		_, err := run(t, tt.args...)
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%v: error = %v, want it to contain %q", tt.args, err, tt.want)
		}
	}
}
