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
// A [mx.ComponentFunc] builds its elements at render time, so this
// build-time walk cannot descend into it — the registry-backed option
// lists (select options, enum-set checkboxes) that need the request
// context are born there. Such a func is wrapped so its deferred output
// is streamed through an [hxTriggerWriter] that applies the same
// injection to the elements as they are written. Any other Component
// type is returned unchanged.
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
	case mx.ComponentFunc:
		return mx.ComponentFunc(func(ctx context.Context, w mx.Writer) error {
			return v(ctx, &hxTriggerWriter{Writer: w})
		})
	}
	return c
}

// isLiveInput reports whether e is the type of input that benefits
// from an hx-trigger=change. Buttons, hidden inputs, and clear-style
// inputs are excluded so they don't fire spurious requests.
func isLiveInput(e *mx.Element) bool {
	typ, hasType := "", false
	if idx := e.AttribIndex("type"); idx >= 0 {
		typ, _ = e.Attribs[idx].AttribValue(context.Background())
		hasType = true
	}
	name := ""
	if idx := e.AttribIndex("name"); idx >= 0 {
		name, _ = e.Attribs[idx].AttribValue(context.Background())
	}
	return liveInputElem(e.Name, hasType, typ, name)
}

// liveInputElem is the shared hx-trigger policy for the build-time
// element walk ([isLiveInput]) and the render-time streaming
// [hxTriggerWriter]. It decides from a bare element identity — tag
// name, whether a type attribute is present and its value, and the name
// attribute — whether the element should receive hx-trigger=change.
func liveInputElem(elem string, hasType bool, typ, name string) bool {
	switch elem {
	case "textarea", "select":
		return true
	case "input":
		if !hasType {
			return true
		}
		switch typ {
		case "hidden", "submit", "reset", "button":
			return false
		}
		// Exclude __clear sentinel checkboxes by name pattern.
		if len(name) > 0 && name[0] == '_' {
			return false
		}
		return true
	}
	return false
}

// hxTriggerWriter wraps a [mx.Writer] and injects hx-trigger="change"
// into live-input start tags (see [liveInputElem]) that don't already
// carry one, applying the same policy as the build-time walk to the
// element events streamed by a deferred [mx.ComponentFunc]. Every other
// Writer call passes straight through untouched.
//
// The render model closes each start tag (CloseElementStartTag, or
// EndElement for a void element) before any child begins, so only the
// most recently opened element's state must be tracked; it is reset on
// each BeginElement. The injection is emitted just before that start tag
// closes, while attributes are still legal.
//
// The wrapper intercepts only the element lifecycle. A live input emitted
// as raw markup (via [mx.Raw] / a direct Write inside an open start tag)
// bypasses injection — but the form deciders always build structured
// [html.Input]/[html.Select]/[html.TextArea] elements, so that boundary is
// unreachable from this package's only caller.
type hxTriggerWriter struct {
	mx.Writer
	inStartTag   bool
	elem         string
	typ          string
	hasType      bool
	name         string
	hasHxTrigger bool
}

func (w *hxTriggerWriter) BeginElement(elem string) error {
	// Track state only after the underlying writer accepts the element, so a
	// rejected BeginElement can't leave the wrapper describing an element the
	// real writer never opened.
	if err := w.Writer.BeginElement(elem); err != nil {
		return err
	}
	w.inStartTag = true
	w.elem = elem
	w.typ = ""
	w.hasType = false
	w.name = ""
	w.hasHxTrigger = false
	return nil
}

func (w *hxTriggerWriter) Attribute(name, value string) error {
	if err := w.Writer.Attribute(name, value); err != nil {
		return err
	}
	switch name {
	case "type":
		w.typ = value
		w.hasType = true
	case "name":
		w.name = value
	case triggerAttrName:
		w.hasHxTrigger = true
	}
	return nil
}

func (w *hxTriggerWriter) CloseElementStartTag() error {
	if err := w.injectTrigger(); err != nil {
		return err
	}
	w.inStartTag = false
	return w.Writer.CloseElementStartTag()
}

func (w *hxTriggerWriter) EndElement() error {
	// A void element (e.g. <input>) reaches EndElement with its start tag
	// still open, so this is its injection point; for a regular element
	// injectTrigger already ran at CloseElementStartTag and is a no-op.
	if err := w.injectTrigger(); err != nil {
		return err
	}
	w.inStartTag = false
	return w.Writer.EndElement()
}

// injectTrigger adds hx-trigger="change" to the currently open start tag
// when it is a live input without one. It is a no-op once the start tag
// has closed or the trigger is already present.
func (w *hxTriggerWriter) injectTrigger() error {
	if !w.inStartTag || w.hasHxTrigger {
		return nil
	}
	if !liveInputElem(w.elem, w.hasType, w.typ, w.name) {
		return nil
	}
	w.hasHxTrigger = true
	return w.Writer.Attribute(triggerAttrName, "change")
}
