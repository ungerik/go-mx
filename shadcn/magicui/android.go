package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/svg"
)

// AndroidOptions configures an [Android]. The zero value is upstream's
// default: an empty 433×882 frame.
type AndroidOptions struct {
	// Width is the SVG's width attribute and viewBox width; 0 means 433.
	Width float64
	// Height is the SVG's height attribute and viewBox height; 0 means 882.
	Height float64
	// Src is an image shown on the screen, scaled to cover it.
	Src string
	// VideoSrc is a video played muted in a loop on the screen. Unlike the
	// other mockups, Android shows an image and a video both if both are
	// set, the video on top, as upstream.
	VideoSrc string
	// ID is the id of the clip path in the shape of the screen. It must be
	// unique in the document; empty derives one with [mx.UniqueID]. Upstream
	// uses the fixed id "clip0_514_20855". It may only contain letters,
	// digits, '_' and '-'.
	ID string
}

// Android renders a mockup of an Android phone as an <svg> with an optional
// image or video on the screen, the video in a <foreignObject>. The caller's
// attributes go to the <svg>, as upstream spreads its props there, so a class
// such as html.Class("size-full") scales it to its container. As upstream,
// Width and Height change the viewBox along with the size, so they crop or
// extend the drawing rather than scaling it. The SVG paths are upstream's.
// Children are appended inside the <svg>, after the screen media.
func Android(opts AndroidOptions, attribsChildren ...any) *mx.Element {
	width, height := or(opts.Width, 433), or(opts.Height, 882)
	id, err := svgRefID(opts.ID)
	if err != nil {
		return mx.NewErrElement(err)
	}
	clip := svg.ClipPathAttr("url(#" + id + ")")
	const frame = "fill-[#E5E5E5] dark:fill-[#404040]"
	root := svg.SVG(attribsChildren...)
	children := root.Children
	// The caller's attributes win, as upstream spreads its props last.
	var own []mx.Attrib
	for _, a := range []mx.Attrib{
		svg.Width(width),
		svg.Height(height),
		svg.ViewBox(0, 0, width, height),
		svg.Fill("none"),
		svg.XMLNS,
	} {
		if root.AttribIndex(a.AttribName()) < 0 {
			own = append(own, a)
		}
	}
	root.Attribs = append(own, root.Attribs...)
	root.Children = []mx.Component{
		svg.Path(svg.D("M376 153H378C379.105 153 380 153.895 380 155V249C380 250.105 379.105 251 378 251H376V153Z"), svg.Class(frame)),
		svg.Path(svg.D("M376 301H378C379.105 301 380 301.895 380 303V351C380 352.105 379.105 353 378 353H376V301Z"), svg.Class(frame)),
		svg.Path(svg.D("M0 42C0 18.8041 18.804 0 42 0H336C359.196 0 378 18.804 378 42V788C378 811.196 359.196 830 336 830H42C18.804 830 0 811.196 0 788V42Z"), svg.Class(frame)),
		svg.Path(svg.D("M2 43C2 22.0132 19.0132 5 40 5H338C358.987 5 376 22.0132 376 43V787C376 807.987 358.987 825 338 825H40C19.0132 825 2 807.987 2 787V43Z"), svg.Class("fill-white dark:fill-[#262626]")),
		svg.G(clip, svg.Path(
			svg.D("M9.25 48C9.25 29.3604 24.3604 14.25 43 14.25H335C353.64 14.25 368.75 29.3604 368.75 48V780C368.75 798.64 353.64 813.75 335 813.75H43C24.3604 813.75 9.25 798.64 9.25 780V48Z"),
			svg.Class("fill-[#E5E5E5] stroke-[#E5E5E5] stroke-[0.5] dark:fill-[#404040] dark:stroke-[#404040]"),
		)),
		svg.Circle(svg.CX(189), svg.CY(28), svg.R(9), svg.Class("fill-white dark:fill-[#262626]")),
		svg.Circle(svg.CX(189), svg.CY(28), svg.R(4), svg.Class(frame)),
	}
	if opts.Src != "" {
		root.Children = append(root.Children, svg.Image(
			svg.Href(opts.Src),
			svg.Width(360),
			svg.Height(800),
			svg.Class("size-full object-cover"),
			svg.PreserveAspectRatio("xMidYMid slice"),
			clip,
		))
	}
	if opts.VideoSrc != "" {
		root.Children = append(root.Children, svg.ForeignObject(
			svg.Width(380),
			svg.Height(820),
			clip,
			svg.PreserveAspectRatio("xMidYMid slice"),
			deviceVideo(opts.VideoSrc, "size-full object-cover", false),
		))
	}
	root.Children = append(root.Children,
		svg.Defs(svg.ClipPath(
			svg.ID(id),
			svg.Rect(
				svg.Width(360),
				svg.Height(800),
				svg.RX(33),
				svg.RY(25),
				svg.Class("fill-white dark:fill-[#262626]"),
				svg.Transform("translate(9 14)"),
			),
		)),
	)
	root.Children = append(root.Children, children...)
	return finish(root, "android", "", "")
}
