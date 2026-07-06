# TODOs

Deferred work. The older `mx` reflection to-do list lives in `README.md`.

## mx reflected forms

Surfaced by the adversarial review of the out-of-list placeholder (2026-07-06):

- [ ] A select with an EMPTY current value and no empty-valued option still
      lets the browser display and submit the first option (the new-record
      case) — the same failure class the out-of-list placeholder fixes, and
      it renders `required` client-side inert. Consider always prepending an
      empty placeholder for required selects.
- [ ] Enum-set checkboxes silently drop checked members that are missing
      from the per-request option list: they are not rendered, so the next
      save removes them from the set (unchecked boxes submit nothing).
- [ ] An authoritative context provider returning an empty list combined
      with a non-empty stored value renders a select containing only the
      disabled placeholder: a required field can then never be submitted,
      and a non-required field clears on save.

## shadcn/cva

Follow-ons from the initial `cva` port (class-variance-authority v0.7.1):

- [ ] Port cva's `compose` — merge several variant resolvers into one
- [ ] Port cva's `defineConfig` / `onComplete` hook
- [ ] Decide on a Go equivalent for cva's `VariantProps` type helper, or
      document that a typed props struct replaces it
