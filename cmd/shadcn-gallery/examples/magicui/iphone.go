package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// IphoneDemo shows an image on the iPhone's screen.
func IphoneDemo() mx.Component {
	return html.DivClass("w-[434px]",
		magicui.Iphone(magicui.IphoneOptions{ID: "iphone-demo-screen", Src: "https://placehold.co/900x1600?text=Hello+World"}),
	)
}

// IphoneVideo plays a muted video loop on the iPhone's screen.
func IphoneVideo() mx.Component {
	return html.DivClass("w-[434px]",
		magicui.Iphone(magicui.IphoneOptions{ID: "iphone-video-screen", VideoSrc: "https://videos.pexels.com/video-files/8946986/8946986-uhd_1440_2732_25fps.mp4"}),
	)
}

// IphoneEmpty shows the bare iPhone frame.
func IphoneEmpty() mx.Component {
	return html.DivClass("w-[434px]", magicui.Iphone(magicui.IphoneOptions{}))
}
