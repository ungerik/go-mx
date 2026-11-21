# TODOS

Deferred work, grouped by component and sorted by priority. The shadcn/ui port
build order lives in `shadcn/TODOS.md`.

Priorities: **P0** blocking · **P1** critical, this cycle · **P2** important ·
**P3** nice-to-have · **P4** someday.

## v1.0 release

Work found by [`V1-AUDIT.md`](V1-AUDIT.md) (2026-11-16) before go-mx tags
v1.0.0 and freezes its API. A title phrased as a question is a decision not
taken yet; answer it before the work under it starts. The audit section named in
each Context has the evidence. The `hx` items are under `## hx` and the `doc`
package question under `## doc`.

### Which modules get their own version line?

**What:** Decide between tagging every module in lockstep at v1.0.0, which
freezes every package of every module, and an exception: the root module goes to
v1.0.0 while new nested modules for `shadcn` (with `cva` and `twmerge`), `pdf`
(with `pdftable`) and `web` (with `schemaorg` and a renamed `mcp`) keep their own
v0.x line until they settle. The audit recommends the exception.

**Why:** Nested modules isolate dependencies, not versions: under lockstep a
breaking change in `shadcn`, such as a Tailwind v5 port, forces `/v2` on every
module, the root included. Moving a package out of a module that already has a
tagged version causes ambiguous-import errors, so the layout must be settled
before the first tag.

**Context:** V1-AUDIT.md §1 and D1. The split moves `cmd/shadcn-gallery`,
`cmd/example-form` and `cmd/example-sse` into the shadcn module, adds
replace + placeholder pairs, and changes the gallery CI job and the Tailwind
`@source` path. `logview` stays in the root module only after it drops its
`shadcn` import; `shadcn/clsx` can move to `shadcn/internal/clsx`.

**Effort:** S to decide; L to split
**Priority:** P1
**Depends on:** None

### Do the reflected forms freeze in v1, and as which package?

**What:** Decide whether the forms machinery (132 of the 233 top-level exports
of the root package) moves into a new `form` package that is frozen at v1, or
gets its own v0 module for more time. If it freezes, the audit's shape: names
without stutter (`form.Handler`, `form.Decider`, `form.Tag`, `form.Texts`, …),
one generic `form.Handler[T]` struct with `ServeHTTP` instead of three
constructors, deciders that get a `context.Context` and one struct argument,
`load`/`onSubmit` that get the `*http.Request`, an explicit render-probe flag,
parse helpers in `form/internal`, and the plain-HTML renderer inside `form`, so
`html` never imports the forms code.

**Why:** The callback signatures and the exported helpers are the parts that
cannot change after v1. Today `load` cannot read `r.PathValue("id")`, and
`IsRenderProbe` hides a marker in a slice's spare capacity.

**Context:** V1-AUDIT.md §3.1 (F1-F14) and D3. The core package needs nothing
from forms except three unexported helpers, so there is no import cycle.

**Effort:** L
**Priority:** P1
**Depends on:** Which modules get their own version line?

### Which configuration stays process-wide?

**What:** Decide which package-level variables survive v1. The audit's rule: a
limit becomes a constant; what varies per use becomes a field on the handler,
config or writer; what is request-scoped goes into the context; at most a couple
of deliberate defaults stay, behind `slog.SetDefault`-style setters
(`ReportError`). Open sub-questions: remove the `AsComponent`/`AsAttribs` hooks
(they run when the tree is built, so they cannot move into the writer or the
context)? Move `TextEscaper` onto `CheckedWriter`? Turn the `FieldDecider`
variables of `html`, `hx` and `shadcn` into `func` declarations and the CDN
scripts and themes into functions that return fresh values?

**Why:** Any imported package can flip these for the whole program, they cannot
differ per handler, and tests that set them cannot run in parallel.
`TextEscaper` lets one dependency weaken escaping everywhere.

**Context:** V1-AUDIT.md §2 C2, with the full list of variables.

**Effort:** M
**Priority:** P1
**Depends on:** None

### What is the final method set of `mx.Writer`?

**What:** Decide the `Writer` methods before the freeze: add explicit methods for
processing instructions and doctypes, or document a raw-write contract, and give
`FieldBehavior` a way to add attributes so `hx` no longer wraps the writer to
inject `hx-trigger`.

**Why:** `Writer` is implemented three times in the repo, so users will
implement or wrap it too, and a method added after v1 breaks them. Processing
instructions are detected today by sniffing a trailing `?>` in `Write`.

**Context:** V1-AUDIT.md §2 C3; `checkedwriter.go`, `logview/search.go`,
`hx/formdecider.go`.

**Effort:** M
**Priority:** P1
**Depends on:** None

### How do `shadcn` and `logview` support a strict Content-Security-Policy?

**What:** Decide the script design. The audit proposes behavior wired through
`data-*` attributes and one delegated script that a `Scripts()` component
renders once per page with the nonce from the context, `commandfor`/`command`
buttons for dialogs, and every browser function under one `window.mx`
namespace.

**Why:** `shadcn` emits 17 inline scripts per component instance, 52 inline
`on*` handlers and 11 inline styles, so sites with a strict CSP cannot use it or
`logview`. 23 of 31 `window` globals have no prefix (`toast`, `tabsSelect`, …),
and `if(!window.toast)` silently yields to an app's own `toast`. Both are
documented behavior once v1 is tagged.

**Context:** V1-AUDIT.md §2 C8 and §3.4 S2-S3.

**Effort:** L
**Priority:** P1
**Depends on:** Which modules get their own version line?

### Fix the `If`/`ElseIf` chain

**What:** Make `If(true, A).ElseIf(false, B).Else(C)` render `A`: carry a
"matched" flag through the chain. Remove `Iff` and `ElseIff`, which evaluate
their closure at once and so add nothing over `If(f(), …)`.

**Why:** `ElseIf` discards its receiver (`conditional.go:50`), so a true first
branch is lost: the chain renders `C`, or `B` for `ElseIf(true, B)`.

**Context:** V1-AUDIT.md §2 C1.

**Effort:** S
**Priority:** P1
**Depends on:** None

### Serve every handler type through `ServeHTTP`

**What:** Rename `HandleHTTP` to `ServeHTTP` on `html.Document`,
`xml.Document`, `web.Robots`, `Sitemap`, `SitemapIndex` and `LLMsTxt`; delete
`html.Serve`, `xml.Serve` and `Document.Serve`; make
`ComponentHTTPHandler` use the default writer for a nil `WriterFactory`;
give the handler constructors one header parameter shape; rename
`RespondNonContextError` and `mx.Middleware` after their purpose.

**Why:** Only `ServeHTTP` makes a type an `http.Handler`; the `Serve` helpers
start a server without timeouts; `ComponentHTTPHandler(c, nil, nil)` answers
every request with 500.

**Context:** V1-AUDIT.md §2 C5 and §3.6 W7.

**Effort:** S
**Priority:** P1
**Depends on:** None

### Remove the unsafe and unused `html`/`xml` API

**What:** Delete `html.Template` and rename `TemplateElem` to `Template`;
delete the `TableView` interface and `WidthPx`/`WidthEm`/`HeightPx`/`HeightEm`;
give `ScriptJS`, `ScriptModuleJS` and `StyleElem` (also in `svg`) an attribute
parameter and refuse `</script`/`</style` at render time; stop escaping script
content; make `xml.Decl` and `xml.Doctype` validate instead of building a `Raw`.

**Why:** `html.Template` runs `text/template` without escaping;
`StyleElem("</style><script>…")` renders a live script, `html.Script("a < b")`
breaks the JavaScript, inline code cannot carry a CSP nonce, `xml.Decl` can be
ended early by its argument, and `WidthPx` emits widths browsers misread.

**Context:** V1-AUDIT.md §2 C8-C9 and §3.2 H1-H5.

**Effort:** M
**Priority:** P1
**Depends on:** None

### Does routing ship in v1?

**What:** Decide whether `Route`, `TypedRoute`, `ComponentFuncHandler`,
`PathValueNames`, `JoinPath`, `JoinAbsPath` and `NameToPath` ship. The audit
recommends leaving them out, or shrinking `Route` to `http.Handler` plus
`Pattern()`.

**Why:** Nothing outside tests uses them except `web` calling `Register`;
`ComponentFuncHandler` binds only string fields and `PathValueNames` returns
`"rest..."` for `{rest...}`. Finishing them after v1 changes behavior.

**Context:** V1-AUDIT.md §2 C6; the `## mx handlers` entry about
`ComponentFuncHandler`.

**Effort:** S
**Priority:** P2
**Depends on:** None

### How are void elements marked?

**What:** Decide whether `Element` gets an explicit `Void bool` (or a separate
type) instead of meaning "void" by `Children == nil`.

**Why:** `br.Children = append(br.Children, x)` renders `<br>x</br>`; meaning
carried by nil vs. empty slice is easy to lose, and changing it after v1 breaks
callers.

**Context:** V1-AUDIT.md §2 C4.

**Effort:** S
**Priority:** P2
**Depends on:** None

### Do library errors keep the go-errs call stack in their text?

**What:** Decide whether go-mx returns plain errors with a package prefix
(`"mx: …"`) plus exported sentinels or types for the conditions callers branch
on, leaving the stack to the caller's own wrap, or keeps go-errs errors whose
`Error()` includes the call stack.

**Why:** The stack lands in users' logs and, with `RevealInternalServerErrors`,
on the page. Apart from `mcp.ErrPageNotFound` callers cannot `errors.Is`/`As`
anything. Error text is not semver API, so this can also wait.

**Context:** V1-AUDIT.md §2 C7 and D5.

**Effort:** M
**Priority:** P2
**Depends on:** None

### Which length unit does `pdf` use?

**What:** Decide between a typed `Length` (`pdf.Mm`, `pdf.In`, `pdf.Pt`) and one
fixed unit for the component layer.

**Why:** Every length is a bare `float64` in a unit chosen at runtime: the
Letter constructor uses inches, the A4 ones and `Document` millimeters, font
sizes are points, and an SVG's pixels become document units (a 24 px icon is
24 mm). Components cannot be reused across documents.

**Context:** V1-AUDIT.md §3.5 P5 and D4.

**Effort:** M
**Priority:** P2
**Depends on:** Which modules get their own version line?

### How do `shadcn` components take variants and state?

**What:** Decide whether variants stay positional parameters (40 functions,
e.g. `Button(variant, size, …)`) or the variant types implement `mx.Attrib` and
go in the variadic arguments, and settle the parameter conventions: the id first
everywhere (not 5th in `Slider` and 4th in `ToggleGroup`), one way to pass state,
hover-card delays once as `time.Duration`.

**Why:** A new upstream variant axis breaks every positional signature after v1.

**Context:** V1-AUDIT.md §3.4 S6-S7.

**Effort:** M
**Priority:** P2
**Depends on:** Which modules get their own version line?

### Should the minimum Go version drop to 1.26?

**What:** Decide whether `go.mod` says `go 1.26.0`, so both maintained Go
releases can use go-mx, with the stdlib-`uuid` test cases behind
`//go:build go1.27` and CI testing 1.26 and 1.27.

**Why:** Only the unused `doc/dbdoc` package and three test files need Go 1.27
(the `uuid` package); the library builds with Go 1.26.4.

**Context:** V1-AUDIT.md §4 item 1.

**Effort:** S
**Priority:** P2
**Depends on:** Finish or delete the unfinished `doc` package before v1.0?

### Which names does the core `mx` package keep?

**What:** Decide the names of the core types and helpers before the freeze:
`Attrib` (interface), `Attribute` (struct) and `Attribs` (map) differ only by
suffix (the struct could become `StaticAttrib`, or be unexported behind
`NewAttrib`); `If` takes `...Component` while element constructors take
`...any`; `ForEach` and `ForEachIter` could be one `iter.Seq` function.

**Why:** Each is a breaking rename after v1.

**Context:** V1-AUDIT.md §2 C10.

**Effort:** S
**Priority:** P2
**Depends on:** None

### Clean up `html` and `svg` naming before the freeze

**What:** One rule for element/attribute name collisions (the element wins, the
attribute gets `Attr`, with `Title`/`Style` as documented exceptions), one
spelling per casing question (`HR`/`Br`, `Readonly`/`AutoFocus`, `HRef`/`Href`,
…), shortcuts named `<Element><Attribute>` (`ButtonSubmit`, `LinkStylesheet`,
…), delete the 17 obsolete attributes and the 50 `…f` variants, add `U`, take
`int` for integer attributes, an ordered slice for `Document.Meta`, and
package doc comments for `html` and `hx`. Fix the float formatting of named
float types and NaN/±Inf.

**Why:** Each of these is a breaking rename or removal after v1.

**Context:** V1-AUDIT.md §3.2 H6-H14.

**Effort:** M
**Priority:** P2
**Depends on:** None

### Fix the `pdf` blockers before it is frozen

**What:** Move the fpdf engine to `pdf/internal/engine` behind a curated
`Renderer` of about 45 methods; remove the broken compare helpers and other
leftovers; replace the `SetDefault*` globals with `Document` fields; translate
text per current font; give `StyleRegular` and `BorderNone` non-empty values;
leave RC4 encryption out of v1.

**Why:** The exported `Renderer` is upstream fpdf's 217-method API; text in a
built-in font is garbled after a UTF-8 font is loaded; the globals are racy.

**Context:** V1-AUDIT.md §3.5 P1-P4, P6-P7; the SHOULD items P8-P11 (panics on
truncated images and fonts, FIPS mode, `Color`, Factur-X typing, Producer) are
in the same section.

**Effort:** L
**Priority:** P2
**Depends on:** Which modules get their own version line?, Which length unit does `pdf` use?

### Drop the release-candidate and unmaintained dependencies of `web`

**What:** Make `GlobPageSource.Dir` a `string` and delete the unused
`Page.Resources` and `Page.PathValues`, which drops go-fs (and with it fsnotify,
pkg/xattr and x/sys); split front matter in-house and decode it with
`go.yaml.in/yaml/v3`, which drops adrg/frontmatter and yaml.v2. Open
sub-question: keep TOML front matter (BurntSushi/toml) or drop it?

**Why:** A v1.0.0 module cannot export a type from go-fs v1.0.0-rc.8, and
yaml.v2 turns a key `n` into `"false"` and leaves dates as strings in
`Page.FrontMatter`, whose value types become API.

**Context:** V1-AUDIT.md §3.6 W1-W3.

**Effort:** M
**Priority:** P2
**Depends on:** None

### Fix the `web` API before the freeze

**What:** `RegisterRoutes(ctx, mux)`; look up `PageTypeRenderer` by media type
without parameters; fold `ServeMarkdown` into a `MarkdownDocument` with
`ServeHTTP`; rename the field `Site.MarkdownAlternates`; read
`Site.LLMsTxtOptions` instead of a duplicate parameter; drop `Site.Routes` and
`DefaultRenderPage`; `schemaorg.Price(amount, decimals)`; rename the `mcp`
package (the audit suggests `webmcp`), whose name collides with the SDK's; add
Examples.

