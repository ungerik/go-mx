# TODOS

Deferred work, grouped by component and sorted by priority. The older `mx`
reflection to-do list lives in `README.md`; the shadcn/ui port build order lives
in `shadcn/TODOS.md`.

Priorities: **P0** blocking · **P1** critical, this cycle · **P2** important ·
**P3** nice-to-have · **P4** someday.

## mx reflected forms

All three were surfaced by the adversarial review of the out-of-list placeholder
(2026-07-06, landed as #23). They share one root: a per-request option list
(`mx.CollectOptions`, `options.go`) and a stored field value can disagree, and
the render path currently resolves that disagreement in a way that loses data or
produces an unsubmittable form.

### Enum-set checkboxes drop members missing from the option list

**What:** An enum-set field silently loses the checked members that are absent
from the per-request option list.

**Why:** Missing members are not rendered at all, and an unchecked box submits
nothing — so the next save removes them from the set. This is silent data loss
on an ordinary round-trip: the user never sees the values, never touches them,
and loses them by pressing Save.

**Context:** Render is `mx.FieldKindEnumSet` in `html/formdecider.go:106` (and
the shadcn mirror in `shadcn/formdecider.go`); parse is `setEnumSet`
(`html/formdecider.go:581`), which rebuilds the set from submitted values only.
A fix has to carry the unrendered members through the round-trip — a hidden
input per missing member, or a parse that merges into the stored set instead of
replacing it. The two paths must agree, so decide the semantics first: is an
option list authoritative over stored members, or only a display filter?

**Effort:** M
**Priority:** P1
**Depends on:** None

### An empty provider list makes a select unsubmittable or clears it

**What:** An authoritative context provider that returns an empty list, combined
with a non-empty stored value, renders a select containing nothing but the
disabled placeholder.

**Why:** A required field can then never be submitted — the form is a dead end
the user cannot escape. A non-required field silently clears on save instead.
Either way the failure is invisible until someone tries to submit.

**Context:** The placeholder comes from the out-of-list branch in
`html/formdecider.go` (~line 295) and `shadcn/formdecider.go`, which returns a
disabled, selected, empty-valued option when the stored value is not in the
list. That branch is correct for a *filtered* list but wrong for an *empty* one.
An empty list from a provider is more likely a failed lookup than a legitimately
empty domain, so the fix is probably to surface it as a render error rather than
emit a form that cannot be used.

**Effort:** M
**Priority:** P1
**Depends on:** None

### A required select with an empty value submits its first option

**What:** A select whose current value is empty and whose option list has no
empty-valued option still lets the browser display and submit the first option.

**Why:** This is the new-record case, so it is the common one. The user submits
a value they never chose, and `required` is inert client-side because the
control is never actually empty — the browser has nothing to complain about.

**Context:** Same failure class the out-of-list placeholder (#23) fixed, from
the other direction: that one handles a value missing from the list, this one
handles a list missing an empty value. The suggested fix is to always prepend an
empty placeholder option for required selects, which also restores client-side
`required` validation. Decide whether it applies to non-required selects too.

**Effort:** S
**Priority:** P1
**Depends on:** None

## shadcn/cva

Follow-ons from the initial `cva` port (class-variance-authority v0.7.1). The
subpackage currently exports only `New(config Config) Variants`
(`shadcn/cva/cva.go:65`), which covers the variant resolution the ported
components actually use.

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

## web

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

## Completed

_Nothing moved here yet. When an item ships, move it to this section unchanged
and append a `**Completed:** <version or PR> (YYYY-MM-DD)` line._
