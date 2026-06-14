package shadcn

import (
	"strings"
	"testing"
)

func TestCarouselComposition(t *testing.T) {
	out := render(t, Carousel(
		CarouselContent(
			CarouselItem("1"),
			CarouselItem("2"),
		),
		CarouselPrevious(),
		CarouselNext(),
	))
	for _, want := range []string{
		`data-slot="carousel"`,
		"relative",
		`data-slot="carousel-content"`,
		"snap-x",
		"snap-mandatory",
		"overflow-x-auto",
		`data-slot="carousel-item"`,
		"basis-full",
		`data-slot="carousel-previous"`,
		`aria-label="Previous slide"`,
		"lucide-chevron-left",
		"scrollBy",
		`data-slot="carousel-next"`,
		`aria-label="Next slide"`,
		"lucide-chevron-right",
		">1<",
		">2<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}
