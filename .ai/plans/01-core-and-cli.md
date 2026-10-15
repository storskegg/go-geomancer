# Plan 01 — Core Modules & CLI

Builds on `index.md`. Scope: the shared core logic and the CLI app. The API server is sketched
only so the core is shaped correctly for it; it gets its own plan later.

## Decisions (from Q&A)

| Topic | Decision |
|---|---|
| Go | 1.27.x, module `github.com/storskegg/go-geomancer` |
| Systems | A System is a Go interface. Ship `traditional` (filled) and `goldendawn` (skeleton). Runtime/user-defined systems are out of scope. |
| System owns | Segment→element, Position→House placement, Figure→Sign. |
| Segment elements | Traditional: Head=Fire, Neck=Air, Body=Water, Feet=Earth. GD: Head=Fire, Neck=Water, Body=Air, Feet=Earth. |
| House placement | Traditional: M1–M4→1–4, D1–D4→5–8, N1–N4→9–12. GD: M1→10, M2→1, M3→4, M4→7, D1→11, D2→2, D3→5, D4→8, N1→12, N2→3, N3→6, N4→9. |
| House signs | House 1 sign = system sign of the House 1 figure; houses 2–12 follow zodiac order. |
| House elements from the Judge | Deferred. |
| Figure content | Fill what is well-established; skeleton the rest for the user to complete. |
| Mothers input | Undecided. v1 ships one simple input method behind an interface (see CLI). |
| CLI stack | Cobra + Lip Gloss (Bubble Tea later if an interactive mode is wanted). |
| Server stack | stdlib `net/http`, JSON, OpenAPI spec. No database. |
| Deferred to v2 | Part of Fortune, company of houses, perfection, way of the point, aspects (trines, sextiles, …). |

### Assumptions (correct me if wrong)

- **Terminology:** a *Figure* is one of the 16 dot patterns (Populus, Via, …) and carries
  name/descriptions. A *Position* is a slot in the chart (M1…N4, WR, WL, J, JM) that holds a
  Figure. "Tetragram" stays as a synonym for Figure in UI copy.
- **The Shield shows the Judgement (JM)** in addition to the items listed in `index.md`.
- **Niece/Nephew** is a display label only (CLI flag/config); the core always uses `N1…N4`.
- **Planet**, like sign and element, is system-owned (`System.Planet`). Both systems give the
  standard attribution, with the North and South Nodes for Caput and Cauda Draconis.
  **Element per figure** varies between sources, so it is system-owned alongside Sign.

## Repository Layout

```
go-geomancer/
├── go.mod
├── figure/              # Segment, Figure, addition, dot counts, the 16 canonical figures
├── zodiac/              # Sign, Element, Planet enums + helpers (Sign.Add(n) for house rotation)
├── system/              # System interface + registry
│   ├── traditional/
│   └── goldendawn/
├── chart/               # Positions, Shield casting, Astrological Square
├── content/             # Long/short descriptions, per-house meanings (embedded data)
├── internal/
│   └── cli/             # Lip Gloss rendering, input parsing
├── cmd/
│   ├── geomancer/       # CLI entry point
│   └── geomancerd/      # API server entry point (later)
└── .ai/plans/
```

Core packages (`figure`, `zodiac`, `system`, `chart`, `content`) are top-level and importable, with no
dependencies on CLI/server code and no third-party deps. Both apps only call into these.

## Core Design

### `figure`

```go
type Segment uint8 // Head, Neck, Body, Feet

type Parity uint8   // Odd (active, 1 dot) / Even (passive, 2 dots)

// Figure is a 4-bit value; a set bit means the segment is Odd.
type Figure uint8

func New(head, neck, body, feet Parity) Figure
func (f Figure) Segment(s Segment) Parity
func (f Figure) Add(g Figure) Figure   // segment-wise parity sum == XOR
func (f Figure) Dots() int             // total
func (f Figure) OddDots() int          // qty odd segments
func (f Figure) EvenDots() int         // qty even segments * 2
func (f Figure) Name() string          // Latin name
func Parse(s string) (Figure, error)   // "1212", "odd,even,odd,even", or a figure name
func All() [16]Figure
```

Odd=1 makes figure addition a plain XOR, and calculating the Daughters a 4×4 bit-matrix transpose.

The 16 figures (Head→Feet, 1 = odd, 2 = even), with the planets both systems give them:

| Figure | Pattern | Planet |
|---|---|---|
| Via | 1111 | Moon |
| Populus | 2222 | Moon |
| Albus | 2212 | Mercury |
| Conjunctio | 2112 | Mercury |
| Puella | 1211 | Venus |
| Amissio | 1212 | Venus |
| Fortuna Major | 2211 | Sun |
| Fortuna Minor | 1122 | Sun |
| Puer | 1121 | Mars |
| Rubeus | 2122 | Mars |
| Acquisitio | 2121 | Jupiter |
| Laetitia | 1222 | Jupiter |
| Carcer | 1221 | Saturn |
| Tristitia | 2221 | Saturn |
| Caput Draconis | 2111 | North Node |
| Cauda Draconis | 1112 | South Node |

