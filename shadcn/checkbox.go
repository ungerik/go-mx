package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// checkboxCheckURL is the lucide check icon as a URL-encoded inline-SVG data
// URL, used as a CSS background-image on the appearance-none input when
// checked. The stroke color is baked to white because background-image does
// not inherit currentColor (matches the default primary-foreground theme).
const checkboxCheckURL = "data:image/svg+xml;utf8," +
	"%3Csvg%20xmlns='http://www.w3.org/2000/svg'%20viewBox='0%200%2024%2024'%3E" +
	"%3Cpath%20d='M20%206%209%2017l-5-5'%20stroke='white'%20stroke-width='3'%20fill='none'%20stroke-linecap='round'%20stroke-linejoin='round'/%3E" +
	"%3C/svg%3E"

// checkboxClasses is shadcn/ui's Checkbox class set adapted to a styled void
// <input type="checkbox">. A native checkbox is a void element and cannot hold
// Radix's child <CheckIcon> indicator; the check is drawn with a CSS
// background-image data URL on the appearance-none input — a deliberate
// divergence, mirroring the Switch thumb / RadioGroup dot. The Radix
// data-[state=checked]:* selectors all become checked:* on a native input.
var checkboxClasses = "peer size-4 shrink-0 appearance-none rounded-[4px] border border-input bg-transparent shadow-xs transition-shadow outline-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive dark:bg-input/30 checked:bg-primary checked:border-primary dark:checked:bg-primary bg-no-repeat bg-center bg-[length:0.75rem_0.75rem] checked:bg-[url(\"" + checkboxCheckURL + "\")]"

// Checkbox renders a shadcn/ui checkbox as a styled void <input type="checkbox">.
// Pass html.Checked to start in the checked state, html.Name/html.Value for
// form submission, html.Disabled to disable, html.ID to link a [Label].
//
// Indeterminate is a JavaScript-only DOM property — HTML has no indeterminate
// attribute — so this component cannot start indeterminate from markup alone.
// Use [CheckboxIndeterminateScript] to emit a one-line script that flips the
// property after the page loads.
//
// Children are not valid on a void element and are dropped.
func Checkbox(attribsChildren ...any) *mx.Element {
	e := html.Element("input", attribsChildren...)
	e.Children = nil // <input> is a void element
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("checkbox"))
	}
	return finish(e, "checkbox", checkboxClasses)
}

// CheckboxIndeterminateScript returns a <script> that flips the indeterminate
// DOM property on the checkbox with the given id. Place it after the checkbox
// in the document (and re-run it after any change that should reset the
// indeterminate visual). id is validated with the shared validateID rules.
//
// HTML has no indeterminate attribute; this script bridges that gap.
func CheckboxIndeterminateScript(id string) *mx.Element {
	validateID(id)
	return html.Script(mx.Raw("document.getElementById('" + id + "').indeterminate=true"))
}
