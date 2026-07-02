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

## Cleanup

- [x] `rbuffer` and the PHP-array `untypedKeyMap` replaced with std types.
- [ ] **`transform.go` misattribution** — the file's only header credits the
      transform authors ("translated from the work of Moritz Wagner and
      Andreas Würmser"), but the file also contains the unrelated renderer
      helpers (`NewRenderer*`, `Str`, `SetTranslator`, `LoadUTF8Font*`,
      `LineHeight`, `ensurePage`, `tr`). Move the helpers to their own file so
      the attribution and file name stay truthful.
- [ ] **errs stragglers** — `getters_test.go` uses `errors.New` and
      `color.go` `Hex` uses `fmt.Errorf`; repo convention is
      `errs.New`/`errs.Errorf` (both files otherwise converted).
- [ ] **`keySortStrings`/`keySortInt`/`keySortArrayRangeMap` triplication** —
      `utf8fontfile.go`: all three are `slices.Sorted(maps.Keys(m))`.
- [ ] **Unpadded license-table row** — `THIRD-PARTY-LICENSES.md`: the new
      go-pdf/fpdf row is unpadded while sibling rows pad (repo markdown rule:
      pad tables up to 50-char columns).
- [ ] **Hand-rolled big-endian readers** — `utf8fontfile.go`
      `readUint16`/`readUint32`/`readInt16`/`getUint16` predate the
      `binary.BigEndian` conversion applied to the write side of the same
      file; `readInt16` carries a dead sign-adjust branch.
- [ ] **Parity-suite duplication** — `fpdf/parity_test.go`:
      `renderLegacy`/`renderNative` are identical bodies (a tiny local
      interface covers both), and `TestParityImages` copy-pastes
      `assertParity`'s comparison block. The determinism knobs must stay in
      lockstep across copies.
- [ ] **Dead field `fontDirStr`** — `def.go`: no readers or writers (the live
      field is `fontpath`).
- [ ] **Minor** — `generateCMAP` duplicates `parseCMAPTable`'s scan and its
      missing-cmap error is created at the caller; `utf8toutf16` (2
      intermediate slices) and `repClosure` (`strings.Builder` would save the
      final copy) allocate more than needed on hot text paths.

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
