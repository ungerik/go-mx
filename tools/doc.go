// Package tools pins the build-time tooling used to maintain go-mx — the
// ungerik/go-enum code generator for the methods of the //#enum types in
// every module of the repository, and gen-stdoptions, which generates the
// stdoptions data from Unicode CLDR (go generate ./stdoptions) — in a
// separate nested module. Keeping them here means go-enum, golang.org/x/text
// and their dependency trees are not part of the shipped
// github.com/ungerik/go-mx module that library consumers import.
//
// Generation is driven by //go:generate directives in the packages that define
// enums, e.g. in svg:
//
//	//go:generate go -C ../tools tool go-enum ../svg/$GOFILE
//
// which runs the go-enum tool pinned in this module against the target file.
// CI checks that the committed methods are up to date with
//
//	go -C tools tool go-enum -validate ../...
//
// plus one run per nested module (../fpdf/..., ../wordpress/..., ../mcp/...),
// because ../... stops at nested modules.
package tools
