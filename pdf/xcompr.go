// Copyright ©2021 The go-pdf Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pdf

import (
	"bytes"
	"compress/zlib"
	"sync"

	"github.com/domonda/go-errs"
)

// xmem pools bytes.Buffer values reused across zlib compress/uncompress
// operations to avoid per-call allocations. Callers must return a buffer
// with release once done reading its Bytes.
var xmem = xmempool{
	Pool: sync.Pool{
		New: func() any { return new(bytes.Buffer) },
	},
}

type xmempool struct{ sync.Pool }

func (pool *xmempool) compress(data []byte) *bytes.Buffer {
	buf := pool.Get().(*bytes.Buffer)
	buf.Grow(len(data))

	zw, err := zlib.NewWriterLevel(buf, zlib.BestSpeed)
	if err != nil {
		panic(errs.Errorf("could not create zlib writer: %w", err))
	}
	_, err = zw.Write(data)
	if err != nil {
		panic(errs.Errorf("could not zlib-compress slice: %w", err))
	}

	err = zw.Close()
	if err != nil {
		panic(errs.Errorf("could not close zlib writer: %w", err))
	}
	return buf
}

func (pool *xmempool) uncompress(data []byte) (*bytes.Buffer, error) {
	zr, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	buf := pool.Get().(*bytes.Buffer)
	buf.Reset()

	_, err = buf.ReadFrom(zr)
	if err != nil {
		pool.release(buf)
		return nil, err
	}

	return buf, nil
}

// release resets buf and returns it to the pool for reuse.
func (pool *xmempool) release(buf *bytes.Buffer) {
	buf.Reset()
	pool.Put(buf)
}
