// Package licenses lists the modules geomancer is built from, with their
// versions and licences.
//
// The list is generated from go.mod and go.sum by internal/tools/licenses
// and compiled in; regenerate it after changing dependencies with
//
//	go generate ./internal/licenses
package licenses

//go:generate go run ../tools/licenses ../.. modules.go

// Module is a module required by go.mod.
type Module struct {
	// Path and Version identify the module, as in go.mod.
	Path    string `json:"path"`
	Version string `json:"version"`
	// Sum is the module's checksum from go.sum.
	Sum string `json:"sum"`
	// License is the SPDX identifier of the module's licence, or "Unknown"
	// if it couldn't be identified.
	License string `json:"license"`
	// Indirect is set for modules required only by other modules.
	Indirect bool `json:"indirect"`
}

// All returns the modules, sorted by path.
func All() []Module { return append([]Module(nil), modules...) }
