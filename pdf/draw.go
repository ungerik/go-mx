package pdf

import (
	"bytes"
	"io"
	"sync"
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
// Because the source has no filename, name is used as the renderer's cache key:
// draw the same image again by passing the same name (the bytes are decoded only
// once), and give distinct images distinct names. imageType gives the encoding.
//
// src is read fully on the first render and buffered in the component, so the
// component can be rendered any number of times, into any number of renderers
// — including concurrently into separate renderers (the one-time read is
// synchronized). A read error is latched: it is recorded on the renderer and
// surfaces from every render, since the partially-drained src cannot be
// re-read.
func ImageReader(name string, imageType ImageType, src io.Reader, x, y, w, h float64) Component {
	readSrc := sync.OnceValues(func() ([]byte, error) {
		return io.ReadAll(src)
	})
	return drawing(func(r *Renderer) {
		data, err := readSrc()
		if err != nil {
			r.SetError(err)
			return
		}
		options := ImageOptions{ImageType: string(imageType)}
		r.RegisterImageOptionsReader(name, options, bytes.NewReader(data))
		r.ImageOptions(name, x, y, w, h, false, options, 0, "")
	})
}

// ImageBytes is [ImageReader] for an in-memory byte slice: the component can
// be rendered any number of times, into any number of renderers.
func ImageBytes(name string, imageType ImageType, data []byte, x, y, w, h float64) Component {
	return drawing(func(r *Renderer) {
		options := ImageOptions{ImageType: string(imageType)}
		r.RegisterImageOptionsReader(name, options, bytes.NewReader(data))
		r.ImageOptions(name, x, y, w, h, false, options, 0, "")
	})
}
