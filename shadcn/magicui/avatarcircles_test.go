package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestAvatarCircles(t *testing.T) {
	out := render(t, AvatarCircles(AvatarCirclesOptions{
		NumPeople: 99,
		Avatars: []AvatarCircle{
			{ImageURL: "https://example.com/a.png", ProfileURL: "https://example.com/a"},
			{ImageURL: "https://example.com/b.png", ProfileURL: "https://example.com/b"},
		},
	}, html.Class("z-20")))
	mustContain(t, out,
		`data-slot="avatar-circles"`, `class="flex -space-x-4 rtl:space-x-reverse z-20"`,
		`<a href="https://example.com/a" target="_blank" rel="noopener noreferrer"><img class="h-10 w-10 rounded-full border-2 border-white dark:border-gray-800" src="https://example.com/a.png" width="40" height="40" alt="Avatar 1"/></a>`,
		`alt="Avatar 2"`, `href="">+99</a>`,
	)
	if out := render(t, AvatarCircles(AvatarCirclesOptions{})); strings.Contains(out, "+") {
		t.Errorf("no +N circle without NumPeople: %s", out)
	}
}
