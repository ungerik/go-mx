# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./html/...

# Run a single test
go test -run TestFunctionName ./...

# Build check
go build ./...
```

## Project Overview

go-mx is a Go library for generating HTML markup programmatically using a component-based architecture. It provides type-safe, composable HTML generation with automatic escaping and validation.

## Architecture

### Core Abstractions (mx package - root)

- **Component** (`component.go`): The fundamental interface - everything that can render HTML implements `Render(context.Context, Writer) error`
- **Element** (`element.go`): Represents an HTML element with name, attributes, and children. Void elements have `nil` Children, regular elements have empty slice for no children.
- **Attrib** (`attrib.go`): Interface for HTML attributes. Implementations: `Attribute` (name/value pair), `ConstAttrib` (format "name=value"), `BoolAttrib` (boolean HTML attributes)
- **Writer** (`writer.go`): Interface for HTML output with element lifecycle methods
- **CheckedWriter** (`checkedwriter.go`): Writer implementation with validation, escaping, and optional indentation

### Packages

- **html/**: HTML5 elements and attributes - `Div()`, `Span()`, `Class()`, `ID()`, etc.
- **hx/**: HTMX integration - `hx.Get()`, `hx.Post()`, `hx.Trigger()`, etc.
- **shadcn/**: Tailwind CSS class merging utilities
- **web/**, **doc/**, **pdf/**: Higher-level abstractions (partially implemented)

### Key Patterns

**Element Creation**:
```go
// Regular elements take attribs and children as variadic any
html.Div(html.Class("container"), html.P("Hello"))

// Void elements take only Attrib (no children allowed)
html.Img(html.Src("/img.png"), html.Alt("Image"))
```

**Component Conversion** (`DefaultAsComponent` in `component.go`):
Children are passed as `...any` and converted dynamically at render-build
time, not checked at compile time.
- `nil` → nil (renders nothing)
- `string` → `Text` (escaped)
- `Component` → passes through
- Functions → wrapped as `ComponentFunc`
- anything else → `Text(fmt.Sprint(value))` — stringified and escaped.
  A non-`Component` value passed by mistake produces no compile error;
  it silently renders as escaped text. Convert non-obvious children to a
  `Component` explicitly.

**Attribute Conversion** (`DefaultAsAttribs` in `attribs.go`):
- Single `Attrib`, slices, maps, structs with `attr` tags

**Conditional Rendering** (`conditional.go`):
```go
mx.If(condition, component1, component2).Else(fallback)
mx.ForEach(slice, func(v T) Component { ... })
```

### Reflection Features

- `ReflectAttribs()`: Extract attributes from struct fields using `attr` tag
- `ReflectFormComponents()`: Generate form inputs from struct fields using `input` tag
- `ReflectStructFields()`: Iterator over flattened struct fields (handles embedded)

## Code Conventions

- Use `any` instead of `interface{}`
- Use `github.com/domonda/go-errs` for errors (`errs.New`, `errs.Errorf`)
- Use `github.com/domonda/go-types/uu` for UUIDs (`uu.ID`, `uu.IDSlice`, `uu.IDNil`)
- SQL strings: prefix with `/*sql*/` and use backticks
- HTML strings: prefix with `/*html*/` and use backticks
