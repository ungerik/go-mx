package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/svg"
)

// The iPhone frame's geometry in SVG user units, from upstream.
const (
	iphoneWidth        = 433
	iphoneHeight       = 882
	iphoneScreenX      = 21.25
	iphoneScreenY      = 19.25
	iphoneScreenWidth  = 389.5
	iphoneScreenHeight = 843.5
	iphoneScreenRadius = 55.75
)

// IphoneOptions configures an [Iphone]. The zero value is an empty frame.
type IphoneOptions struct {
	// Src is an image shown on the screen, cropped to fill it from the top.
	Src string
	// VideoSrc is a video played muted in a loop on the screen; it takes
	// precedence over Src.
	VideoSrc string
	// ID is the id of the SVG mask that cuts the screen out of the frame
	// when there is media to show through it. It must be unique in the
	// document; empty derives one with [mx.UniqueID]. Upstream uses the
	// fixed id "screenPunch". It may only contain letters, digits, '_' and
	// '-'.
	ID string
}

// Iphone renders a mockup of an iPhone as an SVG frame scaled to the width
// of its container, with an optional image or video on the screen. The SVG
// paths are upstream's, the frame is aria-hidden, and the screen image has
// an empty alt text as upstream; pass an aria-label or a visually hidden
// child to describe the mockup. Children are placed after the frame.
//
// The mask is only rendered when there is media, and upstream's unused
// roundedCorners clip path is left out.
func Iphone(opts IphoneOptions, attribsChildren ...any) *mx.Element {
	root := html.Div(attribsChildren...)
	children := root.Children
	root.Children = nil
	hasMedia := opts.VideoSrc != "" || opts.Src != ""
	screenStyle := "left:" + devicePercent(iphoneScreenX, iphoneWidth) +
		";top:" + devicePercent(iphoneScreenY, iphoneHeight) +
		";width:" + devicePercent(iphoneScreenWidth, iphoneWidth) +
		";height:" + devicePercent(iphoneScreenHeight, iphoneHeight) +
		";border-radius:" + devicePercent(iphoneScreenRadius, iphoneScreenWidth) + " / " + devicePercent(iphoneScreenRadius, iphoneScreenHeight)
	switch {
	case opts.VideoSrc != "":
		root.Children = append(root.Children, html.Div(
			html.Class("pointer-events-none absolute z-0 overflow-hidden"),
			html.Style(screenStyle),
			deviceVideo(opts.VideoSrc, "block size-full object-cover", true),
		))
	case opts.Src != "":
		root.Children = append(root.Children, html.Div(
			html.Class("pointer-events-none absolute z-0 overflow-hidden"),
			html.Style(screenStyle),
			html.Img(html.Src(opts.Src), html.Alt(""), html.Class("block size-full object-cover object-top")),
		))
	}

	const frame = "fill-[#E5E5E5] dark:fill-[#404040]"
	var mask, defs []any
	if hasMedia {
		id, err := svgRefID(opts.ID)
		if err != nil {
			return mx.NewErrElement(err)
		}
		mask = []any{svg.MaskAttr("url(#" + id + ")")}
		defs = []any{svg.Defs(svg.Mask(
			svg.ID(id),
			svg.MaskUnitsUserSpaceOnUse,
			svg.Rect(svg.X(0), svg.Y(0), svg.Width(iphoneWidth), svg.Height(iphoneHeight), svg.Fill("white")),
			svg.Rect(
				svg.X(iphoneScreenX), svg.Y(iphoneScreenY),
				svg.Width(iphoneScreenWidth), svg.Height(iphoneScreenHeight),
				svg.RX(iphoneScreenRadius), svg.RY(iphoneScreenRadius),
				svg.Fill("black"),
			),
		))}
	}
	body := svg.G(append(mask,
		svg.Path(svg.D("M2 73C2 32.6832 34.6832 0 75 0H357C397.317 0 430 32.6832 430 73V809C430 849.317 397.317 882 357 882H75C34.6832 882 2 849.317 2 809V73Z"), svg.Class(frame)),
		svg.Path(svg.D("M0 171C0 170.448 0.447715 170 1 170H3V204H1C0.447715 204 0 203.552 0 203V171Z"), svg.Class(frame)),
		svg.Path(svg.D("M1 234C1 233.448 1.44772 233 2 233H3.5V300H2C1.44772 300 1 299.552 1 299V234Z"), svg.Class(frame)),
		svg.Path(svg.D("M1 319C1 318.448 1.44772 318 2 318H3.5V385H2C1.44772 385 1 384.552 1 384V319Z"), svg.Class(frame)),
		svg.Path(svg.D("M430 279H432C432.552 279 433 279.448 433 280V384C433 384.552 432.552 385 432 385H430V279Z"), svg.Class(frame)),
		svg.Path(svg.D("M6 74C6 35.3401 37.3401 4 76 4H356C394.66 4 426 35.3401 426 74V808C426 846.66 394.66 878 356 878H76C37.3401 878 6 846.66 6 808V74Z"), svg.Class("fill-white dark:fill-[#262626]")),
	)...)
	screen := svg.Path(append([]any{
		svg.D("M21.25 75C21.25 44.2101 46.2101 19.25 77 19.25H355C385.79 19.25 410.75 44.2101 410.75 75V807C410.75 837.79 385.79 862.75 355 862.75H77C46.2101 862.75 21.25 837.79 21.25 807V75Z"),
		svg.Class("fill-[#E5E5E5] stroke-[#E5E5E5] stroke-[0.5] dark:fill-[#404040] dark:stroke-[#404040]"),
	}, mask...)...)
	root.Children = append(root.Children, svg.SVG(append([]any{
		svg.ViewBox(0, 0, iphoneWidth, iphoneHeight),
		svg.Fill("none"),
		svg.XMLNS,
		svg.Class("absolute inset-0 size-full"),
		svg.Style("transform:translateZ(0)"),
		ariaHidden,
		body,
		svg.Path(svg.Opacity(0.5), svg.D("M174 5H258V5.5C258 6.60457 257.105 7.5 256 7.5H176C174.895 7.5 174 6.60457 174 5.5V5Z"), svg.Class(frame)),
		screen,
		svg.Path(svg.D("M154 48.5C154 38.2827 162.283 30 172.5 30H259.5C269.717 30 278 38.2827 278 48.5C278 58.7173 269.717 67 259.5 67H172.5C162.283 67 154 58.7173 154 48.5Z"), svg.Class("fill-[#F5F5F5] dark:fill-[#262626]")),
		svg.Path(svg.D("M249 48.5C249 42.701 253.701 38 259.5 38C265.299 38 270 42.701 270 48.5C270 54.299 265.299 59 259.5 59C253.701 59 249 54.299 249 48.5Z"), svg.Class("fill-[#F5F5F5] dark:fill-[#262626]")),
		svg.Path(svg.D("M254 48.5C254 45.4624 256.462 43 259.5 43C262.538 43 265 45.4624 265 48.5C265 51.5376 262.538 54 259.5 54C256.462 54 254 51.5376 254 48.5Z"), svg.Class(frame)),
	}, defs...)...))
	root.Children = append(root.Children, children...)
	return finish(root, "iphone", "relative inline-block w-full align-middle leading-none",
		"aspect-ratio:"+num(iphoneWidth)+"/"+num(iphoneHeight))
}
