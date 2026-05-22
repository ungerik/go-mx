// Package shadcn provides Cn, a Go port of the shadcn/ui `cn` helper.
//
// The shadcn/ui `cn` function is `clsx(inputs)` piped through
// `tailwind-merge`'s `twMerge`. This package is a faithful port of both
// halves against tailwind-merge v3.6.0 (targeting Tailwind CSS v4):
//
//   - The flatten layer (clsx): strings, nested slices and conditional
//     maps are flattened, falsy values dropped.
//   - The merge layer (twMerge): the full default config — every Tailwind
//     utility class group, the class-group trie, modifier/important/postfix
//     parsing, and the conflict tables — so a later class correctly
//     overrides an earlier conflicting one.
//
// The merge algorithm and config are transcribed from the upstream
// TypeScript source; see merge.go, classmap.go, validators.go, parse.go
// and defaultconfig.go.
package shadcn

import (
	"sort"
	"strings"
)

// Cn combines Tailwind CSS classes into a single class string, removing
// earlier classes that a later class overrides.
//
// Accepted argument kinds, matching clsx:
//
//   - string         — used as-is (may contain multiple space-separated classes)
//   - []string       — each element used as-is
//   - []any          — flattened recursively
//   - map[string]bool — keys whose value is true are included
//
// Empty strings, nil, bools and any other kind contribute nothing, so Cn
// can be called with conditional expressions.
//
// Note on maps: clsx applies object keys in insertion order, which a Go
// map cannot reproduce, so keys from a map[string]bool are applied in
// sorted order. For order-dependent conditional classes use a []string.
func Cn(values ...any) string {
	var parts []string
	flatten(values, &parts)
	return twMerge(strings.Join(parts, " "))
}

func flatten(values []any, parts *[]string) {
	for _, v := range values {
		switch x := v.(type) {
		case string:
			if x != "" {
				*parts = append(*parts, x)
			}
		case []string:
			for _, s := range x {
				if s != "" {
					*parts = append(*parts, s)
				}
			}
		case []any:
			flatten(x, parts)
		case map[string]bool:
			keys := make([]string, 0, len(x))
			for k, on := range x {
				if on && k != "" {
					keys = append(keys, k)
				}
			}
			sort.Strings(keys)
			*parts = append(*parts, keys...)
		}
		// nil, false, numbers and other kinds are falsy: contribute nothing.
	}
}
