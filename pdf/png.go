// Copyright ©2023 The go-pdf Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
 * Copyright (c) 2013-2016 Kurt Jung (Gmail: kurt.w.jung)
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package pdf

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/domonda/go-errs"
)

func (r *Renderer) pngColorSpace(ct byte) (colspace string, colorVal int) {
	colorVal = 1
	switch ct {
	case 0, 4:
		colspace = "DeviceGray"
	case 2, 6:
		colspace = "DeviceRGB"
		colorVal = 3
	case 3:
		colspace = "Indexed"
	default:
		r.err = errs.Errorf("unknown color type in PNG buffer: %d", ct)
	}
	return colspace, colorVal
}

func (r *Renderer) parsepngstream(buf *rbuffer, readdpi bool) (info *ImageInfoType) {
	info = r.newImageInfo()
	// 	Check signature
	if string(buf.Next(8)) != "\x89PNG\x0d\x0a\x1a\x0a" {
		r.err = errs.New("not a PNG buffer")
		return info
	}
	// Read header chunk
	_ = buf.Next(4)
	if string(buf.Next(4)) != "IHDR" {
		r.err = errs.New("incorrect PNG buffer")
		return info
	}
	w := buf.i32()
	h := buf.i32()
	bpc := buf.u8()
	if bpc > 8 {
		if r.pdfVersion < pdfVers1_5 {
			r.pdfVersion = pdfVers1_5
		}
	}
	ct := buf.u8()
	var colspace string
	var colorVal int
	colspace, colorVal = r.pngColorSpace(ct)
	if r.err != nil {
		return info
	}
	if buf.u8() != 0 {
		r.err = errs.New("'unknown compression method in PNG buffer")
		return info
	}
	if buf.u8() != 0 {
		r.err = errs.New("'unknown filter method in PNG buffer")
		return info
	}
	if buf.u8() != 0 {
		r.err = errs.New("interlacing not supported in PNG buffer")
		return info
	}
	_ = buf.Next(4)
	dp := fmt.Sprintf("/Predictor 15 /Colors %d /BitsPerComponent %d /Columns %d", colorVal, bpc, w)
	// Scan chunks looking for palette, transparency and image data
	var (
		pal  []byte
		trns []int
		npix = w * h
		data = make([]byte, 0, npix/8)
		loop = true
	)
	for loop {
		n := int(buf.i32())
		// dbg("Loop [%d]", n)
		switch string(buf.Next(4)) {
		case "PLTE":
			// dbg("PLTE")
			// Read palette
			pal = buf.Next(n)
			_ = buf.Next(4)
		case "tRNS":
			// dbg("tRNS")
			// Read transparency info
			t := buf.Next(n)
			switch ct {
			case 0:
				trns = []int{int(t[1])} // ord(substr($t,1,1)));
			case 2:
				trns = []int{int(t[1]), int(t[3]), int(t[5])} // array(ord(substr($t,1,1)), ord(substr($t,3,1)), ord(substr($t,5,1)));
			default:
				pos := strings.Index(string(t), "\x00")
				if pos >= 0 {
					trns = []int{pos} // array($pos);
				}
			}
			_ = buf.Next(4)
		case "IDAT":
			// dbg("IDAT")
			// Read image data block
			data = append(data, buf.Next(n)...)
			_ = buf.Next(4)
		case "IEND":
			// dbg("IEND")
			loop = false
		case "pHYs":
			// dbg("pHYs")
			// png files theoretically support different x/y dpi
			// but we ignore files like this
			// but if they're the same then we can stamp our info
			// object with it
			x := int(buf.i32())
			y := int(buf.i32())
			units := buf.u8()
			// fmt.Printf("got a pHYs block, x=%d, y=%d, u=%d, readdpi=%t\n",
			// x, y, int(units), readdpi)
			// only modify the info block if the user wants us to
			if x == y && readdpi {
				switch units {
				// if units is 1 then measurement is px/meter
				case 1:
					info.dpi = float64(x) / 39.3701 // inches per meter
				default:
					info.dpi = float64(x)
				}
			}
			_ = buf.Next(4)
		default:
			// dbg("default")
			_ = buf.Next(n + 4)
		}
		if loop {
			loop = n > 0
		}
	}
	if colspace == "Indexed" && len(pal) == 0 {
		r.err = errs.New("missing palette in PNG buffer")
	}
	info.w = float64(w)
	info.h = float64(h)
	info.cs = colspace
	info.bpc = int(bpc)
	info.f = "FlateDecode"
	info.dp = dp
	info.pal = pal
	info.trns = trns
	// dbg("ct [%d]", ct)
	if ct >= 4 {
		// Separate alpha and color channels
		mem, err := xmem.uncompress(data)
		if err != nil {
			r.err = err
			return info
		}
		data = mem.Bytes()
		var color, alpha []byte
		if ct == 4 {
			// Gray image
			width := int(w)
			height := int(h)
			length := 2 * width
			sz := height * (width + 1)
			color = data[:0:sz] // reuse decompressed data buffer.
			alpha = make([]byte, 0, sz)
			var pos, elPos int
			for i := range height {
				pos = (1 + length) * i
				color = append(color, data[pos])
				alpha = append(alpha, data[pos])
				elPos = pos + 1
				for range width {
					color = append(color, data[elPos])
					alpha = append(alpha, data[elPos+1])
					elPos += 2
				}
			}
		} else {
			// RGB image
			width := int(w)
			height := int(h)
			length := 4 * width
			sz := width * height
			color = data[: 0 : sz*3+height] // reuse decompressed data buffer.
			alpha = make([]byte, 0, sz+height)
			var pos, elPos int
			for i := range height {
				pos = (1 + length) * i
				color = append(color, data[pos])
				alpha = append(alpha, data[pos])
				elPos = pos + 1
				for range width {
					tmp := data[elPos : elPos+4]
					color = append(color, tmp[0], tmp[1], tmp[2])
					alpha = append(alpha, tmp[3])
					elPos += 4
				}
			}
		}

		xc := xmem.compress(color)
		data = bytes.Clone(xc.Bytes())
		xmem.release(xc)

		// release uncompressed data buffer, after the color buffer
		// has been compressed.
		xmem.release(mem)

		xa := xmem.compress(alpha)
		info.smask = bytes.Clone(xa.Bytes())
		xmem.release(xa)

		if r.pdfVersion < pdfVers1_4 {
			r.pdfVersion = pdfVers1_4
		}
	}
	info.data = data
	return info
}
