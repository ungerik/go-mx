package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Helpers shared by the device mockups [Iphone], [Safari] and [Android].

// devicePercent formats part/whole as a CSS percentage, computed in float64
// steps like upstream's JavaScript so the digits match.
func devicePercent(part, whole float64) string {
	return num(part/whole*100) + "%"
}

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
