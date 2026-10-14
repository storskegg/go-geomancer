package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestMain lets the test binary act as geomancer-export when re-run with
// GEOMANCER_EXPORT_RUN_MAIN set, so tests can exercise the real entry point.
func TestMain(m *testing.M) {
	if os.Getenv("GEOMANCER_EXPORT_RUN_MAIN") != "" {
		os.Args = append([]string{"geomancer-export"}, os.Args[1:]...)
		main()
		return
	}
	os.Exit(m.Run())
}

// An unknown system lists the registered ones, which come from main's
// imports. Failing fast also keeps this test from running a full export.
func TestSystemsRegistered(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-s", "nope", "-o", "-")
	cmd.Env = append(os.Environ(), "GEOMANCER_EXPORT_RUN_MAIN=1")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		t.Fatalf("exit = %v, want code 1", err)
	}
	if !strings.Contains(stderr.String(), "available: goldendawn, traditional") {
		t.Errorf("stderr = %q", stderr.String())
	}
}