### `zodiac`

`Sign` (Aries…Pisces, with `Next()` and `Add(n)` wrapping mod 12), `Element` (Fire/Air/Water/Earth),
`Planet` (7 classical planets + North/South Node). Each has `String()` and JSON marshalling (as
names, not ints), which keeps the API readable.

### `system`

```go
type System interface {
    ID() string   // "traditional", "goldendawn"
    Name() string // "Traditional", "Golden Dawn"

    SegmentElement(figure.Segment) zodiac.Element
    House(position.Position) (int, bool)             // 1–12; false for non-principal positions
    Sign(figure.Figure) (zodiac.Sign, bool)          // false = unmapped
    Element(figure.Figure) (zodiac.Element, bool)    // false = unmapped
    Planet(figure.Figure) (zodiac.Planet, bool)      // false = unmapped
}

// Mapping implements System from lookup tables; both shipped systems use it.
func Validate(System) error // ID, distinct segment elements, houses 1–12 one-to-one

func Register(System)
func Get(id string) (System, bool)
func List() []System
```

Each system subpackage registers itself in `init()`, and the apps blank-import the systems they use.
The `(value, ok)` returns let the Golden Dawn skeleton compile and run before it has data.
`Position` lives in the leaf package `position`, so `system` and `chart` don't import each other.

### `chart`

```go
type Position uint8 // M1..M4, D1..D4, N1..N4, WR, WL, J, JM

type Shield struct {
    Mothers, Daughters, Nieces [4]figure.Figure
    RightWitness, LeftWitness  figure.Figure
    Judge, Judgement           figure.Figure
}

func Cast(mothers [4]figure.Figure) Shield
func (s Shield) At(p Position) figure.Figure

type House struct {
    Number   int
    Position Position
    Figure   figure.Figure
    Sign     zodiac.Sign
    HasSign  bool
}

type Square [12]House

func BuildSquare(s Shield, sys system.System) Square

// Reading is the full result and the API's response payload.
type Reading struct {
    System string
    Shield Shield
    Square Square
}

func NewReading(mothers [4]figure.Figure, sys system.System) Reading
```

Square signs: look up `sys.Sign(figure in House 1)`. If it's unmapped, no house gets a sign;
otherwise House *n* gets `sign.Add(n-1)`.

### `content`

Text is kept out of the code in an embedded JSON file per locale, `content/data/en.json`. JSON
rather than YAML keeps the core free of third-party dependencies.

```json
{"locale": "en", "figures": [
  {"figure": "Acquisitio", "short": "", "long": "", "houses": {"1": "", "2": "", "...": ""}}
]}
```

`content.Default().For(figure.Figure)` returns the `Entry`. The file also holds the twelve
houses' significations (title and text), the same in every system, used to choose the
significator; a reading records it as `Reading.Significator` and the Square gives that house a
heavy border. Loading checks that each of the 16
figures appears exactly once. All text fields ship empty for you to write.

## Figure Sign & Element Data — provisional

Both systems ship with complete tables, each marked `TODO` in the source for verification:

- **Golden Dawn signs:** planetary rulership. Puer–Aries, Amissio–Taurus, Albus–Gemini,
  Populus & Via–Cancer, Fortuna Major & Minor–Leo, Conjunctio–Virgo, Puella–Libra,
  Rubeus–Scorpio, Acquisitio–Sagittarius, Carcer–Capricorn, Tristitia–Aquarius,
  Laetitia–Pisces. Caput and Cauda Draconis have no sign, so when either is in the 1st House
  the Square has no signs.
- **Golden Dawn elements:** the element of each figure's sign, with Caput Draconis = Earth and
  Cauda Draconis = Fire. These were derived, not taken from a published table.
- **Traditional:** copies of both Golden Dawn tables until the traditional source is chosen.
  A recalled, unverified Gerard of Cremona sign table is kept in a comment for reference.

## CLI (`cmd/geomancer`)

```
geomancer cast [M1 M2 M3 M4] [--nephews]  # Shield + Astrological Square; prompts if no args
geomancer figures                       # table of all 16 figures + meta
geomancer figure  <name|pattern>        # detail view incl. house meanings
geomancer systems                       # list registered systems
geomancer suggest QUESTION              # suggest the significator for a question

Global flags: -s/--system (default goldendawn), --json, --color auto|always|never, --symbols unicode|ascii
cast flags:   -q/--significator N|TITLE, --question TEXT, --pdf FILE, --nephews
```

