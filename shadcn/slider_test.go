package shadcn

import (
	"strings"
	"testing"
)

func TestSliderSingleThumb(t *testing.T) {
	out := render(t, Slider(0, 100, 1, []float64{40}, "vol"))
	for _, want := range []string{
		`<input `,
		`data-slot="slider"`,
		`type="range"`,
		`min="0"`,
		`max="100"`,
		`step="1"`,
		`value="40"`,
		"appearance-none",
		"[&amp;::-webkit-slider-thumb]:appearance-none",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// Single-thumb mode must not emit the range wrapper / fill / script.
	for _, gone := range []string{
		`data-slot="slider-track"`,
		`data-slot="slider-range"`,
		"sliderClamp",
	} {
		if strings.Contains(out, gone) {
			t.Errorf("single-thumb slider should not contain %q: %s", gone, out)
		}
	}
}

func TestSliderTwoThumbRange(t *testing.T) {
	out := render(t, Slider(0, 100, 5, []float64{20, 80}, "price"))
	for _, want := range []string{
		`data-slot="slider"`,
		`data-slider="price"`,
		`data-slot="slider-track"`,
		`data-slot="slider-range"`,
		"left: 20%",
		"width: 60%",
		`value="20"`,
		`value="80"`,
		`oninput="sliderClamp('price')"`,
		"[&amp;::-webkit-slider-thumb]:pointer-events-auto",
		"<script>",
		"window.sliderClamp",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if n := strings.Count(out, `type="range"`); n != 2 {
		t.Errorf("expected 2 range inputs, got %d: %s", n, out)
	}
	if n := strings.Count(out, "<script>"); n != 1 {
		t.Errorf("sliderClamp script should be emitted once, got %d: %s", n, out)
	}
}

func TestSliderInvalidLengthPanics(t *testing.T) {
	for _, vs := range [][]float64{{}, {1, 2, 3}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected panic for values=%v", vs)
				}
			}()
			_ = Slider(0, 10, 1, vs, "x")
		}()
	}
}

func TestSliderValidatesID(t *testing.T) {
	for _, bad := range []string{"", "bad id"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected panic for id %q", bad)
				}
			}()
			_ = Slider(0, 10, 1, []float64{5}, bad)
		}()
	}
}
