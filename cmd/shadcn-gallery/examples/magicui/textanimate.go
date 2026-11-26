package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// TextAnimateDemo blurs a text in character by character while it rises.
func TextAnimateDemo() mx.Component {
	return magicui.TextAnimate(magicui.TextAnimateOptions{
		Animation: magicui.TextAnimateBlurInUp,
		By:        magicui.TextAnimateByCharacter,
	}, "Blur in by character")
}

// TextAnimateBlurInHeading blurs an h1 in as a whole.
func TextAnimateBlurInHeading() mx.Component {
	return magicui.TextAnimate(magicui.TextAnimateOptions{
		Animation: magicui.TextAnimateBlurIn,
		As:        magicui.TextAnimateH1,
	}, "Blur in text")
}

// TextAnimateSlideUpByWord slides a text up word by word.
func TextAnimateSlideUpByWord() mx.Component {
	return magicui.TextAnimate(magicui.TextAnimateOptions{
		Animation: magicui.TextAnimateSlideUp,
		By:        magicui.TextAnimateByWord,
	}, "Slide up by word")
}