**Why:** Each is a breaking change after v1; the `PageTypeRenderer` lookup is a
bug today (a `"text/markdown"` key never matches).

**Context:** V1-AUDIT.md §3.6 W4, W6, W8-W12.

**Effort:** M
**Priority:** P2
**Depends on:** Which modules get their own version line?

### Fix the smaller `shadcn`, `highlight` and `logview` API issues

**What:** Delete `PanicOnInvalidID` and always defer an invalid id with
`NewErrElement`; prefix the `ButtonSize` constants; one `Orientation` type;
delete `SectionCard` and unexport `StickToBottomDefaultThresholdPx`; move
`StickToBottom` down to `hx` or `html` so `logview` drops its `shadcn` import;
document the Tailwind `@source` line production builds need; rename
`highlight.GoSource` to `GenerateGo`.

**Why:** All are breaking after v1; the `@source` gap leaves production builds
unstyled today.

**Context:** V1-AUDIT.md §3.4 S4-S5, S8-S12.

**Effort:** M
**Priority:** P2
**Depends on:** Which modules get their own version line?

### Release hygiene for v1.0.0

**What:** CI runs `go vet`, staticcheck, `-race` and govulncheck for every
module, and after the tag `apidiff`/`gorelease` against the latest tag; add
`func Example`s for the main entry points; raise `svg`, `pdf` and `html` above
60% coverage; replace the `[Unreleased]` changelog with a short `[1.0.0]`
entry; write the compatibility statement; give every nested module a LICENSE;
add `docs/go.mod` so the gallery stays out of the module zip; delete `cmd/mx.go`
and the commented-out code.

**Why:** These make the v1 promise checkable and the first release readable.

**Context:** V1-AUDIT.md §0 and §4.

**Effort:** M
**Priority:** P2
**Depends on:** Which modules get their own version line?

### Non-breaking items of the v1.0 audit

**What:** Work through the LATER rows of V1-AUDIT.md (F14, H15, S13-S16, P12,
W13, and the `hx` parameter naming): additive helpers, performance of `Cn`,
browser tests for the untested scripts, docs drift.

**Why:** None of them changes the API incompatibly, so they can follow v1.0.0.

**Context:** V1-AUDIT.md §3. Progress on 2026-11-21: the `Cn` cost of S13
is gone for repeated class lists (`twmerge.Merge` caches its results; a
`Button` takes 1.6 µs instead of 20 µs), the `twmerge` table build at init
is not; S15 now has headless-Chrome tests for Command's keyboard handling
and the Drawer drag, the other scripts it lists still have none.

**Effort:** L
**Priority:** P4
**Depends on:** None

## mx reflected forms

### Core validation messages are English only

**What:** Let the field errors the core form handling writes itself follow
the request's language, as `mx.FormTexts` does for the handler's captions.

**Why:** A translated form shows "Speichern" on its button but "required",
"not a valid value" or "must be at least 3 characters" under a field.

**Context:** `errors.New("required")` in `formhandler.go` (three places) and
`html/formdecider.go`; the sensitive-field mask "not a valid value" in
`formhandler.go` (the `MaskSensitiveError` calls and the `mask` constant),
`errors.go` (`FieldErrorTexts`) and `html/formdecider.go` (its own masks and
`FormMessage("not a valid value")`); "the text is not a valid value" in
`formhandler.go`; the length and range messages in `formtext.go` and
`formnumber.go`. Check what compares against these texts before replacing
them; the messages with a number need a format string, like shadcn's
`Labels.FormDateTimeTooLate`.

**Effort:** M
**Priority:** P3
**Depends on:** None

### Two reflected forms of one struct on a page share element ids

**What:** Let a reflected form prefix the element ids it derives from field
paths, separately from the input names.

**Why:** Every field's id is its path (`Debit`), so two forms of the same
struct on one page (an inline edit form per table row) repeat ids. A
`<label for>` then names the first form's control, and a
`form:"widget=combobox"` or `widget=datepicker` trigger opens the first
form's popover, so a pick changes the other form's value.

**Context:** Found by the /ship review of the Combobox branch
(2026-11-21). `shadcnField`/`FieldLabelFor(string(path))`, `comboboxField`
(`Name: string(c.path)`, the id defaults to it) and `datePickerField` all
take the path as id; a prefix would go into `mx.ReflectFormConfig` or the
render context and be applied where ids are written, not to names.

**Effort:** M
**Priority:** P3
**Depends on:** None

## mx writer

### `CheckedWriter.Comment` writes text that closes the comment early

**What:** Reject, or defer as an error like `xml.Comment` does, comment text
holding `-->` or `--!>`, and decide whether to also enforce the rest of the
HTML comment-text rules (not starting with `>` or `->`, no `<!--`, not ending
in `<!-`).

**Why:** `Comment` writes its text verbatim between `<!-- ` and ` -->`, so a
caller passing text that contains `-->` ends the comment there and the rest
is emitted as markup. `xml.Comment` checks for `--` before calling it; a
caller using the `Writer` directly gets no check.

**Context:** `CheckedWriter.Comment` in `checkedwriter.go`. The padding
spaces it adds already make the start-with-`>` and end-with-`-` rules
harmless, so only the closing sequences can break the document. Found as a
`TODO` comment listing the XML rules, which are stricter (no `--` at all).

**Effort:** S
**Priority:** P3
**Depends on:** None

### Element names are not checked

**What:** Check element names against the HTML/XML name rules in
`CheckedWriter.BeginElement`, which today rejects only an empty name.

**Why:** A name holding a space, `>`, `=` or a quote is written as-is and
breaks the markup. Names come from Go code almost always, but `xml.Element`
and a hand-built `mx.Element` can carry a computed one.

**Context:** The `TODO regex for valid element name` comment in
`checkedwriter.go`. Attribute names are already checked:
`CheckedWriter.Attribute` and `Attribute.Validate` refuse every name
`mx.ValidateAttribName` rejects; an element-name check can follow the same
byte loop. Check its cost on the render hot path; an allow-list writer
(`allowedElems`) already skips the check for elements.

**Effort:** S
**Priority:** P4
**Depends on:** None

### Attribute names: case, UTF-8 and the order of checks

**What:** Four loose ends of the attribute-name check:
- `CheckedWriter.Attribute` refuses a duplicate name only in the same letter
  case, while `html.Document` refuses `class` beside `CLASS`; HTML names are
  ASCII case-insensitive, XML names are case-sensitive, so the writer would
  need to know its target. `strings.EqualFold` in `html.Document` also folds
  Unicode (`claſs` matches `class`), where HTML lowercases only ASCII.
- `mx.ValidateAttribName` accepts invalid UTF-8 and names that are not valid XML
  names (a leading digit, `?`). They cannot break out of the tag, but make XML
  output not well-formed; its doc comment says the check "holds for every
  markup target", which reads as a well-formedness promise.
- `html.Document` checks a name before computing the value, so an
  `mx.ErrAttrib` with an empty `Name` reports "empty attribute name" instead
  of its own error.
- A failing `BodyAttribs` value errors after the `<head>` is written (as
  documented), while `HTMLAttribs` values are all computed before any output.

**Why:** None is reachable from the repo's own callers today; each is a
surprise for a downstream caller building names or values at runtime.

**Context:** `checkedwriter.go` `Attribute`, `attrib.go` `ValidateAttribName`,
`html/document.go` `checkAttribNames`/`Render`. Found by the adversarial
review of the `html.Document` attributes change.

**Effort:** S
**Priority:** P4
**Depends on:** None

## mx handlers

### `ComponentFuncHandler` fills only string fields

**What:** Convert path and form values into the non-string fields of the
handler's argument struct (ints, bools, `encoding.TextUnmarshaler`s), and
answer 400 for a value that does not convert; decide what a non-struct `T`
should mean.

**Why:** A field such as `ID int` is silently left zero today, so a
`/items/{ID}` route renders item 0 instead of failing.

**Context:** `ComponentFuncHandler.ServeHTTP` in `route.go`, at the two
`TODO` comments. The reflected-form binders in the root package already
convert strings to every kind; reuse them rather than writing a third set.

**Effort:** M
**Priority:** P3
**Depends on:** None

## mx components

### A failing attribute passed as a child does not say where it was

**What:** Wrap the error `DefaultAsComponent`'s `Attrib` case returns from
`Render` with the attribute name and the fact that it was passed as a child,
keeping the chain (`%w`) so `errors.Is` and `UserMessage` still work.

**Why:** An `mx.ErrAttribf("autocapitalize", "invalid value %q", v)` passed
to `ModifyOnRender` or `AsComponents` fails the whole page with only
`invalid value "x"`, which reads like a bad attribute on a real element; the
developer searches the element attributes instead of the misplaced child.

**Context:** `component.go`, the `Attrib` case of `DefaultAsComponent`.
`ErrAttribf` errors are plain stdlib errors on purpose (no call stack), so
`fmt.Errorf` may fit better than `errs.Errorf`. Found by the adversarial
review of the attributes-as-children change.

**Effort:** S
**Priority:** P4
**Depends on:** None

### A typed-nil `AttribProvider` that handles a nil receiver is skipped

**What:** Say in the `DefaultAsAttribs` docs that a typed-nil
`AttribProvider` yields no attributes without its `Attribs` method being
called, or call the method when it can handle a nil receiver.

**Why:** A type whose `Attribs` returns defaults for a nil receiver used to
get them; now it gets nothing, silently. No type in the repo does this.

**Context:** `attribs.go`, the `AttribProvider` case (`isNilPointerOrFunc`).
The `Value:` card of `TestDefaultAsAttribsNilFunc` in `attribs_test.go`
still names only the nil funcs, not the iterator and provider rows. Found
by the second /ship review of PR #48.

**Effort:** S
**Priority:** P4
**Depends on:** None

### A reassigned `AsAttribs` is not asked about func children

**What:** Document in the `AsAttribs` and `AsComponent` docs that
`DefaultAsComponent` converts its six func signatures (`func(context.Context,
Writer) error`, `func() Component`, …) before it asks `AsAttribs`, or ask
`AsAttribs` first again.

**Why:** A program-wide `AsAttribs` replacement that recognizes one of
those func types is no longer consulted for it when the value is passed as
a child. Nothing changes under the default `AsAttribs`, which accepts none
of them, and nothing in the repo reassigns it.

**Context:** `component.go`, the order of the func cases and the attribute
case in `DefaultAsComponent`. Found by the pass-2 review of PR #48.

**Effort:** S
**Priority:** P4
**Depends on:** None

## html

### A misplaced `lang` in `NewDocument` fails after `<head>` is written

**What:** Let `Document.Render` return the error of an `mx.ErrAttrib` in
`BodyAttribs` (or `HTMLAttribs`) before it writes anything, as it does for a
nil, invalid or duplicate name.

**Why:** `NewDocument` turns a second or context-dependent `lang` into an
`mx.ErrAttrib`, an error known when the document is built, but `Render`
writes the doctype and the whole `<head>` before it fails at `<body>`. A
caller rendering straight to a `ResponseWriter` sends a truncated 200 page;
the go-mx handlers buffer and are not affected.

**Context:** `html/document.go`, `checkAttribNames` and `NewDocument`.
Documented today as the behavior of a failing value. Found by the
adversarial review of the `NewDocument` attributes change.

**Effort:** S
**Priority:** P3
**Depends on:** None

### The `NewDocument` hint outlives a replaced `BodyAttribs`

**What:** Add the "(NewDocument puts its attribute arguments there)" hint
only for the entries `NewDocument` itself put into `BodyAttribs`.

**Why:** `bodyAttribsFromArgs` stays set when a caller later replaces
`BodyAttribs`, so an error about a name the caller added still blames
`NewDocument`. This affects the error message only.

**Context:** `html/document.go`, `Render` and `bodyAttribsFromArgs`; record
the count `NewDocument` produced instead of a flag. Found by the second
/ship review of PR #48.

**Effort:** S
**Priority:** P4
**Depends on:** None

### The custom `BoolAttrib` example in the how-to uses `hx-boost`

**What:** Replace `html.BoolAttrib("hx-boost")` in `docs/html/how-to.md`
with an attribute whose presence alone is its meaning, such as a `data-`
flag.

**Why:** htmx enables boosting only for `hx-boost="true"`, so markup copied
from the example (`hx-boost="hx-boost"`) silently does nothing.

**Context:** `docs/html/how-to.md`, the custom boolean attribute example.
Found by the second /ship review of PR #48.

**Effort:** S
**Priority:** P4
**Depends on:** None

## hx

### Does `hx` target htmx 2 or htmx 4 at v1.0?

**What:** Decide which htmx major `hx` binds when v1.0 freezes its API:
either keep it as the htmx 2 binding, say so in the package doc, and add
htmx 4 later as a separate package, or port `hx` to htmx 4 before the tag.

**Why:** htmx 4.0.0 shipped on 2026-08-28 and changes about half of what
`hx` exports. `hx-disable` now means what `hx-disabled-elt` meant, so
`hx.Disable` would silently change meaning. `hx-params`, `hx-request`,
`hx-history`, `hx-inherit`, `hx-disinherit` and `hx-ext` are gone, and
`hx-prompt` became an extension. The request header `HX-Trigger` became
`HX-Source`, and `HX-Trigger-Name` and both `HX-Trigger-After-*` response
headers are gone. Every event is renamed (`htmx:after:swap`). After the v1
tag, a port is a breaking change; a separate htmx 4 package is not. npm
still publishes 2.x as `latest` (2.0.11) and 4.x as `next` (4.0.1), so
staying on htmx 2 for v1 is defensible.

**Context:** the CDN pins are in `hx/htmx.go`. The affected API is in
`hx/attributes.go`, `hx/headers.go` and `hx/events.go`. Consumers:
`hx.FieldDecider`, which shadcn uses only for `hx-trigger="change"`;
`logview`; and the SSE examples `cmd/example-sse` and
`cmd/example-logstream`. Check whether htmx 4's streaming support replaces
the `htmx-ext-sse` extension. htmx 2.0.11 already backports the
`<hx-partial>` response tag, a candidate for an `hx.Partial` element either
way. Upstream ships an upgrade checker:
`npx htmx.org@4.0.0 upgrade-check`. Found by the v1.0 audit on 2026-11-16
(V1-AUDIT.md §3.3, D2, which recommends staying on htmx 2); no earlier
evaluation of htmx 4 exists.

**Effort:** S to decide and document; L to port.
**Priority:** P2
**Depends on:** None

### The `hx` event constants do not work with `On`

**What:** Make `On`/`OnHTMX` convert an `htmx:` event constant to the
kebab-case name htmx listens for, or correct the doc of `hx/events.go`, which
shows `On(hx.EventAfterSwap, …)`. Also stop routing `shadcn.FieldDecider`
through `hx.FieldDecider` only to add `hx-trigger="change"`.

**Why:** `On(hx.EventAfterSwap, …)` renders `hx-on:htmx:afterSwap`; browsers
lowercase attribute names, so the handler never fires.

