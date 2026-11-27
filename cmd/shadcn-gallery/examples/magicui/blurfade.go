package magicui

import (
	"fmt"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// blurFadeImages are upstream's nine stock photos, landscape and portrait
// alternating.
func blurFadeImages() []string {
	images := make([]string, 9)
	for i := range images {
		width, height := 600, 800
		if i%2 == 0 {
			width, height = 800, 600
		}
		images[i] = fmt.Sprintf("https://picsum.photos/seed/%d/%d/%d", i+1, width, height)
	}
	return images
}

// blurFadeGallery lays the photos out in columns, each fading in with the
// given options and a delay growing with its index.
func blurFadeGallery(id string, opts magicui.BlurFadeOptions) mx.Component {
	images := blurFadeImages()
	photos := make([]any, 0, len(images)+1)
	photos = append(photos, html.Class("columns-2 gap-4 sm:columns-3"))
	for idx, imageURL := range images {
		opts.Delay = 250*time.Millisecond + time.Duration(idx)*50*time.Millisecond
		photos = append(photos, magicui.BlurFade(opts,
			html.ImgSrc(imageURL, html.Class("mb-4 size-full rounded-lg object-contain"), html.Alt(fmt.Sprintf("Random stock image %d", idx+1))),
		))
	}
	return html.Section(html.ID(id), html.Div(photos...))
}

// BlurFadeDemo fades a photo gallery in while it scrolls into view.
func BlurFadeDemo() mx.Component {
	return blurFadeGallery("photos", magicui.BlurFadeOptions{InView: true})
}

// BlurFadeOnLoad fades the photo gallery in one photo after another when
// the page loads.
func BlurFadeOnLoad() mx.Component {
	return blurFadeGallery("photos-on-load", magicui.BlurFadeOptions{})
}
