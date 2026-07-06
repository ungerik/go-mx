# pdf TODOs

Findings from the code review of the fpdf-inlining branch (all verified by
adversarial re-review; several reproduced with standalone programs).
"Inherited" means the defect existed verbatim in upstream
`codeberg.org/go-pdf/fpdf` v0.12.0.

## Correctness — all fixed

- [x] **CID width-array corruption via `slices.Delete`** — fixed by replacing
      the PHP-array port with the `cidWidthRange` struct: the
      `ws := *cidArray[key]` snapshot in `generateCIDFontMap` now copies a
      plain `interval bool` field instead of aliasing a slice that
      `slices.Delete` zeroes. Regression-tested in `cid_width_test.go` with
      the review repro (`/W [ 1 [ 500 600 700 700 ] 6 6 800 ]`).
- [x] **CI never ran the parity suite** — `.github/workflows/go.yml` now runs
      `go test -C fpdf` (byte-parity suite), `go test -C wordpress`, and
      `go build -C tools` in addition to the root-module tests.
- [x] **Panic on rune U+10000 in width measurement** — `GetStringSymbolWidth`
      guard corrected from `len(Cw) >= intChar` to `>`; the boundary rune now
      falls through to `MissingWidth` (inherited; regression test in
      `renderer_test.go`).
- [x] **TTF parser panics on malformed font bytes** — `parseFile` and
      `GenerateCutFont` recover parser panics into errors at the boundary, and
      `seekTable` reports a missing required table by name; malformed
      (possibly untrusted) font bytes now fail the font load through the
      sticky error instead of crashing the process (inherited; tests in
      `utf8fontfile_test.go`).
- [x] **`New`/`NewCustom` produced cp1252 mojibake** — the cp1252 translator
      is installed in the shared internal constructor, so every public
      constructor yields the same text encoding behavior (test:
      `TestConstructorsInstallTranslator`). Superseded eventually by Phase-3
      automatic text encoding.
- [x] **`parseNAMETable` hard-failed fonts for data it discards** — the parse
      was dead code (its names map was never consumed); deleted entirely, so
      spec-valid format-1 name tables load again.
- [x] **fsType license hard-error** — kept as a deliberate policy (embedding a
      restricted font is a license violation) and documented in
      `parseOS2Table`; see intentional divergences below.
- [x] **`UnicodeTranslator` strictness** — reverted to tolerance: malformed
      lines (comments, headers) in code-page `.map` files are skipped, as the
      legacy engine effectively allowed; only reader errors are returned. The
      doc contract and `util_test.go` updated.
- [x] **`SetCatalogSort` leaked map order for spot colors and page boxes** —
      spot color objects and the `/ColorSpace` resource dict now iterate in
      registration-id order, page boxes in sorted box-name order (inherited;
      determinism tests in `renderer_test.go`).
- [x] **Document header/footer callbacks outlived `Render`** — `applySetup`
      now always sets the callbacks, clearing them with nil when the document
      has no Header/Footer, so a second document rendered into the same
      renderer no longer inherits them. The narrower hazard of a ctx canceled
      between a manual `Render` and a later `Output` remains inherent to the
      callback design and is resolved by the planned Phase-3 header/footer
      Components.
- [x] **`replaceAliases` order-dependence** — aliases now replace longest
      first (ties lexicographic), so interacting aliases produce correct,
      deterministic output (inherited; test `TestReplaceAliasesInteracting`).

## Cleanup — all done

- [x] `rbuffer` and the PHP-array `untypedKeyMap` replaced with std types.
- [x] **`transform.go` misattribution** — the renderer convenience helpers
      (`NewRenderer*`, `Str`, `SetTranslator`, `LoadUTF8Font*`, `LineHeight`,
      `ensurePage`, `tr`) moved to `convenience.go`; `transform.go` now
      contains only the code the Wagner/Würmser attribution covers.
- [x] **errs stragglers** — `getters_test.go` uses `errs.New` and `color.go`
      `Hex` uses `errs.Errorf`.
- [x] **`keySort*` triplication** — replaced by `slices.Sorted(maps.Keys(m))`
      at the call sites; the three helpers are deleted.
- [x] **Unpadded license-table row** — the go-pdf/fpdf row in
      `THIRD-PARTY-LICENSES.md` is padded like its siblings.
- [x] **Hand-rolled big-endian readers** —
      `readUint16`/`readUint32`/`readInt16`/`getUint16` use
      `binary.BigEndian`; `readInt16`'s dead sign-adjust branch is gone.
- [x] **Parity-suite duplication** — `renderLegacy`/`renderNative` share
      `renderDeterministic` (via a small `renderer` interface), and both
      `assertParity` and `TestParityImages` compare through `assertSamePDF`,
      so the determinism knobs and failure reporting live in one place each.
- [x] **Dead field `fontDirStr`** — removed while moving the `Renderer`
      struct.
- [x] **Minor** — `generateCMAP` reuses `parseCMAPTable`'s subtable scan and
      returns the error where it is detected; `utf8toutf16` encodes in a
      single pass without intermediate slices; `repClosure` builds through
      `strings.Builder` (no final copy).
- [x] **Types moved to their matching files** — the `Renderer` struct lives in
      `renderer.go`, `colorMode`/`colorState` in `color.go`, and
      `spotColor`/`cmykColor` in `spotcolor.go`; `def.go` keeps only
      the types without a natural home file.

## Known intentional divergences from the legacy stack (documented, not bugs)

- `utf8toutf16` emits correct surrogate pairs for non-BMP text (legacy emitted
  garbage bytes) — a parity exception for emoji in metadata/bookmarks/UTF-8
  text.
- Fonts whose OS/2 fsType forbids embedding are rejected with an error; the
  legacy engine printed a warning to stdout and embedded them anyway.
- Malformed TrueType font bytes produce an error; the legacy engine panics.
- Spot colors, page boxes, equal-width images, and alias replacement are
  emitted in deterministic order; the legacy engine's order is random per
  process for these.
- `ImageBytes` components are re-renderable (fresh reader per render); legacy
  consumes a single reader. `ImageReader` remains single-use — consider
  buffering on first render instead of documenting the footgun.
