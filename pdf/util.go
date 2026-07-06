// Copyright ©2023 The go-pdf Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
 * Copyright (c) 2013 Kurt Jung (Gmail: kurt.w.jung)
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
	"bufio"
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf16"
)

// utf8toutf16 converts s to UTF-16BE bytes as required by PDF text strings,
// by default prefixed with a byte order mark. Runes above the Basic
// Multilingual Plane are encoded as surrogate pairs. This helper is on the
// hot path of every text operation with a UTF-8 font, so it encodes in a
// single pass without intermediate rune/uint16 slices.
func utf8toutf16(s string, withBOM ...bool) string {
	bom := len(withBOM) == 0 || withBOM[0]
	res := make([]byte, 0, 2+2*len(s))
	if bom {
		res = append(res, 0xFE, 0xFF)
	}
	for _, r := range s {
		if r <= 0xFFFF {
			res = append(res, byte(r>>8), byte(r))
		} else {
			hi, lo := utf16.EncodeRune(r)
			res = append(res, byte(hi>>8), byte(hi), byte(lo>>8), byte(lo))
		}
	}
	return string(res)
}

// pdfTextToUTF8 decodes a metadata text string as stored on the renderer back
// to UTF-8. isUTF16 states the stored encoding explicitly (the setters record
// it): UTF-16BE with a byte order mark (surrogate pairs included), or Latin-1
// where each byte is its code point. The encoding is not sniffed from a BOM,
// because a Latin-1 string may legitimately start with the bytes "þÿ".
func pdfTextToUTF8(s string, isUTF16 bool) string {
	if !isUTF16 {
		runes := make([]rune, 0, len(s))
		for i := range len(s) {
			runes = append(runes, rune(s[i]))
		}
		return string(runes)
	}
	b, _ := strings.CutPrefix(s, "\xFE\xFF")
	u := make([]uint16, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		u = append(u, uint16(b[i])<<8|uint16(b[i+1]))
	}
	return string(utf16.Decode(u))
}

func repClosure(m map[rune]byte) func(string) string {
	return func(str string) string {
		var b strings.Builder
		b.Grow(len(str))
		for _, r := range str {
			ch := byte('.')
			if r < 0x80 {
				ch = byte(r)
			} else if c, ok := m[r]; ok {
				ch = c
			}
			b.WriteByte(ch)
		}
		return b.String()
	}
}

// UnicodeTranslator returns a function that can be used to translate, where
// possible, utf-8 strings to a form that is compatible with the specified code
// page. The returned function accepts a string and returns a string.
//
// r is a reader that should read a buffer made up of content lines that
// pertain to the code page of interest. Each line is made up of three
// whitespace separated fields. The first begins with "!" and is followed by
// two hexadecimal digits that identify the glyph position in the code page of
// interest. The second field begins with "U+" and is followed by the unicode
// code point value. The third is the glyph name. A number of these code page
// map files are packaged with the gfpdf library in the font directory.
//
// Lines that do not conform to the expected format (comments, headers) are
// skipped. An error is returned only if reading from r fails; in that case
// the returned function is valid but does not perform any rune translation.
func UnicodeTranslator(r io.Reader) (func(string) string, error) {
	m := make(map[rune]byte)
	var uPos, cPos uint32
	var nameStr string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		lineStr := strings.TrimSpace(sc.Text())
		if len(lineStr) == 0 {
			continue
		}
		if _, err := fmt.Sscanf(lineStr, "!%2X U+%4X %s", &cPos, &uPos, &nameStr); err != nil {
			continue // skip non-mapping lines
		}
		if cPos >= 0x80 {
			m[rune(uPos)] = byte(cPos)
		}
	}
	if err := sc.Err(); err != nil {
		return returnStringUnchanged, err
	}
	return repClosure(m), nil
}

// UnicodeTranslatorFromFile returns a function that can be used to translate,
// where possible, utf-8 strings to a form that is compatible with the
// specified code page. See UnicodeTranslator for more details.
//
// fileStr identifies a font descriptor file that maps glyph positions to names.
//
// If an error occurs reading the file, the returned function is valid but does
// not perform any rune translation.
func UnicodeTranslatorFromFile(fileStr string) (func(string) string, error) {
	fl, err := os.Open(fileStr)
	if err != nil {
		return returnStringUnchanged, err
	}
	f, err := UnicodeTranslator(fl)
	return f, errors.Join(err, fl.Close())
}

