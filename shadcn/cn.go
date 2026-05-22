// Package shadcn is a Go port of shadcn/ui components, built on the go-mx html
// primitives. It renders HTML on the server with no client runtime, so it is a
// port, not a wrapper: markup and Tailwind classes are reproduced in Go.
//
// The class-handling utilities that shadcn/ui pulls in as separate npm
// packages live in subpackages, each a faithful port of its upstream:
//
//   - clsx    — flattens class-value arguments into one class string
//   - twmerge — resolves Tailwind utility-class conflicts
//   - cva     — class-variance-authority, builds variant class strings
//
// [Cn] is the thin shadcn cn helper that composes clsx and twmerge.
package shadcn

import (
	"github.com/ungerik/go-mx/shadcn/clsx"
	"github.com/ungerik/go-mx/shadcn/twmerge"
)

// Cn is a Go port of the shadcn/ui cn helper: it flattens its arguments with
// [clsx.Join] and resolves Tailwind class conflicts with [twmerge.Merge], so a
// later class overrides an earlier conflicting one.
//
//	Cn("px-2 py-1", "p-4")   // "p-4"
//	Cn("text-sm", "text-lg") // "text-lg"
//
// See [clsx.Join] for the accepted argument kinds.
func Cn(values ...any) string {
	return twmerge.Merge(clsx.Join(values...))
}
