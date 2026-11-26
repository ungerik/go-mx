package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// AndroidDemo shows an image on the Android phone's screen.
func AndroidDemo() mx.Component {
	return html.DivClass("relative",
		magicui.Android(magicui.AndroidOptions{
			ID:  "android-demo-screen",
			Src: "https://images.unsplash.com/photo-1730326405863-c6fa7e499a6e?q=80&w=1964&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D",
		}, html.Class("size-full")),
	)
}

// AndroidVideo plays a muted video loop on the Android phone's screen.
func AndroidVideo() mx.Component {
	return html.DivClass("relative",
		magicui.Android(magicui.AndroidOptions{
			ID:       "android-video-screen",
			VideoSrc: "https://videos.pexels.com/video-files/14993748/14993748-uhd_1296_2304_30fps.mp4",
		}, html.Class("size-full")),
	)
}

// AndroidEmpty shows the bare Android frame.
func AndroidEmpty() mx.Component {
	return html.DivClass("relative",
		magicui.Android(magicui.AndroidOptions{ID: "android-empty-screen"}, html.Class("size-full")),
	)
}