// UnicodeTranslatorFromDescriptor returns a function that can be used to
// translate, where possible, utf-8 strings to a form that is compatible with
// the specified code page. See UnicodeTranslator for more details.
//
// cpStr identifies a code page. A descriptor file in the font directory, set
// with the fontDirStr argument in the call to New(), should have this name
// plus the extension ".map". If cpStr is empty, it will be replaced with
// "cp1252", the gofpdf code page default.
//
// If an error occurs reading the descriptor, the returned function is valid
// but does not perform any rune translation.
//
// The CellFormat_codepage example demonstrates this method.
func (r *Renderer) UnicodeTranslatorFromDescriptor(cpStr string) func(string) string {
	if r.err != nil {
		return returnStringUnchanged
	}
	cpStr = cmp.Or(cpStr, "cp1252")
	emb, err := embFS.Open("font_embed/" + cpStr + ".map")
	if err == nil {
		defer emb.Close()
		rep, err := UnicodeTranslator(emb)
		r.err = err
		return rep
	}
	rep, err := UnicodeTranslatorFromFile(filepath.Join(r.fontpath, cpStr) + ".map")
	r.err = err
	return rep
}

// Transform moves a point by given X, Y offset
func (p *Point) Transform(x, y float64) Point {
	return Point{p.X + x, p.Y + y}
}

// Orientation returns the orientation of a given size:
// "P" for portrait, "L" for landscape
func (s *Size) Orientation() string {
	if s == nil || s.Ht == s.Wd {
		return ""
	}
	if s.Wd > s.Ht {
		return "L"
	}
	return "P"
}

// ScaleBy expands a size by a certain factor
func (s *Size) ScaleBy(factor float64) Size {
	return Size{s.Wd * factor, s.Ht * factor}
}

// ScaleToWidth adjusts the height of a size to match the given width
func (s *Size) ScaleToWidth(width float64) Size {
	height := s.Ht * width / s.Wd
	return Size{width, height}
}

// ScaleToHeight adjusts the width of a size to match the given height
func (s *Size) ScaleToHeight(height float64) Size {
	width := s.Wd * height / s.Ht
	return Size{width, height}
}

// Condition font family string to PDF name compliance. See section 5.3 (Names)
// in https://resources.infosecinstitute.com/pdf-file-format-basic-structure/
func fontFamilyEscape(familyStr string) string {
	return strings.ReplaceAll(familyStr, " ", "#20")
}

// escapeName escapes the characters that cannot appear literally in a PDF
// name object (ISO 32000-1 7.3.5) with the #xx number-sign notation, e.g.
// the MIME type "text/xml" becomes the name "text#2Fxml".
func escapeName(s string) string {
	var b strings.Builder
	for i := range len(s) {
		c := s[i]
		if c <= ' ' || c > '~' || strings.IndexByte("#/%()<>[]{}", c) >= 0 {
			fmt.Fprintf(&b, "#%02X", c)
		} else {
			b.WriteByte(c)
		}
	}
	return b.String()
}

// pdfDocEncode renders s for a PDF byte string that must be ASCII/PDFDocEncoded
// rather than UTF-16, such as a file specification's /F entry. ASCII and Latin-1
// code points map to their single-byte value (which equals PDFDocEncoding across
// that range); anything above U+00FF is replaced with '?', since the exact
// Unicode name is carried separately by /UF. The common all-ASCII case returns s
// unchanged.
func pdfDocEncode(s string) string {
	ascii := true
	for _, r := range s {
		if r > 0x7E {
			ascii = false
			break
		}
	}
	if ascii {
		return s
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 0x20 && r <= 0x7E, r >= 0xA1 && r <= 0xFF:
			b.WriteByte(byte(r))
		default:
			b.WriteByte('?')
		}
	}
	return b.String()
}

// pdfDate formats tm as a PDF date string. withTimezone appends the UTC
// offset, fully specifying the instant as PDF/A-relevant dates require;
// without it the timezone is left unknown, the legacy fpdf format.
func pdfDate(tm time.Time, withTimezone bool) string {
	if !withTimezone {
		return "D:" + tm.Format("20060102150405")
	}
	_, offset := tm.Zone()
	sign := '+'
	if offset < 0 {
		sign = '-'
		offset = -offset
	}
	return fmt.Sprintf("D:%s%c%02d'%02d'", tm.Format("20060102150405"), sign, offset/3600, offset%3600/60)
}

func returnStringUnchanged(s string) string { return s }
