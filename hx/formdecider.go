package hx

import (
	"context"
	"reflect"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// FieldDecider is the HTMX layer of the layered form rendering chain.
// It delegates rendering, parsing, and validation to
// [html.FieldDecider], then enriches the rendered element with the
// `hx-trigger="change"` attribute so callers who wire `hx-post` /
// `hx-target` at the form level get live submission as users edit
// fields.
//
// The hx layer is intentionally minimal in v1: deeper HTMX semantics
// (per-field live validation, OOB swaps, server-driven indicators)
// are deferred to a follow-on PR. The layering invariant is what
// matters today — shadcn can wrap hx without referencing html, and a
// future hx feature lands here without touching html or shadcn.
var FieldDecider mx.FieldDecider = func(path mx.FieldPath, field reflect.StructField, value reflect.Value) mx.FieldBehavior {
	base := html.FieldDecider(path, field, value)
	if base.Render == nil {
		return base
	}
	originalRender := base.Render
	base.Render = func(path mx.FieldPath, field reflect.StructField, value reflect.Value, errs []error) mx.Component {
		rendered := originalRender(path, field, value, errs)
		return withHXAttribs(rendered)
	}
	return base
}

// withHXAttribs walks the rendered Component looking for the primary
// input element (input / textarea / select) and adds
// `hx-trigger="change"`. It is a non-invasive enhancement: callers
// who don't wire hx-post on the surrounding form see no behavior
// change.
//
// Component types that are not [mx.Components], [*mx.Element], or
// [mx.ComponentFunc] are returned unchanged.
func withHXAttribs(c mx.Component) mx.Component {
	switch v := c.(type) {
	case nil:
		return nil
	case mx.Components:
		out := make(mx.Components, len(v))
		for i, child := range v {
			out[i] = withHXAttribs(child)
		}
		return out
	case *mx.Element:
		if isLiveInput(v) && v.AttribIndex("hx-trigger") < 0 {
			v.Attribs = append(v.Attribs, Trigger("change"))
		}
		if v.Children != nil {
			for i, child := range v.Children {
				v.Children[i] = withHXAttribs(child)
			}
		}
		return v
	}
	return c
}

// isLiveInput reports whether e is the type of input that benefits
// from an hx-trigger=change. Buttons, hidden inputs, and clear-style
// inputs are excluded so they don't fire spurious requests.
func isLiveInput(e *mx.Element) bool {
	switch e.Name {
	case "textarea", "select":
		return true
	case "input":
		idx := e.AttribIndex("type")
		if idx < 0 {
			return true
		}
		typ, _ := e.Attribs[idx].AttribValue(context.Background())
		switch typ {
		case "hidden", "submit", "reset", "button":
			return false
		}
		// Exclude __clear sentinel checkboxes by name pattern.
		ni := e.AttribIndex("name")
		if ni >= 0 {
			name, _ := e.Attribs[ni].AttribValue(context.Background())
			if len(name) > 0 && name[0] == '_' {
				return false
			}
		}
		return true
	}
	return false
}
