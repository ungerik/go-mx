package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestBreadcrumb(t *testing.T) {
	out := render(t, Breadcrumb(
		BreadcrumbList(
			BreadcrumbItem(BreadcrumbLink(html.HRef("/"), "Home")),
			BreadcrumbSeparator(),
			BreadcrumbItem(BreadcrumbPage("Current")),
		),
	))
	for _, want := range []string{
		`data-slot="breadcrumb"`,
		`aria-label="breadcrumb"`,
		`data-slot="breadcrumb-list"`,
		`data-slot="breadcrumb-item"`,
		`data-slot="breadcrumb-link"`,
		`data-slot="breadcrumb-separator"`,
		`data-slot="breadcrumb-page"`,
		`aria-current="page"`,
		`href="/"`,
		">Home<",
		">Current<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

// TestBreadcrumbSeparatorDefaultIcon checks the default lucide chevron and
// that a caller-supplied child replaces it.
func TestBreadcrumbSeparatorDefaultIcon(t *testing.T) {
	out := render(t, BreadcrumbSeparator())
	for _, want := range []string{`aria-hidden="true"`, "<svg", "lucide-chevron-right"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	custom := render(t, BreadcrumbSeparator("/"))
	if strings.Contains(custom, "<svg") || !strings.Contains(custom, ">/<") {
		t.Errorf("caller child should replace the default chevron: %s", custom)
	}
}

func TestBreadcrumbEllipsis(t *testing.T) {
	out := render(t, BreadcrumbEllipsis())
	for _, want := range []string{
		`data-slot="breadcrumb-ellipsis"`,
		"lucide-ellipsis",
		"sr-only",
		">More<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}
