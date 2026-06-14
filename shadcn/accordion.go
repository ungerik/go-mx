package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/svg"
)

// Accordion renders a shadcn/ui accordion container as a <div>. It is purely
// structural; the items inside are independent native <details> elements (see
// [AccordionItem]).
//
// shadcn's Accordion has a type prop (single / multiple) that Radix uses to
// coordinate exclusivity. The native equivalent is the <details name>
// exclusive-group attribute: items that share a name are mutually exclusive
// (single mode), items with no name behave independently (multiple mode). That
// choice is made per item by the groupName passed to [AccordionItem].
func Accordion(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "accordion", "")
}

// AccordionItem renders one collapsible item as a native <details>. groupName
// is the items' shared exclusive-group identifier:
//
//   - same groupName across items → single mode (only one open at a time, via
//     the native <details name=...> attribute)
//   - "" → multiple mode (the name attribute is omitted, items are independent)
//
// A non-empty groupName is validated with the shared validateID rules. The
// element carries the group base class so the [AccordionTrigger] chevron can
// rotate via group-open:rotate-180.
//
// Pass html.Open to start the item open.
func AccordionItem(groupName string, attribsChildren ...any) *mx.Element {
	e := html.Details(attribsChildren...)
	if groupName != "" {
		validateID(groupName)
		if e.AttribIndex("name") < 0 {
			e.Attribs = append(e.Attribs, html.Name(groupName))
		}
	}
	return finish(e, "accordion-item", "border-b last:border-b-0 group")
}

// accordionTriggerClasses is shadcn/ui's AccordionTrigger class string, minus
// flex-1 (the wrapping AccordionPrimitive.Header is not ported — a <summary>
// is its own header), plus list-none + the webkit marker hider so the browser's
// default disclosure triangle is suppressed in favor of the chevron child.
// The Radix [&[data-state=open]>svg]:rotate-180 is moved to the chevron itself
// as group-open:rotate-180.
const accordionTriggerClasses = "focus-visible:border-ring focus-visible:ring-ring/50 flex items-start justify-between gap-4 rounded-md py-4 text-left text-sm font-medium transition-all outline-none hover:underline focus-visible:ring-[3px] disabled:pointer-events-none disabled:opacity-50 list-none cursor-pointer [&::-webkit-details-marker]:hidden"

// AccordionTrigger renders the always-visible trigger as a <summary>. When
// no caller children include a chevron, a default lucide chevron-down icon is
// appended, carrying group-open:rotate-180 so it rotates when the parent
// [AccordionItem] is open. To use a different indicator, pass it as a child.
func AccordionTrigger(attribsChildren ...any) *mx.Element {
	e := html.Summary(attribsChildren...)
	// Always append the chevron — shadcn's trigger always renders one.
	chevron := icon("chevron-down",
		"text-muted-foreground pointer-events-none size-4 shrink-0 translate-y-0.5 transition-transform duration-200 group-open:rotate-180",
		svg.Path(svg.D("m6 9 6 6 6-6")))
	e.Children = append(e.Children, chevron)
	return finish(e, "accordion-trigger", accordionTriggerClasses)
}

// AccordionContent renders the open-state content as a <div>. shadcn's
// Radix-driven open/close height animation is dropped; the content snaps in
// when the parent [AccordionItem] is open.
func AccordionContent(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "accordion-content", "pb-4 pt-0 text-sm")
}
