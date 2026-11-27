package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// heroVideoDialogPair renders upstream's light and dark thumbnail variants
// of one HeroVideoDialog, each shown in its color scheme.
func heroVideoDialogPair(id string, style magicui.HeroVideoAnimationStyle) mx.Component {
	const video = "https://www.youtube.com/embed/qh3NGpYRG3I?si=4rb-zSdDkVK9qxxb"
	return html.DivClass("relative",
		magicui.HeroVideoDialog(magicui.HeroVideoDialogOptions{
			ID:             id + "-light",
			AnimationStyle: style,
			VideoSrc:       video,
			ThumbnailSrc:   "https://startup-template-sage.vercel.app/hero-light.png",
			ThumbnailAlt:   "Hero Video",
		}, html.Class("block dark:hidden")),
		magicui.HeroVideoDialog(magicui.HeroVideoDialogOptions{
			ID:             id + "-dark",
			AnimationStyle: style,
			VideoSrc:       video,
			ThumbnailSrc:   "https://startup-template-sage.vercel.app/hero-dark.png",
			ThumbnailAlt:   "Hero Video",
		}, html.Class("hidden dark:block")),
	)
}

// HeroVideoDialogDemo opens the video growing from the center.
func HeroVideoDialogDemo() mx.Component {
	return heroVideoDialogPair("hero-video-demo", magicui.HeroVideoFromCenter)
}

// HeroVideoDialogTopInBottomOut slides the video in from the top and out at
// the bottom.
func HeroVideoDialogTopInBottomOut() mx.Component {
	return heroVideoDialogPair("hero-video-top-in-bottom-out", magicui.HeroVideoTopInBottomOut)
}
