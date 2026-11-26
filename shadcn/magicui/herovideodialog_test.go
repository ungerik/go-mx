package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestHeroVideoDialogDefault(t *testing.T) {
	out := render(t, HeroVideoDialog(HeroVideoDialogOptions{ID: "hv", VideoSrc: "https://example.com/embed/x", ThumbnailSrc: "/thumb.png"}))
	mustContain(t, out,
		`data-slot="hero-video-dialog"`, `class="relative"`,
		`data-slot="dialog-trigger"`, `aria-label="Play video"`, `document.getElementById('hv').showModal()`,
		`<img src="/thumb.png" alt="Video thumbnail" width="1920" height="1080"`,
		`<dialog data-slot="dialog-content" id="hv"`,
		`<iframe src="https://example.com/embed/x" title="Hero Video player"`, `loading="lazy"`, `allowfullscreen`,
		"f.src=f.src", // stops the video on close
		"scale-50 starting:open:scale-50", "backdrop:backdrop-blur-md", "md:max-w-4xl",
		"[&amp;>[data-slot=dialog-close]:last-child]:hidden",
		"absolute -top-16 right-0 rounded-full", `<span class="sr-only">Close</span>`,
	)
	// The dialog box drops the shadcn card look and zoom.
	dialog := out[strings.Index(out, "<dialog"):]
	dialog = dialog[:strings.Index(dialog, ">")]
	for _, gone := range []string{"bg-background", " p-6", "sm:max-w-lg", "scale-95", " duration-200 "} {
		if strings.Contains(dialog, gone) {
			t.Errorf("dialog still has %q: %s", gone, dialog)
		}
	}
}

func TestHeroVideoDialogAnimationStyles(t *testing.T) {
	for style, want := range map[HeroVideoAnimationStyle]string{
		HeroVideoFromBottom:     "translate-y-full starting:open:translate-y-full open:translate-y-0",
		HeroVideoTopInBottomOut: "translate-y-full starting:open:-translate-y-full open:translate-y-0",
		HeroVideoLeftInRightOut: "translate-x-full starting:open:-translate-x-full open:translate-x-0",
		HeroVideoFade:           "scale-100 starting:open:scale-100",
		"bogus":                 "scale-50 starting:open:scale-50",
	} {
		out := render(t, HeroVideoDialog(HeroVideoDialogOptions{ID: "hv", AnimationStyle: style}))
		mustContain(t, out, want)
	}
	for _, style := range HeroVideoAnimationStyle("").Enums() {
		if heroVideoAnimationClasses[style] == "" {
			t.Errorf("no classes for %s", style)
		}
	}
}

func TestHeroVideoDialogCaller(t *testing.T) {
	out := render(t, HeroVideoDialog(HeroVideoDialogOptions{ID: "hv", ThumbnailAlt: "Hero"}, html.Class("block dark:hidden")))
	mustContain(t, out, `class="relative block dark:hidden"`, `alt="Hero"`)
	// An invalid id renders an error like the other magicui ID options,
	// instead of reaching shadcn's id check, which panics by default.
	if renderErr(t, HeroVideoDialog(HeroVideoDialogOptions{ID: "bad id"})) == nil {
		t.Error("invalid dialog id accepted")
	}
	// Without an ID every dialog gets its own.
	a, b := render(t, HeroVideoDialog(HeroVideoDialogOptions{})), render(t, HeroVideoDialog(HeroVideoDialogOptions{}))
	if a == b {
		t.Error("two dialogs without ID share an id")
	}
}
