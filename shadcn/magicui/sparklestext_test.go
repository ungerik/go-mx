package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestSparklesTextDefault(t *testing.T) {
	out := render(t, SparklesText(SparklesTextOptions{}, "Magic UI"))
	mustContain(t, out, `data-slot="sparkles-text"`, `class="text-6xl font-bold"`,
		`style="--sparkles-first-color:#9E7AFF;--sparkles-second-color:#FE8BBB"`,
		`<span class="relative inline-block">`, "<strong>Magic UI</strong>",
		"pointer-events-none absolute z-20 opacity-0 animate-sparkles-text", `viewBox="0 0 21 21"`, `fill="currentColor"`)
	if n := strings.Count(out, "<svg"); n != 10 {
		t.Errorf("%d sparkles, want 10", n)
	}
}

func TestSparklesTextOptions(t *testing.T) {
	out := render(t, SparklesText(SparklesTextOptions{SparklesCount: 3, Colors: SparklesTextColors{First: "red", Second: "blue"}},
		html.Class("text-4xl"), html.Em("x")))
	mustContain(t, out, `class="font-bold text-4xl"`, "--sparkles-first-color:red;--sparkles-second-color:blue", "<strong><em>x</em></strong>")
	if n := strings.Count(out, "<svg"); n != 3 {
		t.Errorf("%d sparkles, want 3", n)
	}
	if err := renderErr(t, SparklesText(SparklesTextOptions{Colors: SparklesTextColors{First: "red;x:y"}}, "x")); err == nil {
		t.Error("an unsafe color rendered without error")
	}
}

func TestSparklesTextSeed(t *testing.T) {
	a, b := render(t, SparklesText(SparklesTextOptions{Seed: 1}, "x")), render(t, SparklesText(SparklesTextOptions{Seed: 1}, "x"))
	if a != b {
		t.Error("the same seed rendered different markup")
	}
	if render(t, SparklesText(SparklesTextOptions{Seed: 2}, "x")) == a {
		t.Error("another seed rendered the same markup")
	}
	mustContain(t, a, "left:", "top:", "--sparkles-text-scale:", "--sparkles-text-delay:")
}

func TestSparklesTextNegativeCount(t *testing.T) {
	if renderErr(t, SparklesText(SparklesTextOptions{SparklesCount: -5}, "Magic UI")) == nil {
		t.Error("negative SparklesCount accepted")
	}
}
