package magicui

import (
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx/html"
)

func TestWordRotate(t *testing.T) {
	out := render(t, WordRotate(WordRotateOptions{Words: []string{"Word", "Rotate"}}, html.Class("text-4xl font-bold")))
	mustContain(t, out, `data-slot="word-rotate"`, `class="overflow-hidden py-2"`,
		`<h1 data-slot="word-rotate-text" class="text-4xl font-bold">`,
		"--word-rotate-cycle:5s", "--word-rotate-step:2.5s", "--word-rotate-lead:0.25s",
		// visible from 2*lead to step, gone after step+lead
		"--word-rotate-fade-easing:linear(0 0%, 0 5%, 0.3781 6.25%, 0.6846 7.5%, 0.9065 8.75%, 1 10%, 1 50%,",
		"0 55%, 0 100%)",
		`<span class="sr-only">Word Rotate</span>`,
		`<span class="col-start-1 row-start-1 animate-word-rotate" style="--i:0">Word</span>`,
		`<span class="col-start-1 row-start-1 animate-word-rotate opacity-0" style="--i:1">Rotate</span>`)

	out = render(t, WordRotate(WordRotateOptions{Words: []string{"a", "b", "c"}, Duration: 400 * time.Millisecond}))
	mustContain(t, out, "--word-rotate-cycle:1.2s", "--word-rotate-lead:0.1s")

	// A negative Duration means the default, as in the other components,
	// not a zero-length cycle with negative easing stops.
	out = render(t, WordRotate(WordRotateOptions{Words: []string{"a", "b"}, Duration: -time.Second}))
	mustContain(t, out, "--word-rotate-cycle:5s", "--word-rotate-step:2.5s")
}

func TestWordRotateSingleWord(t *testing.T) {
	out := render(t, WordRotate(WordRotateOptions{Words: []string{"Still"}}))
	if strings.Contains(out, "animate-word-rotate") || strings.Contains(out, "style=\"--word-rotate") {
		t.Errorf("a single word should not rotate: %s", out)
	}
}

func TestWordRotateChildren(t *testing.T) {
	if err := renderErr(t, WordRotate(WordRotateOptions{Words: []string{"a"}}, "child")); err == nil {
		t.Error("a child rendered without error")
	}
}
