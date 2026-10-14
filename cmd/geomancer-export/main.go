// Command geomancer-export writes a reading for every possible chart under
// each system, as CSV, JSON Lines, Parquet or SQLite.
package main

import (
	"os"

	"github.com/storskegg/go-geomancer/internal/exportcli"
	_ "github.com/storskegg/go-geomancer/system/goldendawn"
	_ "github.com/storskegg/go-geomancer/system/traditional"
)

func main() { os.Exit(exportcli.Execute()) }
