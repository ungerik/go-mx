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
	"unicode/utf16"
)

// utf8toutf16 converts s to UTF-16BE bytes as required by PDF text strings,
// by default prefixed with a byte order mark.
func utf8toutf16(s string, withBOM ...bool) string {
	bom := len(withBOM) == 0 || withBOM[0]
	units := utf16.Encode([]rune(s))
	res := make([]byte, 0, 2+2*len(units))
	if bom {
		res = append(res, 0xFE, 0xFF)
	}
	for _, u := range units {
		res = append(res, byte(u>>8), byte(u))
	}
	return string(res)
}

func repClosure(m map[rune]byte) func(string) string {
	return func(str string) string {
		buf := make([]byte, 0, len(str))
		for _, r := range str {
			ch := byte('.')
			if r < 0x80 {
				ch = byte(r)
			} else if b, ok := m[r]; ok {
				ch = b
			}
			buf = append(buf, ch)
		}
		return string(buf)
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
func (p *PointType) Transform(x, y float64) PointType {
	return PointType{p.X + x, p.Y + y}
}

// Orientation returns the orientation of a given size:
// "P" for portrait, "L" for landscape
func (s *SizeType) Orientation() string {
	if s == nil || s.Ht == s.Wd {
		return ""
	}
	if s.Wd > s.Ht {
		return "L"
	}
	return "P"
}

// ScaleBy expands a size by a certain factor
func (s *SizeType) ScaleBy(factor float64) SizeType {
	return SizeType{s.Wd * factor, s.Ht * factor}
}

// ScaleToWidth adjusts the height of a size to match the given width
func (s *SizeType) ScaleToWidth(width float64) SizeType {
	height := s.Ht * width / s.Wd
	return SizeType{width, height}
}

// ScaleToHeight adjusts the width of a size to match the given height
func (s *SizeType) ScaleToHeight(height float64) SizeType {
	width := s.Wd * height / s.Ht
	return SizeType{width, height}
}

// Condition font family string to PDF name compliance. See section 5.3 (Names)
// in https://resources.infosecinstitute.com/pdf-file-format-basic-structure/
func fontFamilyEscape(familyStr string) string {
	return strings.ReplaceAll(familyStr, " ", "#20")
}

func returnStringUnchanged(s string) string { return s }
