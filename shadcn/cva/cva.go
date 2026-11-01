// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go port of class-variance-authority (cva) v0.7.1,
// Copyright 2022 Joe Bell, licensed under the Apache License, Version 2.0.
// Config.Composes is ported from its successor cva v1.0.0-beta.12, same
// author and license. The library has been reimplemented in Go (a modified
// derivative work). See THIRD-PARTY-LICENSES.md for the full license text.

// Package cva ports class-variance-authority (cva) to Go.
//
// cva compiles a [Config] — a base class string plus a table of variants,
// optional compound variants and defaults — into a [Variants] function that
// resolves a props map to a concatenated class string. It is a faithful port
// of class-variance-authority v0.7.1, plus [Config.Composes] from its
// successor cva v1.0.0-beta.12, which builds a config on other configs.
//
// Like the original, cva only concatenates classes; it does not resolve
// Tailwind conflicts. Follow it with a tailwind-merge step (such as
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
	// Composes lists the configs this one builds on, so a component can share
	// another component's variants instead of restating them. Each composed
	// config resolves first, in order, with the caller's variant props (minus
	// "class") over the effective defaults, then this config's own classes
	// follow. The effective defaults are the composed configs' defaults, a
	// later one overriding an earlier one, overridden by this config's own
	// DefaultVariants: a composer can retune a composed default. Own variants
	// and compound variants resolve against the effective defaults too, so a
	// compound variant may test a composed variant.
	Composes []Config
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
// entry in When, the resolved variant value is one of the listed values. The
// resolved value is the caller's prop, or the default when the prop is omitted
// or empty. An empty When matches unconditionally.
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
// Resolution follows class-variance-authority v0.7.1, preceded by the composed
// configs of cva v1.0.0-beta.12 (see [Config.Composes]): the composed configs,
// then base, then one class set per variant (the caller's value, or the default
// when the prop is omitted or empty), then matching compound variants, then the
// props "class" override. An empty prop counts as omitted at every step, so the
// composed configs, the variants and the compound variants all resolve the
// same value.
// Boolean variants need no special handling — pass "true"/"false" as the prop
// value and key the config the same way.
//
// Three deliberate divergences from the JavaScript original: variant class
// order follows sorted variant names rather than declaration order (Go maps are
// unordered; the resolved class set is identical, and callers tailwind-merge
// it anyway); compound variants treat an empty prop as omitted, where cva
// matches them against a "" that overrides the default even though the variant
// lookup falls back to it (Go callers pass "" for an unset variant, so cva's
// rule would make a compound on a default miss whenever the variant is unset);
// and cva's defineConfig options and VariantProps helper are not ported (the
// package README says what replaces VariantProps).
func New(config Config) Variants {
	composed := make([]Variants, len(config.Composes))
	for i, c := range config.Composes {
		composed[i] = New(c)
	}
	defaults := effectiveDefaults(config)

	variantNames := make([]string, 0, len(config.Variants))
	for name := range config.Variants {
		variantNames = append(variantNames, name)
	}
	slices.Sort(variantNames)
	needResolved := len(composed) > 0 || len(config.CompoundVariants) > 0

	return func(props map[string]string) string {
		// resolved is the defaults overlaid with the caller's non-empty props,
		// what the composed configs and the compound variants see. Withholding
		// "class" keeps a caller override from repeating once per composed
		// config and keeps it last, where a tailwind-merge step lets it win.
		// The defaults are copied only when a prop overrides one, so neither
		// they nor the caller's map change; a resolver compiled by New never
		// writes to its props.
		var resolved map[string]string
		if needResolved {
			resolved = defaults
			copied := false
			for name, value := range props {
				if name == "class" || value == "" || resolved[name] == value {
					continue
				}
				if !copied {
					resolved = make(map[string]string, len(defaults)+len(props))
					maps.Copy(resolved, defaults)
					copied = true
				}
				resolved[name] = value
			}
		}

		parts := make([]string, 0, len(composed)+len(variantNames)+3)
		for _, resolve := range composed {
			if cls := resolve(resolved); cls != "" {
				parts = append(parts, cls)
			}
		}
		if config.Base != "" {
			parts = append(parts, config.Base)
		}

		for _, name := range variantNames {
			// An omitted or empty prop falls back to the default, matching
			// cva's falsyToString(prop) || falsyToString(default).
			value := props[name]
			if value == "" {
				value = defaults[name]
			}
			if cls := config.Variants[name][value]; cls != "" {
				parts = append(parts, cls)
			}
		}

		for _, comp := range config.CompoundVariants {
			if comp.Class == "" {
				continue
			}
			match := true
			for key, accepted := range comp.When {
				if !slices.Contains(accepted, resolved[key]) {
					match = false
					break
				}
			}
			if match {
				parts = append(parts, comp.Class)
			}
		}

		if cls := props["class"]; cls != "" {
			parts = append(parts, cls)
		}

		return strings.Join(parts, " ")
	}
}

// effectiveDefaults returns the default variants config resolves with: those
// of its composed configs in order, a later one overriding an earlier one,
// then its own DefaultVariants. It returns nil when there are none.
func effectiveDefaults(config Config) map[string]string {
	if len(config.Composes) == 0 {
		return config.DefaultVariants
	}
	var defaults map[string]string
	for _, c := range config.Composes {
		if d := effectiveDefaults(c); len(d) > 0 {
			if defaults == nil {
				defaults = make(map[string]string, len(d)+len(config.DefaultVariants))
			}
			maps.Copy(defaults, d)
		}
	}
	if len(config.DefaultVariants) > 0 {
		if defaults == nil {
			return config.DefaultVariants
		}
		maps.Copy(defaults, config.DefaultVariants)
	}
	return defaults
}
