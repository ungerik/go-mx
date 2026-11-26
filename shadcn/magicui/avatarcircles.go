package magicui

import (
	"strconv"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// AvatarCircle is one avatar of [AvatarCircles]: the image and the profile
// it links to.
type AvatarCircle struct {
	// ImageURL is the avatar image.
	ImageURL string
	// ProfileURL is the page the avatar links to, opened in a new tab.
	ProfileURL string
}

// AvatarCirclesOptions configures [AvatarCircles].
type AvatarCirclesOptions struct {
	// Avatars are the overlapping avatar images, in order.
	Avatars []AvatarCircle
	// NumPeople, if positive, appends a "+NumPeople" circle for the people
	// not shown.
	NumPeople int
}

// AvatarCircles renders a row of overlapping round avatars, each linking to
// its profile in a new tab, optionally followed by a "+N" circle. Images
// get the alt text "Avatar 1", "Avatar 2", … as upstream. Children are
// placed after the avatars. Like upstream, the "+N" circle is a link with
// an empty href, so it points to the current page.
func AvatarCircles(opts AvatarCirclesOptions, attribsChildren ...any) *mx.Element {
	root := html.Div(attribsChildren...)
	children := root.Children
	root.Children = make([]mx.Component, 0, len(opts.Avatars)+1+len(children))
	for i, avatar := range opts.Avatars {
		root.Children = append(root.Children, html.A(
			html.HRef(avatar.ProfileURL),
			html.Target("_blank"),
			html.Rel("noopener", "noreferrer"),
			html.Img(
				html.Class("h-10 w-10 rounded-full border-2 border-white dark:border-gray-800"),
				html.Src(avatar.ImageURL),
				html.Width(40),
				html.Height(40),
				html.Alt("Avatar "+strconv.Itoa(i+1)),
			),
		))
	}
	if opts.NumPeople > 0 {
		root.Children = append(root.Children, html.A(
			html.Class("flex h-10 w-10 items-center justify-center rounded-full border-2 border-white bg-black text-center text-xs font-medium text-white hover:bg-gray-600 dark:border-gray-800 dark:bg-white dark:text-black"),
			html.HRef(""),
			"+"+strconv.Itoa(opts.NumPeople),
		))
	}
	root.Children = append(root.Children, children...)
	return finish(root, "avatar-circles", "z-10 flex -space-x-4 rtl:space-x-reverse", "")
}
