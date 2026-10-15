# Plan 05 — Time and Place of the Reading

`cast` asks for the date, local time and place of the reading, and works out the querent's
coordinates and time zone, offline. They're shown for now; later work will use them.

## Decisions

| Topic | Decision |
|---|---|
| Order | System → Date → Time → Place → (which time, if ambiguous) → Question → Significator → Mothers |
| Required | Date, time and place, with no defaults (not "now") |
| Date and time input | Forgiving text with a live confirmation line ("→ Thursday, 15 October 2026"). The one Bubble Tea date picker found (`EthanEFung/bubble-datepicker`) targets Bubble Tea v0.24, not v2 |
| Places | GeoNames (CC BY 4.0): 168,209 populated places in the 50 states and DC, most populous first; built by `internal/tools/places` into `place/data/places.tsv.gz` (3.6 MB) |
| Labels | "City, ST"; where a state has two places of a name, the county too: "Springfield, PA (Delaware County)" |
| Time zone | `github.com/ringsaturn/tzf` (MIT, data ODbL), offline, from the coordinates; offsets and daylight saving from Go's embedded `time/tzdata` |
| Daylight saving | Daylight and standard time are separate zones (CDT, CST). A time in the spring-forward gap is rejected with the equivalent later time; one in the fall-back overlap asks which (or `--time "1:30 am CST"`) |
| Flags | `--date`, `--time`, `--place`. With prompts, one that can't be used is reported and asked for. Without prompts (Mothers as arguments), the three go together, all or none |
| Display | "When" and "Where" beside the question at the top of the terminal output and in the PDF header; `circumstances` in the JSON |

## Code

- `when`: `ParseDate`, `ParseClock`, `Resolve` (the instants a local time means in a zone: one,
  none with a `GapError`, or two).
- `place`: `Index.Search` (typeahead: the first word begins the name; others begin a word of the
  name, the state's abbreviation, which must match exactly, its name, or the county; "st", "ft",
  "mt" are "saint", "fort", "mount"), `Index.Find`, `TimeZone`, `Preload`.
- `chart.Circumstances` and `chart.Location` on `Reading`.
- The place prompt is a custom field over huh's Input: huh's Select renders every option each
  frame, too slow for 168,000 places.

## Tests

`tzf`'s zone agrees with GeoNames' recorded zone for all of a 3,000-place sample. DST gaps and
overlaps are covered in `when`, the prompts (by key presses) and the flags.

## Later

Extract places, vectors and other tables into updatable external support files.
