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
  name/planet/descriptions. A *Position* is a slot in the chart (M1…N4, WR, WL, J, JM) that holds a
  Figure. "Tetragram" stays as a synonym for Figure in UI copy.
- **The Shield shows the Judgement (JM)** in addition to the items listed in `index.md`.
- **Niece/Nephew** is a display label only (CLI flag/config); the core always uses `N1…N4`.
- **Planet** is intrinsic to the Figure (traditional and GD agree), not system-owned.
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
func (f Figure) Planet() zodiac.Planet
func Parse(s string) (Figure, error)   // "1212", "odd,even,odd,even", or a figure name
func All() [16]Figure
```

Odd=1 makes figure addition a plain XOR, and calculating the Daughters a 4×4 bit-matrix transpose.

The 16 figures (Head→Feet, 1 = odd, 2 = even), with their planets:

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
    House(chart.Position) int                        // 1–12, principal positions only
    Sign(figure.Figure) (zodiac.Sign, bool)          // false = unmapped
    Element(figure.Figure) (zodiac.Element, bool)    // false = unmapped
}

func Register(System)
func Get(id string) (System, bool)
func List() []System
```

Each system subpackage registers itself in `init()`, and the apps blank-import the systems they use.
The `(value, ok)` returns let the Golden Dawn skeleton compile and run before it has data.
`House` takes `chart.Position`, so `Position` should live in a small leaf package (or `figure`) to
avoid an import cycle. That gets settled at implementation time.

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

Text is kept out of the code: an embedded data file (`content/data/en.yaml` or JSON) with one entry per
figure:

```yaml
- figure: Acquisitio
  short: ""          # qualitative one-liner
  long: ""           # long description
  houses:            # 1..12 qualitative meanings
    1: ""
    ...
```

`content.For(figure.Figure) Entry`. Keying the files by locale (`en`) keeps translation possible
without committing to it. I'll fill in only the names (the planets are already in `figure`), and
leave the text fields blank for you to write.

## Traditional System Data — needs your confirmation

Segment elements and house placement are confirmed above. **Figure→Sign and Figure→Element are not**:
sources disagree, and I'm not confident enough to commit values without your source. There are two
common sign tables:

- **Gerard of Cremona (medieval).** This is my recollection, *please verify*: Acquisitio–Aries,
  Laetitia–Taurus, Amissio–Taurus, Fortuna Minor–Taurus, Puer–Gemini, Rubeus–Gemini,
  Albus–Cancer, Via–Leo, Caput Draconis–Virgo, Puella–Libra, Tristitia–Scorpio,
  Cauda Draconis–Sagittarius, Populus–Capricorn, Fortuna Major–Aquarius, Carcer–Pisces,
  Conjunctio–Virgo or Pisces (unsure).
- **Rulership-based (Agrippa / commonly given for Golden Dawn).** Puer–Aries, Amissio–Taurus,
  Albus–Gemini, Populus & Via–Cancer, Fortuna Major & Minor–Leo, Conjunctio–Virgo, Puella–Libra,
  Rubeus–Scorpio, Acquisitio–Sagittarius, Carcer–Capricorn, Tristitia–Aquarius, Laetitia–Pisces;
  Caput/Cauda Draconis have no sign.

**Action:** tell me which table "traditional" means (or paste yours), and the figure→element table
you use. Until then the traditional system keeps these as clearly marked TODO tables, and tests
cover only the confirmed mappings.

## CLI (`cmd/geomancer`)

```
geomancer cast    --system traditional --mothers 1212,2112,1111,2221 [--nephews] [--json]
geomancer figures                       # table of all 16 figures + meta
geomancer figure  <name|pattern>        # detail view incl. house meanings
geomancer systems                       # list registered systems
```

- **Input:** `internal/cli/input` defines `type MothersSource interface { Mothers() ([4]figure.Figure, error) }`.
  v1 implements only `--mothers` (accepts patterns or names via `figure.Parse`). Random generation,
  mark-counting and interactive entry can be added later as further implementations without
  touching the core.
- **Rendering (Lip Gloss):**
  - Shield in the traditional layout: Mothers and Daughters across the top read right to left
    (M1 at the far right), Nieces below, then the Witnesses, with the Judge (and JM) at the bottom.
    Odd = `●`, even = `● ●`.
  - Square as a 3×4 grid of house cells (number, sign, figure glyph, name) to start. A
    traditional square-chart drawing can come later.
  - `--json` prints the `chart.Reading` so the CLI and API output can be compared.
  - Colour degrades gracefully when there's no TTY (Lip Gloss handles this).

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

1. Which figure→sign table "traditional" uses, and the figure→element table. *(blocks the
   traditional sign/element data only; everything else can proceed)*
2. Confirm the assumptions above: Figure/Position terms, JM in the Shield, planet intrinsic vs
   element system-owned.
3. Mothers input methods beyond `--mothers`. *(not blocking)*
