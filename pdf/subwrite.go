// Copyright ©2023 The go-pdf Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pdf

// Adapted from http://www.fpdf.org/en/script/script61.php by Wirus and released with the FPDF license.

// SubWrite prints text from the current position in the same way as Write().
// ht is the line height in the unit of measure specified in New(). str
// specifies the text to write. subFontSize is the size of the font in points.
// subOffset is the vertical offset of the text in points; a positive value
// indicates a superscript, a negative value indicates a subscript. link is the
// identifier returned by AddLink() or 0 for no internal link. linkStr is a
// target URL or empty for no external link. A non--zero value for link takes
// precedence over linkStr.
//
// The SubWrite example demonstrates this method.
func (r *Renderer) SubWrite(ht float64, str string, subFontSize, subOffset float64, link int, linkStr string) {
	if r.err != nil {
		return
	}
	// resize font
	subFontSizeOld := r.fontSizePt
	r.SetFontSize(subFontSize)
	// reposition y
	subOffset = (((subFontSize - subFontSizeOld) / r.k) * 0.3) + (subOffset / r.k)
	subX := r.x
	subY := r.y
	r.SetXY(subX, subY-subOffset)
	//Output text
	r.write(ht, str, link, linkStr)
	// restore y position
	subX = r.x
	subY = r.y
	r.SetXY(subX, subY+subOffset)
	// restore font size
	r.SetFontSize(subFontSizeOld)
}
