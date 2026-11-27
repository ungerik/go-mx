package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Helpers shared by the device mockups [Iphone], [Safari] and [Android].

// deviceVideo is the muted, looping, inline-playing <video> of the device
// mockups; withPreload adds upstream's preload="metadata", which Android
// does not set.
func deviceVideo(src, class string, withPreload bool) *mx.Element {
	video := html.Video(
		html.Class(class),
		html.Src(src),
		html.AutoPlay,
		html.Loop,
		html.Muted,
		html.PlaysInline,
	)
	if withPreload {
		video.Attribs = append(video.Attribs, mx.NewAttrib("preload", "metadata"))
	}
	return video
}
