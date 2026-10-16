# Plan 08 — The Heavens at the Moment of the Reading

From the reading's date, time and place, work out the astrology of the moment: where the Sun,
Moon, planets and nodes are, their conditions and aspects, the Moon's phase and lunations, and
the planetary day and hour. Shown as tables in the terminal and on a page of the PDF. Relating
them to the question (the hour suiting the matter, and so on) is the next plan.

## Decisions

| Topic | Decision |
|---|---|
| Bodies | The seven classical planets and the Moon's two nodes (`zodiac.Planet`) |
| Zodiac | Tropical |
| Positions | Apparent geocentric ecliptic longitude, as sign, degree and minute: "Moon 14°22′ Scorpio". No houses, Ascendant or Midheaven for now |
| Nodes | The mean node (the traditional one; the true node differs by up to 1.5°); the South Node opposite. Nodes have no conditions and are always retrograde in mean motion, so motion isn't shown for them |
| Retrograde | At the moment only: a planet whose longitude is decreasing. No stations or shadow periods |
| Conditions | Essential dignity by sign, with traditional rulerships: domicile, exaltation, detriment, fall, or peregrine. And, for planets other than the Sun, closeness to the Sun: cazimi (within 0°17′), combust (within 8°30′), under the beams (within 17°) |
| Aspects | Ptolemaic: conjunction, sextile, square, trine, opposition, between the seven planets, within the sum of the two planets' moieties (half orbs, after Lilly: Sun 7°30′, Moon 6°, Mercury 3°30′, Venus 3°30′, Mars 3°45′, Jupiter 4°30′, Saturn 4°30′). Each with its orb and whether it's applying or separating. With the nodes, conjunctions only, within the planet's moiety |
| Moon phase | One of eight phases by the Moon's elongation from the Sun, in 45° sectors from the conjunction (the astrological convention: new, crescent, first quarter, gibbous, full, disseminating, last quarter, balsamic), with the percentage illuminated. Plus the previous and next new and full Moons, in the reading's local time |
| Planetary day | Ruled by the planet of the weekday, beginning at sunrise: before sunrise it's still the previous day's |
| Planetary hour | Day and night each divided into twelve equal hours, sunrise to sunset and sunset to the next sunrise, ruled in Chaldean order (Saturn, Jupiter, Mars, Sun, Venus, Mercury, Moon) from the day's ruler in the first hour. Shown: the hour's ruler, its number and whether it's of the day or night, and its start and end |
| Sunrise, sunset | Almanac definition: the Sun's upper edge on the horizon, allowing for refraction (centre at −0°50′) |
| No sunrise or sunset | Possible in northern Alaska. The day takes the weekday's ruler; the hour is shown as unavailable, with why |
| No date, time and place | Only possible with the Mothers as arguments and none of `--date`, `--time`, `--place`; the section is left out |
| Ephemeris | VSOP87 (chosen over JPL's approximate elements for accuracy), credited in NOTICE.md. `github.com/soniakeys/meeus/v3` (MIT) for the Moon, mean node, nutation, sidereal time and lunations |

## Display

### Terminal

A "The Heavens" section after the Square, before the Reading:

- A summary line: "Day of Mercury · Hour of Venus (3rd of the night, 7:12–8:16 pm) · Moon
  waxing gibbous, 78%".
- Bodies table: Body (glyph and name), Position, Motion (direct / retrograde), Condition
  ("Domicile · Combust").
- Aspects table: Aspect ("Moon □ Saturn"), Orb ("2°14′"), Applying / separating.
- Lunations: previous and next new and full Moons, with dates and local times.

`--json` adds `sky` to the reading.

### PDF

- Page 1's header gains one line: "Day of Mercury · Hour of Venus · Moon waxing gibbous in
  Scorpio".
- A new page, "The Heavens", after the Shield: the bodies table, the aspects, the Moon's phase
  and lunations, and the day and hour. In the report's classical style; planet glyphs from the
  symbol font if it has them, otherwise drawn as aspect symbols are.

## Code

- New core package `sky` (public, for the future server): `At(t time.Time, lat, lon float64)
  (Sky, error)`, returning positions, aspects, the Moon, and the planetary day and hour. No
  dependency on content or the CLI.
- `chart.Reading` gains `Sky *sky.Sky`, set by `cast` when the circumstances are known.
- `render.Heavens`, and the report page.
- NOTICE.md: credit for the ephemeris data, and the licence command lists the new module.

## Tests

- The library's own tests cover Meeus's worked examples.
- Positions at a handful of moments checked against JPL Horizons, to the arcminute; sunrise and
  sunset against USNO; lunation times against published tables, to the minute.
- Planetary day and hour: before and after sunrise, either side of sunset, at midnight, across a
  daylight-saving change, and in polar day and night.
- Aspects: orbs at the limits, applying versus separating, the nodes.

## Outcome

Built as planned.

- `sky.At(t, lat, lon)`: about 2 ms. The Sun and planets come from our own evaluation of the
  VSOP87D series (ecliptic of date, so no precession step), embedded as `sky/vsop87.go`: 2,544
  of 25,659 terms, cut at A × 0.2^k ≥ 1e-7 by `internal/tools/vsop87`, which leaves at most
  0.46″ of heliocentric error between 1800 and 2200. meeus's `planetposition` only reads the
  files from disk, so it isn't used. Then light-time, FK5, aberration and nutation (Meeus ch. 32,
  33). The Moon and mean node are meeus's.
- ΔT: Espenak and Meeus's NASA polynomials from 2005 (meeus's own after 2000 is ~34 s high
  for 2026), meeus's table from 1620 to 2005.
- Sunrise and sunset: the Sun's altitude sampled every ten minutes over three days and each
  horizon crossing bisected to under a second, which also finds polar days.
- Checked against the VSOP87D check values (CDS `vsop87.chk`), JPL Horizons longitudes for all
  seven planets (within 10″) and rise and set times (within the minute Horizons gives), and
  midsummer and midwinter at Utqiaġvik.
- Terminal: "The Heavens" after the Square. Sign glyphs are left out of the position column:
  ♈–♓ can be drawn as emoji, two columns wide, which breaks table alignment in some terminals.
- PDF: page 3, between the Shield and the Reading, and a line in page 1's header. The symbol
  font lacks ☉, so the Sun is drawn as a circle with a point; aspects use the Square's vector
  symbols, and ☌ the font's.