**Context:** `hx/events.go:3-5`, `hx/on.go`, `shadcn/formdecider.go:33`;
V1-AUDIT.md §3.3.

**Effort:** S
**Priority:** P2
**Depends on:** Does `hx` target htmx 2 or htmx 4 at v1.0?

## shadcn

### `Combobox` options have no groups and cannot be disabled

**What:** Let a `shadcn.Combobox` list its options under group headings
(`CommandGroup`) and dim single options (`data-disabled="true"`).

**Why:** A chart of accounts reads best grouped by account class, and some
accounts may be closed for booking while they stay listed.

**Context:** `mx.NamedOption` has only `Name` and `Value`, and
`ComboboxOptions.Options` builds one untitled `CommandGroup`. A group
field, or a separate `ComboboxOptions.Groups`, would also serve
`<optgroup>` in the reflected select. Command already skips disabled items.

**Effort:** S
**Priority:** P3
**Depends on:** None

### A `Combobox`'s own `aria-invalid` outlives an htmx history restore

**What:** Clear the `aria-invalid` a `shadcn.Combobox` script set after a
failed required check also when the markup was restored from a snapshot.

**Why:** The script remembers that it set the mark only in the `mxInvalid`
property, which an htmx history snapshot drops. After a restore the mark
reads as one the server set and stays after a valid pick: the trigger
looks and is announced invalid. The submitted value is not affected.

**Context:** Found by the /ship adversarial review (2026-11-21). A
`data-mx-invalid` attribute in `comboboxScript`, set and cleared with the
mark, would survive the snapshot.

**Effort:** S
**Priority:** P4
**Depends on:** None

### `Combobox` renders every option into the page

**What:** Load a `Combobox`'s matches from the server as the user types,
for option lists of thousands of entries.

**Why:** Each option is a `CommandItem` in the markup, so a long list
makes the page large in every form that shows the field; a few hundred
entries are fine.

**Context:** `Command` already picks up items an htmx swap puts into its
`CommandList` (its `MutationObserver`), so the search input could carry
`hx-get` with a `keyup changed delay` trigger. The combobox script reads
the label of a pick from the item, so swapped items work as they are; the
selected option's label must still render server-side.

**Effort:** M
**Priority:** P4
**Depends on:** None

### Scripted `shadcn` components repeat their script in every instance

**What:** Emit the definition of a component's client script once per
render (or once per page), and only a short init call per instance.

**Why:** Every `Combobox` carries the whole `comboboxScript` (about 6 KB)
plus, through its `Command`, the whole `commandScript`; their guards keep
the definitions from running twice, but the bytes are repeated. A form with
eight `widget=combobox` fields, or a table of inline row-edit forms, sends
eight copies of both in every response and every htmx swap. `Tabs`,
`ToggleGroup`, `Calendar`, `Menu`, `Tooltip` and the other scripted
components follow the same pattern.

**Context:** Found by the Combobox code review (2026-11-21). mx has no
render-scoped state to record that a script was already written, and an
htmx fragment rendered alone still needs the definition, so the flag
belongs to one render, not to the page: a context value the render entry
points set (a `*sync.Map` or a small registry), consulted by a
`scriptOnce(name, js)` component, with the definition emitted as is where
the context has no registry. The strict-CSP question in `## v1.0 release`
(how `shadcn` and `logview` support a Content-Security-Policy) may move the
scripts into an external asset instead, which would settle this too.

**Effort:** M
**Priority:** P3
**Depends on:** The strict-CSP decision under `## v1.0 release`

### A `Combobox` that picks several values

**What:** A multi-select combobox, upstream Base UI's chips variant.

**Why:** Tags and similar sets want search too; the reflected form renders
an enum set as a checkbox grid.

**Context:** The single-option `<select>` would become a `<select
multiple>`, the trigger would show chips, and the popover would stay open
after a pick, as a `CalendarMultiple` does.

**Effort:** M
**Priority:** P4
**Depends on:** None

## web

### A body of only nil entries hides the page content

**What:** Let `PageTypeRenderer.RenderPage` treat a `BodyChildren` whose
entries are all nil like an empty one.

**Why:** A page renderer returning `mx.Components{x}` with a nil `x` (an
unset optional part) counts as having a body, so `Page.Content` is skipped
and the page renders an empty `<body>` without an error.

**Context:** `web/pagetyperenderer.go`, the `len(doc.BodyChildren) > 0`
check. Found by the second /ship review of PR #48.

**Effort:** S
**Priority:** P4
**Depends on:** None

### HTML pages of a `GlobPageSource` have no title

**What:** Read the `<title>` (or the first `<h1>`) of an `.html` page found
by a `GlobPageSource` into `Page.Title`.

**Why:** A Markdown page takes its title from front matter and a `.txt` page
from its file name, but an HTML page gets none, so its sitemap entry,
llms.txt line and MCP `list_pages` result show an empty title.

**Context:** The `TODO parse HTML title` in `web/globpagesource.go`. The
content is read only `withContent`, so the title needs a bounded read of the
head, or the file-name fallback the `.txt` case uses.

**Effort:** S
**Priority:** P3
**Depends on:** None

## doc

### Finish or delete the unfinished `doc` package before v1.0?

**What:** Finish or remove `doc/` and `doc/dbdoc` before the v1.0 freeze (the
v1.0 audit recommends deleting them): `DefaultRenderer` is nil,
so `RendererFromContext` returns nil for a context without a renderer and
the caller panics; `dbdoc.Color` is a placeholder; the `Font.LineHeight`
column is tagged `line_hight`.

**Why:** Untested, partial API reads as supported to a newcomer, like the
dead route code removed on 2026-11-01.

**Context:** The `TODO`s on `DefaultRenderer` in `doc/config.go` and on
`Color` in `doc/dbdoc/font.go`. Nothing in the repo imports `doc`. A v1.0.0 tag
would freeze it as it is; V1-AUDIT.md §1 and §3.6 W5.

**Effort:** S to delete; M to finish
**Priority:** P2
**Depends on:** None

## Completed

### `NewDocument` repeats `NewElement`'s argument sorting

**What:** Share the sorting of `attribsChildren ...any` into attributes and
children between `mx.NewElement` and `html.NewDocument`.

**Why:** `NewDocument` re-implements the loop (nil skipped, `AsAttribs`,
`AsComponent` with nil dropped) with two differences: a `Component` is
checked before `AsAttribs`, and `lang` is moved to `<html>`. A later change
to one loop can silently miss the other.

**Context:** `element.go` `NewElement`, `html/document.go` `NewDocument`.
A shared helper would need a hook for the `lang` handling and the
`Component`-first order. Found by the review of the `NewDocument`
attributes change.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-15 — the new exported `mx.AsAttribsOrComponent` sorts one argument; `NewElement` is its loop, and `NewDocument` checks for a `Component` first and post-processes the attributes (nil skip, `lang`).

### Edge values of a `lang` argument to `NewDocument`

**What:** Decide how `NewDocument` handles `Lang("")` (consumed silently
and counted as the one allowed `lang`, so a later `Lang("de")` is an error),
`ConstAttrib("lang")` (renders `<html lang="lang">`) and a `*mx.Attribute`
lang (refused although its value is static).

**Why:** Each is a plausible caller mistake that now renders something
unexpected or refuses a valid value, without a clear error.

**Context:** `html/document.go`, `NewDocument`'s lang branch. Found by the
adversarial review of the `NewDocument` attributes change.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-15 — decided won't fix: `ConstAttrib("lang")` keeps rendering `lang="lang"`, and `Lang("")` keeps being consumed as the one allowed `lang` (it also blocks a later one). A `*mx.Attribute` lang stays refused; `Document.Lang` is the way to set a computed language.

### Attribute text bypasses the writer's text escaping

**What:** Let the attributes `DefaultAsComponent` renders as children go
through the `Writer`'s text escaping instead of `mx.Raw`.

**Why:** `AttribStringContext` escapes name and value with the
single-quote attribute escaper, and the result is written raw, so a writer
with a different `TextEscaper` (or a text context with other rules) does not
get its own escaping for this text. Escaping is meant to be the writer's job.

**Context:** `component.go`, the attribute case of `DefaultAsComponent`;
`attrib.go`, `formatEscapedAttrib`. Writing the unescaped `name='value'` as
`Text` would escape once through the writer, but `AttribStringContext` would
then need an unescaped variant. Found by the review of the
attributes-as-children change.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-15 — decided won't fix: the `name='value'` text of attributes passed as a child stays escaped by `AttribStringContext` and written as `mx.Raw`, not through the writer's `TextEscaper`.

### A typed-nil pointer attribute panics

**What:** Treat a typed-nil pointer `Attrib` (a `*T` with pointer receivers
holding nil) like a nil `Attrib`, or report it as an error, in
`html.NewDocument`, `DefaultAsComponent` and `Element.Render`.

**Why:** `html.NewDocument(title, (*T)(nil))` calls `AttribName` while it
builds the document and panics; `mx.AsComponents((*T)(nil))` and an element
holding one panic when rendered. Handlers recover, direct callers do not.

