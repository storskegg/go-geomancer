# Notices

go-geomancer is licensed under the Apache License 2.0 (see LICENSE). It builds in the following
data, under their own licences.

| Data | Where | Source | Licence |
|---|---|---|---|
| US places: names, coordinates, populations | `place/data/places.tsv.gz` | [GeoNames](https://www.geonames.org), built by `internal/tools/places` | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) |
| Time zone boundaries | the `github.com/ringsaturn/tzf-dist` module | [timezone-boundary-builder](https://github.com/evansiroky/timezone-boundary-builder), via [tzf](https://github.com/ringsaturn/tzf) | [ODbL 1.0](https://opendatacommons.org/licenses/odbl/1-0/) (data); MIT (code) |
| Word vectors | `question/data/vectors.bin` | [GloVe](https://nlp.stanford.edu/projects/glove/) 6B, 100 dimensions, built by `internal/tools/vectors` | [PDDL 1.0](https://opendatacommons.org/licenses/pddl/1-0/) |
| Fonts | `report/fonts` | Libre Caslon Display, Crimson Text, Noto Sans Symbols | [SIL Open Font License 1.1](https://openfontlicense.org) (licences beside the fonts) |
| Time zone rules | Go's `time/tzdata` | the [IANA tz database](https://www.iana.org/time-zones) | public domain |
| Planetary theory | `sky/vsop87.go` | VSOP87D, P. Bretagnon and G. Francou, *Astron. Astrophys.* 202, 309 (1988), from [CDS catalogue VI/81](https://cdsarc.cds.unistra.fr/viz-bin/cat/VI/81), truncated by `internal/tools/vsop87` | published theory; no licence stated |

Places data © GeoNames, used under CC BY 4.0 and modified: filtered to populated places in the
50 states and DC, deduplicated, and reformatted. Time zone boundaries © OpenStreetMap contributors
and timezone-boundary-builder contributors, available under the Open Database License.

The Sun's and planets' positions use the VSOP87D series of the Bureau des Longitudes (now IMCCE),
truncated to the terms that matter between 1800 and 2200. The Moon, its node and the lunations use
the algorithms of Jean Meeus, *Astronomical Algorithms* (2nd ed., 1998), as implemented by
[github.com/soniakeys/meeus](https://github.com/soniakeys/meeus) (MIT).
