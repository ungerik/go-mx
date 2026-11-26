package magicui

import (
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx/html"
)

func TestShimmerButtonDefault(t *testing.T) {
	out := render(t, ShimmerButton(ShimmerButtonOptions{}, "Shimmer"))
	mustContain(t, out,
		`data-slot="shimmer-button"`, `type="button"`,
		`--spread:90deg;--shimmer-color:#ffffff;--radius:100px;--speed:3s;--cut:0.05em;--bg:rgba(0, 0, 0, 1)`,
		"animate-shimmer-slide", "animate-spin-around", "[background:var(--bg)]", "inset-(--cut)",
	)
	// The label sits between the spark and the highlight, as upstream.
	if spark, label := strings.Index(out, "animate-shimmer-slide"), strings.Index(out, "Shimmer<"); spark < 0 || label < spark {
		t.Errorf("label should follow the spark: %s", out)
	}
}

func TestShimmerButtonOptions(t *testing.T) {
	out := render(t, ShimmerButton(ShimmerButtonOptions{
		ShimmerColor: "#ff0", ShimmerSize: "2px", BorderRadius: "8px",
		ShimmerDuration: 1500 * time.Millisecond, Background: "blue",
	}, html.Class("px-2 shadow-2xl"), html.Type("submit"), "Go"))
	mustContain(t, out, "--shimmer-color:#ff0", "--radius:8px", "--speed:1.5s", "--cut:2px", "--bg:blue",
		`type="submit"`, "px-2", "shadow-2xl")
	if strings.Contains(out, "px-6") {
		t.Errorf("caller px-2 should replace px-6: %s", out)
	}
	if err := renderErr(t, ShimmerButton(ShimmerButtonOptions{ShimmerColor: "red;x:y"})); err == nil {
		t.Error("an invalid shimmer color should be a render error")
	}
}
