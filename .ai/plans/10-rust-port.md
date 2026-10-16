# Plan 10 — Port to Rust

Port the core logic and the `geomancer` CLI to Rust, behaving as the Go version does. Not ported:
`geomancer-export` (and with it Parquet, SQLite and CSV), and the build-only tools in
`internal/tools` (their outputs, `places.tsv.gz`, `vectors.bin` and the VSOP87 series, are data
we keep or no longer need).

The Go tests are the parity truth: their cases are ported one for one, and the golden files in
`internal/cli/testdata` are copied into the Rust repository unchanged.

This plan lives here; the Rust code will live in its own repository. Nothing is ported yet.

## Size

About 9,000 lines of Go to port (CLI 2,400, report 1,150, sky 880, chart 660, question 530, system
430, when 350, content 310, figure 320, place 290, the rest under 250 each), and 7,000 lines of
tests.

## Crates

All permissive. No GPL-family licence is needed anywhere: the Go build's GPL and LGPL modules
(freetype, textprocessing) came in through tdewolff/canvas, which this port replaces.

| Need | Go | Rust | Licence |
|---|---|---|---|
| Ephemeris: planets, Moon, mean node, speeds, ΔT, nutation, sunrise and sunset, event search | VSOP87D (ours) + soniakeys/meeus | `xalen-ephem` 0.6 (VSOP87A, ELP2000-82 per Meeus 47, IAU 2006/2000B, SMH 2016 ΔT, rise/set at −0°50′) | Apache-2.0 |
| Time, zones, DST gaps and overlaps | `time` + `time/tzdata` | `jiff` (bundled tzdb) | Unlicense OR MIT |
| Time zone from coordinates | `ringsaturn/tzf` | `tzf-rs` (same author) | MIT; data ODbL-1.0 as now |
| Command line | cobra | `clap` 4 | MIT OR Apache-2.0 |
| JSON | `encoding/json` | `serde`, `serde_json` | MIT OR Apache-2.0 |
| Stemming | kljensen/snowball | `rust-stemmers` (Snowball English) | MIT / BSD-3-Clause |
| Unicode normalization | x/text/norm | `unicode-normalization` | MIT OR Apache-2.0 |
| Places file | `compress/gzip` | `flate2` (miniz_oxide backend, pure Rust) | MIT OR Apache-2.0 |
| Terminal styling, colour profiles | lipgloss, colorprofile | `anstyle`, `anstream`; tables our own (see below) | MIT OR Apache-2.0 |
| Display width | (lipgloss) | `unicode-width` | MIT OR Apache-2.0 |
| Interactive prompts | huh, bubbletea | `crossterm` and a small form layer of our own (see below) | MIT |
| PDF | tdewolff/canvas | `krilla` (PDF, font subsetting) with `skrifa` metrics for our own line layout | MIT OR Apache-2.0 |
| PDF rendering in tests | canvas rasterizer | `hayro` (PDF to pixels, for the shading test and previews) | Apache-2.0 OR MIT |
| Tests | testing | built-in, plus `similar` for readable golden diffs | Apache-2.0 |
| Licence checking in CI | — | `cargo-deny`, with an allow-list | MIT OR Apache-2.0 |

### The ephemeris: xalen-ephem, behind a trait

`xalen-ephem` covers what `sky` needs from an ephemeris: apparent geocentric tropical longitudes
and speeds for the Sun, Moon and planets, the mean node, ΔT, nutation, Sun rise and set with the
same horizon we use (and the same method: ten-minute samples, then bisection, with
circumpolar and never-rises flags), and crossing and ingress search for lunations and the
Moon's course. Pure Rust, no data files, no network by default (the DE440 download is an
opt-in feature). Its README claims ≤0.76″ against DE440 for Mercury to Saturn and ~3″ RMS for
the Moon.

Risks: it's young (one release line, 0.6.0 of June 2026, single vendor; the umbrella crate's
releases are all yanked), so:

- `sky` reaches it only through a small `Ephemeris` trait (longitude and speed of a body at a
  moment, ΔT, Sun altitude), so it can be replaced.
- Version pinned exactly; `cargo-deny` watches it.
- The Go `sky` tests' references (CDS check values are VSOP87-specific and don't carry over; JPL
  Horizons longitudes within 10″, rise and set within the minute, Utqiaġvik's polar days) are
  the gate, whatever the backend.
