package shadcn

import (
	"strconv"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// sliderInputBase is the styling shared by single- and two-thumb slider inputs.
// It removes the native chrome (appearance-none) and rebuilds shadcn's track +
// thumb look via arbitrary ::-webkit-* / ::-moz-* pseudo-element utilities.
// The thumb's negative top margin centers a size-4 (16px) thumb on a h-1.5
// (6px) track: (16-6)/2 = 5px = -mt-1.25.
const sliderInputBase = "appearance-none bg-transparent outline-none disabled:pointer-events-none disabled:opacity-50 focus-visible:outline-hidden" +
	" [&::-webkit-slider-runnable-track]:bg-muted [&::-webkit-slider-runnable-track]:h-1.5 [&::-webkit-slider-runnable-track]:rounded-full" +
	" [&::-moz-range-track]:bg-muted [&::-moz-range-track]:h-1.5 [&::-moz-range-track]:rounded-full" +
	" [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:size-4 [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-background [&::-webkit-slider-thumb]:border [&::-webkit-slider-thumb]:border-primary [&::-webkit-slider-thumb]:shadow-sm [&::-webkit-slider-thumb]:-mt-1.25" +
	" [&::-moz-range-thumb]:appearance-none [&::-moz-range-thumb]:size-4 [&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:bg-background [&::-moz-range-thumb]:border [&::-moz-range-thumb]:border-primary [&::-moz-range-thumb]:shadow-sm"

// sliderClampScript is the once-emitted client function used by the two-thumb
// range mode. It re-derives the low/high values from the two inputs (so either
// thumb can move past the other and they swap roles), then positions the
// range-fill <div> by left% and width%. The script is a no-op for slider
// instances with fewer than 2 inputs.
const sliderClampScript = /*js*/ `if(!window.sliderClamp){window.sliderClamp=function(id){var r=document.querySelector('[data-slider="'+id+'"]');if(!r)return;var i=r.querySelectorAll('input[type=range]');if(i.length<2)return;var lo=Math.min(+i[0].value,+i[1].value);var hi=Math.max(+i[0].value,+i[1].value);var mn=+i[0].min,mx=+i[0].max,rg=mx-mn;if(rg<=0)return;var f=r.querySelector('[data-slot="slider-range"]');if(f){f.style.left=((lo-mn)/rg*100)+'%';f.style.width=((hi-lo)/rg*100)+'%';}};}`

// Slider renders a shadcn/ui slider. With len(values)==1 it ships a single-thumb
// native <input type="range">. With len(values)==2 it ships a two-thumb range:
// two <input>s overlaid on a shared track + fill, with the shared sliderClamp
// script kept the fill in sync as either thumb moves. Any other len panics.
//
// id is a stable identifier (validated) used as the data-slider attribute that
// scopes the script to this instance. min/max/step go on each native input.
func Slider(min, max, step float64, values []float64, id string, attribsChildren ...any) *mx.Element {
	if err := validateID(id); err != nil {
		return mx.NewErrElement(err)
	}
	switch len(values) {
	case 1:
		return sliderSingle(min, max, step, values[0], attribsChildren...)
	case 2:
		return sliderRange(min, max, step, values[0], values[1], id, attribsChildren...)
	default:
		panic("shadcn: Slider values must have length 1 or 2")
	}
}

func sliderSingle(min, max, step, value float64, attribsChildren ...any) *mx.Element {
	e := html.Element("input", attribsChildren...)
	e.Children = nil // <input> is a void element
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("range"))
	}
	if e.AttribIndex("min") < 0 {
		e.Attribs = append(e.Attribs, html.Min(min))
	}
	if e.AttribIndex("max") < 0 {
		e.Attribs = append(e.Attribs, html.Max(max))
	}
	if e.AttribIndex("step") < 0 {
		e.Attribs = append(e.Attribs, html.Step(step))
	}
	if e.AttribIndex("value") < 0 {
		e.Attribs = append(e.Attribs, html.Value(value))
	}
	return finish(e, "slider", "w-full "+sliderInputBase)
}

func sliderRange(min, max, step, lo, hi float64, id string, attribsChildren ...any) *mx.Element {
	rangeSize := max - min
	leftPct, widthPct := "0%", "100%"
	if rangeSize > 0 {
		leftPct = fmtFloat((lo-min)/rangeSize*100) + "%"
		widthPct = fmtFloat((hi-lo)/rangeSize*100) + "%"
	}

	track := finish(html.Div(
		finish(html.Div(html.Style("left: "+leftPct+"; width: "+widthPct)),
			"slider-range", "bg-primary absolute h-full"),
	), "slider-track", "bg-muted relative h-1.5 w-full rounded-full overflow-hidden")

	// Two overlaid inputs. pointer-events-none on the input lets clicks pass
	// to the underlying track, but pointer-events-auto on the thumb pseudo-
	// element keeps each thumb independently draggable.
	mkInput := func(val float64) *mx.Element {
		in := html.Element("input",
			html.Type("range"),
			html.Min(min),
			html.Max(max),
			html.Step(step),
			html.Value(val),
			html.OnInput("sliderClamp('"+id+"')"),
			html.Class("absolute inset-0 w-full pointer-events-none [&::-webkit-slider-thumb]:pointer-events-auto [&::-moz-range-thumb]:pointer-events-auto "+sliderInputBase),
		)
		in.Children = nil
		return in
	}

	e := html.Div(attribsChildren...)
	if e.AttribIndex("data-slider") < 0 {
		e.Attribs = append(e.Attribs, html.DataAttr("slider", id))
	}
	e.Children = append(e.Children,
		track,
		mkInput(lo),
		mkInput(hi),
		html.ScriptJS(sliderClampScript),
	)
	return finish(e, "slider", "relative flex w-full touch-none items-center select-none h-4")
}

func fmtFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
