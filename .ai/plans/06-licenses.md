# Plan 06 — Module Licences

`geomancer license` lists the modules the binary is built from, with their versions and
licences, as a table or with `--json`.

## Decisions

| Topic | Decision |
|---|---|
| Source | go.mod (requirements, direct or indirect) and go.sum (checksums), read when the list is generated and compiled into the binary; nothing is read at run time |
| Generation | `go generate ./internal/licenses` runs `internal/tools/licenses`, which writes `internal/licenses/modules.go`. `go:embed` can't reach go.mod from a subpackage, and licences need the module sources, so a generated file is used. `make tidy` regenerates it |
| Licence files | The first `LICENSE`, `LICENCE` or `COPYING` (optionally `.md`/`.txt`) at the module's root in the module cache (`go mod download -json`); `vendor/` is gitignored, so it isn't used |
| Identification | The file's own `SPDX-License-Identifier`, if any; otherwise phrases from the common licences (Apache-2.0, MIT, BSD-2/3-Clause, ISC, MPL-2.0, (L)GPL, Unlicense, WTFPL, and freetype's FTL-or-GPL). Anything else is `Unknown` |
| Staleness | `TestCurrent` checks modules.go against go.mod and go.sum, so a dependency change without regenerating fails the tests |
| Replace directives | Not supported; the generator stops with an error |
| Scope | Modules only. Built-in data (places, time zones, vectors, fonts) stays in NOTICE.md, which the output points to |

## Code

- `internal/licenses`: `Module` (path, version, sum, licence, indirect) and `All()`, sorted by
  path.
- `internal/tools/licenses`: the generator.
- `internal/cli/licenses.go`, `render.Licenses`: the command and its table.

## Known

`github.com/srwiley/scanx` has no licence file, so it's listed as `Unknown`.
