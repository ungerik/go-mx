package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func AccordionDemo() mx.Component {
	return shadcn.Accordion(html.Class("w-full max-w-md"),
		shadcn.AccordionItem("faq",
			shadcn.AccordionTrigger("Is it accessible?"),
			shadcn.AccordionContent("Yes. It adheres to the WAI-ARIA design pattern."),
		),
		shadcn.AccordionItem("faq",
			shadcn.AccordionTrigger("Is it styled?"),
			shadcn.AccordionContent("Yes. It comes with default styles that match the other components."),
		),
		shadcn.AccordionItem("faq",
			shadcn.AccordionTrigger("Is it animated?"),
			shadcn.AccordionContent("A native <details> snaps open; the open/close animation is opt-in CSS."),
		),
	)
}
