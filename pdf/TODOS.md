# pdf TODOs

Findings from the code review of the fpdf-inlining branch (all verified by
adversarial re-review; several reproduced with standalone programs). Ordered
most severe first. "Inherited" means the defect exists verbatim in upstream
`codeberg.org/go-pdf/fpdf` v0.12.0; everything else was introduced on this
branch.

## Correctness

- [ ] **CID width-array corruption via `slices.Delete`** — `util.go`
      `untypedKeyMap.delete` now uses `slices.Delete`, which zeroes the vacated
      tail slot of the shared backing array. `generateCIDFontMap`
      (`renderer.go`, around the `ws := *cidArray[key]` snapshot) relies on the
      snapshot surviving the delete: upstream's reslice-truncation left it
      intact, `slices.Delete` writes `nil` into the `"interval"` slot, so
      `ws.getIndex("interval")` returns -1 and the `nextKey--` correction is
      skipped. Repro: widths CID1=500, CID2=600, CID3=CID4=700, CID5 unused,
      CID6=800 → upstream `/W [1 [500 600 700 700] 6 6 800]`, branch
      `/W [1 [500 600 700 700 800]]` — wrong glyph advances in embedded UTF-8
      fonts. Fix: capture `ws.getIndex("interval")`/`cws` before the delete, or
      deep-copy the snapshot slices.
- [ ] **CI never runs the parity suite** — `.github/workflows/go.yml` runs
      `go test -v ./...` from the repo root, which covers only root-module
      packages; the `fpdf` module (byte-parity suite), `wordpress` and `tools`
      are never built or tested in CI. Add per-module test steps.
- [ ] **Panic on rune U+10000 in width measurement** — `renderer.go`
      `GetStringSymbolWidth`: `len(r.currentFont.Cw) >= intChar` should be `>`.
      With the standard 65536-entry `Cw`, `GetStringWidth("\U00010000")`
      panics (index out of range). Reachable from user strings via
      `Text`/`Cell`/`MultiCell`. Inherited, but the branch rewrote these lines.
- [ ] **TTF parser panics on malformed font bytes** — `utf8fontfile.go`:
      `fileReader.Read` slices without bounds checks, `seekTable` derefs a nil
      `tableDescriptions[name]`, `GenerateCutFont` slices `postTable[4:16]`
      unchecked. `LoadUTF8FontBytes` with truncated/corrupt input crashes the
      process instead of setting the sticky error (repro: garbage after a valid
      magic). Inherited, but contradicts the branch's "errors reach the sticky
      error state" plumbing. Fix at the reader layer (bounds-checked reads that
      return errors).
- [ ] **`New`/`NewCustom` produce mojibake for cp1252 text** — `renderer.go`:
      both leave the `translate` field nil, so text components pass raw UTF-8
      to cp1252 core fonts (`"café €10"` → `cafÃ© â‚¬10`), while `NewRenderer`
      installs the translator. Either install the cp1252 translator in
      `fpdfNew`, or document loudly — resolved anyway by the planned Phase-3
      automatic text encoding.
- [ ] **`parseNAMETable` hard-fails fonts for data it discards** —
      `utf8fontfile.go`: the parsed `names` map is local and never used, yet
      spec-valid format-1 name tables (and odd-size records) now abort the
      whole document. Skip the quirk (or drop the dead parse entirely) instead
      of returning an error.
- [ ] **fsType license check now fails documents that legacy rendered** —
      `utf8fontfile.go` `parseOS2Table`: restricted-license fonts (fsType
      0x0002 / 0x0300 bits) previously embedded with a stdout warning; now the
      whole document errors. Probably the right policy — but decide explicitly
      and document it as a divergence from the legacy module (parity tests
      never load UTF-8 fonts, so nothing guards this).
- [ ] **`UnicodeTranslator` strictness poisons the renderer** — `util.go`:
      first-error-wins (per the doc contract) plus
      `UnicodeTranslatorFromDescriptor` assigning `r.err` means one malformed
      line in a user-supplied code-page `.map` file now fails the entire
      document, where legacy skipped the line. Consider skipping malformed
      lines (with the scanner error still fatal), or documenting the break.
- [ ] **`SetCatalogSort` still leaks map order for spot colors and page
      boxes** — `spotcolor.go` `putSpotColors`/`spotColorPutResourceDict`
      range over `spotColorMap` (object numbering varies run to run);
      `renderer.go` `putpages` ranges over `pageBoxes[n]`. Same disease as the
      equal-width image flake fixed on this branch. Inherited. Fix with a
      shared sorted-iteration helper for every map that feeds output
      (`spotColorType.id` already provides a deterministic key).
- [ ] **Document header/footer callbacks outlive `Render`** — `document.go`
      `applySetup` installs closures capturing `ctx` and never uninstalls
      them: a second Document rendered into the same Renderer inherits the
      first's header/footer, and a ctx canceled between `Render` and `Output`
      fails `Close`'s final footer with `context.Canceled`. Built-in flows
      (`Output`/`Bytes`/`OutputFile`/`ServeHTTP`) are safe. Inherited from the
      wrapper design — resolved properly by the planned Phase-3 header/footer
      Components.
- [ ] **`replaceAliases` is order-dependent for interacting aliases** —
      `renderer.go`: iterates `aliasMap` in random map order; when one alias is
      a substring of another (or a replacement contains another alias) output
      differs run to run. Standard single-`{nb}` usage is unaffected.
      Inherited. Sort keys (longest first) when iterating.

## Cleanup

- [ ] **`transform.go` misattribution** — the file's only header credits the
      transform authors ("translated from the work of Moritz Wagner and
      Andreas Würmser"), but the file now also contains the unrelated renderer
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
      file; `readInt16` carries a dead sign-adjust branch. Convert when adding
      the bounds-checked reader (see TTF panic item above).
- [ ] **Parity-suite duplication** — `fpdf/parity_test.go`:
      `renderLegacy`/`renderNative` are identical bodies (a tiny local
      interface covers both), and `TestParityImages` copy-pastes
      `assertParity`'s comparison block. The determinism knobs must stay in
      lockstep across copies.
- [ ] **Dead field `fontDirStr`** — `def.go`: no readers or writers (the live
      field is `fontpath`).
- [ ] **Minor** — `document.go` `NewRenderer`'s empty-value defaults duplicate
      the engine's own `""` fallbacks; `generateCMAP` duplicates
      `parseCMAPTable`'s scan and its missing-cmap error is created at the
      caller; `utf8toutf16` (2 intermediate slices) and `repClosure`
      (`strings.Builder` would save the final copy) allocate more than needed
      on hot text paths.

## Known intentional divergences from the legacy stack (document, don't "fix")

- `utf8toutf16` now emits correct surrogate pairs for non-BMP text (legacy
  emitted garbage bytes) — a parity exception for emoji in
  metadata/bookmarks/UTF-8 text.
- `putimages` orders equal-width images deterministically by content hash;
  legacy's order is random per process.
- `ImageBytes` components are re-renderable (fresh reader per render);
  legacy consumes a single reader. `ImageReader` remains single-use — consider
  buffering on first render instead of documenting the footgun.
