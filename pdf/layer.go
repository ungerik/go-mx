// Copyright ©2023 The go-pdf Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
 * Copyright (c) 2014 Kurt Jung (Gmail: kurt.w.jung)
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

// Routines in this file are translated from
// http://www.fpdf.org/en/script/script97.php

import (
	"fmt"
	"strings"
)

type layer struct {
	name    string
	visible bool
	objNum  int // object number
}

type layerRec struct {
	list          []layer
	currentLayer  int
	openLayerPane bool
}

// AddLayer defines a layer that can be shown or hidden when the document is
// displayed. name specifies the layer name that the document reader will
// display in the layer list. visible specifies whether the layer will be
// initially visible. The return value is an integer ID that is used in a call
// to BeginLayer().
func (r *Renderer) AddLayer(name string, visible bool) (layerID int) {
	layerID = len(r.layer.list)
	r.layer.list = append(r.layer.list, layer{name: name, visible: visible})
	return layerID
}

// BeginLayer is called to begin adding content to the specified layer. All
// content added to the page between a call to BeginLayer and a call to
// EndLayer is added to the layer specified by id. See AddLayer for more
// details.
func (r *Renderer) BeginLayer(id int) {
	r.EndLayer()
	if id >= 0 && id < len(r.layer.list) {
		r.outf("/OC /OC%d BDC", id)
		r.layer.currentLayer = id
	}
}

// EndLayer is called to stop adding content to the currently active layer. See
// BeginLayer for more details.
func (r *Renderer) EndLayer() {
	if r.layer.currentLayer >= 0 {
		r.out("EMC")
		r.layer.currentLayer = -1
	}
}

// OpenLayerPane advises the document reader to open the layer pane when the
// document is initially displayed.
func (r *Renderer) OpenLayerPane() {
	r.layer.openLayerPane = true
}

func (r *Renderer) layerEndDoc() {
	if len(r.layer.list) == 0 {
		return
	}
	if r.pdfVersion < pdfVers1_5 {
		r.pdfVersion = pdfVers1_5
	}
}

func (r *Renderer) layerPutLayers() {
	for j, l := range r.layer.list {
		r.newobj()
		r.layer.list[j].objNum = r.n
		r.outf("<</Type /OCG /Name %s>>", r.textstring(utf8toutf16(l.name)))
		r.out("endobj")
	}
}

func (r *Renderer) layerPutResourceDict() {
	if len(r.layer.list) == 0 {
		return
	}
	r.out("/Properties <<")
	for j, layer := range r.layer.list {
		r.outf("/OC%d %d 0 R", j, layer.objNum)
	}
	r.out(">>")
}

func (r *Renderer) layerPutCatalog() {
	if len(r.layer.list) == 0 {
		return
	}
	var onStr strings.Builder
	var offStr strings.Builder
	for _, layer := range r.layer.list {
		fmt.Fprintf(&onStr, "%d 0 R ", layer.objNum)
		if !layer.visible {
			fmt.Fprintf(&offStr, "%d 0 R ", layer.objNum)
		}
	}
	r.outf("/OCProperties <</OCGs [%s] /D <</OFF [%s] /Order [%s]>>>>", onStr.String(), offStr.String(), onStr.String())
	if r.layer.openLayerPane {
		r.out("/PageMode /UseOC")
	}
}
