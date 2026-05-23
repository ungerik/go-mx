package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestAccordionSingleMode(t *testing.T) {
	// Same groupName across items = single mode (native <details name=...>).
	out := render(t, Accordion(
		AccordionItem("faq", AccordionTrigger("Q1"), AccordionContent("A1")),
		AccordionItem("faq", AccordionTrigger("Q2"), AccordionContent("A2")),
	))
	for _, want := range []string{
		`data-slot="accordion"`,
		`data-slot="accordion-item"`,
		"<details ",
		`name="faq"`,
		`data-slot="accordion-trigger"`,
		"<summary ",
		"group-open:rotate-180",
		"lucide-chevron-down",
		`data-slot="accordion-content"`,
		">Q1<",
		">A1<",
		"border-b",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "data-[state=open]") {
		t.Errorf("Radix data-[state=open] should have been rewritten: %s", out)
	}
	// Two items with the same group name.
	if n := strings.Count(out, `name="faq"`); n != 2 {
		t.Errorf("expected 2 name=\"faq\" attributes, got %d: %s", n, out)
	}
}

func TestAccordionMultipleMode(t *testing.T) {
	// Empty groupName = multiple mode (no name attribute).
	out := render(t, Accordion(
		AccordionItem("", AccordionTrigger("A")),
		AccordionItem("", AccordionTrigger("B")),
	))
	if strings.Contains(out, "name=") {
		t.Errorf("multiple mode (empty groupName) should not emit a name attribute: %s", out)
	}
}

func TestAccordionItemValidateGroupName(t *testing.T) {
	for _, bad := range []string{"bad name", "x.y"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected panic for groupName %q", bad)
				}
			}()
			AccordionItem(bad)
		}()
	}
	// "" must NOT panic (it's the multiple-mode opt-out).
	_ = AccordionItem("")
}

func TestAccordionItemOpenPassesThrough(t *testing.T) {
	out := render(t, AccordionItem("g", html.Open, AccordionTrigger("x")))
	if !strings.Contains(out, "open") {
		t.Errorf("html.Open should pass through: %s", out)
	}
}
