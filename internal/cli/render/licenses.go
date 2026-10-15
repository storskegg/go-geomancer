package render

import (
	"github.com/storskegg/go-geomancer/internal/licenses"
)

// Licenses lists the modules geomancer is built from, with their versions
// and licences.
func Licenses(mods []licenses.Module) string {
	t := newTable("Module", "Version", "License", "Required")
	for _, m := range mods {
		license := m.License
		if license == "Unknown" {
			license = mutedStyle.Render(license)
		}
		required := "directly"
		if m.Indirect {
			required = mutedStyle.Render("indirectly")
		}
		t.Row(m.Path, m.Version, license, required)
	}
	return Heading("Licenses") + "\n\n" +
		"geomancer is licensed under Apache-2.0 and built from these modules.\n" +
		mutedStyle.Render("See NOTICE.md for the data it builds in.") + "\n\n" +
		t.Render()
}
