# Plan 03 — PDF Report (draft)

`geomancer cast --pdf FILE` writes a report of the reading as a PDF, in addition to the usual
terminal output.

## Decisions so far

| Topic | Decision |
|---|---|
| Flag | `--pdf FILE` on `cast`; stdout output is unchanged |
| Paper | US Letter; other sizes later |
| Style | Classical: parchment background, dark-brown ink, earthy accents (oxblood, olive, ochre, slate), a serif title and text face, double-rule page border, ornamental rule under the title |
| Square | The traditional square chart (outer square, centre square, diamond through its corners, corner diagonals: twelve equal triangles, counterclockwise from House 1 on the left). The terminal keeps the 5×5 layout |
| Aspects | Lines across the centre square from the significator to each aspected house: red stressful, green easing, solid Dexter, dashed Sinister, each ending in the aspect's symbol |
| Library | `github.com/tdewolff/canvas` (pure Go; PDF, and SVG/PNG for later use) |
| Fonts | Embedded open-licensed fonts, so output doesn't depend on the system. The samples used macOS fonts (Big Caslon, Iowan Old Style, Apple Symbols) |
| Symbols | Aspect symbols and figure dots are drawn as vector shapes; sign and element glyphs come from the symbol font |

## Samples

`.ai/samples/style-{classical,modern}.{pdf,png}` (untracked). Classical was chosen.

## Open

1. A prompt for the question in the querent's own words, for the report header (to revisit).
2. Contents and order of pages: header, Square, Shield, Reading, legend, figure notes.
3. Which open-licensed fonts to embed (text, titles, symbols).
