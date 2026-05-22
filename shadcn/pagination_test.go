package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestPagination(t *testing.T) {
	out := render(t, Pagination(
		PaginationContent(
			PaginationItem(PaginationPrevious(html.HRef("?page=1"))),
			PaginationItem(PaginationLink(false, "", html.HRef("?page=1"), "1")),
			PaginationItem(PaginationLink(true, "", html.HRef("?page=2"), "2")),
			PaginationItem(PaginationEllipsis()),
			PaginationItem(PaginationNext(html.HRef("?page=3"))),
		),
	))
	for _, want := range []string{
		`data-slot="pagination"`,
		`role="navigation"`,
		`aria-label="pagination"`,
		`data-slot="pagination-content"`,
		`data-slot="pagination-item"`,
		`data-slot="pagination-link"`,
		`data-slot="pagination-ellipsis"`,
		">Previous<",
		">Next<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

// TestPaginationLinkActive checks the active page uses the outline variant and
// is marked with aria-current, while an inactive link uses the ghost variant.
func TestPaginationLinkActive(t *testing.T) {
	active := render(t, PaginationLink(true, "", "2"))
	for _, want := range []string{`aria-current="page"`, `data-active="true"`, "border"} {
		if !strings.Contains(active, want) {
			t.Errorf("active link missing %q: %s", want, active)
		}
	}

	inactive := render(t, PaginationLink(false, "", "1"))
	if strings.Contains(inactive, "aria-current") {
		t.Errorf("inactive link should not set aria-current: %s", inactive)
	}
	for _, want := range []string{`data-active="false"`, "hover:bg-accent"} {
		if !strings.Contains(inactive, want) {
			t.Errorf("inactive link missing %q: %s", want, inactive)
		}
	}
}

func TestPaginationPreviousNextIcons(t *testing.T) {
	prev := render(t, PaginationPrevious())
	if !strings.Contains(prev, "lucide-chevron-left") || !strings.Contains(prev, `aria-label="Go to previous page"`) {
		t.Errorf("unexpected PaginationPrevious output: %s", prev)
	}
	next := render(t, PaginationNext())
	if !strings.Contains(next, "lucide-chevron-right") || !strings.Contains(next, `aria-label="Go to next page"`) {
		t.Errorf("unexpected PaginationNext output: %s", next)
	}
}
