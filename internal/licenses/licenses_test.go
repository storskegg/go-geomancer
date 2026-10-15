package licenses

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestCurrent checks modules.go against go.mod and go.sum, so it can't fall
// behind a dependency change.
func TestCurrent(t *testing.T) {
	gomod, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Fatal(err)
	}
	gosum, err := os.ReadFile("../../go.sum")
	if err != nil {
		t.Fatal(err)
	}

	// Requirements are "path version" lines inside require blocks, or after
	// "require" on one line.
	require := regexp.MustCompile(`(?m)^(?:require\s+|\t)(\S+)\s+(v\S+)(\s*// indirect)?\s*$`)
	want := map[string]bool{}
	for _, m := range require.FindAllStringSubmatch(string(gomod), -1) {
		want[m[1]+" "+m[2]] = m[3] != ""
	}
	sums := map[string]string{}
	for line := range strings.Lines(string(gosum)) {
		if f := strings.Fields(line); len(f) == 3 {
			sums[f[0]+" "+f[1]] = f[2]
		}
	}

	stale := "; run go generate ./internal/licenses"
	all := All()
	if len(all) != len(want) {
		t.Errorf("%d modules, go.mod requires %d%s", len(all), len(want), stale)
	}
	for i, m := range all {
		key := m.Path + " " + m.Version
		indirect, ok := want[key]
		switch {
		case !ok:
			t.Errorf("%s isn't required by go.mod%s", key, stale)
		case m.Indirect != indirect:
			t.Errorf("%s: Indirect = %t, go.mod says %t%s", key, m.Indirect, indirect, stale)
		case m.Sum != sums[key]:
			t.Errorf("%s: Sum = %s, go.sum has %s%s", key, m.Sum, sums[key], stale)
		}
		if m.License == "" {
			t.Errorf("%s has no licence", key)
		}
		if i > 0 && all[i-1].Path >= m.Path {
			t.Errorf("modules not sorted: %s before %s", all[i-1].Path, m.Path)
		}
	}
}

func TestAllCopies(t *testing.T) {
	All()[0].Path = "changed"
	if All()[0].Path == "changed" {
		t.Error("All returned the package's own slice")
	}
}
