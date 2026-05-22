package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestAvatar(t *testing.T) {
	out := render(t, Avatar(
		AvatarImage(html.Src("/me.png"), html.Alt("Me")),
		AvatarFallback("ME"),
	))
	for _, want := range []string{
		`data-slot="avatar"`,
		`data-slot="avatar-image"`,
		`data-slot="avatar-fallback"`,
		`src="/me.png"`,
		"rounded-full",
		">ME<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

// TestAvatarImage covers the native-fallback behavior: the image is a void
// element, overlays the fallback (absolute), and hides itself on load error.
func TestAvatarImage(t *testing.T) {
	out := render(t, AvatarImage(html.Src("/broken.png")))
	for _, want := range []string{"absolute", "inset-0", "onerror="} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "</img>") {
		t.Errorf("avatar image must render as a void element: %s", out)
	}
}

func TestAvatarImageCallerOnErrorOverride(t *testing.T) {
	out := render(t, AvatarImage(html.Src("/x.png"), html.OnError("custom()")))
	if !strings.Contains(out, `onerror="custom()"`) || strings.Contains(out, "this.style.display") {
		t.Errorf("caller onerror should override the default: %s", out)
	}
}
