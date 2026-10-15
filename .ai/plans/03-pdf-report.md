# Plan 03 — PDF Report

`geomancer cast --pdf FILE` writes a report of the reading as a PDF, in addition to the usual
terminal output.

## Decisions

| Topic | Decision |
|---|---|
| Flag | `--pdf FILE` on `cast`; stdout output is unchanged |
| Paper | US Letter; other sizes later |
| Style | Classical: parchment background, dark-brown ink, earthy accents (oxblood, olive, ochre, slate), a serif title and text face, double-rule page border, ornamental rule under the title |
| Square | The traditional square chart (outer square, centre square, diamond through its corners, corner diagonals: twelve equal triangles, counterclockwise from House 1 on the left). The terminal keeps the 5×5 layout |
| Aspects | Lines across the centre square from the significator to each aspected house: red stressful, green easing, solid Dexter, dashed Sinister. Each ends towards its house (in the direction of the house's contents) in the aspect's symbol, with the house's number beyond it; the significator's end is numbered too. Lines running out into the houses were tried and rejected: they cross the figures and names |
| Library | `github.com/tdewolff/canvas` (pure Go; PDF, and SVG/PNG for later use) |
| Fonts | Embedded open-licensed fonts, so output doesn't depend on the system. The samples used macOS fonts (Big Caslon, Iowan Old Style, Apple Symbols) |
| Symbols | Aspect symbols and figure dots are drawn as vector shapes; sign and element glyphs come from the symbol font |

## Question

`cast` asks for the question in the querent's own words after the system and before the
significator (free text, optional; `--question` skips the prompt). It's stored as
`Reading.Question`, shown above the Shield in the terminal, and in the report's header and
document title. Plan 04 uses it to suggest the significator.

## Pages

1. **Header and Square.** Title; the question, in balanced lines; system, significator and
   time of casting; an ornament. The traditional square chart, with the significator's house
   shaded and the aspect lines across the centre. Legends for the elements and aspects.
2. **The Shield,** drawn as a heraldic shield: eight Mothers and Daughters across the top (M1 on
   the right), four Nieces, two Witnesses, the Judge in the point, and the Judgement beside it.
3. **The Reading,** as in the terminal (`chart.Reading.Focus`), continuing onto further pages as
   needed.

Every page has the parchment ground, double-rule border and a page number.

## Fonts

Embedded from `report/fonts`, all under the SIL Open Font License 1.1 (licences beside them):
Libre Caslon Display (titles), Crimson Text regular, italic and semibold (text), and Noto Sans
Symbols (sign and element glyphs). Aspect symbols and figure dots are vector shapes.

## Code

- Package `report` (public, so the future server can produce reports): `Pages` draws the pages
  as canvases; `WritePDF` writes them as one PDF.
- `cast --pdf FILE` checks the directory before prompting, then writes through a temporary file
  renamed into place.
- `REPORT_PREVIEW_DIR=dir go test ./report -run TestPreview` renders sample reports as PNG and
  PDF for checking the design.

## Samples

`.ai/samples/style-{classical,modern}.{pdf,png}` (untracked): the style samples; classical was
chosen. `.ai/samples/report*.{pdf,png}` (untracked): the finished report.
