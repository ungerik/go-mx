package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/svg"
)

// BacklightOptions configures a [Backlight]. The zero value is upstream's
// default.
type BacklightOptions struct {
	// ID is the id of the SVG <filter>, which the wrapper of the children
	// references as url(#ID). Empty means a process-unique id from
	// [mx.UniqueID]; pass a fixed one where the markup has to be
	// deterministic. It may only contain letters, digits, '_' and '-'.
	ID string
	// Blur is the standard deviation of the glow's Gaussian blur in pixels;
	// 0 means 20.
	Blur float64
}

// Backlight makes its children glow in their own colors, like an ambient
// light behind a screen: an SVG filter draws a blurred, four times
// saturated copy of them behind the sharp original. It works on images,
// videos, iframes and SVG alike. The children are wrapped in a <div> that
// applies the filter; the root <div> takes the caller's attributes.
func Backlight(opts BacklightOptions, attribsChildren ...any) *mx.Element {
	id, err := svgRefID(opts.ID)
	if err != nil {
		return mx.NewErrElement(err)
	}
	root := html.Div(attribsChildren...)
	children := root.Children
	root.Children = []mx.Component{
		svg.SVG(svg.Width(0), svg.Height(0), ariaHidden,
			svg.Filter(svg.ID(id), svg.Y("-50%"), svg.X("-50%"), svg.Width("200%"), svg.Height("200%"),
				svg.FeGaussianBlur(svg.In("SourceGraphic"), svg.StdDeviation(or(opts.Blur, 20)), svg.Result("blurred")),
				svg.FeColorMatrix(svg.Type("saturate"), svg.In("blurred"), svg.Values("4")),
				svg.FeComposite(svg.In("SourceGraphic"), svg.Operator("over")),
			),
		),
		html.Div(append([]any{html.Style("filter:url(#" + id + ")")}, anys(children)...)...),
	}
	return finish(root, "backlight", "", "")
}
