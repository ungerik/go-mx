# cva

A Go port of [class-variance-authority](https://github.com/joe-bell/cva) (cva)
v0.7.1 — the variant-builder that shadcn/ui uses to turn a component's props
into a Tailwind class string.

`cva.New` compiles a [`Config`](cva.go) (a base class plus a table of variants,
optional compound variants and defaults) into a `Variants` resolver: a function
that maps a props map to a concatenated class string. Like the original, it only
*concatenates* classes — it does not resolve Tailwind conflicts. Compose it with
a tailwind-merge step (the repo's [`twmerge`](../twmerge/), via
[`shadcn.Cn`](../cn.go)) the way shadcn/ui composes cva with its `cn` helper,
or [choose a concatenator](#choosing-the-concatenator) that merges.

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

## Composing configs

`Config.Composes` builds a config on other configs, so a component can share
another component's variants instead of restating them:

```go
box := cva.Config{
    Variants:        map[string]map[string]string{"shadow": {"sm": "shadow-sm", "md": "shadow-md"}},
    DefaultVariants: map[string]string{"shadow": "sm"},
}
stack := cva.Config{
    Variants: map[string]map[string]string{"gap": {"1": "gap-1", "2": "gap-2"}},
}
card := cva.New(cva.Config{
    Composes: []cva.Config{box, stack},
    Base:     "rounded-lg border",
})

card(map[string]string{"gap": "2", "class": "w-full"})
// → "shadow-sm gap-2 rounded-lg border w-full"
```

The composed configs resolve first, in order, then the config's own base,
variants and compound variants, then the `"class"` override, once. Each
composed config gets the caller's props minus `"class"`, laid over the
effective defaults: the composed configs' `DefaultVariants`, a later one
overriding an earlier one, overridden by the config's own. So a composer can
retune a composed default, and its own variants and compound variants can
test a composed variant. `Composes` is ported from cva v1.0.0-beta.12 (the
successor package to class-variance-authority, same author and license),
since v0.7.1 has nothing like it; it takes `Config` values because, as
upstream accepts only `cva()` components, a composed config must expose its
variants and defaults, which a compiled `Variants` function does not.

## Choosing the concatenator

`cva.DefineConfig` returns a `Factory` whose resolvers end in a concatenator of
your choosing instead of the plain space join of `cva.New`. It is the one place
to post-process every class string the factory generates — wire tailwind-merge
in once instead of at each call site:

```go
var tw = cva.DefineConfig(cva.Options{
    CX: func(classes ...string) string {
        return twmerge.Merge(strings.Join(classes, " "))
    },
})

button := tw.New(cva.Config{
    Base:            "font-semibold bg-gray-200 border rounded",
    Variants:        map[string]map[string]string{"intent": {"primary": "bg-blue-500 text-white"}},
    DefaultVariants: map[string]string{"intent": "primary"},
})

button(map[string]string{"class": "bg-red-500"})
// → "font-semibold border rounded text-white bg-red-500"

tw.CX("bg-gray-200", "bg-blue-500") // the concatenator on its own
// → "bg-blue-500"
```

The concatenator receives the assembled class strings verbatim, one argument
each, in resolution order: each composed config's result, the base, one class
set per variant, the matching compound variants, then the `"class"` override.
Empty strings are left out, so it never sees one, and it is still called when
nothing is left, with no arguments. Composed configs are compiled by the same
factory, so the concatenator runs for each of them first and their results
reach the composer's call as single arguments. A nil `CX`, like the zero
`Factory`, is the space join: `cva.New(config)` is `cva.Factory{}.New(config)`,
the way upstream's own `cva` export is `defineConfig({ cx: clsx })`.

`DefineConfig` is ported from cva v1.0.0-beta.12, where a `cx` option replaced
the earlier `hooks.onComplete` post-processor; an `onComplete` hook is the
concatenator above that joins first and then transforms the result. The
shadcn components do not use it — they merge with `shadcn.Cn` at the component
boundary — so it is for a caller who wants to intercept class generation
program-wide.

## Typed variant props instead of `VariantProps`

In TypeScript, `VariantProps<typeof buttonVariants>` derives a props type from
the cva config, so a caller who writes `variant="outlin"` gets a compile error.
This package has no equivalent helper, on purpose: the Go answer is a named
string type per variant, declared next to the component, which the component
takes as typed parameters and converts to the props map at one place.

```go
type ButtonVariant string //#enum

const (
    ButtonDefault ButtonVariant = "default"
    ButtonOutline ButtonVariant = "outline"
)

func ButtonClasses(variant ButtonVariant, size ButtonSize) string {
    return shadcn.Cn(buttonVariants(map[string]string{
        "variant": string(variant.ValidOr(ButtonDefault)),
        "size":    string(size.ValidOr(SizeDefault)),
    }))
}
```

This is what [`shadcn.Button`](../button.go) and the other ported components
do. A caller sees the variant names as constants in the documentation and in
completion, a value of the wrong variant type does not compile, and the
go-enum `ValidOr` resolves an unknown value (an untyped string constant still
converts implicitly, and a value may come from a request) to the default at
run time. A generic helper could not do better: Go generics cannot derive a
struct type or a set of constants from the keys of a runtime `Config` map, so
it would either reflect over a props struct at run time, which checks nothing
at compile time, or restate the variant names, which the typed constants
already do.

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
| `config.composes` (v1.0.0-beta.12) | `Config.Composes`  |
| `defineConfig({ cx })` (v1.0.0-beta.12) | `DefineConfig(Options{CX: ...})` |
| the `{ cva, cx }` it returns       | `Factory.New`, `Factory.CX` |

The Go types are: `Config.Variants` is `map[string]map[string]string`,
`CompoundVariants` is `[]Compound`, `DefaultVariants` is `map[string]string`,
and `Variants` is `func(props map[string]string) string`.

Resolution order matches cva v0.7.1: **base**, then **one class set per
variant** (the caller's value, or the default when the prop is omitted or
empty — mirroring cva's `falsyToString(prop) || falsyToString(default)`), then
**matching compound variants** (evaluated against defaults overlaid with the
non-empty props), then the **`"class"` override**. Boolean variants need no special
handling: pass `"true"`/`"false"` as the prop value and key the config the same
way (Go has no distinct boolean prop type, so cva's boolean-key coercion
collapses to plain string keys).

The unit tests in [`cva_test.go`](cva_test.go) are derived from
class-variance-authority v0.7.1's own test suite, and those of `Composes` and
`DefineConfig` from cva v1.0.0-beta.12's.

### Deliberate divergences

- **Variant class order** follows sorted variant names rather than the original
  declaration order. Go maps are unordered, so a stable order is chosen by
  sorting the variant names. The resolved class *set* is identical to cva's, and
  callers tailwind-merge the result anyway, so the order does not affect the
  final styling.
- **An empty prop counts as omitted in compound variants too.** cva falls
  back to the default for an empty variant prop but matches compound
  variants against defaults overlaid with every defined prop, so there `""`
  overrides the default. Go callers pass `""` for an unset variant (shadcn's
  components do), which under cva's rule would make a compound variant on a
  default value miss exactly when the variant is unset, and would make a
  config resolve differently once wrapped in `Composes`, which forwards only
  non-empty props. Here the composed configs, the variants and the compound
  variants all resolve the same value. A compound variant listing `""` still
  matches a variant that neither the props nor the defaults set.
- **The concatenator gets strings only, never an empty one.** Upstream's `cx`
  receives class values of its own grammar (clsx arrays and objects) and
  passes an authored `""` or `null` through; a Go `Config` holds strings, and
  an empty one is indistinguishable from an absent one, so it is left out.
  A composed config is compiled by the composer's `Factory` rather than
  keeping the concatenator of the factory that defined it, since `Composes`
  takes `Config` values, not compiled resolvers.
- **Not ported:** the `VariantProps` type helper, whose role typed variant
  parameters fill (see [above](#typed-variant-props-instead-of-variantprops)).

## License

Because this package is a port (a modified derivative work) of cva, it carries
cva's license rather than the repository's MIT license. It is distributed under
the **Apache License, Version 2.0**, Copyright 2022 Joe Bell — see the
[`LICENSE`](LICENSE) file in this directory and the repository's
[THIRD-PARTY-LICENSES.md](../../THIRD-PARTY-LICENSES.md). The rest of go-mx is
licensed under the [MIT License](../../LICENSE).
