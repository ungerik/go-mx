//go:generate go -C ../tools tool go-enum ../shadcn/$GOFILE

package shadcn

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// NativeSelectSize selects a [NativeSelect]'s height.
type NativeSelectSize string //#enum

const (
	// NativeSelectSizeDefault is the default height, matching [Input].
	NativeSelectSizeDefault NativeSelectSize = "default"
	// NativeSelectSizeSM is the small height.
	NativeSelectSizeSM NativeSelectSize = "sm"
)

// Valid indicates if n is any of the valid values for NativeSelectSize
func (n NativeSelectSize) Valid() bool {
	switch n {
	case
		NativeSelectSizeDefault,
		NativeSelectSizeSM:
		return true
	}
	return false
}

// Validate returns an error if n is none of the valid values for NativeSelectSize
func (n NativeSelectSize) Validate() error {
	if !n.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.NativeSelectSize", n)
	}
	return nil
}

// ValidOr returns n if it is any of the valid values for NativeSelectSize, else defaultVal
func (n NativeSelectSize) ValidOr(defaultVal NativeSelectSize) NativeSelectSize {
	if n.Valid() {
		return n
	}
	return defaultVal
}

// Enums returns all valid values for NativeSelectSize
func (NativeSelectSize) Enums() []NativeSelectSize {
	return []NativeSelectSize{
		NativeSelectSizeDefault,
		NativeSelectSizeSM,
	}
}

// EnumStrings returns all valid values for NativeSelectSize as strings
func (NativeSelectSize) EnumStrings() []string {
	return []string{
		"default",
		"sm",
	}
}

// String implements the fmt.Stringer interface for NativeSelectSize
func (n NativeSelectSize) String() string {
	return string(n)
}

// nativeSelectClasses is the <select>'s skeleton from native-select.tsx plus
// the style-vega.css .cn-native-select rule.
const nativeSelectClasses = "outline-none disabled:pointer-events-none disabled:cursor-not-allowed border-input placeholder:text-muted-foreground selection:bg-primary selection:text-primary-foreground dark:bg-input/30 dark:hover:bg-input/50 focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive dark:aria-invalid:border-destructive/50 h-9 w-full min-w-0 appearance-none rounded-md border bg-transparent py-1 pr-8 pl-2.5 text-sm shadow-xs transition-[color,box-shadow] select-none focus-visible:ring-3 aria-invalid:ring-3 data-[size=sm]:h-8"

// NativeSelect is a Go port of shadcn/ui's NativeSelect: a plain
// appearance-none <select> with a chevron drawn over its right edge, inside a
// wrapper <div>. Class strings are reconstructed from the post-July-2026
// registry (native-select.tsx + the style-vega.css .cn-native-select* rules) —
// see "Upstream restructure (July 2026)" in TODOS.md.
//
// It complements [Select] rather than replacing it: Select styles the open
// picker too, through appearance: base-select, but only in browsers that
// support it and falls back to the native control everywhere else, while
// NativeSelect styles only the closed control and looks the same in every
// browser; the open list is the platform's own.
//
// size may be "" for the default; an unknown value also resolves to the
// default. As upstream does with className, a caller class goes to the
// wrapper, which sizes the control (it is w-fit, the <select> fills it);
// every other attribute — id, name, html.Disabled, aria-invalid — and the
// children go to the <select>. Fill it with [NativeSelectOption] and
// [NativeSelectOptGroup].
func NativeSelect(size NativeSelectSize, attribsChildren ...any) *mx.Element {
	s := string(size.ValidOr(NativeSelectSizeDefault))
	sel := html.Select(attribsChildren...)
	wrapperClasses := takeClassAttribs(sel)
	sel.Attribs = append(sel.Attribs, html.DataAttr("size", s))

	chevron := iconChevronDown()
	chevron.Attribs = append(chevron.Attribs, html.Attrib("aria-hidden", "true"))

	wrapper := html.Div(append(wrapperClasses,
		html.DataAttr("size", s),
		finish(sel, "native-select", nativeSelectClasses),
		finish(chevron, "native-select-icon",
			"pointer-events-none absolute select-none text-muted-foreground top-1/2 right-2.5 size-4 -translate-y-1/2"),
	)...)
	return finish(wrapper, "native-select-wrapper",
		"group/native-select relative w-fit has-[select:disabled]:opacity-50")
}

// NativeSelectOption renders one <option> of a [NativeSelect] or
// [NativeSelectOptGroup]. value is the submitted value; the children are the
// option's text. Pass html.Selected to mark the initial selection,
// html.Disabled to disable it. Its Canvas colors keep the platform's open list
// readable in dark mode.
func NativeSelectOption(value string, attribsChildren ...any) *mx.Element {
	e := html.Option(attribsChildren...)
	if e.AttribIndex("value") < 0 {
		e.Attribs = append(e.Attribs, html.Value(value))
	}
	return finish(e, "native-select-option", "bg-[Canvas] text-[CanvasText]")
}

// NativeSelectOptGroup renders an <optgroup> of a [NativeSelect]. label is the
// group's heading, which the element is invalid without.
func NativeSelectOptGroup(label string, attribsChildren ...any) *mx.Element {
	e := html.OptGroup(attribsChildren...)
	if e.AttribIndex("label") < 0 {
		e.Attribs = append(e.Attribs, html.LabelAttr(label))
	}
	return finish(e, "native-select-optgroup", "bg-[Canvas] text-[CanvasText]")
}
