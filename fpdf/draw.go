package fpdf

import (
	"bytes"
	"io"

	"codeberg.org/go-pdf/fpdf"
)

// Line strokes a straight line from (x1, y1) to (x2, y2) using the current draw
// color and line width.
func Line(x1, y1, x2, y2 float64) Component {
	return drawing(func(r *Renderer) {
		r.Line(x1, y1, x2, y2)
	})
}

// Rect paints a rectangle at (x, y) of width w and height h with the given
// paint operation (Stroke, FillShape or FillStroke).
func Rect(x, y, w, h float64, op DrawOp) Component {
	return drawing(func(r *Renderer) {
		r.Rect(x, y, w, h, string(op))
	})
}

// RoundedRect paints a rectangle with all four corners rounded to radius. Use
// the raw fpdf RoundedRectExt for per-corner radii.
func RoundedRect(x, y, w, h, radius float64, op DrawOp) Component {
	return drawing(func(r *Renderer) {
		r.RoundedRect(x, y, w, h, radius, "1234", string(op))
	})
}

// Circle paints a circle of radius rad centered at (x, y).
func Circle(x, y, rad float64, op DrawOp) Component {
	return drawing(func(r *Renderer) {
		r.Circle(x, y, rad, string(op))
	})
}

// Ellipse paints an ellipse centered at (x, y) with horizontal radius rx and
// vertical radius ry, rotated degRotate degrees counter-clockwise.
func Ellipse(x, y, rx, ry, degRotate float64, op DrawOp) Component {
	return drawing(func(r *Renderer) {
		r.Ellipse(x, y, rx, ry, degRotate, string(op))
	})
}

// Polygon paints a closed polygon through the given points (at least three).
func Polygon(op DrawOp, points ...Point) Component {
	return drawing(func(r *Renderer) {
		r.Polygon(points, string(op))
	})
}

// Image draws the image file scaled into the box at (x, y) of width w and
// height h. A zero w or h preserves the image's aspect ratio from the other
// dimension; both zero uses the image's natural size at 72 dpi. The format is
// inferred from the file extension. To draw an image held in memory, use
// [ImageReader] or [ImageBytes].
func Image(file string, x, y, w, h float64) Component {
	return drawing(func(r *Renderer) {
		r.Image(file, x, y, w, h, false, "", 0, "")
	})
}

// ImageReader draws an image read from src into the box at (x, y) of width w and
// height h, without touching the filesystem. Sizing follows [Image].
//
// Because the source has no filename, name is used as fpdf's cache key: draw the
// same image again by passing the same name (the bytes are decoded only once),
// and give distinct images distinct names. imageType gives the encoding.
func ImageReader(name string, imageType ImageType, src io.Reader, x, y, w, h float64) Component {
	return drawing(func(r *Renderer) {
		options := fpdf.ImageOptions{ImageType: string(imageType)}
		r.RegisterImageOptionsReader(name, options, src)
		r.ImageOptions(name, x, y, w, h, false, options, 0, "")
	})
}

// ImageBytes is [ImageReader] for an in-memory byte slice.
func ImageBytes(name string, imageType ImageType, data []byte, x, y, w, h float64) Component {
	return ImageReader(name, imageType, bytes.NewReader(data), x, y, w, h)
}
