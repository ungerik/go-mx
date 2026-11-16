# go-mx v1.0 readiness audit

Audit of what to change before go-mx tags v1.0.0 and freezes its API, taken on
2026-11-16 at `0e6784d`. It covers the core `mx` package, cross-cutting checks
(vet, staticcheck, race detector, govulncheck, coverage, exported-API scans, a
Go 1.26 build) and package reviews of the forms machinery; `html`/`svg`/`xml`/`hx`;
`shadcn`/`highlight`/`logview`; `pdf`/`pdftable`; and
`web`/`schemaorg`/`stdoptions`/`doc`/`mcp`. Surprising findings were reproduced
with throwaway programs before being written down.

Already done since the audit (not repeated below): the `fpdf` and `wordpress`
nested modules moved to their own repositories,
[`go-mx-fpdf`](https://github.com/ungerik/go-mx-fpdf) and
[`go-mx-wordpress`](https://github.com/ungerik/go-mx-wordpress); `pdf`'s parity
test reads its own `testdata/legacy_reference.pdf`; `hx` loads htmx 2.0.11, with
an opt-in test of the CDN integrity hashes, and the htmx 2 vs. htmx 4 decision
is tracked in [`TODOS.md`](TODOS.md).

The actionable items and the open decisions of this audit are tracked in
[`TODOS.md`](TODOS.md) under `## v1.0 release`, as questions where no decision
has been taken yet.

Tags: **BLOCKER** = a breaking change that must land before the v1 freeze.
**SHOULD** = better before v1 (often breaking, smaller). **LATER** = additive or
non-breaking, can wait. File and line references are as of the audit commit.

---

## 0. Health baseline

- `go vet` is clean, `go test -race` is clean, govulncheck finds no
  vulnerabilities.
- staticcheck (latest) has 5 findings: unused fields in `attribs_test.go:78` and
  `validation_chain_test.go:26`, S1016 in `highlight/highlighter.go:76`, and
  capitalized error strings (ST1005) in `web/robots.go:188,196`.
- Every exported top-level identifier has a doc comment. Struct fields less so:
  `web` 13 undocumented, `mcp` 7, `doc` 56 of 57.
- `func Example`: 7 in the whole repo (xml 3, html 1, entity 1, schemaorg 1,
  mcp 1). None in `mx`, `svg`, `hx`, `shadcn`, `pdf`, `web`, `logview`,
  `highlight`.
- Coverage: root 85.7%, stdoptions 100, mcp 97.9, highlight 96.6, web 94.0,
  logview 93.1, pdftable 89.1, shadcn 88.7, schemaorg 88.5, xml 69.4, hx 66.7.
  Below the 60% minimum: **html 56.7, pdf 54.2, svg 10.8**.
- The root module exports about 4,070 identifiers (methods included): html 816,
  svg 789, shadcn 745, pdf 572, mx 298, entity 279, hx 151, web 123,
  schemaorg 76, logview 63, xml 43, highlight 36, pdftable 33.

Already v1-grade: the deferred-error design (`ErrAttrib`, `NewErrElement`),
escaping by the writer, the doc-comment discipline, the SSE response
(`SSEResponse`, `KeyedID`, resumable ids), `highlight`, the design of
`pdftable`, the go-enum convention, and the test depth of the root package.

---

## 1. Module and package layout

The biggest v1 question. A v1 tag on the root module freezes every package in
it: today a markup core, a reflected-forms framework, a shadcn/ui port that
tracks a moving upstream and Tailwind, an inlined PDF engine, a site framework
with a release-candidate dependency, and an htmx binding whose upstream shipped
a new major in August 2026.

| Package | Recommendation | Why |
|---|---|---|
| `mx` core (Component, Element, Attrib, Writer, CheckedWriter, Text/Raw, If/ForEach, SSE, ids) | **v1, root** | The product. Fix §2 first. |
| `mx` forms (132 of 233 top-level exports, 57%) | **New package `form`**, root module | The core needs nothing from forms (three unexported helpers flow the other way); no import cycles. See §3.1. |
| `html`, `html/entity`, `svg`, `xml` | **v1, root** | After the mechanical cleanup in §3.2. |
| `hx` | **v1 only after the htmx decision** (D2) | htmx 4.0.0 shipped 2026-08-28; `hx` binds htmx 2. |
| `highlight` | **v1, root** | Small, stable, depends only on `mx` and `html`. |
| `logview` | **v1, root, after dropping its `shadcn` import** | It uses `shadcn` only for `ScrollArea` + `StickToBottom`; move `StickToBottom` down to `hx` or `html`. |
| `stdoptions` | **v1, root** | Depends only on `mx`/`form`; CLDR data updates are minor releases. |
| `shadcn` (+ `cva`, `twmerge`) | **Own nested module** `go-mx/shadcn` | Tracks the upstream new-york-v4 and vega snapshots, Tailwind v4 and the tailwind-merge 3.7 tables; `twmerge` builds 623 KB of tables at init in every binary that imports it. |
| `shadcn/clsx` | **`shadcn/internal/clsx`** (or fold into `cn.go`) | One function, one importer. |
| `pdf` (+ `pdftable`) | **Own nested module** `go-mx/pdf`, engine under `pdf/internal/engine` | 26% of the library lines and 15% of the exports; the exported engine API is upstream fpdf's (§3.5). No dependencies beyond the root module. |
| `web` (+ `schemaorg`) | **Own nested module** `go-mx/web` | Only `mcp` imports it; it carries the YAML/TOML dependencies and changes fast. |
| `mcp` | **Nested module, renamed `webmcp`** (directory, path, package) | The package name collides with the SDK's `mcp`, so every user has to alias it. |
| `doc`, `doc/dbdoc` | **Delete** | Unused; `doc.Text.Render` panics on a nil `DefaultRenderer`; `db:"line_hight"` typo. |
| `cmd/mx.go` | **Delete** | A stub CLI that only prints "gen-routes" or "jsx-to-mx". |
| `docs/` | **Add `docs/go.mod`** | 5.2 MB (5.1 MB of gallery) of the root module's tracked files ship in every `go get`. |

**The lockstep catch.** Nested modules isolate *dependencies*, not *versions*,
as long as every module is tagged in lockstep (the policy at the top of
`CHANGELOG.md`): a breaking change in `shadcn`, a Tailwind v5 port say, would
force `/v2` on every module, the root included. Independent stability needs an
exception to the lockstep policy (D1).

**Order matters.** Moving a package out of a module that already has a tagged
version causes ambiguous-import errors for users who still require the old root
version, so every layout move has to land before the first tag.

The split moves `cmd/shadcn-gallery`, `cmd/example-form` and `cmd/example-sse`
into the shadcn module (otherwise root → shadcn → root), gives `webmcp`
replace + placeholder pairs for both root and `web`, and changes the "Gallery up
to date" CI job and the Tailwind `@source` path.

---

## 2. Core `mx`: cross-cutting blockers

**C1 BLOCKER: `If`/`ElseIf` drops a true first branch.**
`mx.If(true, A).ElseIf(false, B).Else(C)` renders `C`, and
`If(true, A).ElseIf(true, B)` renders `B`: `ElseIf` discards its receiver
(`conditional.go:50`). Fix the chain (carry a "matched" flag), and remove `Iff`
and `ElseIff`: they evaluate eagerly, so the closure buys nothing over
`If(f(), …)`, and "iff" reads as if-and-only-if.

**C2 BLOCKER: process-wide mutable configuration.**
- root: `DefaultWriterFactory`, `TextEscaper`, `AsAttribs`, `AsComponent`,
  `RevealInternalServerErrors`, `ReportError`, `SingleValueTypes`,
  `DefaultMaxMemory`, `DefaultMaxUploadSize`, `MaxFieldErrorLen`,
  `DefaultFormTexts`, `DefaultMaxRows`;
- `html`, `hx`, `shadcn`: `var FieldDecider` (reassignable; `hx` and `shadcn`
  read the `html` variable at call time);
- `shadcn`: `PanicOnInvalidID`, `UseUtilities`, `DefaultLabels`;
- `hx`: `ScriptFromCDN`, `ScriptDebugFromCDN`, `ScriptSSEFromCDN` are shared
  mutable `*mx.Element` values;
- `highlight`, `logview`: the themes, `DefaultHighlighted` and `DefaultLevels`
  share their maps and slices with every copy;
- `web`: `DefaultPageRenderer`, `DefaultMaxPageFileSize`;
- `pdf`: `SetDefaultCompression/CatalogSort/CreationDate/ModificationDate`
  write an unlocked package struct that `newRenderer` reads (racy).

Any imported package can flip these for the whole process, they cannot differ
per handler, and tests that set them cannot run in parallel. `TextEscaper` is
the sharpest: one dependency can weaken escaping everywhere. Rule of thumb for
the fix: a limit becomes a constant; what varies per use becomes a field on the
handler, config or writer (`form.Handler.MaxRows`,
`CheckedWriter.WithTextEscaper`); what is request-scoped goes into the context
(`FormTexts`, language); at most a couple of deliberate process defaults stay,
behind `slog.SetDefault`-style setters (`ReportError`). The deciders become
`func` declarations; themes and CDN scripts become functions that return fresh
values. `AsComponent`/`AsAttribs` run when the tree is built, not when it is
rendered, so they cannot move into the writer or the context: freeze the
conversion rules and remove the hooks.

**C3 BLOCKER: the method set of `Writer`.** It is implemented three times in
the repo (`CheckedWriter`, the text writer in `logview/search.go`, and
`hxTriggerWriter` in `hx/formdecider.go`, which wraps a writer to inject
`hx-trigger`). Users will wrap it too, so a method added after v1 breaks them.
Processing instructions are detected by sniffing a trailing `?>` in `Write`.
Decide the final set now (explicit `ProcInst`/`Doctype`, or a documented raw
contract), and give `FieldBehavior` a way to add attributes so `hx` needs no
writer-wrapping trick.

**C4 SHOULD (decide now): void elements are `Children == nil`.**
`br := html.Br(); br.Children = append(br.Children, x)` renders `<br>x</br>`.
Meaning carried by nil vs. empty slice is un-Go-like (DeepEqual, JSON and
`append` all blur it). Use an explicit `Void bool` or a separate type; changing
it after v1 is breaking.

**C5 BLOCKER: the HTTP surface is inconsistent.**
- `html.Document`, `xml.Document`, `web.Robots`, `Sitemap`, `SitemapIndex` and
  `LLMsTxt` have `HandleHTTP`; `pdf.Document` and `ComponentFuncHandler` have
  `ServeHTTP`. Rename all to `ServeHTTP` so they are `http.Handler`s.
- `html.Serve`, `xml.Serve` and `Document.Serve` call `ListenAndServe` without
  timeouts: delete them.
- `mx.ComponentHTTPHandler(c, nil, nil)` answers every request with 500 (the nil
  `WriterFactory` panics and the panic is recovered), while
  `NewComponentFuncHandler` documents nil as "use the default".
- The three handler constructors take `header http.Header` vs.
  `headers ...http.Header`.
- `RespondNonContextError` names an exclusion, not its purpose; `mx.Middleware`
  is too generic a name for "install a FieldDecider", and its doc contradicts
  itself.

**C6 BLOCKER: routing is unfinished.** `Route` (5 methods), `TypedRoute`,
`ComponentFuncHandler` (binds only string fields, by Go field name, and silently
skips the rest), `PathValueNames` (returns `"rest..."` for `{rest...}`),
`JoinPath`, `JoinAbsPath`, `NameToPath`: no users outside tests except `web`
calling `Register`. Leave them out of v1, or shrink `Route` to
`http.Handler` + `Pattern()`.

**C7 SHOULD: library errors.** Every go-errs error carries its call stack in
`Error()`, for example
`"can't EndElement without BeginElement\ngithub.com/ungerik/go-mx.(*CheckedWriter).EndElement\n    github.com/ungerik/go-mx/checkedwriter.go:298\n"`.
It lands in users' logs, and on the page with `RevealInternalServerErrors`.
Prefixes are inconsistent, and apart from `mcp.ErrPageNotFound` there are no
exported sentinel errors or error types, so callers cannot `errors.Is`/`As`
anything. Proposal: plain `"mx: …"`/`"html: …"` errors in library code, plus
sentinels or types for the conditions callers branch on (invalid attribute
name, writer misuse, invalid id); the caller's own wrap adds a stack. Message
text is not semver API, so this is not a hard blocker, but it is the first thing
users see.

**C8 BLOCKER: Content-Security-Policy and inline code.**
- `html.ScriptJS(js ...string)`, `ScriptModuleJS` and `StyleElem(css)` (in
  `html` and `svg`) take no attributes, so inline code cannot carry a `nonce`.
- `StyleElem("</style><script>alert(1)</script>")` renders a live script, and
  `html.Script("a < b")` escapes the JavaScript to `a &lt; b`: script and style
  content has to be raw but checked, not escaped.
- `shadcn` emits 17 inline scripts per component instance (the 6.4 KB calendar
  script repeats for every calendar), 52 inline `on*` handlers in 27 files and
  11 inline styles, so sites with a strict CSP cannot use `shadcn` or `logview`.
- Fix: `ScriptJS(js string, attribs ...mx.Attrib)` and the same for `StyleElem`,
  refusing `</script`/`</style` (any case) at render time like `JSONLD` does;
  `shadcn` behavior through `data-*` attributes and one delegated script that a
  `Scripts()` component renders once per page with the nonce from the context;
  `commandfor`/`command="show-modal"` for dialog triggers.

**C9 BLOCKER: `html.Template`** runs `text/template`, whose output is not
escaped (an XSS hole with any user data), re-parses its glob on every render,
is unused, and takes the name `Template`, which forces the element to be
`TemplateElem`. Delete it and rename `TemplateElem` to `Template`.

**C10 SHOULD: naming in the core.** `Attrib` (interface), `Attribute` (struct)
and `Attribs` (map) differ only by suffix; the struct could become
`StaticAttrib`, or be unexported behind `NewAttrib`. `If` takes `...Component`
while element constructors take `...any`. `ForEach`/`ForEachIter` could be one
`iter.Seq` function.

---

## 3. Per-package findings

### 3.1 Forms (`mx` → `form`)

**Split plan.** Stays in `mx`: `ReportError`, `ReportHiddenError`,
`RevealInternalServerErrors`, `UserMessage`, `UserMessageError`, a renamed
`RespondNonContextError`, and a new exported `ContextWithoutReports`. Moves to
`form`: the 18 forms files plus `FieldErrorTexts`, `BoundFieldError`,
`MaskSensitiveError`, `ContextWithFieldPath`, `FieldPathFromContext`,
`FormErrorText` and `FormMessage`, renamed without stutter: `form.Handler`,
`form.Config`, `form.Fields`, `form.Tag`, `form.Texts`, `form.Decider`,
`form.Behavior`, `form.Message`.

Inside `form`, two tiers: a narrow public interface for renderers (`Decider`,
`Behavior`, `Path`, `Kind`, `Tag`, `ClearSentinelName`, `FieldErrorTexts`, the
option-preface functions, `DateTimeInputSpec`/`DateTimeError`, ideally one
precomputed `FieldView`), and `form/internal` for the parse helpers
(`DriverValue*`, `TextProblem`, `ParseDateTimeInput`, `DateTimeInputLayout`,
`DateTimeFormatError`, `Check*Range`, `CheckTextLength`, `MaxDateTimeLen`,
`TextIsRFC3339`, `FormMaxMemory`, `BoundFieldError`, `MaskSensitiveError`). The
default Parse/Validate, today shared by all three deciders through
`html.FieldDecider`, moves into `form`; `html`, `hx` and `shadcn` only render.
Put the plain-HTML renderer in `form` itself (`form` imports `html`), so the
`html` vocabulary never imports the forms code and is not linked with its
reflection machinery; the `hx` and `shadcn` deciders import `form`. `html` then
stops exporting `FieldTextProblem`, `FileTagAttribs` and `TextTagAttribs`.

| # | Tag | Finding | Fix |
|---|---|---|---|
| F1 | BLOCKER | `FieldDecider func(path, field, value)` and Render/Parse/Validate get no `context.Context` and take fixed positional parameters (`fieldbehavior.go:360,374,379,414`); hiding a field per user needs a decider built per request | `type Decider func(ctx, Field) Behavior`; each callback takes one struct so it can grow |
| F2 | BLOCKER | `IsRenderProbe` hides a marker in a slice's spare capacity (`fieldbehavior.go:392-408`); a wrapper that copies the slice loses it | an explicit `Probe bool` field |
| F3 | BLOCKER | `load`/`onSubmit` get only a context (`formhandler.go:157-161,208-212`), so `load` cannot read `r.PathValue("id")`; every example edits one fixed record | pass the `*http.Request` |
| F4 | BLOCKER | three constructors (`ReflectFormHandler`, `…With`, `…Redirecting`), with a variadic decider of which only `[0]` is read | one generic `form.Handler[T]` struct with `ServeHTTP` and fields Load, Submit (redirect, err), Decider, limits, Redirect, Action, SubmitLabel, Message, WriterFactory |
| F5 | BLOCKER | `ReflectFormFields(target, decider ...)` ignores its parameter; the `path` parameter of `DetectField` is unused | remove both |
| F6 | SHOULD | hidden-input names and row commands are exported (`Sentinel{Present,Clear,Row}`, `…SentinelName`, `Parse…Sentinel`, `AddRowCommand`, `DeleteRowCommand`) although `__cmd__` is unexported | unexport all but `ClearSentinelName` |
| F7 | SHOULD | dead or test-only exports: `IsZero` (ignores `Zeroable`, contrary to its doc), `SetField`, `ImplementsText{Un,}Marshaler`, `FormatOptionValue`, `RunValidationChain`, `MaxNestingDepth` | unexport or delete |
| F8 | SHOULD | `SingleValueTypes` is an exported variable of an unexported type | `RegisterSingleValueType(reflect.Type)` |
| F9 | SHOULD | `FormTag.Widget`/`InputMode` are strings with closed sets; `ParseFormTagString` silently ignores `requird`, `maxlength=abc`, `widget=texarea` | a go-enum `Widget`; return an error or add `form.CheckTags[T]()` |
| F10 | SHOULD | `Validator{Valid() bool}` vs. `SingleErrValidator{Validate() error}` are named the wrong way round; the `SingleErr*` prefix | `Validator{Validate() error}`, `Normalizer{Normalize() error}`, `MultiNormalizer`, `BoolValidator` |
| F11 | SHOULD | `FormMessage` names an error type, a config field and a `shadcn` component | `form.Message`; field `MessageComponent`; `shadcn.FormMessageAlert` |
| F12 | SHOULD | names: `SetElemType`/`SetMemberValues`/`SetField` read as setters; `DriverValueType`/`MultipleFiles` are predicates without `Is`; `NamedOption.Name` is the label; `FieldKindUnknown` = `FieldKindCatchAll`; structural kinds mixed with widget kinds; the zero `DateTimeProblem` is a problem while `TextProblem` starts at 1; argument order varies (`CheckNumberRange(value, raw, tag)` vs. `CheckTextLength(tag, raw)`) | rename and reorder with the split |
| F13 | SHOULD | three language settings: `mx.ContextWithFormTexts`, `shadcn.ContextWithLabels`, `stdoptions.ContextWithLanguage` | one `mx.ContextWithLanguage` |
| F14 | LATER | no decider configured gives the error text inside a 200; `MaxMemory <= 0` vs. `MaxUploadSize == 0` default rules differ; `RegisterNamedOptions` is one global namespace that panics on duplicates; commented-out code in `errors.go:439-447`, `paths.go:50-52`, `funccomponent.go` | |

### 3.2 `html`, `html/entity`, `svg`, `xml`

| # | Tag | Finding | Fix |
|---|---|---|---|
| H1 | BLOCKER | `html.Template` (C9) | delete; `TemplateElem` → `Template` |
| H2 | BLOCKER | the `TableView` interface (`tables.go:13`) is unused, and interfaces cannot grow after v1 | delete |
| H3 | BLOCKER | `WidthPx/WidthEm/HeightPx/HeightEm` emit `width="2em"`, which browsers read as 2 px; unused | delete |
| H4 | BLOCKER | the signatures of `ScriptJS`/`ScriptModuleJS`/`StyleElem` and the breakout (C8) | take attributes, check at render time |
| H5 | BLOCKER | `xml.Decl` builds a `Raw` with `%q`, so `Decl("1.0\"?><evil x=\"", "")` ends the declaration early; `Doctype(text)` is unchecked raw text | validating Components like `ProcInst` |
| H6 | SHOULD | element/attribute name collisions are resolved four ways (`CiteAttr`… vs. `TitleElem`/`StyleElem` vs. the type `Template` vs. `ContentAttr` with nothing to collide with); `html.Title` is an attribute but `svg.Title` an element; `html.Text` is `mx.Text` but `svg.Text` is `<text>` | one rule: the element wins, the attribute gets `Attr`; document `Title`/`Style` as exceptions; `ContentAttr` → `Content` |
| H7 | SHOULD | casing: `HR` vs. `Br`/`WBr`, `Readonly` vs. `AutoFocus`, `OnCanplay` vs. `OnCanPlayThrough`, `CharSet`/`AcceptCharset`/`MetaCharset`, `html.HRef` vs. `svg.Href`/`XLinkHref` | settle one spelling each |
| H8 | SHOULD | shortcuts that break the `<Element><Attribute>` rule: `SubmitButton`, `ResetButton`, `StyleSheet`, `Icon` | `ButtonSubmit`, `ButtonReset`, `LinkStylesheet`, `LinkIcon` |
| H9 | SHOULD | 17 obsolete exports (`Align`, `Background`, `BGColor`, `Border`, `Color`, `CSP`, `IntrinsicSize`, `Language`, `NoEmbed`, `RB`, `RTC` and their `…f` variants) while the valid `<u>` is commented out | delete; add `U` |
| H10 | SHOULD | 50 `…f` attribute variants cover an arbitrary subset (none for `HRef`, `Src`, `For`; none in `svg`/`xml`/`hx`) | delete; `fmt.Sprintf` does the same |
| H11 | SHOULD | `TabIndex`, `Size`, `Start`, `SpanAttr` are generic over `AttribValue`, though the package's own rule says integer attributes take `int` | `int` |
| H12 | SHOULD | `attribValueString` switches on the concrete `float32`/`float64` types: a named float type renders `1e-05`, a `~string` type with a `String()` method renders `String()`, NaN/±Inf render as `NaN`/`+Inf` (also in `svg.ViewBox`) | switch on `reflect.Kind`; `ErrAttrib` for non-finite values |
| H13 | SHOULD | `Document.Meta`/`MetaProperty` are `map[string]string`, so repeated Open Graph properties are impossible (`web/site.go:503-508` already works around it) | an ordered slice of pairs |
| H14 | SHOULD | `html` and `hx` have no package doc comment | add them |
| H15 | LATER | 31 identifiers are shared by `html` and `svg` (dot-import collisions); `html.BoolAttrib` duplicates a bare `mx.ConstAttrib` (`hx` imports `html` only for it); `xml` escaping lets through characters XML forbids, and element names are unchecked (`xml.Element("a b")`); no ARIA helpers; `svg` enums reject `inherit` | document "dot-import at most one package"; fold `BoolAttrib` into `mx` |

Verdicts: `html` needs mechanical work, `entity` is ready, `svg` nearly (`Href`
casing, `StyleElem`), `xml` is ready after H5.

### 3.3 `hx`

- **BLOCKER: the htmx major.** `hx` binds htmx 2 (2.0.11 since this audit).
  htmx 4.0.0 shipped 2026-08-28 and changes about half of what `hx` exports:
  `hx-disable` now means what `hx-disabled-elt` meant, so `hx.Disable` would
  silently change meaning; `Params`, `Request`, `History`, `Inherit`,
  `Disinherit` and `Ext` are removed and `Prompt` became an extension;
  `HX-Trigger-Name` is gone, the request header `HX-Trigger` became `HX-Source`,
  both `HX-Trigger-After-*` headers are gone; every event is renamed
  (`htmx:after:swap`). npm still publishes 2.x as `latest` and 4.x as `next`.
  The decision is tracked in [`TODOS.md`](TODOS.md) under `## hx`.
- SHOULD: `hx/events.go:3-5` documents `On(hx.EventAfterSwap, …)`, which renders
  `hx-on:htmx:afterSwap`; browsers lowercase attribute names, so the handler
  never fires.
- SHOULD: `shadcn` goes through `hx.FieldDecider` only to add
  `hx-trigger="change"` (`shadcn/formdecider.go:33`); decouple it.
- LATER: `Put`/`Patch`/`Delete` name their parameter `value`, `Get`/`Post` `url`.

### 3.4 `shadcn`, `highlight`, `logview`

| # | Tag | Finding | Fix |
|---|---|---|---|
| S1 | BLOCKER | module boundary and version line (§1, D1) | decide |
| S2 | BLOCKER | the scripts define 31 `window` globals, 8 with the `mx` prefix and 23 without (`toast`, `tabsSelect`, `menuKeyNav`, `calendarPick`, `sidebarToggle`, `resizeStart`, …); `toast()` and `calendarPick()` are documented API; `if(!window.toast)` silently yields to an app's own `toast` | one namespace, `window.mx.*` |
| S3 | BLOCKER | no CSP support (C8) | `data-*` attributes and a delegated `Scripts()` with a nonce |
| S4 | SHOULD | `PanicOnInvalidID = true` (48 call sites) panics by default, against the no-panic constructor policy of `mx` (`doc.go:16`) | delete it; always defer with `NewErrElement` |
| S5 | SHOULD | the `ButtonSize` constants have no type prefix (`SizeDefault`, `SizeXS` … `SizeIconLG`), unlike the other 39 enums | `ButtonSizeDefault`, … |
| S6 | SHOULD | variants are positional in 40 functions (`Button(variant, size, …)`, `ToggleGroup(type, variant, size, id, …)`); a new upstream variant axis breaks the signatures | decide: accept that, or let the variant types implement `mx.Attrib` and go in the variadic |
| S7 | SHOULD | the id comes first except in `Slider` (5th) and `ToggleGroup` (4th); state is a positional bool in some components and an HTML attribute in others; the radio `name` goes on the group and on each item; `HoverCardTrigger` and `HoverCardContent` both take `openDelayMs, closeDelayMs int` | id first; delays once, as `time.Duration`; one state convention |
| S8 | SHOULD | five orientation types | one `Orientation` (`Field` keeps its own) |
| S9 | SHOULD | `SectionCard` is unused outside the package and its doc is stale; `StickToBottomDefaultThresholdPx` is exported only for `logview` | delete / unexport |
| S10 | SHOULD | `docs/shadcn/how-to.md:46` says Tailwind finds the classes on its own, but they live in the Go module cache, which Tailwind v4 does not scan, so production builds come out unstyled | document an `@source` line (the path depends on the module split) or ship a class-list generator |
| S11 | SHOULD | `highlight.GoSource` becomes ambiguous once a second language exists (adding one is otherwise non-breaking through a `Highlighter.Tokenize` field) | `GenerateGo` |
| S12 | SHOULD | `logview` imports `shadcn` only for `ScrollArea` + `StickToBottom` | move `StickToBottom` to `hx` or `html` |
| S13 | LATER | building one `Button` takes 28.7 µs, 24.6 KB and 295 allocations (`Cn` re-merges the base class string); the `twmerge` tables cost 1.1 ms, 623 KB and 6,558 allocations at init | a memo cache in `Cn`; build the tables with `sync.OnceValue` |
| S14 | LATER | `finish` merges the caller's class with `context.Background()` when the component is built (`component.go:33`) | merge at render time |
| S15 | LATER | the headless-Chrome tests cover only Calendar, the CodeBlock copy button, the file preview and `StickToBottom`; no browser test for Tabs, ToggleGroup, Toggle, Slider, InputOTP, Command, the menus' keyboard navigation, Tooltip, HoverCard, Resizable, Drawer drag, Sidebar, Toaster, Questionnaire keys or the `logview` script | add them |
| S16 | LATER | docs drift: `shadcn/TODOS.md:23` (the `…Classes` helpers), `shadcn/README.md:309` ("dialogID is not sanitized"), a dead `var _ = http.MethodGet` (`formdecider.go:875-877`); `twmerge` cannot be extended for custom theme scales | fix; an additive `Extend` later |

### 3.5 `pdf`, `pdftable`

| # | Tag | Finding | Fix |
|---|---|---|---|
| P1 | BLOCKER | the fpdf engine **is** the public API: `Renderer` has 217 methods (208 named as in upstream fpdf), 55 take string flags beside the typed enums, colors are int triples beside `Color`, error state lives in `Err()/Ok()/Error()/SetError/SetErrorf/ClearError`, `RawWriteStr/Buf` are exported; `pdf.Text` flows text while `r.Text(x,y,s)` positions it | engine → `pdf/internal/engine`; a curated `Renderer` of about 45 methods (`pdftable` uses 35) |
| P2 | BLOCKER | broken leftovers: `CompareBytes("abc","abcd")` returns nil, `ComparePDFs` dereferences nil (`compare.go:117`), both print to stdout; `TtfParse`, `UTF8CutFont`, `Tickmarks`, `ImageInfo.Gob*`, the `Normalize*` legacy aliases, `New/NewCustom/Init`, `AddFont*`/`FontLoader` needing makefont JSON that no longer exists, `Size.Orientation()` returning "P"/"L", `OutputIntent_GTS_*` | remove or internalize |
| P3 | BLOCKER | the `SetDefault*` globals (C2); `Document` has no date fields, so reproducible output needs the globals | `Document.CreationDate`, `ModDate`, `Compress` |
| P4 | BLOCKER | one text translator per renderer: after `LoadUTF8FontBytes`, core-font text is written as raw UTF-8 into WinAnsi ("Grüße" shows as "GrÃ¼ÃŸe"); `pdftable` has to call `r.Str` | translate per current font; remove `Str`, `SetTranslator`, `UnicodeTranslator*` |
| P5 | BLOCKER | units: bare `float64` in a unit chosen at runtime (`NewRendererLetterPortrait` uses inches, the A4 constructors and `Document` mm); font sizes are always points; an SVG's intrinsic px are used as document units (a 24 px icon becomes 24 mm); components cannot be reused across documents; `Pt(x,y)` reads as "points" | a typed `Length` (`pdf.Mm`, `pdf.In`, `pdf.Pt`) or one fixed unit for the component layer (D4) |
| P6 | BLOCKER | encryption is RC4-40 only (`/V 1 /R 2`) while the README claims 40/128-bit; the owner password comes from `math/rand`; the `CnProtect*` flags are untyped bytes | leave it out of v1; AES-256 later as new API |
| P7 | BLOCKER | `StyleRegular == ""` and `BorderNone == ""` cannot express "unset", which forces `*pdf.FontStyle` in `pdftable` | non-empty values |
| P8 | SHOULD | a truncated PNG (`png.go:248,256`) or TTF (`ttfparser.go:359-396`) panics | recover at the registration boundary |
| P9 | SHOULD | with `GODEBUG=fips140=only` a hello-world `Document.Bytes` panics: SHA-1/MD5 compute font and image ids and the attachment checksum | `hash/fnv` or `maphash` for the ids |
| P10 | SHOULD | `MoveDown` says "keeping x" but resets x to the margin; `Color{R,G,B int}` clamps silently and blocks CMYK later; Factur-X/PDF/A conformance levels are strings; `SetXmpMetadata` vs. `XMPMetadata`; `ImageOptions.ImageType` is a string beside the enum; `ImageBytes` needs a manual cache name though the content is hashed; `Symbol` renders as ZapfDingbats; the Producer is "FPDF 1.7"; `pdftable.AsCell` turns a `pdf.Component` into the text "pdf.ComponentFunc"; `pdf.ContentType` duplicates `mx.ContentTypePDF`; `Document` has no `Fonts`/`Lang` | fix before the freeze |
| P11 | SHOULD | the inlined file headers say "license in the LICENSE file", which is go-mx's; the nested `mcp/` module has no LICENSE of its own | point the headers at `THIRD-PARTY-LICENSES.md`; a LICENSE per module |
| P12 | LATER | invalid enum values pass silently; no Example tests; 41 doc comments point to upstream examples that do not exist, 20 still say "fpdf" | |

### 3.6 `web`, `schemaorg`, `stdoptions`, `mcp`, `doc`

| # | Tag | Finding | Fix |
|---|---|---|---|
| W1 | BLOCKER | a type of go-fs **v1.0.0-rc.8** is in the exported API (`GlobPageSource.Dir`, `Page.Resources`); go-fs supplies one call (`Dir.LocalPath()`), the file work already uses `os.OpenRoot`, `root.FS()` and `io/fs` | `Dir string`; delete `Resources`; drops go-fs, fsnotify, pkg/xattr and x/sys |
| W2 | BLOCKER | `Page.FrontMatter` holds whatever yaml.v2 produces: a key `n` becomes `"false"`, dates stay strings, nested maps are `map[any]any` | split the delimiters in-house (about 80 lines) and decode with `go.yaml.in/yaml/v3` (plus BurntSushi/toml, or drop TOML); drops adrg/frontmatter and yaml.v2 |
| W3 | BLOCKER | the exported fields `Page.PathValues` (its doc is a commented-out line) and `Page.Resources` are unused | delete |
| W4 | BLOCKER | `Site.RegisterRoutes(mux)` takes no context but reads the sources with `context.Background()` (`site.go:531,577`) | `RegisterRoutes(ctx, mux)` |
| W5 | BLOCKER | `doc`/`doc/dbdoc` (§1) | delete |
| W6 | BLOCKER | the `mcp` package name collides with the SDK's | `webmcp` |
| W7 | SHOULD | `HandleHTTP` (C5) | `ServeHTTP` |
| W8 | SHOULD | bug: `PageTypeRenderer` looks up the full Content-Type including the charset (`pagetyperenderer.go:45`), so a `"text/markdown"` key never matches | `mime.ParseMediaType` |
| W9 | SHOULD | `ServeMarkdownDocument` has 6 positional parameters and uses `page` only for `Indexable()` | fields on `MarkdownDocument` plus `ServeHTTP`; fold `ServeMarkdown` in |
| W10 | SHOULD | the field `Site.MarkdownAlternates` vs. the method `Site.MarkdownAlternate`; `Site.LLMsTxtOptions` duplicates the `opts` of `LLMsTxt`; `Site.Routes []mx.Route` registers the caller's own routes; `DefaultRenderPage` duplicates `DefaultPageRenderer` | rename / read the field / drop |
| W11 | SHOULD | `schemaorg.Price(cents)` is fixed at 2 decimals (JPY, KWD/BHD/TND) | `Price(amount int64, decimals int)` |
| W12 | SHOULD | no Examples in `web` (123 exported identifiers); undocumented fields | add them |
| W13 | LATER | `schemaorg.Availability` has 6 of 12 values; no Article/Product/BreadcrumbList; 14 fields typed `any` (narrowing them later breaks); the sitemap fails past 50,000 URLs; the `pagesAtSource` optimization is private to `GlobPageSource`; no `lastmod` front-matter key | additive |

`stdoptions.Register()` (explicit, guarded by `sync.Once`) is fine. Nothing in
`web` knows what a site sells.

---

## 4. Release hygiene: what is missing for a clean v1.0.0

1. **Go version floor.** `go.mod` says `go 1.27.0`, but the library builds with
   Go 1.26.4 once `doc/dbdoc` is gone; only three test files also import the
   Go 1.27 `uuid` package (`formvalue_test.go`, `html/formdecider_test.go`,
   `shadcn/formdecider_test.go`). Set `go 1.26.0` to support both maintained Go
   releases, put the uuid cases behind `//go:build go1.27` (or use a local test
   type), test 1.26 and 1.27 in CI, and update the README's "Requires Go 1.27".
2. **CI** builds, tests, validates the enums and checks gallery drift. Missing:
   `go vet`, staticcheck, `-race`, govulncheck, the same for each nested module,
   and after the tag `apidiff`/`gorelease` against the latest tag on every pull
   request, so the v1 promise is checked by a machine.
3. **Examples**: 7 `func Example` in the repo. pkg.go.dev is the front door; add
   them for `Element`/`If`/`ForEach`, `html.Document`, `form.Handler`,
   `SSEResponse`, `KeyedID`, a `shadcn` component, `pdf.Document`, `web.Site`.
4. **Coverage** under the 60% minimum: `svg` 10.8%, `pdf` 54.2%, `html` 56.7%.
5. **CHANGELOG**: more than 1,100 lines under `[Unreleased]` describe changes to
   an API that never had a release. Replace them with a short `[1.0.0]` entry
   (packages, compatibility promise); the history stays in git.
6. **Compatibility statement** in the README and package docs: what v1 covers,
   which modules stay at v0 (if D1 goes that way), the Go version policy, and
   which htmx, Tailwind and shadcn baselines are bound.
7. **Deletions**: `cmd/mx.go`, `doc/`, and commented-out code
   (`funccomponent.go`, `errors.go:439-447`, `paths.go:50-52`).
8. **Module zip**: a `docs/go.mod` keeps the 5.1 MB gallery out of every
   `go get`.
9. **LICENSE** in every nested module.
10. **Tagging**: the root as `v1.0.0`, the nested modules per D1; `tools/` is
    never tagged.
11. **AGENTS.md / README** follow-ups after the layout change (package list, the
    browser-test claim, the `doc` mention, "Nothing is tagged yet").

---

## 5. Decisions needed

- **D1, version lines.** Lockstep everywhere means every package of every
  module is frozen at v1 and has to clear its blockers first (`shadcn` S2-S3,
  `pdf` P1-P7, `web` W1-W4). The alternative is an exception to the lockstep
  policy: the root goes to v1.0.0 while the `shadcn`, `pdf` and `web` (with
  `webmcp`) modules keep their own v0.x line until they settle. Recommendation:
  the exception.
- **D2, the htmx major of `hx`.** Freeze `hx` as the htmx 2 binding and add
  htmx 4 later as a separate package (additive), or port to htmx 4 before the
  tag. Recommendation: freeze as htmx 2, since npm still publishes 2.x as
  `latest`. Tracked in `TODOS.md`.
- **D3, forms in v1.** Freeze them after the `form` split and F1-F5, or give
  them their own v0 module for more time (a v0 package cannot live inside a v1
  module).
- **D4, `pdf` units** (P5): a typed `Length` vs. one fixed unit.
- **D5, go-errs in library errors** (C7).

## 6. Suggested order

1. Layout and version decisions (D1, D3): they must precede the first tag.
2. Deletions (cheap, and they shrink the frozen surface).
3. Core fixes C1-C10.
4. The `form` split and F1-F13.
5. `html`/`svg`/`xml` cleanup; `hx` per D2.
6. `shadcn` S2-S10 (if it ships at v1).
7. `pdf` P1-P11 (if it ships at v1).
8. `web` W1-W12.
9. Hygiene items 1-11.
10. Tag, and record the apidiff baseline.