- Fallback: port our Go pipeline (truncated VSOP87D, regenerated as Rust by the existing tool,
  and Meeus's lunar theory) behind the same trait. About 500 lines.

Ours, not xalen's, stay: dignities, closeness to the Sun, aspects with Lilly's moieties,
applying and separating, phases, void of course, planetary day and hour, and the timing
judgements. They're our rules, tested by our tests; xalen's versions (string-typed planetary
hours, its own orbs) differ.

### Terminal tables

The goldens are lipgloss's output to the byte. A table crate wouldn't match it, so a small
module of our own draws them: rounded borders, padding, faint grey borders, column widths by
`unicode-width` (counting the sign glyphs as lipgloss does).

### Prompts

huh's forms are the hardest part to port. They need:

- a filtered select (system, significator, Mothers by name or pattern)
- free-text fields with a live confirmation line ("→ Thursday, 15 October 2026")
- a typeahead over 168,000 places
- the daylight-saving choice
- going back to earlier steps, keeping answers
- the accessible, line-based mode

No crate does all of it (`inquire` and `dialoguer` have no back-navigation or live previews). So:
a small form layer on `crossterm`, aiming for the same steps, defaults, validation and wording,
not huh's exact look. The accessible mode is matched exactly, since tests read its transcript.

### Licence command

A build script reads `Cargo.lock` and `cargo metadata` at compile time and embeds each
package's name, version, checksum and declared SPDX licence. That's simpler and more exact than
the Go version's guessing from licence files. Same table and `--json`.

## Parity

1. **Unit tests**, ported case for case from the Go tests, package by package.
2. **Golden files**: copied from `internal/cli/testdata` into the Rust repository, with a
   script that checks the copies against this repository's (by hash) so they can't drift.
   The Rust CLI tests read the copies. Text must
   match byte for byte. JSON too, which means writing floats as Go does (shortest digits; an
   exponent below 1e-6 or from 1e21, with at least two exponent digits: `1e-07`), keeping Go's
   field order, and RFC 3339 times without fractions.
3. **Differential runs**, a script in the Rust repository driving both binaries, building the
   Go one from a checkout of this repository named by an environment variable:
   - `cast --json` for all 65,536 sets of Mothers under each system (131,072 charts), with and
     without significators: identical.
   - `suggest --json` for every question in the three test sets: identical.
   - The sky for 2,000 random moments and places from 1900 to 2100: positions within 2″; every
     displayed minute, time and judgement identical, except where Go's value is within 2″ (or 2
     seconds of time) of a boundary, which is reported but allowed.
4. **The PDF** isn't compared byte for byte: page count, the shading test, text extracted with
   `hayro`, and a side-by-side look at the previews.

The ephemeris, the stemmer and tzf are the likely sources of mismatch:

- **Ephemeris:** ΔT (SMH 2016 against our Espenak–Meeus, ~6 s in 2026, ~3″ of Moon).
- **Stemmer:** `rust-stemmers` may follow a different Snowball revision. Stem the whole lexicon
  and every test question in both first, and fix any differences before porting `question`.
- **tzf:** check its data format against the version Go pins.

## Phases

Each ends with its tests passing and, from phase 1, its goldens.

0. **Workspace and CI**: crates, `cargo test`, `clippy`, `rustfmt`, `cargo-deny`; the
   differential script's skeleton.
1. **Core**: `zodiac`, `figure`, `position`, `aspect`, `names`, `system` with Golden Dawn and
   traditional, `chart` (Shield, Square, focus, matter). `figures`, `figure`, `systems`, and
   `cast` from arguments, without the sky.
2. **Content and questions**: `content` (JSON as is), `question` (vectors file as is, the
   stemmer checked first). The accuracy thresholds and `suggest` goldens.
3. **Time and place**: `when` on jiff, `place` with tzf-rs; the 3,000-place zone agreement test.
4. **Sky and timing**, on xalen-ephem behind the trait; `cast` with `--date`, `--time`,
   `--place`.
5. **CLI**: all commands and flags, colour handling (`--color`, `NO_COLOR`, non-terminals), JSON,
   the licence command.
6. **Prompts**: the form layer, then each step; the accessible transcripts.
7. **PDF report**: fonts embedded as now, our line layout, the pages, the previews.
8. **Differential runs** in full, NOTICE and README, then cutover.

## Layout

Its own repository (the name is still to choose), a Cargo workspace:

- `geomancer` (library: zodiac, figure, position, aspect, system, chart, content, question,
  when, place, sky, timing)
- `geomancer-report` (PDF)
- `geomancer-cli` (the `geomancer` binary: commands, rendering, prompts, licences)

The data files (`places.tsv.gz`, `vectors.bin`, `en.json`, the fonts) are copied over and
embedded with `include_bytes!`; the same sync check covers them, and their notices go to the
new repository's NOTICE. Edition 2024 (xalen needs it), so Rust 1.85 or later.

## Settled

| Question | Answer |
|---|---|
| Where | The plan here; the code in a new repository |
| Prompts | Same behaviour (steps, order, defaults, validation, wording, back-navigation, the suggestion hint); accessible mode byte for byte; the look close, not exact |
| JSON | Byte for byte with Go's: Go's float formatting, field order and time format |
| The Go version | Kept as the reference, frozen, so the differential runs can be repeated, until there's confidence in the port |

## Still open

- The new repository's name.
- Whether `xalen-ephem` holds up: decided in phase 4 by the sky tests. If it doesn't, the
  fallback (our pipeline, ported) takes its place behind the trait.
- How the stemmer compares: decided at the start of phase 2.
