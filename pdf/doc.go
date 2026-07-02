// Package pdf is the native go-mx PDF rendering package. It will provide the
// same component model as the html package (Component, Document, If, ForEach,
// typed enums) rendering to PDF instead of markup, with no external PDF
// dependency.
//
// The PDF-generation engine in this package is inlined from
// [codeberg.org/go-pdf/fpdf] v0.12.0 (MIT license, see
// THIRD-PARTY-LICENSES.md in the repository root): the [Renderer] type is
// fpdf's Fpdf with the contrib bridges (gofpdi templates), the makefont
// tooling, and the HTML/SVG/grid extras removed. The component layer on top —
// mirroring the previous wrapper module, now kept at fpdf/ purely as the
// byte-for-byte parity baseline — is ported in a subsequent step, after which
// the engine's exported surface is aligned with it (typed enums instead of
// strings, header/footer components instead of callbacks, automatic UTF-8
// text handling).
package pdf
