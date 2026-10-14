// Command geomancer casts and interprets geomantic charts.
package main

import (
	"os"

	"github.com/storskegg/go-geomancer/internal/cli"
	_ "github.com/storskegg/go-geomancer/system/goldendawn"
	_ "github.com/storskegg/go-geomancer/system/traditional"
)

func main() { os.Exit(cli.Execute()) }
