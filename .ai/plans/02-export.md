# Plan 02 — Export App

A second CLI, `geomancer-export`, that casts every possible chart under each known system and
writes the results to a file.

## Scope

- Every set of Mothers: 16⁴ = 65,536, under each registered system (131,072 readings for two).
- Formats: `csv` (default), `jsonl`, `parquet`, `sqlite`.

## Decisions

| Topic | Decision |
|---|---|
| Binary | `cmd/geomancer-export`, standalone, Cobra for flags and help |
| Code | `internal/export`; the core packages stay free of third-party dependencies |
| Chart ID | Each set of Mothers has an ID 0–65535: `M1<<12 \| M2<<8 \| M3<<4 \| M4`, using the figures' 4-bit values. Stable across formats and systems, and decodable by hand. |
| Order | Systems by ID, then chart ID ascending |
| Figures | Names (`Fortuna Major`) in CSV, JSONL and Parquet; SQLite uses integer figure IDs (0–15, the 4-bit value) with a `figures` lookup table |
| Libraries | `github.com/parquet-go/parquet-go`, `modernc.org/sqlite` (pure Go, no cgo) |

### Flags

```
geomancer-export [--format csv|jsonl|parquet|sqlite] [--system ID]... [-o PATH] [--force]
```

- `--system` may be repeated; default is every registered system.
- `-o` defaults to `readings.<ext>`. `-o -` writes to stdout (not allowed for SQLite).
- An existing output file is not overwritten without `--force`.
- A one-line summary goes to stderr.

## Flat formats (CSV, Parquet)

One row per (system, chart), with columns in this order:

```
system, chart,
m1..m4, d1..d4, n1..n4, wr, wl, j, jm,     # the Shield
h1_figure..h12_figure,                     # the Square, by house
h1_sign..h12_sign                          # empty / null when the system gives House 1 no sign
```

The house figures repeat Shield values (placement is fixed per system), but they're included so
flat-file users don't have to know each system's placement.

## JSONL

One `chart.Reading` per line, the same JSON as `geomancer cast --json`, with a `chart` ID added.

## SQLite

Normalized tables:

| Table | Rows | Columns |
|---|---|---|
| `figures` | 16 | `id` (0–15), `name`, `pattern`, `planet`, `dots`, `odd_dots`, `even_dots` |
| `signs` | 12 | `id` (0 = Aries), `name` |
| `positions` | 16 | `code` (M1…JM), `ordinal`, `principal` |
| `systems` | per system | `id`, `name`, `head_element`, `neck_element`, `body_element`, `feet_element` |
| `system_figures` | 16 per system | `system_id`, `figure_id`, `sign_id` (nullable), `element` (nullable) |
| `system_houses` | 12 per system | `system_id`, `position`, `house` |
| `charts` | 65,536 | `id`, `m1`…`jm` (figure IDs); system-independent |
| `chart_houses` | 12 per chart per system | `system_id`, `chart_id`, `house`, `position`, `figure_id`, `sign_id` (nullable) |

`chart_houses` could be computed from the others, but materializing it makes queries simple
(e.g. "every chart where Carcer is in House 7 under Golden Dawn"). Indexed on
`(system_id, house, figure_id)`.

## Testing

- Each writer is tested on a small set of readings (round-trip read-back for CSV, JSONL, Parquet,
  and SQL queries for SQLite).
- An end-to-end test runs the full export for every format and checks row counts.

## Outcome

Implemented as planned. A full export of both systems (131,072 readings) on an Apple-silicon Mac:

| Format | Size | Time |
|---|---|---|
| CSV | 47 MB | 0.5 s |
| JSONL | 144 MB | 1.3 s |
| Parquet (zstd) | 0.6 MB | 0.5 s |
| SQLite | 89 MB | 8.7 s |

SQLite loads in one transaction with no journal, inserts each Square in one statement, and
builds its indexes after loading. Most of its time is the 1.57M `chart_houses` rows and their
index.

Output is written to a temporary file and renamed into place, so a failed export leaves nothing
behind. `make test` uses `-short` to skip the full-size export tests, which are slow under
`-race`. `make test-full` runs them.
