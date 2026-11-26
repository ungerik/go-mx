package main

import (
	"embed"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"runtime"
	"strings"

	"github.com/ungerik/go-mx"
)

// exampleFS holds the source of every example function so each preview can be
// shown next to the exact Go code that produced it. The files are parsed once
// at startup (see [newSourceMap]); the running gallery never re-reads them.
//
//go:embed examples/*.go examples/magicui/*.go
var exampleFS embed.FS

// exampleDirs are the directories of exampleFS holding example packages:
// the shadcn/ui examples and the Magic UI examples.
var exampleDirs = []string{"examples", "examples/magicui"}

// Example is one labeled preview of a component: a function that renders the
// live component and the Go source of that function's body, extracted from
// [exampleFS]. Source is filled in by [NewRegistry]; callers only supply Name
// and Func, plus ShowSource when the Code tab should show another function's
// source than Func's, such as the month-taking function behind a calendar
// example whose months the gallery also serves as fragments.
type Example struct {
	Name       string
	Func       func() mx.Component
	ShowSource any
	Source     string
}

// ComponentDoc is one component's documentation page: a slug for its route, a
// title, a one-line description (transcribed from the upstream docs), the
// ordered list of examples shown on the page, and the library group the
// sidebar and the index list it under ("" for shadcn/ui).
type ComponentDoc struct {
	Slug        string
	Title       string
	Description string
	Examples    []Example
	Group       string
}

// DocGroup is a titled run of component pages, such as the Magic UI ones.
type DocGroup struct {
	Title string
	Docs  []ComponentDoc
}

// groupTitle is the heading of a ComponentDoc.Group.
func groupTitle(group string) string {
	if group == "" {
		return "shadcn/ui"
	}
	return group
}

// Groups returns the pages grouped by ComponentDoc.Group, in the order each
// group first appears.
func (r *Registry) Groups() []DocGroup {
	var groups []DocGroup
	index := map[string]int{}
	for _, d := range r.Docs {
		i, ok := index[d.Group]
		if !ok {
			i = len(groups)
			index[d.Group] = i
			groups = append(groups, DocGroup{Title: groupTitle(d.Group)})
		}
		groups[i].Docs = append(groups[i].Docs, d)
	}
	return groups
}

// Registry is the ordered set of component pages plus a slug index. It owns the
// source map so every Example.Source is resolved exactly once.
type Registry struct {
	Docs   []ComponentDoc
	bySlug map[string]*ComponentDoc
}

// NewRegistry resolves the Go source for every example and builds the slug
// index. The input docs are taken by value and stored with their Source fields
// populated.
func NewRegistry(docs []ComponentDoc) *Registry {
	src := newSourceMap()
	r := &Registry{bySlug: make(map[string]*ComponentDoc, len(docs))}
	for _, d := range docs {
		for i := range d.Examples {
			var shown any = d.Examples[i].Func
			if d.Examples[i].ShowSource != nil {
				shown = d.Examples[i].ShowSource
			}
			d.Examples[i].Source = src[funcName(shown)]
		}
		r.Docs = append(r.Docs, d)
	}
	for i := range r.Docs {
		slug := r.Docs[i].Slug
		if _, dup := r.bySlug[slug]; dup {
			// Pages share one URL space (components/<slug>/) and one id
			// space (tab ids), so a slug must be unique across groups.
			panic("duplicate component slug " + slug)
		}
		r.bySlug[slug] = &r.Docs[i]
	}
	return r
}

// Lookup returns the page for a slug, or nil if there is none.
func (r *Registry) Lookup(slug string) *ComponentDoc {
	return r.bySlug[slug]
}

// funcName returns the package name and name of a function value, e.g.
// "examples.ButtonDefault" for examples.ButtonDefault and
// "magicui.MarqueeDemo" for the Magic UI example. Names are unique within an
// example package, so this is a stable key into the source map.
func funcName(fn any) string {
	full := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	if i := strings.LastIndex(full, "/"); i >= 0 {
		return full[i+1:]
	}
	return full
}

// newSourceMap parses every embedded example file and maps each top-level
// function, keyed "<package>.<name>" like [funcName], to the dedented source
// shown in the "Code" tab. Parse errors and non-function declarations are
// skipped.
func newSourceMap() map[string]string {
	m := make(map[string]string)
	fset := token.NewFileSet()
	for _, dir := range exampleDirs {
		entries, _ := exampleFS.ReadDir(dir)
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
				continue
			}
			src, err := exampleFS.ReadFile(dir + "/" + e.Name())
			if err != nil {
				continue
			}
			file, err := parser.ParseFile(fset, e.Name(), src, parser.SkipObjectResolution)
			if err != nil {
				continue
			}
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv != nil || fn.Body == nil {
					continue
				}
				lo, hi := snippetBounds(fset, fn.Body)
				if lo < 0 || hi > len(src) {
					continue
				}
				m[file.Name.Name+"."+fn.Name.Name] = dedent(strings.Trim(string(src[lo:hi]), "\n"))
			}
		}
	}
	return m
}

// snippetBounds returns the byte range of the source to display for a function
// body. For the common single `return <expr>` body it returns just <expr>, so
// the snippet is the bare component constructor with no `return` keyword;
// otherwise it returns the full body between the braces.
func snippetBounds(fset *token.FileSet, body *ast.BlockStmt) (lo, hi int) {
	if len(body.List) == 1 {
		if ret, ok := body.List[0].(*ast.ReturnStmt); ok && len(ret.Results) == 1 {
			lo = fset.Position(ret.Results[0].Pos()).Offset
			hi = fset.Position(ret.Results[0].End()).Offset
			if lo < hi {
				return lo, hi
			}
		}
	}
	lo = fset.Position(body.Lbrace).Offset + 1
	hi = fset.Position(body.Rbrace).Offset
	if lo >= hi {
		return -1, -1
	}
	return lo, hi
}

// dedent removes one leading tab from every line — the indentation gofmt adds
// inside a function body — so the displayed snippet starts at column zero while
// keeping its relative indentation.
func dedent(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimPrefix(line, "\t")
	}
	return strings.Join(lines, "\n")
}