**Context:** `html/document.go` (`NewDocument`'s attribute loop),
`component.go`, `element.go` (`Render`'s attribute loop). Needs a reflect
nil-pointer check. Found by the adversarial review of the `NewDocument`
attributes change; `Element.Render` predates it.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-15 — the new `mx.IsNilAttrib` catches a nil interface and a typed nil pointer or func. `Element.Render` refuses one with its "nil Attrib" error, `Element.AttribIndex`, `NewDocument` and `DefaultAsComponent` leave it out, and `html.Document.Render` refuses one in `BodyAttribs`/`HTMLAttribs`.

### The `NewDocument` hint also appears for `Document` literals

**What:** Add the "(NewDocument puts its attribute arguments there)" hint
only to errors about a document `NewDocument` built.

**Why:** A caller who filled `BodyAttribs` in a `Document` literal gets a
hint about a constructor they never called.

**Context:** `html/document.go`, `newDocumentHint` in `Render`; needs a flag
on the `Document` that `NewDocument` sets. Found by the adversarial review
of the `NewDocument` attributes change.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-15 — `Document` remembers in an unexported field that `NewDocument` filled `BodyAttribs`, and `Render` adds the hint only then, and never to a nil entry, which `NewDocument` leaves out.

### A `ConstAttrib` without `=` panics

**What:** Make `ConstAttrib.AttribName` and `AttribValue` handle a string
without `=` (for example `mx.ConstAttrib("hidden")`) instead of slicing at
index -1.

**Why:** `string(a)[:strings.IndexByte(string(a), '=')]` panics with
`slice bounds out of range [:-1]`, so one malformed constant crashes the
render of the whole page; `html.Document` reaches it before writing anything,
through the `BodyAttribs`/`HTMLAttribs` name check.

**Context:** `attrib.go`, `ConstAttrib`. Either treat the whole string as the
name with an empty value (a boolean-style attribute) or return an empty name,
which `mx.ValidateAttribName` then refuses with an error. Found by the
adversarial review of the `html.Document` attributes change; predates it.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-15 — a `ConstAttrib` without `=` is both name and value, as a boolean attribute: `ConstAttrib("hidden")` renders `hidden="hidden"`. A non-boolean one missing its value (`ConstAttrib("data-id")`) therefore renders `data-id="data-id"` instead of failing; the adversarial review noted this, and it was kept deliberately.

### An empty body from the page renderer hides the page's content

**What:** Let `PageTypeRenderer.RenderPage` treat a `BodyChildren` that is an
empty `mx.Components` like a nil one, so the content renderer still runs.

**Why:** It checks `doc.BodyChildren != nil`; `html.NewDocument(title)` with no
body arguments stores an empty `mx.Components`, a non-nil interface, so the
page renders an empty `<body>` without an error and the `Page.Content` is lost.

**Context:** `web/pagetyperenderer.go:40`. Found by the adversarial review of
the `html.Document` attributes change; predates it (the field was `Body`).

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-15 — fixed at the cause instead: `html.NewDocument` leaves `BodyChildren` nil when it gets no children (also when it gets only body attributes), so `PageTypeRenderer` fills in the content. `BodyChildren` is now an `mx.Components` slice, and `PageTypeRenderer` checks `len(doc.BodyChildren) > 0`, so an explicitly empty one no longer hides the content either.

### Core repeatable texts are English only

**What:** Let the texts the core package renders for a repeatable field,
"Add row", "Remove" and the empty readonly field's "None"
(`RepeatableNoRowsText`), follow the request's language.

**Why:** A shadcn form translates every built-in string through `Labels`,
but these three come from the core `renderRepeatable`, which has no access
to them, so a German form shows English row buttons.

**Context:** `repeatableAddButton`, `renderRepeatableRow` and the
`RepeatableNoRowsText` line in `repeatable.go`. A context hook read at render
time (as shadcn's `ContextWithLabels`) would let a renderer install them.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-14 — the new `mx.FormTexts` (`ContextWithFormTexts`, `FormTextsFromContext`, `DefaultFormTexts`) holds the texts the handler renders around the fields: the row buttons and "None", and also the submit caption, the generic form error and the retype notice, which were English-only package variables too. It replaces `FormSubmitLabel`, `FormErrorMessage`, `FormRetypeSensitiveMessage` and `RepeatableNoRowsText`.

### The date picker trigger is not described by its errors and help

**What:** Give the trigger button of a `form:"widget=datepicker"` field the
`aria-describedby` of its help and error messages, and the `hx-trigger` the
htmx decider gives every other control.

**Why:** A screen reader announces an invalid date picker without saying why:
its `role="alert"` error is not linked to the button, so on a full page load
it is never read. And an htmx form that validates on change skips the picker.

**Context:** `datePickerField` in `shadcn/formdecider.go` builds its markup in
a `mx.ComponentFunc`, because it reads `ContextWithDatePickerOptions` while
rendering, so `mx.AddDescribedBy` and the hx decider cannot find the trigger.
Pass the ids into the closure, or build the trigger outside it. Found by the
/ship review (2026-11-01).

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `datePickerField` is a component type that `shadcnFieldMarked` hands the help, stored-value and error ids, so the trigger links them with `aria-describedby`; `DatePickerCalendar` gives its hidden input `hx-trigger="change"`, which the calendar script fires on a picked day, also in a `MonthURL` fragment.

### A file uploaded with a failing submit is shown as kept

**What:** After a submit that fails validation, do not tell the user that a
file field "keeps" the file just uploaded with it, or keep that upload for
the next submit.

**Why:** The re-rendered form shows the upload of a `*multipart.FileHeader`
or several-files field as stored, with the note that leaving the input blank
keeps it. But the request's temporary files are removed when the response
ends and the next submit loads the record without them, so an optional
field's upload is lost without a word. The `[]byte` and `string` file fields
behaved the same before.

**Context:** `mx.StoredFile` in `fieldbehavior.go` and the stored-file note
of the html and shadcn file inputs. Telling a value loaded from the record
from one bound by the failing request needs the handler to pass that along.
Found by the /ship review (2026-11-01).

**Effort:** M
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — the handler copies the loaded values of the file fields, also in repeatable rows, before it binds a POST, and puts them back before every failed submit renders the form again (validation errors, `FieldErrors` or another error from `onSubmit`, and a row command); a file field of a row that was not loaded becomes empty. The form shows what is stored, and a required file input asks for the file again.

### Port cva's `defineConfig` / `onComplete` hook

**What:** Port the `defineConfig` factory and its `onComplete` hook.

**Why:** Upstream uses it to post-process every generated class string in one
place — the natural seam for wiring `twmerge` in globally instead of per call
site.

**Context:** The most speculative of the three: this repo already merges classes
through `Cn` at the component boundary, so the hook has no current job. Revisit
only if a caller needs to intercept class generation program-wide.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — Ported cva v1.0.0-beta.12's `defineConfig`, whose `cx` concatenator replaced the `onComplete` hook: `cva.DefineConfig(cva.Options{CX: …})` returns a `Factory` with `New` and `CX`, and `cva.New` is the zero `Factory` (space join); tests transcribed from upstream's `config.test.ts`.

### A carriage return is a line break, not a terminal overwrite

**What:** A progress-bar line that rewrites itself with `\r` renders as several
lines instead of one.

**Why:** It is a display-semantics choice, not corruption: `mx.SSEResponse`
already normalizes CR and CRLF to LF on the wire, so the bytes are intact.

**Context:** Collapsing `\r` runs — keeping only the text after the last one in
a segment — is a few lines in `Config.Line`, and would want to be a field rather
than the default, since a log that uses `\r` as a plain separator would lose
content.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — `Config.OverwriteCR` (opt-in) keeps only the text after the last carriage return of each LF-separated segment, raw lines and decoded record strings alike, before `MaxLineLen`; a trailing CR or CRLF still ends the line.

### Expandable pretty-printed detail for a record

**What:** A collapsed line that expands to an indented, highlighted rendering of
the whole JSON record.

**Why:** A record with several nested objects gets long on one line, and the
compact `{k=v k=v}` form is harder to read the deeper it nests.

**Context:** `<details>` around the line, with the summary being what
`Config.Line` renders today and the body an indenting variant of the same value
renderer. The parse already produces an ordered tree, so only the rendering side
is new. An indenting writer leaves the inside of a `<pre>` as written, so a
block with element children is safe in it.

**Effort:** M
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — `Config.Expandable` (opt-in) wraps each JSON record in a `<details>` whose summary is the compact line and whose body is a `<pre class="log-pretty">` with the whole record indented in the same `key=value` syntax and classes; searched by the filter and `Matches` too.

### `GlobPageSource` drops front matter fields it does not know

**What:** Since the front matter is no longer part of a Markdown page's
`Content`, fields other than title, description, lang, date, publishDate,
draft and tags (an `image`, `layout`, `aliases`) are lost.

**Why:** A `ContentRenderer` that parsed the front matter itself for fields of
its own has nothing to read anymore. The CHANGELOG records the loss.

**Context:** Expose the parsed front matter, e.g. a `Page.FrontMatter
map[string]any` filled by `GlobPageSource` in `web/globpagesource.go`, or an
option keeping the raw file as content.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — `Page.FrontMatter map[string]any` holds every front matter field of a Markdown page, mapped ones included, with nested objects normalized to `map[string]any` for YAML, TOML and JSON.

### The context is not checked before the glob

**What:** Pages checks the context after `filepath.Glob`, so an already
cancelled request still stats `Dir` and runs the whole glob, and with a
missing `Dir` returns that error instead of the context's.

**Why:** Wasted work for a request nobody waits for.

**Context:** A `ctx.Err()` check at the top of the iterator. Found in round 3
of the /ship review (2026-10-26).

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — `GlobPageSource.Pages` checks `ctx.Err()` before anything else, so a cancelled request neither stats `Dir` nor runs the glob and gets the context's error.

### Only `Dir` is checked to exist, not the fixed prefix of `Pattern`

**What:** A missing `Dir` fails the source, but a missing directory in the
fixed part of `Pattern`, with or without `Dir` (`blog/*.md` with no `blog`),
still matches nothing without an error, although the `Dir` doc calls the two
forms the same result written differently.

**Why:** The same silent empty source, one level down.

**Context:** Stat the fixed directory prefix of the joined pattern instead of
`Dir` alone, or document that only `Dir` is checked. Found in round 3 of the
/ship review (2026-10-26).

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — `GlobPageSource` checks that the fixed directory part of `Pattern` (before the first wildcard, below `Dir`, through the `os.Root` with `Confine`) exists and is a directory, and fails with a `GlobPageSource.Pattern` error otherwise.

### No `og:locale:alternate` for translations

**What:** `Site.AddPageMetadata` emits hreflang links for translations but no
Open Graph `og:locale:alternate` tags.

**Why:** Link previews (Facebook, LinkedIn) read the Open Graph locales, not
hreflang, to offer a translated preview.

**Context:** `html.Document.MetaProperty` is a `map[string]string`, which holds
one value per property, while `og:locale:alternate` repeats once per locale.
It needs a document field that allows repeated properties, or the tags rendered
into `HeadCustom` next to the alternate links from the same list.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — `Site.AddPageMetadata` renders an `og:locale:alternate` meta tag per translation (no x-default, not the document's own `og:locale`) into `HeadCustom` next to the alternate links, from the same `Page.Alternates`.

### `mcp.GuardHandler` sends no CORS headers

**What:** `GuardHandler` lets requests from `AllowedOrigins` through but neither
answers a CORS preflight nor sets `Access-Control-Allow-Origin`, so a page on
another origin cannot call the endpoint even when its origin is allowed.

**Why:** Which headers, methods and credentials an endpoint allows is a policy
of the site, and same-origin pages and non-browser clients — what a public
page endpoint serves — need no CORS at all. The guard only rejects; the doc
comment and the mcp README say so.

**Context:** CORS for the Streamable HTTP transport means answering `OPTIONS`
for allowed origins with `GET, POST, DELETE` and the `Content-Type`,
`Mcp-Session-Id`, `Mcp-Protocol-Version` and `Last-Event-ID` request headers,
and exposing `Mcp-Session-Id` on responses. A `GuardOptions.CORS` flag reusing
`AllowedOrigins` would cover it.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — `GuardOptions.CORS`: `GuardHandler` answers preflights of `AllowedOrigins` with 204 (GET, POST, DELETE; Content-Type, Mcp-Session-Id, Mcp-Protocol-Version, Last-Event-ID; Max-Age 7200) and sets `Access-Control-Allow-Origin`, `Expose-Headers: Mcp-Session-Id` and `Vary: Origin`, without credentials.

### A readonly repeatable field with no rows renders an empty box

**What:** Render a short "none" line inside the fieldset of a
`form:"readonly"` repeatable field that has no rows.

**Why:** A readonly repeatable field has no add-row button, so with no rows
its fieldset shows only the legend, and a user cannot tell "no entries" from a
broken render.

**Context:** `renderRepeatable` in `repeatable.go` builds the fieldset in the
core package, which has no localizable labels; the text would need a hook, such
as a shadcn `Labels` field passed down or a `ReflectFormConfig` option.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — the fieldset shows a `<p data-repeatable-empty>` with `mx.RepeatableNoRowsText` ("None", reassignable like `FormErrorMessage`). The core repeatable texts were not per-request localizable then; "Core repeatable texts are English only" (completed 2026-11-14) made them `mx.FormTexts` fields.

### shadcn readonly inputs look editable

**What:** Give readonly text, number and date inputs in shadcn forms a
readonly look, such as `read-only:bg-muted/50 read-only:cursor-default`,
or mark the Field root `data-readonly`.

**Why:** The `Input` classes have no `read-only:` variant, so a readonly
input looks like an editable one, while readonly checkboxes, selects and
file inputs in the same readonly section render disabled and dimmed.

**Context:** The attrib builders in `shadcn/formdecider.go`; the `Input`
component stays faithful to upstream.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — readonly text, number, date and textarea inputs of `shadcn.FieldDecider` get `readonlyInputClasses` (`read-only:bg-muted/50 read-only:cursor-default`, no focus ring); `Input` itself stays upstream's.

### A time-of-day type whose zero is midnight cannot hold 00:00

**What:** Let a time-of-day type such as `civil.Time`, whose `IsZero` is true
at 00:00:00, show and store midnight in a time input.

**Why:** `DateTimeInput` renders a zero value as no value, so a stored
midnight shows an empty input, and the handler's required check reads the
bound zero value as missing, so a required time field rejects a posted 00:00.
An optional field saves midnight but shows it empty again.

**Context:** `marshaledText` in `formdatetime.go` honours the type's own
`IsZero`; `isEffectivelyEmpty` in `formhandler.go` decides required. A fix would
show a time-of-day value's text whenever it is not null and decide required
from the raw submission for date and time fields. The type cannot tell an
unset value from midnight itself, so either way it needs a pointer or a
nullable wrapper to mean "no time".

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — a time widget shows a time-of-day type's text whenever it is not null (`marshaledText`'s `zeroIsValue`), and `sentBlank` decides a parsed time field's blank from the submitted text, so a required time takes 00:00.

### A `database/sql` text holding a NUL byte cannot be saved

**What:** Decide the `Scan` path of a `database/sql` field by its type, not by
whether its stored text fits an input, and show stored text an input cannot
hold (a NUL byte, invalid UTF-8) as an invalid text input the user can
replace, with an error naming the cause.

**Why:** `isScanner` returns `mx.DriverValueText`'s ok, which is false for such
text, so `setScalar` skips `Scan` and every save of a `sql.NullString` holding
`"a\x00b"` fails with "implement encoding.TextUnmarshaler or
database/sql.Scanner", though the type implements `Scanner`; a newly typed
valid value cannot be saved either.

A tab or line break in such text is not caught at all: the attribute
escaper turns them into spaces and browsers strip line breaks from text
inputs, so an unchanged save rewrites `"a\tb"` as `"a  b"` and `"k1\nk2"`
as `"k1 k2"`; excluding them in `inputText` would make such values
unsaveable through the same `isScanner` path, so both need the type-based
`Scan` decision, and multi-line text a textarea.

**Context:** `isScanner` and `setScalar` in `html/formdecider.go`;
`DriverValueText` and `inputText` in `nullable.go`. Skipped in the date
widget branch's review as rare; the tab and line break case came from its
adversarial review.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — `mx.DriverValueType` decides the `Scan` path by type; `DriverValueText` writes text an input cannot hold with U+FFFD for NUL and invalid UTF-8, and `setScanned` keeps the stored text when that shown text, or the text without its line breaks, comes back unchanged. `mx.DriverValueTextProblem` drives a note under the input in both renderers (shadcn `Labels.FormTextNUL`, `FormTextInvalidUTF8`, `FormTextLineBreak`).

### An unsent required enum whose stored zero is an option reads as blank

**What:** On a POST that leaves out a required enum's `__present` sentinel,
count a stored zero that is one of its options (a `Low = iota`, an option
`"0"`) as a value, not as blank.

**Why:** The handler then reports "required" under a select that shows the
option selected, and the save fails. A browser always sends the sentinel, so
only partial or programmatic POSTs hit it.

**Context:** `isBlank` in `formhandler.go` (the unsent path of `parseField`);
the sent path decides from the submitted text (`sentBlank`). Knowing whether
the zero is listed needs the field's options (`mx.CollectOptions`). Skipped
in the date widget branch's review.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — `zeroIsOption` collects the field's options for a zero enum that was not sent, and a listed zero (a non-empty text such as "0") is a value, not blank.

### A nil request or context panics in the error helpers

**What:** Fall back to `context.Background()` in `RespondNonContextError`
for a nil request and in `ReportHiddenError` and `FormErrorText` for a nil
context.

**Why:** A nil context panics only once `mx.ReportError` is set, which tests
that leave the hook nil never see.

**Context:** `errors.go`; Go's convention forbids a nil context, so this is
defense against misuse.

**Effort:** XS
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — `ReportHiddenError` (and through it `FormErrorText` and `MaskSensitiveError`) treats a nil context as `context.Background()`, as `RespondNonContextError` does a nil request.

### `mx.ReportError` gets a field's error without the field's path

**What:** Tell the hook which field a reported field error belongs to: an
error no input shows, a `form:"sensitive"` field's masked error.

**Why:** A log line "required" or a strconv error says nothing about the field;
for a `mx.FieldErrors` entry under a misspelled path, the path is the bug.

**Context:** `ReportHiddenError` gets the error as it is, so the hook can
`errors.Is` it. Wrapping it with the path (`fmt.Errorf("%s: %w", path, err)`)
keeps that, but `html`'s `guardBind` does not know the path and would need it
passed.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — the field's path rides in the context instead of the error: `mx.ContextWithFieldPath`/`FieldPathFromContext`, set by the form handler while parsing (a decider's `Parse` gets it in its request), validating and showing each field, also for an error under a path no input shows.

### Backtick HTML literals in tests lack the `/*html*/` prefix

**What:** Prefix the backtick HTML literals in the reflected-form and calendar
tests with `/*html*/`, on its own line where the literal follows another
argument.

**Why:** AGENTS.md asks for the prefix on every backtick HTML literal; the
date widget branch added about 100 without it (`html/formdecider_test.go`,
`shadcn/formdecider_test.go`, `shadcn/calendar_test.go`,
`formhandler_test.go`), and `main` already had a few.

**Context:** Mechanical, but it reflows many table rows. Skipped in the date
widget branch's review.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — 174 literals prefixed in `html/formdecider_test.go`, `shadcn/formdecider_test.go`, `shadcn/calendar_test.go` and `formhandler_test.go` by a `go/scanner` pass (own line after a comma).

### Text-length, autocomplete and input-mode tag options

**What:** `form:"…"` tag options for the input attributes a reflected field
cannot set today: `minlength`/`maxlength` (checked by the parser too, like
`pattern`), `autocomplete` and `inputmode`.

**Why:** A length limit is the most common constraint after `required`, and
today it needs a `Validate` method on the struct while `min`/`max` of a
number come from the tag. `autocomplete` decides whether a browser or password
manager fills a field correctly (`one-time-code`, `new-password`, `email`).

**Context:** `FormTag` in `formtag.go` holds widget, label, placeholder, help,
section, pattern, options, min/max/step and the boolean flags. The server-side
check belongs with the shared form rules (see the completed "Form rules
duplicated across `mx`, `html` and `shadcn`"), and a forged over-long value
must be bounded the way `MaxFieldErrorLen` bounds its error. Carried over from
the README's old "More ReflectInputOptions" line.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — `form:"minlength=…,maxlength=…,autocomplete=…,inputmode=…"`: `FormTag.MinLength`/`MaxLength`/`Autocomplete`/`InputMode`, rendered by `html.TextTagAttribs` in both renderers, and `mx.CheckTextLength` holds text, textarea and catch-all fields to the lengths on the server, in UTF-16 code units.

### Rich file upload widget

**What:** A file field with a preview of the chosen image, several files
(`multiple` on a slice field) and upload progress.

**Why:** A reflected `file` widget is a bare `<input type="file">`: one file,
no preview, and a large upload shows nothing until the response comes back.

**Context:** The bind side is in `fieldbehavior.go` and `formhandler.go`
(multipart parsing, temp files, the "file field holding a file" state). A
preview needs a small inline script (`URL.createObjectURL`), progress needs
`XMLHttpRequest` or htmx's `htmx:xhr:progress` event, and multi-file needs a
slice-of-file field kind on the bind side. shadcn has no file input component
to transcribe; the look would follow `Input` and `Empty`.

**Effort:** M
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — a `[]string`, `[][]byte` or `[]*multipart.FileHeader` field (`mx.MultipleFiles`) gets a multiple input and binds every upload, a `*multipart.FileHeader` field needs no tag, and `form:"accept=…"` sets the picker's types (`html.FileTagAttribs`). shadcn's editable file input previews the chosen files (image thumbnails, other names) and shows a progress bar for an htmx upload (`Labels.FormUploadProgress`), from one small script.

### Out-of-band fragments in reflected form responses

**What:** Let a `ReflectFormHandler` answer an htmx submit with the form plus
out-of-band fragments (`hx-swap-oob`), such as a toast, an updated list row or
a counter elsewhere on the page.

**Why:** After a successful save the page around the form is often stale, and
the caller has to replace the whole page or send `HX-Trigger` events and
reload the parts in separate requests.

**Context:** `writeFormResponse` in `formhandler.go` renders one component.
`hx.SwapOOB`/`hx.SwapOOBTrue` exist, and `mx.KeyedID` gives the targets
stable ids. The shape could be an `onSubmit` result carrying extra components,
rendered after the form only when `hx.IsRequest`.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — `mx.AddOutOfBand(ctx, comps...)` from `onSubmit` appends fragments to the response of an htmx submit (`HX-Request: true`); a successful one with fragments and no destination of its own is answered with the form re-rendered from a fresh load instead of the redirect. A plain submit ignores them.

### Standard option lists as named options providers

**What:** A subpackage that registers common option lists with
`mx.RegisterNamedOptions`: ISO 4217 currencies, ISO 639 languages, ISO 3166
countries, so a field takes `form:"options=iso4217"`.

**Why:** Every application that has a currency or country field writes the
same list, with the same label and sort mistakes.

**Context:** `RegisterNamedOptions` and `form:"options=…"` in `options.go`
are the hook. Labels need localization, so the providers would read the
language from the context (`NamedOptionsFunc` gets it). The data must come
from a maintained source (CLDR) and be regenerated, not typed in; keep it out
of the root package so binaries that do not use it do not carry it.

**Effort:** M
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — the `stdoptions` package registers ISO 4217 currencies, ISO 639-1 languages and ISO 3166-1 alpha-2 countries as `iso4217`, `iso639-1` and `iso3166-1` with an explicit `stdoptions.Register()`; English and German names picked per request with `stdoptions.ContextWithLanguage`, presorted by CLDR collation. Generated from cldr-json 49.0.0 by `tools/gen-stdoptions`; withdrawn, non-tender, user-assigned, exceptionally reserved and aliased codes are left out the way CLDR marks them.

### Attribute values lose their line breaks and tabs

**What:** Escape a line break in an attribute value as `&#10;` (and a tab as
`&#9;`) instead of replacing it with a space (two spaces).

**Why:** The replacement is lossy: a value read back with `getAttribute` or
`dataset` is not the value written. `shadcn.CodeBlockCopyButton` hit it — a
multi-line copy value reached the clipboard on one line — and now
percent-encodes its value and decodes it in its script to get around it.
Any other attribute carrying text for a script (`data-*`, `title`, a
`value` of a hidden input) loses line breaks the same way.

**Context:** `doubleQuoteAttribEscaper` and `singleQuoteAttribEscaper` in
`attrib.go`. The replacement may be there to keep a long `class` or `style`
on one line in indented output; check who relies on it before changing,
and the golden outputs of the gallery and the tests.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — the attribute escapers write `&#10;`, `&#13;` and `&#9;`, so the value read back is the value written; `CodeBlockCopyButton` dropped its percent-encoding. Nothing relied on the spaces: the gallery's attributes held no line breaks.

### Indentation changes the content of `<pre>` and `<textarea>`

**What:** Make an indenting `CheckedWriter` stop breaking lines and
indenting inside elements whose whitespace is content, at least `pre`,
`textarea` and `script`.

**Why:** `WithIndent` writes a newline and indentation before every child
element, which inside a `<pre>` becomes part of the shown code. The
gallery works around it with `rawPre` (rendering code panes with a second,
non-indenting writer into raw HTML), and `shadcn.CodeBlockContent` with an
`unindented` wrapper that does the same for its `<code>`; every other
`<pre>` with element children is still broken in indented output.

**Context:** `BeginElement`, `Comment`, `CDATA` and `EndElement` in
`checkedwriter.go` call `Newline` when `indent` is set; they would check
the element stack for such an element first. Once fixed, both workarounds
can go.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — the `CheckedWriter` marks `pre`, `textarea`, `script`, `style`, `listing`, `xmp` and `plaintext` and everything inside them as preformatted and breaks, indents and adds no line after a processing instruction there. The gallery's `rawPre` and `CodeBlockContent`'s `unindented` are gone.

### A file field holding a file shows "No file chosen"

**What:** Show that a file field holds a file, such as its filename, next to
the file picker, readonly or editable, and on an editable one that leaving
it empty keeps the stored file; link the note with `aria-describedby`.

**Why:** The browser labels a file input "No file chosen" even when a file is
stored. A readonly one tells the user there is no file; an editable one,
which keeps the stored file on an empty submit and so is no longer
HTML-`required` when it holds one, gives no sign of it, so the user uploads
again or thinks the file is gone.

**Context:** `fileInput` in `html/formdecider.go` and `fileAttribs` in
`shadcn/formdecider.go`; `mx.BrowserTag`'s `hasStoredFile` tells whether one is stored. A
string field holds a filename to show; a `[]byte` field has none, so it
would need a caption, which needs localizable text (shadcn `Labels`). The
editable case was found in the date widget branch's review.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `mx.StoredFile` returns the stored file's name, and both renderers show a note next to the picker linked with `aria-describedby`: the current file, or that one is stored, and on an editable field that a blank keeps it (shadcn `Labels.FormStoredFile`, `FormStoredFileUnnamed`, `FormKeepStoredFile`).

### Date format hints and date errors are English only

**What:** Let a localized form translate the format it asks for and the
messages of a rejected date or time.

**Why:** The format in the placeholder and hint of a date field's text
fallback (`YYYY-MM-DD`, `YYYY-MM-DDTHH:MM`, `HH:MM`) and the errors of a
rejected or out-of-range date ("is not a date in the format …", "earlier
than …") are fixed English. A German shadcn form shows "Erwartetes Format:
YYYY-MM-DD", a translated `form:"placeholder=JJJJ-MM-TT"` disagrees with the
hint, and after a submit the only statement that the value is invalid is the
English field error.

**Context:** `timeTexts` and `DateTimeFormatError` in `formdatetime.go`,
`CheckDateTimeRange`, and the hint in `shadcn/formdecider.go` (`Labels`
captions already exist). A typed error the shadcn renderer can map to
`Labels`, and one displayed format string used by both the placeholder and the
hint, would cover it. The parser's layouts stay internal.

**Effort:** M
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `mx.DateTimeFormatError` and `mx.CheckDateTimeRange` return `*mx.DateTimeError` (a `DateTimeProblem`, the bounds and the format), a user message that `MaskSensitiveError` keeps as it is; shadcn translates it and the format from new `Labels` fields and shows the same translated format as the placeholder and in the hint. The html renderer stays English.

### `aria-describedby` on a field leaves out its errors and help

**What:** Give each field's help text and error messages ids and reference
them all from the control's `aria-describedby` (or `aria-errormessage`), in
the html and shadcn renderers.

**Why:** A date field's text fallback references only its format hint, a
select or checkbox set only its note on options the list no longer offers
(`optionNoteID`), and other fields reference nothing; the error and help ids
must join those in the same attribute. After a failed submit, a screen reader user
who tabs to the field hears "invalid" but not the error, and `role="alert"`
on server-rendered markup is not announced on a full-page load.

**Context:** `renderField` in `html/formdecider.go` and `shadcnFieldMarked` /
`appendFieldErrors` in `shadcn/formdecider.go`. It changes the markup of every
field, so the gallery output changes too.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — both renderers give the help and each error message ids and add them to the control's `aria-describedby` with `mx.AddDescribedBy`, after the ids it had (format hint, option note).

### `form:"pattern=…"` is checked only by the browser

**What:** Check a string field's `form:"pattern=…"` on the server too,
anchored like the HTML attribute (`^(?:pattern)$`), compiled once per tag,
with a sensitive field's error masked; or document it as browser-only.

**Why:** `FormTag.Pattern` is documented as a constraint on string inputs,
but it only becomes the `pattern` attribute, so a scripted POST stores any
string. `min`/`max` of numbers and dates are now checked on the server, so
`pattern` is the last declarative constraint only the browser enforces.

**Context:** The string bind in `html/formdecider.go` (`parseField`'s
`FieldKindString` case). Predates the date widget work; found in its review.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `checkPattern` in `html/formdecider.go` checks string and catch-all fields once bound, anchored as the attribute, compiled once per pattern; an empty text and a pattern RE2 cannot compile are not checked.

### A missing `__present` sentinel skips `required` and `min` of a number

**What:** Reject a required number that is no pointer, and check a number's
range, when a rendered field's `__present` sentinel is missing and the value
is the zero a submit-only form's `new(T)` starts with.

**Why:** A browser always sends the sentinel. A scripted POST that leaves it
out keeps the stored value unchecked, and for `new(T)` that is a 0 that
`form:"required,min=1"` should reject. The missing-sentinel required check
treats a number as never blank, and the range check lives in the parser,
which does not run.

**Context:** The missing-sentinel branch of `parseField` (`formhandler.go`)
and `checkNumberRange` (`html/formdecider.go`). Exposing the range check
as the decider's `Validate` would let that branch run it. A partial update
that posts a subset of fields must keep working.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — the range check moved to `mx.CheckNumberRange`; `parseField`'s missing-sentinel branch rejects the zero of a non-pointer number when it is required or outside its range, while a stored number is kept.

### Form-level messages are unstyled in shadcn forms

**What:** Let a renderer style the form-level `<p role="alert">` messages,
for example with a `data-slot="form-message"` hook that the shadcn renderer
styles like a destructive `Alert`.

**Why:** The messages above a reflected form (an `onSubmit` error, an error
no input shows) are bare paragraphs, so in a shadcn form they look like
body text next to the red field errors, though they are often the only sign
that a save failed.

**Context:** `formMessageComponent` in `formhandler.go` builds them in the
core package, which knows no Tailwind classes.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `ReflectFormConfig.FormMessage` renders them, `shadcn.FormMessage` as a destructive `Alert`, and the default paragraph carries `data-slot="form-message"`; `cmd/example-form` uses the shadcn one.

### A sensitive input says nothing about its stored value

**What:** Tell the user what a blank sensitive input does: a caption such
as "Leave blank to keep the current value" while a value is stored, and a
note on an add-row or remove-row re-render that a typed value must be typed
again.

**Why:** A sensitive input never shows its value, so a user cannot tell a
stored secret from none, does not know a blank keeps it, and loses a value
typed before a row command without notice.

**Context:** `renderField` (html) and `shadcnFieldMarked` (shadcn); the
captions need shadcn `Labels` fields.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — while a value is stored both renderers show "Leave blank to keep the current value." linked to the control (shadcn `Labels.FormKeepStoredValue`), and a row command after text was typed into a sensitive input re-renders with `mx.FormRetypeSensitiveMessage`.

### Fractional seconds in date and time inputs

**What:** Accept and keep fractional seconds in datetime and time inputs, so
a `form:"step=…"` under one second works.

**Why:** A sub-second step makes the browser submit values such as
`2026-07-06T09:05:30.123`. A field holding text rejects them, since its parser
reads only whole seconds, and stored text with a fraction renders as an
invalid text input. A `time.Time` accepts them, because `time.Parse` tolerates a
fraction, but renders without it, so an unchanged save truncates. For now the
README and `FormTag.Step` document that the finest step is one second.

**Context:** `timeText.parse` and `shortest` in `formdatetime.go` read and
write the input forms; a `.999999999` layout variant would keep the fraction,
and `dateTimeInput` would format a `time.Time` with it. The `step="1"` rule in
`DateTimeInput` would then count sub-second remainders from `min` too.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — the datetime and time layouts write and read a fraction of up to three digits, the most an input's value holds (`2006-01-02T15:04:05.999`), trailing zeros allowed; a value with a sub-second offset from `min` gets `step="any"`.

### `database/sql` types render as plain text inputs

**What:** Give a `sql.Null*` field the widget of the value it holds: a
nullable checkbox or a true/false/(none) select for `sql.NullBool`, the
datetime widget for `sql.NullTime`, and a number input with `min`, `max` and
`step` for the integer and float types.

**Why:** A catch-all `sql.Null*` field is a text input holding the driver
value, so a user types "true", a raw RFC 3339 time with no picker, or a
number without a numeric keypad, and the browser checks none of it; the
server does (`setScanned`).

**Context:** `mx.DriverValueText` and `setScanned`/`checkScannedRange` in
`html/formdecider.go` bind them today. `DetectField` would map the type by
its driver value kind (from a zero value's `Scan` targets or a field-name
convention such as `Bool`/`Time`/`Int64`) to the matching `FieldKind`, and
the binders of those kinds would need a null for a blank.

**Effort:** M
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `DetectField` rule 14.5 (`driverValueKind`) gives a database/sql type the kind of what its `Scan` reads: bool, number or date-time; the html binders scan those, a blank as the null, and a `sql.NullBool` gets a clear control for its null. A text-holding type stays a text input.

### A text-held date type that reads only its own spelling cannot be saved

**What:** Let a string-kind date type whose `UnmarshalText` reads only its own
spelling (a `type DEDate string` that takes `02.01.2006`) render as a text
input of its own text and bind through `UnmarshalText` of the raw text, as a
non-string own-spelling type does.

**Why:** Such a field shows its stored text as "Not in the expected format:
YYYY-MM-DD"; posting it back fails in `mx.ParseDateTimeInput`, and posting
ISO text passes that but is rejected by the type, so nothing can be saved.

**Context:** `readsOwnSpellingOnly` in `formdatetime.go` exempts every string
kind; `setTimeText` in `html/formdecider.go` calls `mx.ParseDateTimeInput`
before `UnmarshalText`. Found in the date widget branch's review; deferred.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `readsOwnSpellingOnly` covers a string kind with an `UnmarshalText`, which renders as a text input of its own text, and `setTimeText` binds the posted text through its `UnmarshalText`.

### Edge cases of the user message at a field

**What:** Close the gaps the reviews of the `UserMessageError` change left
in `mx.FieldErrorTexts` and the sensitive masks: an error that supplies its
`UserMessageError` only through an `As` method shows its `Error()` at a
field (it does above the form); an empty `UserMessage` stops `errors.As` and
hides a real one deeper in the chain, as in
`errors.Join(FormMessage(""), FormMessage("x"))`; a `form:"sensitive"` field
shows only the first of several joined messages (`validateField`,
`mx.MaskSensitiveError`); and an error that gives no text, such as an empty
`FieldErrors` from a decider's Validate, marks a field invalid with no
message. The shadcn and repeatable renderers of several messages per error
have no test of their own.

**Why:** Each shows the user less than the error carries, or a wrapper's
internal text; none is worse than before the change, where fields showed
`Error()`.

**Context:** `FieldErrorTexts` and `UserMessage` in `errors.go`; a tree walk
that skips empty messages and asks `errors.As` at each node would fix the
first two. Found in the review passes of this change; deferred at the
review cycle limit.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `UserMessage` walks the tree, skipping empty messages and asking `As` methods; `FieldErrorTexts` finds a message an `As` method supplies, skips joined parts without text and reads "not a valid value" for an error without text; `MaskSensitiveError` shows every message; an empty `FieldErrors` from `Validate` is no error. The shadcn and repeatable renderers of several messages have tests.

### A failing upload's temp file answers the client's 400

**What:** Tell a server-side failure of the multipart parser (a full or
missing temp directory) apart from a malformed body, answer it with a 500
and report it.

**Why:** All uploads then fail with "could not parse form" and nothing
reaches `mx.ReportError`, which excludes the 400s as the client's.

**Context:** `parseFormBody` and the status switch in
`ReflectFormHandlerRedirecting`; predates this change.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — a `*fs.PathError` from the multipart parser, a temp file it could not create or write, is answered by `RespondNonContextError`: a 500 that is reported.

### An internal error joined with a user message is dropped

**What:** Report the rest of an error that also carries a user message or a
`mx.FieldErrors`, such as `errors.Join(FieldErrors{"Name": taken}, auditErr)`
returned from `onSubmit`.

**Why:** The form shows the user message, and the internal error reaches
neither the page nor `mx.ReportError`, so a failed audit write is lost.

**Context:** The `ReportError` contract does not report an error the user
sees, which an `errors.As` match makes the whole chain; reporting every
wrapped user message instead would log each validation failure. Needs a rule
for which parts count as hidden.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `reportHiddenParts` reports each joined part that holds no user message and no `FieldErrors`, in `FormErrorText`, `MaskSensitiveError` and for `onSubmit`'s `FieldErrors`; the text wrapping a message is not reported.

### A visible number field rejects a float that needs more than 64 characters

**What:** Parse a number before applying `maxNumberLen`, as hidden fields
now do through `guardBind`'s quoting bound.

**Why:** The renderer writes floats with `strconv.FormatFloat(..., 'f', -1, ...)`,
so a value of 1e64 or more round-trips as more than 64 characters and can
never be saved again.

**Context:** The `FieldKindNumber` branch of `html.parseField`; predates this
change.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `maxNumberLen` is 327, the longest float64 `strconv.FormatFloat(f, 'f', -1, 64)` writes.

### A panic recovered while binding a field is not reported

**What:** Pass the panic `guardBind`, `unmarshalTime` and the other
field-binding guards in `html` recover to `mx.ReportHiddenError`, with its
stack.

**Why:** The field shows "cannot be stored as …" and the panic, a bug in the
application's type such as a promoted method of a nil embedded pointer, is
dropped; the handler-level panics of `ComponentHTTPHandler` and
`ComponentFuncHandler` are reported.

**Context:** The guards turn the panic into a validation error the field
shows, so the user sees an error; only the panic itself is lost.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `guardBind`, `unmarshalTime` and `setTimeText` report the panic with its stack (`errs.AsErrorWithDebugStack`).

### A long error of long posted text is not reported

**What:** Report a bounded stand-in, such as the error's type and the length
of its text and of the posted text, where `mx.BoundFieldError` hides a long
error of a long posted text without reporting it.

**Why:** Whether a hidden error is reported depends on the length of the
posted text, for a checkbox set of all its values joined, so a legitimate
long value's long error, a real bug, never reaches `mx.ReportError`.

**Context:** `boundFieldError` in `errors.go` reports only when the posted
text is within the bound, as a forged value of any size must not reach the
hook. Split out of the binder-quoting item (2026-11-01).

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `reportHiddenFieldError` reports a stand-in naming the error's type and the lengths of its text and of the posted text.

### Port cva's `compose`

**What:** Merge several variant resolvers into one, as npm cva's `compose` does.

**Why:** Lets a component build on another component's variants instead of
restating them, which is how upstream shares a base look across a family.

**Context:** No component in this repo needs it yet — every port declares its own
`Config`. Worth doing when the first component would otherwise copy another's
variant table.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — first as `cva.Compose`, then, the same day, as `cva.Config.Composes` in `shadcn/cva/cva.go`, because cva v1.0.0-beta.11 removed `compose()` for the `composes` option; ported from v1.0.0-beta.12 with its tests and a README section.

### Decide on a Go equivalent for cva's `VariantProps`

**What:** Provide a Go equivalent of the `VariantProps` type helper, or document
that a typed props struct replaces it.

**Why:** `VariantProps` is how TypeScript callers get compile-time checking of
variant names. Go's answer is a typed struct per component, and the port should
say so explicitly rather than leave readers looking for the missing helper.

**Context:** This is mostly a documentation decision — the existing components
already take typed variant parameters (`ButtonVariant`, `ButtonSize`), which is
the Go equivalent in practice. Write it down in `shadcn/cva/README.md` and close
the item unless a real generics-based helper turns out to be worth the surface.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — documented in `shadcn/cva/README.md` that typed variant parameters (go-enum types like `ButtonVariant`) are the Go equivalent; no generic helper, because generics cannot derive types from a runtime `Config` map.

### ANSI escape sequences render as literal text

**What:** A log line carrying SGR sequences (`\x1b[31m`) shows them as `[31m`
garbage instead of as color.

**Why:** Real streams carry them constantly — CI output, `docker logs`, anything
using a colorizing logger. They are noise in every line they appear in, and the
one thing they encode, severity, is exactly what the view already colors.

**Context:** Two shapes: strip SGR sequences in `Config.Line` before parsing, or
map them to spans with their own classes. Stripping is a few lines and fixes the
noise; mapping is more work and would need a decision about how ANSI's 16 colors
relate to a `Theme`. Start with stripping, behind a `Config` field, because a
caller who pipes raw terminal output usually wants it gone rather than rendered.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `Config.Line` strips SGR/CSI/OSC/ESC sequences from the raw line (before parsing and the length cap) and from decoded JSON strings; on by default because it drops no text, `Config.KeepANSI` opts out.

### The filter searches the DOM, not the backlog

**What:** The filter matches the lines currently in the view. Lines the
`MaxLines` cap already dropped, and everything behind the stream, are not
searched.

**Why:** A reader who types `panic`, sees nothing and concludes there was no
panic has been misled. The control is labelled "Filter" rather than "Search" for
that reason, but the honest fix is a real search.

**Context:** It belongs on the server, as a parameter of the stream URL, which
is a smaller change than any client-side alternative — the handler already
answers "everything after this id" and would answer "everything matching this"
the same way. Changing the filter would then have to reconnect the stream, since
`hx.SSEConnect` takes a static URL, so the view needs a way to swap its
connecting element.

**Effort:** M
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `Config.SearchParam` reconnects the stream with the filter as a URL parameter by swapping the view's inner connecting element through `htmx.swap`; `Config.Matches` gives handlers the browser filter's test; `cmd/example-logstream` answers it.

### No docs-site page for `hx` or the streaming surface

**What:** The site at <https://ungerik.github.io/go-mx/> has tutorial, how-to
and reference pages for `html` and `shadcn` only. `hx` has none, and neither
does the streaming surface (`mx.SSEResponse`, `mx.KeyedID`, the `hx.SSE*`
attributes, `shadcn.StickToBottom`, `logview`).

**Why:** Streaming is documented — `README.md`, `hx/README.md`,
`logview/README.md` and two worked commands all cover it — but none of that is
on the site. A reader who starts at the docs site finds nothing and concludes
go-mx cannot stream, while the feature is one repo click away. Reference and
how-to coverage exist; the tutorial quadrant is empty for the whole surface.

**Context:** The gap predates the streaming work: `hx` never had a site page
either, so this is the site's structure trailing the repo rather than anything
this cycle removed. `docs/index.md` names `logview` and `mx.SSEResponse` in the
package table, which is the only site-level mention. The existing
`docs/html/{index,tutorial,how-to}.md` trio is the shape to copy; the gallery
under `docs/shadcn/gallery/` is generated and unaffected. `cmd/example-sse` and
`cmd/example-logstream` are ready-made tutorial material.

**Effort:** M
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `docs/hx/{index,tutorial,how-to}.md` cover `hx` and the streaming surface, linked from `docs/index.md`.

### `Site.RegisterRoutes` serves neither llms.txt nor the Markdown alternates

**What:** `Site.RegisterRoutes` registers robots.txt and the sitemap, but not
`LLMsTxtPath` or the `.md` URLs that `MarkdownPath` derives.

**Why:** Both are opt-in publishing decisions — which pages an agent may read,
whether the alternate is served at its own URL, under the page URL by
`AcceptsMarkdown`, or both, and with which revision as ETag — so a site wires
them explicitly instead of getting them by default.

**Context:** A `Site.LLMsTxtOptions` field plus a catch-all `GET /{path...}.md`
route resolving the page by `MarkdownPath` would make it one line for the common
case. ServeMux cannot match a suffix, so the route needs a pattern per page or
a handler that strips `.md` itself.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `Site.LLMsTxtOptions` and `Site.MarkdownAlternates` make `RegisterRoutes` serve llms.txt (`Site.HandleLLMsTxt`) and the `.md` alternates, each with an exact route registered from the sources; each request looks the page up again and checks that it is public. A page path added after registration has no route until the routes are registered again.

### `list_pages` returns every public page at once

**What:** The MCP `list_pages` tool has no limit or cursor; it returns all
public pages in one response, which the SDK sends twice (structured content
plus its JSON text fallback).

**Why:** A site with thousands of pages answers with megabytes that overflow
an agent's context, while `search_pages` and `get_page` are bounded. The spec
asked for "all matching pages", so the first version keeps that.

**Context:** A `limit` with a `truncated` flag like `search_pages`, or a
`cursor`, in `mcp/pagetools.go` `listPages`. Decided in the /ship review of the
knowledge layer (2026-10-15) to keep the spec'd behavior for now.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `list_pages` takes `limit` (default 50, max 200) and `cursor`, and returns `next_cursor`; the source is still read to the end for its errors.

### Page tools redo per-page work on every call

**What:** `Page.MarkdownSource` copies string content into a new `[]byte` on
every call, and `get_page` builds the full `PageInfo` (link validation, URL
building) for every public page while looking for one.

**Why:** With a cached source of 1000 pages of 16 KiB string content,
`list_pages` and `get_page` allocate about 22 MB per call, and `get_page` costs
about 4.6 ms where a path comparison per page would do.

**Context:** Give ok-only callers (`Site.LLMsTxt`, the filter in
`mcp.publicPages`) a check without the copy and keep string content as a
string in `publicPage`; in `getPage`, normalize the input to a page path once
(strip the BaseURL, map a `.md` path back) and build `PageInfo` only for the
match, while still draining the source for errors. Measured by the performance
specialist of the /ship review (2026-10-15).

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `Page.HasMarkdownSource` serves the ok-only callers without a copy; `get_page` compares candidate paths and builds `PageInfo` only for a match, still draining the source.

### A matched path shaped like a URI is read by go-fs as a URI

**What:** Pages stats a matched path with `os.Stat` but reads it through
`fs.File(local).ReadAll`, and go-fs parses a path like `file://localhost/x` as
a URI. A pattern without wildcards returns its spelling unchanged, so a local
decoy can be stat'ed while go-fs reads another file.

**Why:** The stat and the read must name the same file. Needs a pattern that
spells a URI, so it is contrived.

**Context:** Read the matched path with `os` instead of go-fs; the context is
already checked before each file, which is all go-fs's `ReadAll` does with it.
That also removes the `./` workaround for a leading `~`. Found in round 3 of
the /ship review (2026-10-26).

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — fixed by `localFiles.open` reading with `os`; `TestGlobPageSourceURIShapedPath` pins it, and no `~` workaround remains.

### The glob and the reads cannot be cancelled

**What:** `filepath.Glob`, `os.Stat` and go-fs's local `ReadAll` take no
context; the context is checked only between them.

**Why:** A huge directory or a stalled NFS or FUSE mount holds the request
past its deadline, and repeated requests pile up goroutines.

**Context:** A directory walk that checks the context per entry instead of
`filepath.Glob`; reads would need a context-aware reader. Deferred in the
/ship review of the go-fs bump (2026-10-26).

**Effort:** M
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `globFS` checks the context per directory and per batch of 1024 entries (`readDirContext`), reads go through `contextReader` in 64 KiB chunks; a single stalled system call still blocks, as documented on `Pages`.

### Glob characters in `Dir` act as wildcards

**What:** `globPattern` joins `Dir` with `Pattern` unescaped, so a directory
named `content[prod]` or `site[1]` is matched as a pattern and reads another
directory or nothing.

**Why:** A legal directory name changes which files are published.

**Context:** Glob below the directory literally, e.g.
`fs.Glob(os.DirFS(dir), filepath.ToSlash(Pattern))`, as `Confine` already
does through its `os.Root`. Deferred in the /ship review of the go-fs
bump (2026-10-26).

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — the glob runs below `Dir` literally: only `Pattern` is split at its first wildcard, with and without `Confine`.

### A page file can change between its stat and its read

**What:** Pages stats a matched file to skip pipes and devices, then go-fs opens
it again by name for the read, so a file swapped for a named pipe in between
blocks the read without a timeout.

**Why:** Needs write access below `Dir`, which is trusted, so it is low risk.

**Context:** Open once, check the open file's `Stat`, and read from it; the
open itself must not block on a pipe (`O_NONBLOCK` on Unix). Deferred in the
/ship review of the go-fs bump (2026-10-26).

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `openPageFile` opens once (O_NONBLOCK on Unix through build-tagged files), stats the open file and skips directories, pipes, sockets and devices; the Markdown head re-read seeks back on the same file.

### The `Dir` error redacts only a password

**What:** The error for a non-local `Dir` passes the URI through
`fsimpl.RedactURI`, which masks a password but not a token in the user info
(`https://ghp_TOKEN@github.com/x`) or in the query (`?access_token=…`), and on
Windows `file://user:pass@server/share` is a local UNC path printed by the
other `Dir` errors.

**Why:** The errors go to logs, like the mcp module's `OnError`.

**Context:** The message needs no URI: print the scheme only, or drop the user
info and query after `url.Parse`. Found in round 3 of the /ship review
(2026-10-26).

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — `redactURI` drops user info, query and fragment (only the scheme for a URI that does not parse), and a local `Dir` URI with user info is refused before any error prints it.

### Remove or finish `reflectroutes.go`, `writereflect.go` and `TypedRoute.Path`

**What:** Three pieces of the root package are dead: `reflectroutes.go` (and
its `reflectroutes_test.go`) and `writereflect.go` are entirely commented out,
and `TypedRoute.Path` at `route.go:141` is a stub returning the raw pattern,
with its implementation commented out under a `TODO`.

**Why:** Dead code reads as API to a newcomer and to an agent searching the
repo, and `Path` is worse than dead: it is part of the `Route` interface, so a
caller building a URL with it silently gets `"/blog/{slug}"` back, placeholders
and all.

**Context:** Each has a replacement already. `ReflectRoutes` (route struct
fields registered by reflection) gave way to explicit `mx.Route`/`TypedRoute`
registration on `http.ServeMux` patterns and `web.Site.RegisterRoutes`.
`writereflect.go` (struct fields written as attributes) gave way to the typed
attribute constructors in `html`/`svg` and the `Attrib` interface; decide the
README "To Do" item `ReflectMarkup()` together with it, since that is the same
idea. For `Path`, either implement it with the existing `JoinAbsPath` and
`FormatPathValue` in `paths.go` (the commented-out body is most of it), or drop
`Path` from the `Route` interface, since sites build URLs with
`web.Site.URL`/`PageURL`.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-11-01 — removed the commented-out files and nested routes; `TypedRoute.Path` fills `{name}`, `{name...}` and `{$}` with `FormatPathValue` and returns `(string, error)`; the README `ReflectMarkup()` item is dropped.

### An unreadable `Dir` is an empty source

**What:** `globPattern` checks `Dir` with `os.Stat`, which needs only search
permission on the parent, and `filepath.Glob` ignores a directory it cannot
open, so a `Dir` with mode `000` yields no page and no error. An unreadable
subdirectory matched by a wildcard (`*/*.md`) drops out the same way.

**Why:** An empty sitemap, llms.txt and MCP page list with a nil error, the
failure the `Dir` check was added to prevent; likely in non-root containers
where another user owns the content.

**Context:** Open the directory instead (`os.Open`, `Stat`, `Close`). The
subdirectory case needs a walk that reports read errors instead of
`filepath.Glob`. Found in round 3 of the /ship review (2026-10-26).

**Effort:** S
**Priority:** P2
**Depends on:** None
**Completed:** 2026-11-01 — `globPattern` opens `Dir`, and `globFS` replaces `filepath.Glob`: a
directory it has to search but cannot read fails the source, also below an
`os.Root`, and a hidden directory is not searched unless the pattern names
its dot, so an unreadable `.cache` is no error.

### An escaped dot in a `Pattern` without `Dir` gives pages without a path

**What:** `matchedHidden` accepts `\.` as naming the dot on Unix, but `baseDir`
cuts the fixed prefix at the first `*?[` without unescaping it, so for a
`Pattern` like `/srv/site/\.well-known/*.txt` and no `Dir` the page path is
empty and `Site.Sitemap` fails.

**Why:** With `Dir` set the same pattern works, so the two documented forms
disagree.

**Context:** On non-Windows, cut the prefix at the first unescaped wildcard and
unescape it in `baseDir`. Escapes in a `Dir`-less fixed prefix were already
broken before the branch. Found in round 3 of the /ship review (2026-10-26).

**Effort:** S
**Priority:** P2
**Depends on:** None
**Completed:** 2026-11-01 — `baseDir` takes the names before the first one holding a wildcard or an
escape (`splitGlobBase`, shared with the glob), so `\.well-known` stays below
the base.

### Markdown files are read whole, without a size limit

**What:** Every `.md` file is read completely, also for a sitemap, which needs
only the front matter, and no file has a size limit.

**Why:** A huge, sparse or growing file costs its size in memory on every
sitemap, llms.txt or MCP request.

**Context:** A `MaxFileSize` field with a default that fails larger files, and
reading only the front matter when no content is wanted. Deferred in the
/ship review of the go-fs bump (2026-10-26).

**Effort:** S
**Priority:** P2
**Depends on:** None
**Completed:** 2026-11-01 — `GlobPageSource.MaxFileSize`, default `DefaultMaxPageFileSize` (10 MiB),
fails a larger page file, checked on its stat and while reading; without
content a Markdown file's first 64 KiB are read for its front matter, and the
whole file only when the front matter does not end within them.

### No limit on the number of submitted repeatable rows

**What:** Add a row limit (a `ReflectFormConfig.MaxRows` with a default, or a
`form:"maxrows=N"` tag) that rejects a POST with more rows before binding and
hides the add-row button at the limit.

**Why:** Rows are bounded only by Go's form limits (10,000 urlencoded values,
1,000 multipart parts). One 218 KB urlencoded POST with 9,990 `__row__`
markers renders a 6 MB page and allocates about 127 MB, and wider row structs
scale linearly. A row with a required cell answers such a POST with every row
in error: 9,000 markers against `cmd/example-form` gave a 49 MB page in 1.5 s.
Each forged row whose hidden cell fails to parse, or whose `form:"sensitive"`
cell fails its validation, is also one `mx.ReportError` call: 3,300 rows gave
3,300 synchronous calls, so a hook sending to an error tracker can spend its
quota on one request.

**Context:** `repeatableRowIndices` and `parseRepeatable` in `repeatable.go`;
`applyRepeatableCommand` for the add-row side. Predates the date widget work;
found in its review.

**Effort:** S
**Priority:** P2
**Depends on:** None
**Completed:** 2026-11-01 — `form:"maxrows=N"` (`FormTag.MaxRows`) and `mx.DefaultMaxRows` (1000):
the add-row button is gone at the cap, a forged add-row click past it adds
nothing, and `parseRepeatable` rejects more rows before binding any.

### Validation errors are not bounded by `mx.MaxFieldErrorLen`

**What:** Apply the bound to every error a field shows or reports, not only
to `html.FieldDecider`'s parse errors: the validation chain's (a go-enum
`Validate` quotes the value with `%#v`), a decider's `Validate`, the parts
of a joined error next to a user message, which exempts all of them, and a
user message itself, which a domain type may build by quoting the value
without knowing of the bound.

**Why:** A forged string enum of 200 KB in a visible select shows whole at
the field, and in a hidden one reaches `mx.ReportError` whole; a checkbox
set gets one error per forged element, so 9,990 values give a 930 KB page.

**Context:** `guardBind` in `html/formdecider.go` holds the only bound;
moving it into `mx`'s `parseField`, per error part, would cover every
decider. Found in the review of the bound.

**Effort:** S
**Priority:** P2
**Depends on:** None
**Completed:** 2026-11-01 — the bound moved into `mx` as `BoundFieldError`, applied part by part,
user messages included, by `parseField` to every decider's parse errors and
by `validateField` to validation errors; `html`'s `guardBind` calls it too.

### Three binders quote a forged value as they build its error

**What:** Build the quoting error lazily, or quote at most
`mx.MaxFieldErrorLen` bytes, in `scanError` (its `fmt.Sprintf("%q is", raw)`
runs before the length check), in `setScanned` (a `database/sql` value is
scanned as a string and again as `[]byte`, and each `Scan` error quotes it),
in `rejectUnlistedOptions`, and in `guardBind`'s recovered panic.

**Why:** `boundFieldError` skips building the text of a long value's error,
but these errors hold the quote already: a forged 12 MB value allocated about
1.1 GB in a `sql.NullInt64` field, and about 380 MB in a context-option select
or a type whose UnmarshalText panics.

**Context:** All predate the `UserMessageError` change; found in the review of
its bound. Also: whether a long error is reported depends on the length of
the posted text, for a checkbox set the length of all its values joined, so
a legitimate long value's long error is not reported; reporting a bounded
stand-in (the error's type and the text's length) would close that.

**Effort:** S
**Priority:** P2
**Depends on:** None
**Completed:** 2026-11-01 — `quotedValue` quotes the posted text only within the bound, in
`rejectUnlistedOptions`, `guardBind`'s recovered panic and `stringToType`;
`scanError` checks the length before quoting; `setScanned` rejects text
longer than any number before a type that scans only numbers or bools
(`scansOnlyNumbers`) sees it. Not done: reporting a bounded stand-in for a
long error of a long value (see "A long error of long posted text is not
reported").

### A sensitive field shows `onSubmit`'s `FieldErrors` text unmasked

**What:** Mask an `onSubmit` `mx.FieldErrors` entry for a `form:"sensitive"`
field as its parse and validation errors are: its `UserMessage`, else "not
a valid value".

**Why:** After a blank submission the record holds the stored secret, and an
entry such as `fmt.Errorf("token %q rejected", r.Token)` renders it at the
field.

**Context:** `perField` in `ReflectFormHandlerRedirecting` passes entries to
the renderers as they are; predates the `UserMessageError` change.

**Effort:** S
**Priority:** P2
**Depends on:** None
**Completed:** 2026-11-01 — `ReflectFormHandlerRedirecting` masks the entries of sensitive fields and
row cells (`sensitiveFieldPaths`) with `mx.MaskSensitiveError`.

### A server-side timeout counts as a client that went away

**What:** Tell the request's own cancellation (`ctx.Err() != nil`) apart from
a `context.DeadlineExceeded` that a call with its own timeout returned, and
treat only the first as a disconnected client. A join of a real error with
`context.Canceled`, as fan-out results collected after a cancel give, is
the same case.

**Why:** A database query under `context.WithTimeout` fails with
`DeadlineExceeded` while the client still waits: `mx.RespondNonContextError`
then answers nothing (an empty 200), and `mx.ReportError` never hears of it,
also from a form's `onSubmit` or load function.

**Context:** `isContextError` in `errors.go` decides by the error alone, as
the `ReportError` contract and `TestRespondNonContextError_SilentForContextErrors`
require; the 500 path predates the hook. Found in the review of the hook.

**Effort:** S
**Priority:** P2
**Depends on:** None
**Completed:** 2026-11-01 — `clientGone` replaces `isContextError`: a context error counts as a
disconnect only while the request's context is done, in `ReportHiddenError`
and the 500 and SSE error paths.

### A static enum without a validator accepts any submitted value

**What:** Check a submitted single-value enum against its static options
(`Enums`, `EnumStrings`, `NamedOptions`, `Options`) on the server when the
type has no `Validate`/`Valid` method, as a context option list already is.

**Why:** Only a type's own validator rejects a value its list does not
offer: a forged `"42"`, or a forged `"0"` for a required enum whose zero is
no constant, binds and saves. The branch's own enums (go-enum) validate
themselves; a hand-written `Enums()` type does not.

**Context:** `validateField` in `formhandler.go` runs the validation chain;
`validateOptionMembership` in `html/formdecider.go` checks context lists
only. Predates the date widget branch; found in its last review pass.

**Effort:** S
**Priority:** P2
**Depends on:** None
**Completed:** 2026-11-01 — `validateField` checks an enum or enum-set element type without a
validation chain (`hasValidationChain`) against its static option list
(`checkStaticOptions`).

### A forged row origin pairs one row's readonly values with another's key

**What:** Tie a rebuilt repeatable row's origin to its hidden key cells: a
row whose `__row__` marker names a loaded row but whose submitted hidden
cells (such as a per-row `ID`) differ from that row's starts as a new row,
or rows are matched by a declared key (`form:"hidden,key"`) that must be one
of the loaded keys.

**Why:** A rebuilt row starts as a copy of the loaded row its marker names,
so its readonly and sensitive cells are kept, while its hidden key is parsed
from the POST; both are client-controlled and nothing checks that they
belong together. A tampered POST naming row 0 with row 1's key saves row 0's
readonly price (or stored secret) under row 1's key, and an `onSubmit` that
updates by key writes it to the wrong record.

**Context:** `parseRepeatable` and `rowOrigin` in `repeatable.go`. Keeping
the loaded hidden cells instead would break a hidden optimistic-locking
`UpdatedAt` per row, which must come from the client, and comparing parsed
values needs `Equal` for times (a round-tripped time has another location).
Until then, an `onSubmit` must restore readonly row values by key itself.
Found by the date widget branch's adversarial review.

**Effort:** M
**Priority:** P1
**Depends on:** None
**Completed:** 2026-11-01 — a row type with `form:"key"` cells is matched by
key: a key renders hidden and is never parsed, a submitted row starts from
the loaded row whose key it sends (`keyedRowOrigin` in `repeatable.go`), and
a key naming no loaded row makes a new row with a zero key. Keys compare as
the text `mx.FormValueText` shows, so a time key needs no `Equal`, and a
hidden `UpdatedAt` that is no key still comes from the client. A row type
without a key keeps the position markers.

### `GlobPageSource.Dir` is no containment boundary

**What:** The glob and the reads follow symlinks, so a symlink below `Dir`
publishes the file it points to, and a `..` in `Pattern` reaches outside `Dir`.
A `leak.txt -> /etc/hosts` came out as the page `/leak` with its content; on
Linux `/proc/self/environ` reads like a regular file.

**Why:** Kept trusted on purpose (the `Dir` doc says so): deploy tools symlink
content directories, and confining by default would break them. It matters
once people who should not read the server can write below `Dir`, like a
contributed content repo.

**Context:** An opt-in confinement, e.g. `GlobPageSource.Confine`, built on
`os.OpenRoot(dir)`: `fs.Glob(root.FS(), Pattern)` and reads through the root,
so a symlink out of `Dir` fails and `..` is rejected. Codex rated it P1 in the
/ship review of the go-fs v1.0.0-rc.1 bump (2026-10-26).

**Effort:** M
**Priority:** P1
**Depends on:** None
**Completed:** 2026-11-01 — `GlobPageSource.Confine` globs, stats and reads
through an `os.Root` on `Dir`: a symlink is followed only if it is relative
and stays below `Dir`, a file or directory reached through another fails
the source, and a `Pattern` reaching above `Dir` is an error. `Dir` stays trusted by default.

### Form rules duplicated across `mx`, `html` and `shadcn`

**What:** Keep each reflected-form rule in one place in `mx`: when an input
drops HTML `required` (`browserTag` and `hasStoredFile`, copied in `html`
and `shadcn`), the RFC 3339 check of a time wrapper (`textIsRFC3339` in
`html`, a copy of `mx`'s), the "not a valid value" of a `form:"sensitive"`
field's error (masked in `mx`'s `validateField` and `html`'s `maskSensitive`),
and the 64-byte date text limit
(`maxDateTimeLen` in `html`, `maxQuotedDateTimeLen` in `mx`). Also inline
`validateBound`, which has a single caller, `validateField`, and fold
`mx.SelectedOption` into `mx.SelectPreface` (both renderers call one with
the other's result).

**Why:** Each copy can drift from the others: `hasStoredFile` already differs
from the handler's `isEffectivelyEmpty` on a non-nil `*string` holding "".

**Context:** Advisory findings of the date widget branch's review, skipped to
land it; each is a small refactor with an exported helper in `mx`.

**Effort:** S
**Priority:** P4
**Depends on:** None
**Completed:** 2026-11-01 — each rule is one exported `mx` helper both
renderers call: `BrowserTag`, `TextIsRFC3339`, `MaskSensitiveError` (also in
`validateField`, which now holds `validateBound`'s body) and `MaxDateTimeLen`;
`SelectPreface` takes the value and returns the option it shows, replacing
`SelectedOption`. The same change moved the renderers' other copies into `mx`:
`FormValueText`, `OptionComponents`, `NamedOption.Label`, `SetElemType`,
`SetMemberValues`, `FieldLabel` and `StringInputType`. The `*string` drift
of `hasStoredFile` is moot: the html parser rejects an upload into a pointer,
so a file field is a `string` or `[]byte`.

### Date and time inputs drop the stored time zone

**What:** Bind a submitted date or time into a `time.Time` (or a type that
wraps it) in the location of the value it replaces, instead of UTC.

**Why:** A `time.Time` stored as 2026-07-06T09:00+02:00 shows 09:00 in the
input, which has no zone, and an unchanged save stores 09:00 UTC: the time
moves by the offset on every save. Types that wrap `time.Time` behave the same
way since they render and bind like it.

**Context:** `setTime` in `html/formdecider.go` parses with `time.Parse`, which
gives UTC; `time.ParseInLocation` with the loaded value's `Location()` keeps the
wall clock in its zone. It changes what existing `time.Time` fields store, so it
was left out of the string-kind date work. Fields holding text are not
affected: their text has no zone, and text with one falls back to a text input.

**Effort:** S
**Priority:** P2
**Depends on:** None
**Completed:** 2026-10-20 — `setTime` keeps the stored value on an unchanged
submission and reads a changed one in the zone the input showed it in
(`mx.DateTimeInputSpec.Time`): a date at the first instant of its day, a time
on the stored date; a new record still reads UTC.

### Server-side `min`/`max` for `time.Time` date fields

**What:** Check `form:"min=…,max=…"` on the server when binding a `time.Time`
(or other non-text) date, datetime or time field, as `mx.ParseDateTimeInput`
already does for fields that hold the value as text.

**Why:** For a `time.Time` field the range only reaches the browser as the
input's `min`/`max`, so a tampered or scripted POST stores a date outside it.

**Context:** `ParseDateTimeInput` (`formdatetime.go`) parses the bounds with the
input type's layouts and compares times; `setTime` in `html/formdecider.go`
parses `time.Time` with its own layout list. Reusing the same bound check there
keeps both paths consistent. Left out of the string-kind date work because the
text path needed it (its fallback cannot enforce the range in the browser) and
the `time.Time` path was to stay unchanged.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-10-19 — `setTime` checks the bounds with
`mx.CheckDateTimeRange` for `time.Time`, types defined as it, wrappers and date
structs.

### `datetime` and `time` widgets on string-kind fields

**What:** Make `form:"widget=datetime"` and `form:"widget=time"` prefill and
bind a field that holds its value as text (`string`, `*string`, a named string
type), the way `form:"widget=date"` does.

**Why:** A string field tagged `widget=datetime` or `widget=time` rendered an
empty input, and every non-empty submit failed with
`cannot assign time to <type>`, so such a field could not be edited at all.

**Context:** Settled formats: no zone, because these inputs are local times
and attaching or dropping a zone would change what the stored text means; and
seconds only when they are not zero, the form a `datetime-local` input uses
itself (`2006-01-02T15:04[:05]`, `15:04[:05]`). `mx.ParseDateTimeInput`
normalizes on bind, so a round trip does not depend on whether the browser
submits zero seconds. A type with its own spelling converts with `Format` and
`UnmarshalText`. Text the input cannot show, a time with a zone included,
falls back to a text input with the format as placeholder and hint.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-10-19

### A tag-free opt-in for date types

**What:** Let a string-kind date type declare itself a date once, so its
fields render a date input without a `form:"widget=date"` tag on each of
them.

**Why:** A date type is a date wherever it appears, and a field that forgets
the tag renders as a plain text input, which accepts any spelling.

**Context:** `mx.FormWidgetHint` was the opt-in already but did not work for
dates: both renderers and the parser read only `FormTag.Widget`, so a type
returning `"date"` rendered `type="datetime-local"` and failed to bind.
`mx.DetectField` now returns the hint's widget as the tag's `Widget` when the
hint decides the kind, so every renderer and the parser read one value. That
also fixed a `"checkbox"` hint rendering a shadcn switch, an explicit
`widget=text` tag being overridden by an `"email"` hint, and a panic on a nil
pointer to a type with a value-receiver `FormWidget`.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-10-19

### Enum-set checkboxes drop members missing from the option list

**What:** An enum-set field silently loses the checked members that are absent
from the per-request option list.

**Why:** Missing members are not rendered at all, and an unchecked box submits
nothing — so the next save removes them from the set. This is silent data loss
on an ordinary round-trip: the user never sees the values, never touches them,
and loses them by pressing Save.

**Context:** The per-request list is the trust boundary for the set. Members
it does not offer are not rendered and not echoed, and a save does not keep
them: the submitted checkboxes replace the set, so a filtered-out id cannot
be written back and an uncheck cannot be undone by the list changing before
POST. A value outside the list is still rejected. A static enum list is the
whole domain and replaces the set the same way.

**Effort:** M
**Priority:** P1
**Depends on:** None
**Completed:** 2026-10-16

### An empty provider list makes a select unsubmittable or clears it

**What:** An authoritative context provider that returns an empty list, combined
with a non-empty stored value, renders a select containing nothing but the
disabled placeholder.

**Why:** A required field can then never be submitted — the form is a dead end
the user cannot escape. A non-required field silently clears on save instead.
Either way the failure is invisible until someone tries to submit.

**Context:** An empty context-dependent list still renders. A stored select
value shows the generic disabled placeholder, a blank required select shows
"Choose…", and an enum set with members the list does not offer shows the
note and no checkboxes. A provider that failed returns its own error, as
before. The page stays usable; the field does not invent a value.

**Effort:** M
**Priority:** P1
**Depends on:** None
**Completed:** 2026-10-16

### A required select with an empty value submits its first option

**What:** A select whose current value is empty and whose option list has no
empty-valued option still lets the browser display and submit the first option.

**Why:** This is the new-record case, so it is the common one. The user submits
a value they never chose, and `required` is inert client-side because the
control is never actually empty — the browser has nothing to complain about.

**Context:** A select whose list has no empty option gets a leading empty
option: disabled and selected "Choose…" when the field is required and empty,
selected "(none)" when it is optional and empty, and an unselected "(none)"
when it is optional and already has a listed value so it can be cleared.
Required fields that already have a listed value get no extra option. The
out-of-list placeholder is unchanged. The same leading option applies to
static enum lists.

**Effort:** S
**Priority:** P1
**Depends on:** None
**Completed:** 2026-10-16

### `GlobPageSource.Dir` does not scope the glob

**What:** `Dir` is the base directory page paths are derived from, but the glob
itself still runs as a raw `filepath.Glob(Pattern)` against the process working
directory (`web/globpagesource.go`), so `Pattern` has to repeat the directory.

**Why:** Two spellings of the same directory that disagree yield either no
pages or pages whose `Page.Path` is relative to the wrong root, which lands in
the sitemap as wrong URLs. A single source of truth for the directory removes a
class of silent misconfiguration.

**Context:** Making `Dir` scope the match means resolving `Pattern` relative to
it (and deciding whether an absolute `Pattern` stays legal). The related
`content.IsDir()` branch in the same loop is still an empty `// TODO`: a
directory that matches the glob is silently skipped instead of being walked.

**Effort:** M
**Priority:** P3
**Depends on:** None
**Completed:** #31 (2026-08-25) — `Pattern` is matched below `Dir` and an
absolute `Pattern` with a `Dir` set is an error rather than a silent choice
between the two. A matching directory is now skipped explicitly; walking it
recursively was not added, `filepath.Match` has no `**`.

### SSE response type with a per-event flush loop

**What:** A `text/event-stream` response type that renders a `Component` per
event, flushes, and keeps the connection open. No flushing primitive exists in
go-mx today: `http.Flusher` and `Flush(` appear nowhere in the module.

**Why:** Without it there is no token streaming and no server-pushed message
append. `Component.Render(ctx, Writer) error` is pull-based and the context
already carries cancellation for a disconnected client, so nothing else about
the render path blocks streaming — this is the one load-bearing piece.

**Context:** The type owns the flush loop so callers cannot half-use it:
`NewSSEResponse(w, factory)` failing at handler entry if `w` cannot flush, then
`Send`/`SendError`/`Keepalive`/`Close`. Four details that are easy to get wrong:
`data:` is line-delimited, so rendered markup containing a newline must be
re-split into one `data:` line per line (it collides with `CheckedWriter.Newline`
and `WithIndent`); the headers must include `X-Accel-Buffering: no` because
on-premise deployments sit behind the customer's own reverse proxy; `Keepalive`
writes an SSE comment to defeat proxy idle timeouts; and `SendError` is the
in-band error contract — a reserved event name the client binds to, to be
documented together with `hx.EventSSEError` (`hx/events.go`), which is how htmx's
`sse` extension surfaces transport failures.

**Effort:** L
**Priority:** P0
**Depends on:** None
**Completed:** (2026-09-10) — `mx.SSEResponse` in `sse.go`. `NewSSEResponse`
probes the flush capability through the `Unwrap` chain *before* writing any
header, so a non-flushable writer is still answerable with a 500. `Send` renders
into a buffer and returns a render error before a byte of the frame is written,
and re-splits the output on CRLF/CR/LF into one `data:` line each. Frames are
written under a mutex so a `Keepalive` ticker cannot interleave with the
producer. `SendError` follows `RespondNonContextError`: generic message unless
`RevealInternalServerErrors`, silent on a context error, but with the ctx
cancellation stripped so the report still reaches a client that is reading.
Worked example in `cmd/example-sse`, verified in Chrome end to end.

### Content-derived stable element ids

**What:** A deterministic `id` `Attrib` derived from caller-supplied key parts,
alongside the existing counter-based `UniqueID`.

**Why:** `UniqueID()` (`uniqueid.go`) draws from a process-lifetime atomic
counter formatted in base 36. Two renders of the same entity produce different
ids, and a process restart restarts the sequence. Out-of-band swaps target by
id, so appending a token to "message N, part M" needs an id that is the same in
the initial render and in every later event. With only `UniqueID` available, a
streaming append cannot find its own target.

**Context:** Prefer a sanitising join over a hash — `_msg-<uuid>-3` is readable
in devtools where a hash is not, and debuggability is most of the value. Keep the
`_` prefix convention so the result is a valid HTML id that does not start with a
digit. Signature along the lines of `func KeyedID(parts ...any) Attrib`,
documented as "stable across renders and processes" in explicit contrast to
`UniqueID`.

**Effort:** S
**Priority:** P1
**Depends on:** None
**Completed:** (2026-09-10) — `mx.KeyedID` / `mx.KeyedIDValue` in `uniqueid.go`.
A sanitising join rather than a hash, so the id stays readable in devtools:
`KeyedID("msg", id, 3)` renders `_msg-<uuid>-3`. The plan's pure join was
amended on one point: reducing a character away makes distinct keys collide
(`"a.b"` and `"a/b"` become one id, and a collision is invisible — an
out-of-band swap just lands in the first match), so an FNV-1a digest of the
exact parts is appended *only when* a rune was reduced. Keys built from usable
characters alone keep the fully readable form the plan asked for.
`KeyedIDValue` was added beyond the sketch because the stated use case needs
the selector string (`"#"+KeyedIDValue(...)` as an `hx-swap-oob` target), and getting it
out of the `Attrib` otherwise means calling `AttribValue` with a context for a
value that has no context dependency.

### Typed `sse-connect` / `sse-swap` / `sse-close` attributes

**What:** Three typed attribute helpers for the htmx SSE extension.

**Why:** `hx/attributes.go` carries 30+ typed `hx-*` helpers and these are
absent, so every call site hand-writes `mx.NewAttrib("sse-connect", url)` — no
naming, no doc comment, no discoverability. The package already knows about the
extension: `hx/events.go` defines `EventSSEError` with a comment that htmx 2.0
moved SSE out of core.

**Context:** `sse-swap` takes one or more event names, so it should accept a
variadic and join on comma, mirroring how `SwapOOB` takes variadic selectors
(`attributes.go`). `hx.Ext("sse")` already loads the extension, so no new
machinery is needed beyond the attribs.

**Effort:** S
**Priority:** P1
**Depends on:** None
**Completed:** (2026-09-10) — `hx/sse.go`, documented in `hx/README.md`
alongside the two distinct error paths (`hx.EventSSEError` for a transport
failure, `mx.SSEEventError` for one the server reports in-band). `SSESwap` with
no event names defers an error instead of emitting a subscription to nothing.

### Scroll anchoring for `ScrollArea`

**What:** Stick-to-bottom behaviour for a scroll container whose content grows:
follow new content unless the user has scrolled up.

**Why:** It is the single most-noticed behaviour in a chat UI, and its absence is
noticed as a bug rather than a missing feature. `shadcn/scrollarea.go` is pure
classes (`scrollAreaClasses`) with no script.

**Context:** This needs **no new go-mx API**. `hx.OnHTMX` plus the existing
`hx.EventOOBAfterSwap` / `hx.EventAfterSettle` constants (`hx/events.go`) can
drive it, and `ScrollArea(attribsChildren ...any)` composes an extra attrib
naturally. The precedent for a small inline script shipped with a component is
`tabsSelectScript` (`shadcn/tabs.go`). Open question worth deciding before
writing it: library component or consumer-side recipe. A recipe is honest if it
stays ten lines; a `StickToBottom` attrib is better if the near-bottom threshold
needs tuning, because then every consumer would otherwise copy the same tuning.

**Effort:** S
**Priority:** P2
**Depends on:** SSE response type (there is nothing to anchor until content streams)
**Completed:** (2026-09-10) — `StickToBottom` / `StickToBottomThreshold` in
`shadcn/scrollarea.go`. The open question resolved to a library attrib: the
threshold does need tuning, so every consumer would otherwise copy the same
script, and it travels in the attribute value (`data-stick-to-bottom="96"`) so
one shared script serves every instance. It watches DOM mutations rather than
the htmx events the sketch suggested, which makes it work for an SSE stream, an
ordinary swap or any other script, and on a page with no htmx at all. Whether to
follow is recorded on scroll, not measured at mutation time — after a large
append every scroll position looks far from the bottom. Verified in Chrome
against a live `mx.SSEResponse` stream: pinned while content grew (gap 0 as
scrollHeight went 383→533), stayed put after scrolling up mid-stream (gap grew
255→405 instead of snapping back), and resumed following on scrolling back down.
