# cva

A Go port of [class-variance-authority](https://github.com/joe-bell/cva) (cva)
v0.7.1 — the variant-builder that shadcn/ui uses to turn a component's props
into a Tailwind class string.

`cva.New` compiles a [`Config`](cva.go) (a base class plus a table of variants,
optional compound variants and defaults) into a `Variants` resolver: a function
that maps a props map to a concatenated class string. Like the original, it only
*concatenates* classes — it does not resolve Tailwind conflicts. Compose it with
a tailwind-merge step (the repo's [`twmerge`](../twmerge/), via
[`shadcn.Cn`](../cn.go)) the way shadcn/ui composes cva with its `cn` helper.

## Usage

```go
import "github.com/ungerik/go-mx/shadcn/cva"

buttonVariants := cva.New(cva.Config{
    Base: "inline-flex items-center justify-center rounded-md text-sm font-medium",
    Variants: map[string]map[string]string{
        "variant": {
            "default":     "bg-primary text-primary-foreground",
            "destructive": "bg-destructive text-white",
            "outline":     "border bg-background",
        },
        "size": {
            "default": "h-9 px-4 py-2",
            "sm":      "h-8 px-3",
            "lg":      "h-10 px-6",
        },
    },
    DefaultVariants: map[string]string{
        "variant": "default",
        "size":    "default",
    },
})

// Pick variants via the props map:
buttonVariants(map[string]string{"variant": "outline", "size": "lg"})
// → "inline-flex items-center justify-center rounded-md text-sm font-medium h-10 px-6 border bg-background"

// Omitted props fall back to DefaultVariants; the reserved "class" key
// appends a caller override last:
buttonVariants(map[string]string{"class": "w-full"})
// → base + default variant + default size + "w-full"
```

## How it was ported

cva's TypeScript API maps onto Go types one-to-one:

| class-variance-authority (TS)     | this package (Go)          |
| --------------------------------- | -------------------------- |
| `cva(base, config)`               | `New(Config{...})`         |
| the returned variant function     | `Variants`                 |
| `config.variants`                 | `Config.Variants`          |
| `config.compoundVariants`         | `Config.CompoundVariants`  |
| `config.defaultVariants`          | `Config.DefaultVariants`   |
| `props` object                    | `props` map argument       |
| `props.class` / `props.className` | the reserved `"class"` key |

The Go types are: `Config.Variants` is `map[string]map[string]string`,
`CompoundVariants` is `[]Compound`, `DefaultVariants` is `map[string]string`,
and `Variants` is `func(props map[string]string) string`.

Resolution order matches cva v0.7.1: **base**, then **one class set per
variant** (the caller's value, or the default when the prop is omitted or
empty — mirroring cva's `falsyToString(prop) || falsyToString(default)`), then
**matching compound variants** (evaluated against defaults overlaid with
props), then the **`"class"` override**. Boolean variants need no special
handling: pass `"true"`/`"false"` as the prop value and key the config the same
way (Go has no distinct boolean prop type, so cva's boolean-key coercion
collapses to plain string keys).

The unit tests in [`cva_test.go`](cva_test.go) are derived from
class-variance-authority v0.7.1's own test suite.

### Deliberate divergences

- **Variant class order** follows sorted variant names rather than the original
  declaration order. Go maps are unordered, so a stable order is chosen by
  sorting the variant names. The resolved class *set* is identical to cva's, and
  callers tailwind-merge the result anyway, so the order does not affect the
  final styling.
- **Not ported:** cva's `compose`, the `defineConfig` / `onComplete` hooks, and
  the `VariantProps` type helper. In Go a typed props struct or named constants
  fill the role of `VariantProps`.

## License

Because this package is a port (a modified derivative work) of cva, it carries
cva's license rather than the repository's MIT license. It is distributed under
the **Apache License, Version 2.0**, Copyright 2022 Joe Bell — see the
[`LICENSE`](LICENSE) file in this directory and the repository's
[THIRD-PARTY-LICENSES.md](../../THIRD-PARTY-LICENSES.md). The rest of go-mx is
licensed under the [MIT License](../../LICENSE).
