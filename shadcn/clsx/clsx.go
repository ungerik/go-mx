// Package clsx is a Go port of the clsx npm package: [Join] flattens a mix of
// class-value arguments into a single space-separated class string, dropping
// falsy values.
//
// It does not resolve Tailwind conflicts — compose it with package twmerge for
// that, the way the shadcn cn helper does. See
// [github.com/ungerik/go-mx/shadcn.Cn].
package clsx

import (
	"sort"
	"strings"
)

// Join flattens its arguments into a single space-separated class string.
//
// Accepted argument kinds, matching clsx:
//
//   - string          — used as-is (may contain multiple space-separated classes)
//   - []string        — each element used as-is
//   - []any           — flattened recursively
//   - map[string]bool — keys whose value is true are included
//
// Empty strings, nil, bools and any other kind contribute nothing, so Join
// can be called with conditional expressions.
//
// Note on maps: clsx applies object keys in insertion order, which a Go map
// cannot reproduce, so keys from a map[string]bool are applied in sorted
// order. For order-dependent conditional classes use a []string.
func Join(values ...any) string {
	var b strings.Builder
	flatten(values, &b)
	return b.String()
}

func flatten(values []any, b *strings.Builder) {
	for _, v := range values {
		switch x := v.(type) {
		case string:
			writeClass(b, x)
		case []string:
			for _, s := range x {
				writeClass(b, s)
			}
		case []any:
			flatten(x, b)
		case map[string]bool:
			keys := make([]string, 0, len(x))
			for k, on := range x {
				if on && k != "" {
					keys = append(keys, k)
				}
			}
			sort.Strings(keys)
			for _, k := range keys {
				writeClass(b, k)
			}
		}
		// nil, false, numbers and other kinds are falsy: contribute nothing.
	}
}

// writeClass appends class to b, separating it from any previous content with a
// single space. Empty strings contribute nothing.
func writeClass(b *strings.Builder, class string) {
	if class == "" {
		return
	}
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(class)
}
