# Plan 09 — Timing: the Moment Against the Matter

Judges the moment of the reading (plan 08) against what the question is about (plan 07), in the
traditional manner: each point says whether it helps or hinders the matter, and why. No overall
verdict.

## Decisions

| Topic | Decision |
|---|---|
| Techniques | The planetary hour and day against the matter's natural planets; the Moon's condition. Not the lords of the houses, or their aspects (horary-style perfection): possible later |
| Natural planets | Per topic in `content/data/en.json` (`topics[].planets`), drafted from Lilly's general significators and the houses' traditional associations: marriage Venus, sickness Mars and Saturn, career the Sun, small animals Mercury… |
| The matter | The question's topics in the significator's house (for a derived house, in the subject's own: "the husband's career" is House 10's career, the Sun). If the topic isn't known, the house's first topic, its principal matter. No significator: only the Moon |
| Hour and day | Of a natural planet: helps. Otherwise of a benefic (Jupiter, Venus): helps. Of a malefic (Saturn, Mars): hinders. Of the Sun, Moon or Mercury: neither. No hour (polar days): noted |
| Void of course | No exact Ptolemaic aspect to a planet (not the nodes) before the Moon leaves its sign, the usual modern definition (Lilly's counts aspects within orbs). Hinders |
| Next aspect | Judged by kind (easy: sextile, trine; hard: square, opposition) and the planet's nature. Benefic: easy or conjunction helps, hard is mixed. Malefic: easy is mixed, hard or conjunction hinders. Sun, Mercury: easy helps, hard hinders, conjunction neither. Noted if the planet is a significator of the matter |
| Waxing, waning | Waxing helps (increase, beginnings); waning is mixed (decrease, endings) |
| Dignity, Sun | The Moon in domicile or exaltation helps, in detriment or fall hinders; cazimi helps; combust and under the beams hinder |
| Display | Terminal: "Timing" after the Heavens, each line marked ✓ ✗ ± · (ASCII + - ~ .). PDF: "Timing" below the lunations on the Heavens page if it fits, otherwise at the head of the Reading's page. JSON: `timing` on the reading |

## Code

- `sky.Moon`: `LeavesSign`, `NextAspect` (planet, kind, exact time) and `VoidOfCourse`, found by
  following the Moon hourly to the next cusp and narrowing each event to under a minute.
- `content.Topic.Planets`, and `Signification.Planets(topics)` with the fallback to the first.
- `chart.Timing(sky, subject, planets) []Judgement`; `chart.Reading.Timing`. The texts are
  English, written in code rather than content: a translation would need them moved.
- `internal/cli`: `matterPlanets` chooses the matter; `render.Timing`; `report` draws it.