- **Input:** `internal/cli/input` defines `type MothersSource interface { Mothers() ([4]figure.Figure, error) }`.
  - `input.Prompt` (default): first a select for the System (skipped if `--system` is given;
    shows each system's segment elements and Mother houses), then the Significator, the house
    of the question (skipped if `--significator` is given; filterable by subject, not by number,
    which the list shows as circled digits), then one huh v2 select per Mother, scrollable, with typeahead filtering on
    by name or pattern and a dot preview of the highlighted figure. Shift+Tab goes back. It draws
    on stderr, so `cast --json > file` stays clean. `ACCESSIBLE=1` switches to numbered,
    line-based prompts. Without arguments and outside a terminal, `cast` fails with a hint.
  - `input.Args`: four positional arguments, each a pattern (`1212`), parities
    (`odd,even,odd,even`) or a name. Kept for scripts and tests.
  - Random generation and mark-counting can be added later as further implementations.
- **Rendering (Lip Gloss v2, `internal/cli/render`):**
  - Shield in the traditional layout: Mothers and Daughters across the top read right to left
    (M1 at the far right), Nieces below, then the Witnesses, with the Judge centred and JM to its
    right. Each figure is centred under the two it was summed from. 88 columns wide.
  - Square as bordered house cells around the edge of a 5×5 grid, three per side, corners and
    the 3×3 centre empty, running counterclockwise from House 1 at the middle of the left side.
    Houses are 18 columns wide with 1-column gaps, so the Square is 94 columns wide and 60 lines
    tall. Sign symbols run down each house's left edge and elements down its right: the house's
    sign flanked by its symbol and element; the figure's dots, the top row flanked by the
    figure's sign symbol and element, the second row's right edge reserved for a further
    element icon; then the figure's name and position. A legend explains the symbols. A
    traditional square-chart drawing can come later.
  - `--symbols unicode|ascii`: signs as ♈–♓ or two-letter abbreviations, elements as alchemical
    triangles (🜂 🜁 🜄 🜃) or letters, for fonts that lack the triangles.
  - Geomantic aspects to the significator (package `aspect`, the same in every system): counting
    around the houses and including the significator, the Sextile is 2 houses away, the Square
    3, the Trine 4, the Opposition 6. Each but the Opposition falls on both sides; the Dexter
    (clockwise, counting back through the house numbers) is the stronger of the pair. The
    Opposition and Square are stressful; the Trine and Sextile ease the situation, the Sextile
    mildly benevolent. A reading lists them in `Reading.Aspects`. On the Square, the line under
    each house's sign labels the significator and each aspect (e.g. "□ Square ↻"), coloured red
    for stressful and green for easing; a summary line and legend explain them.
  - The relationship between a figure and its own house is still deferred.
  - After the Square, a "Reading" section gives the meaning of the figure in each house that
    bears on the question (`chart.Reading.Focus`): House 1 (the querent), the significator's
    house, then the houses aspecting it, Opposition first and Dexter before Sinister; a house
    with several parts is listed once. Without a significator, all twelve houses.
  - `--json` prints the `chart.Reading` so the CLI and API output can be compared.
  - Output goes through a `colorprofile` writer, so styling is stripped or downsampled
    automatically. `--color` overrides that.
- **Tests:** golden files in `internal/cli/testdata`; `make golden` rewrites them.

## API Server (outline only, later plan)

`cmd/geomancerd` serves `POST /v1/readings` (mothers + system → `chart.Reading`), `GET /v1/figures`,
`GET /v1/figures/{id}`, `GET /v1/systems`. An `openapi.yaml` lives at the repo root. All logic is
in the core, so the handlers are thin.

## Testing

- `figure`: XOR addition truth table, dot counts for all 16, Parse/format round-trips, name↔pattern
  uniqueness.
- `chart`: the `index.md` worked example; golden test for one full shield worked by hand;
  **exhaustive test over all 16⁴ = 65,536 Mother sets** checking invariants. One invariant: the Judge
  always has an even total dot count. (This is a known property of geomancy, and a good self-check.)
- `system`: placement is a bijection onto houses 1–12 for every system; sign rotation wraps
  correctly; an unmapped sign gives no house signs.
- CLI: golden-file tests for the rendered output (with colours stripped) and for the JSON.

## Milestones

Milestones 1–6 are complete.

1. **Scaffold:** `go.mod` (go 1.27), the directory layout, `golangci-lint` config, a `Makefile`
   (`test`, `lint`, `build`).
2. **`zodiac` + `figure`:** types, the 16 figures, addition, Parse, tests.
3. **`chart`:** Positions, `Cast`, Shield tests including the exhaustive invariants.
4. **`system`:** interface, registry, `traditional` (confirmed data plus TODO tables),
   `goldendawn` skeleton, `BuildSquare`, tests.
5. **`content`:** embedded data skeleton plus loader.
6. **CLI:** `cast` (text + `--json`), then `figures`, `figure`, `systems`, then golden tests.
7. *(Later plan)* API server.

## Open Items

1. Verify the provisional sign and element tables (see above), especially Traditional.
2. Confirm the assumptions above: Figure/Position terms, JM in the Shield.
3. Mothers input methods beyond the prompt and arguments. *(not blocking)*
4. House signs when Caput or Cauda Draconis is in the 1st House. Currently no house gets a sign.
5. Review the draft meanings of each figure in each house (`content/data/en.json`, 192 entries).
   They're original summaries from each figure's traditional character, not taken from a source.
   Rubeus and Cauda Draconis in House 1 note that the chart is traditionally not to be judged.
