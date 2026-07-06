# TODOs

Deferred work. The older `mx` reflection to-do list lives in `README.md`.

## mx reflected forms

Deferred from the per-request select options ship review (2026-07-06):

- [ ] Render a disabled placeholder option when a select's current value is
      not in the per-request option list. Today no option is marked selected
      (the browser shows the first one), and saving requires re-picking
      because POST membership validation rejects out-of-list values.

## shadcn/cva

Follow-ons from the initial `cva` port (class-variance-authority v0.7.1):

- [ ] Port cva's `compose` — merge several variant resolvers into one
- [ ] Port cva's `defineConfig` / `onComplete` hook
- [ ] Decide on a Go equivalent for cva's `VariantProps` type helper, or
      document that a typed props struct replaces it
