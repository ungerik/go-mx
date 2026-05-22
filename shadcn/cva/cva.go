// Package cva ports class-variance-authority (cva) to Go.
//
// cva compiles a [Config] — a base class string plus a table of variants,
// optional compound variants and defaults — into a [Variants] function that
// resolves a props map to a concatenated class string. It is a faithful port
// of class-variance-authority v0.7.1.
//
// Like the original, cva only concatenates classes; it does not resolve
// Tailwind conflicts. Compose it with a tailwind-merge step (such as
// shadcn.Cn) the way shadcn/ui composes cva with its cn helper.
package cva

import (
	"maps"
	"slices"
	"strings"
)

// Config declares a component's style variants. Its fields mirror the cva
// configuration object.
type Config struct {
	// Base classes always included in the output.
	Base string
	// Variants maps a variant name to a value-to-classes table.
	Variants map[string]map[string]string
	// CompoundVariants are classes applied only when several variants combine.
	CompoundVariants []Compound
	// DefaultVariants gives the value used for a variant the caller omits.
	DefaultVariants map[string]string
}

// Compound is a conditional class selection. Class is appended when, for every
// entry in When, the resolved variant value is one of the listed values. An
// empty When matches unconditionally.
type Compound struct {
	When  map[string][]string
	Class string
}

// Variants is a compiled variant resolver. It takes a props map — variant name
// to value, plus an optional "class" key for a caller override — and returns
// the concatenated class string.
type Variants func(props map[string]string) string

// New compiles a [Config] into a [Variants] resolver.
//
// Resolution follows class-variance-authority v0.7.1: base, then one class set
// per variant (the caller's value, or the default when the prop is omitted or
// empty), then matching compound variants, then the props "class" override.
// Boolean variants need no special handling — pass "true"/"false" as the prop
// value and key the config the same way.
//
// Two deliberate divergences from the JavaScript original: variant class order
// follows sorted variant names rather than declaration order (Go maps are
// unordered; the resolved class set is identical, and callers tailwind-merge
// it anyway), and cva's compose, defineConfig hooks and VariantProps helper
// are not ported.
func New(config Config) Variants {
	variantNames := make([]string, 0, len(config.Variants))
	for name := range config.Variants {
		variantNames = append(variantNames, name)
	}
	slices.Sort(variantNames)

	return func(props map[string]string) string {
		parts := make([]string, 0, len(variantNames)+3)
		if config.Base != "" {
			parts = append(parts, config.Base)
		}

		for _, name := range variantNames {
			// An omitted or empty prop falls back to the default, matching
			// cva's falsyToString(prop) || falsyToString(default).
			value := props[name]
			if value == "" {
				value = config.DefaultVariants[name]
			}
			if cls := config.Variants[name][value]; cls != "" {
				parts = append(parts, cls)
			}
		}

		if len(config.CompoundVariants) > 0 {
			// Compound conditions match against defaults overlaid with props.
			merged := make(map[string]string, len(config.DefaultVariants)+len(props))
			maps.Copy(merged, config.DefaultVariants)
			maps.Copy(merged, props)
			for _, comp := range config.CompoundVariants {
				if comp.Class == "" {
					continue
				}
				match := true
				for key, accepted := range comp.When {
					if !slices.Contains(accepted, merged[key]) {
						match = false
						break
					}
				}
				if match {
					parts = append(parts, comp.Class)
				}
			}
		}

		if cls := props["class"]; cls != "" {
			parts = append(parts, cls)
		}

		return strings.Join(parts, " ")
	}
}
