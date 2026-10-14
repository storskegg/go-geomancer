package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"testing"
)

// TestMain lets the test binary act as geomancer when re-run with
// GEOMANCER_RUN_MAIN set, so tests can exercise the real entry point.
func TestMain(m *testing.M) {
	if os.Getenv("GEOMANCER_RUN_MAIN") != "" {
		os.Args = append([]string{"geomancer"}, os.Args[1:]...)
		main()
		return
	}
	os.Exit(m.Run())
}

func geomancer(t *testing.T, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(os.Args[0], args...)
	cmd.Env = append(os.Environ(), "GEOMANCER_RUN_MAIN=1")
	out, err := cmd.Output()
	if exit := (*exec.ExitError)(nil); errors.As(err, &exit) {
		return string(out), exit.ExitCode()
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(out), 0
}

// Both systems are registered by main's imports.
func TestSystemsRegistered(t *testing.T) {
	out, code := geomancer(t, "systems", "--json")
	if code != 0 {
		t.Fatalf("exit code %d", code)
	}
	var systems []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(out), &systems); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if len(systems) != 2 || systems[0].ID != "goldendawn" || systems[1].ID != "traditional" {
		t.Errorf("systems = %+v", systems)
	}
}

func TestExitCodes(t *testing.T) {
	if _, code := geomancer(t, "cast", "--json", "1212", "1122", "1111", "2222"); code != 0 {
		t.Errorf("cast: exit code %d, want 0", code)
	}
	if _, code := geomancer(t, "figure", "nobody"); code != 1 {
		t.Errorf("bad figure: exit code %d, want 1", code)
	}
}
