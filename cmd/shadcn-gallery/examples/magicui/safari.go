package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// SafariDemo shows an image as the page in a Safari window.
func SafariDemo() mx.Component {
	return html.DivClass("w-[1203px]",
		magicui.Safari(magicui.SafariOptions{ID: "safari-demo-page", URL: "magicui.design", ImageSrc: "https://placehold.co/1200x750?text=Hello+World"}),
	)
}

// SafariVideo plays a muted video loop as the page in a Safari window.
func SafariVideo() mx.Component {
	return html.DivClass("w-[1203px]",
		magicui.Safari(magicui.SafariOptions{ID: "safari-video-page", URL: "magicui.design", VideoSrc: "https://videos.pexels.com/video-files/27180348/12091515_2560_1440_50fps.mp4"}),
	)
}

// SafariSimple shows an empty Safari window without the toolbar buttons.
func SafariSimple() mx.Component {
	return html.DivClass("w-[1203px]",
		magicui.Safari(magicui.SafariOptions{URL: "magicui.design", Mode: magicui.SafariModeSimple}),
	)
}
