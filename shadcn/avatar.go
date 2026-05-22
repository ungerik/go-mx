package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Avatar renders a shadcn/ui avatar container as a <span>. Place an
// [AvatarImage] and an [AvatarFallback] inside it.
func Avatar(attribsChildren ...any) *mx.Element {
	return finish(html.Span(attribsChildren...), "avatar",
		"relative flex size-8 shrink-0 overflow-hidden rounded-full")
}

// AvatarImage renders the avatar's image as a void <img>. Pass html.Src and
// html.Alt the normal way.
//
// shadcn/ui's Avatar mounts either the image or the fallback based on a
// JavaScript load check. This port renders both: the image is positioned
// absolute inset-0 so it overlays the fallback, and it carries a default
// onerror handler that hides the image when its src fails to load, revealing
// the fallback beneath. The absolute inset-0 is a deliberate divergence from
// shadcn's verbatim "aspect-square size-full" — the native replacement for
// Radix's mount/unmount. Pass your own onerror attribute to override the
// default. Children are not valid on a void element and are dropped.
func AvatarImage(attribsChildren ...any) *mx.Element {
	e := html.Element("img", attribsChildren...)
	e.Children = nil // <img> is a void element
	if e.AttribIndex("onerror") < 0 {
		e.Attribs = append(e.Attribs, html.OnError("this.style.display='none'"))
	}
	return finish(e, "avatar-image", "absolute inset-0 aspect-square size-full")
}

// AvatarFallback renders the avatar's fallback content (typically initials) as
// a <span>. It shows when no [AvatarImage] is present or the image fails to
// load.
func AvatarFallback(attribsChildren ...any) *mx.Element {
	return finish(html.Span(attribsChildren...), "avatar-fallback",
		"bg-muted flex size-full items-center justify-center rounded-full")
}
