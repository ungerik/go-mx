// Copyright ©2023 The go-pdf Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright (c) Kurt Jung (Gmail: kurt.w.jung)
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY
// SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION
// OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF OR IN
// CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

// Adapted from http://www.fpdf.org/en/script/script89.php by Olivier PLATHEY

package pdf

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/domonda/go-errs"
)

func byteBound(v byte) byte {
	if v > 100 {
		return 100
	}
	return v
}

// AddSpotColor adds an ink-based CMYK color to the gofpdf instance and
// associates it with the specified name. The individual components specify
// percentages ranging from 0 to 100. Values above this are quietly capped to
// 100. An error occurs if the specified name is already associated with a
// color.
func (r *Renderer) AddSpotColor(nameStr string, c, m, y, k byte) {
	if r.err == nil {
		_, ok := r.spotColorMap[nameStr]
		if !ok {
			id := len(r.spotColorMap) + 1
			r.spotColorMap[nameStr] = spotColorType{
				id: id,
				val: cmykColorType{
					c: byteBound(c),
					m: byteBound(m),
					y: byteBound(y),
					k: byteBound(k),
				},
			}
		} else {
			r.err = errs.Errorf("name \"%s\" is already associated with a spot color", nameStr)
		}
	}
}

func (r *Renderer) getSpotColor(nameStr string) (clr spotColorType, ok bool) {
	if r.err == nil {
		clr, ok = r.spotColorMap[nameStr]
		if !ok {
			r.err = errs.Errorf("spot color name \"%s\" is not registered", nameStr)
		}
	}
	return clr, ok
}

// SetDrawSpotColor sets the current draw color to the spot color associated
// with nameStr. An error occurs if the name is not associated with a color.
// The value for tint ranges from 0 (no intensity) to 100 (full intensity). It
// is quietly bounded to this range.
func (r *Renderer) SetDrawSpotColor(nameStr string, tint byte) {
	clr, ok := r.getSpotColor(nameStr)
	if !ok {
		return
	}
	r.color.draw.mode = colorModeSpot
	r.color.draw.spotStr = nameStr
	r.color.draw.str = fmt.Sprintf("/CS%d CS %.3f SCN", clr.id, float64(byteBound(tint))/100)
	if r.page > 0 {
		r.out(r.color.draw.str)
	}
}

// SetFillSpotColor sets the current fill color to the spot color associated
// with nameStr. An error occurs if the name is not associated with a color.
// The value for tint ranges from 0 (no intensity) to 100 (full intensity). It
// is quietly bounded to this range.
func (r *Renderer) SetFillSpotColor(nameStr string, tint byte) {
	clr, ok := r.getSpotColor(nameStr)
	if !ok {
		return
	}
	r.color.fill.mode = colorModeSpot
	r.color.fill.spotStr = nameStr
	r.color.fill.str = fmt.Sprintf("/CS%d cs %.3f scn", clr.id, float64(byteBound(tint))/100)
	r.colorFlag = r.color.fill.str != r.color.text.str
	if r.page > 0 {
		r.out(r.color.fill.str)
	}
}

// SetTextSpotColor sets the current text color to the spot color associated
// with nameStr. An error occurs if the name is not associated with a color.
// The value for tint ranges from 0 (no intensity) to 100 (full intensity). It
// is quietly bounded to this range.
func (r *Renderer) SetTextSpotColor(nameStr string, tint byte) {
	clr, ok := r.getSpotColor(nameStr)
	if !ok {
		return
	}
	r.color.text.mode = colorModeSpot
	r.color.text.spotStr = nameStr
	r.color.text.str = fmt.Sprintf("/CS%d cs %.3f scn", clr.id, float64(byteBound(tint))/100)
	r.colorFlag = r.color.fill.str != r.color.text.str
}

func (r *Renderer) returnSpotColor(clr colorType) (name string, c, m, y, k byte) {
	name = clr.spotStr
	if name != "" {
		spotClr, ok := r.getSpotColor(name)
		if ok {
			c = spotClr.val.c
			m = spotClr.val.m
			y = spotClr.val.y
			k = spotClr.val.k
		}
	}
	return name, c, m, y, k
}

// GetDrawSpotColor returns the most recently used spot color information for
// drawing. This will not be the current drawing color if some other color type
// such as RGB is active. If no spot color has been set for drawing, zero
// values are returned.
func (r *Renderer) GetDrawSpotColor() (name string, c, m, y, k byte) {
	return r.returnSpotColor(r.color.draw)
}

// GetTextSpotColor returns the most recently used spot color information for
// text output. This will not be the current text color if some other color
// type such as RGB is active. If no spot color has been set for text, zero
// values are returned.
func (r *Renderer) GetTextSpotColor() (name string, c, m, y, k byte) {
	return r.returnSpotColor(r.color.text)
}

// GetFillSpotColor returns the most recently used spot color information for
// fill output. This will not be the current fill color if some other color
// type such as RGB is active. If no fill spot color has been set, zero values
// are returned.
func (r *Renderer) GetFillSpotColor() (name string, c, m, y, k byte) {
	return r.returnSpotColor(r.color.fill)
}

// spotColorNames returns the spot color names ordered by registration id, so
// that object numbering and resource dictionaries do not depend on the map
// iteration order.
func (r *Renderer) spotColorNames() []string {
	return slices.SortedFunc(maps.Keys(r.spotColorMap), func(a, b string) int {
		return r.spotColorMap[a].id - r.spotColorMap[b].id
	})
}

func (r *Renderer) putSpotColors() {
	for _, k := range r.spotColorNames() {
		v := r.spotColorMap[k]
		r.newobj()
		r.outf("[/Separation /%s", strings.ReplaceAll(k, " ", "#20"))
		r.out("/DeviceCMYK <<")
		r.out("/Range [0 1 0 1 0 1 0 1] /C0 [0 0 0 0] ")
		r.outf("/C1 [%.3f %.3f %.3f %.3f] ", float64(v.val.c)/100, float64(v.val.m)/100,
			float64(v.val.y)/100, float64(v.val.k)/100)
		r.out("/FunctionType 2 /Domain [0 1] /N 1>>]")
		r.out("endobj")
		v.objID = r.n
		r.spotColorMap[k] = v
	}
}

func (r *Renderer) spotColorPutResourceDict() {
	r.out("/ColorSpace <<")
	for _, k := range r.spotColorNames() {
		clr := r.spotColorMap[k]
		r.outf("/CS%d %d 0 R", clr.id, clr.objID)
	}
	r.out(">>")
}

// spotColorType specifies a named spot color value.
type spotColorType struct {
	id, objID int
	val       cmykColorType
}

// cmykColorType specifies an ink-based CMYK color value.
type cmykColorType struct {
	c, m, y, k byte // 0% to 100%
}
