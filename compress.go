// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"bufio"
	"compress/flate"
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"strings"
)

type compressWriter struct {
	w  io.Writer
	rw http.ResponseWriter
	hj http.Hijacker
}

func (cw *compressWriter) Write(bs []byte) (int, error) {
	return cw.w.Write(bs)
}

func (cw *compressWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return cw.hj.Hijack()
}

func (cw *compressWriter) Header() http.Header {
	return cw.rw.Header()
}

func (cw *compressWriter) WriteHeader(code int) {
	cw.rw.WriteHeader(code)
}

type compress struct {
	h http.Handler
}

func NewCompress(h http.Handler) http.Handler {
	return &compress{h: h}
}

func (c *compress) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var encoding string
	var gzw io.Writer

	hj, ok := w.(http.Hijacker)
	if !ok {
		c.h.ServeHTTP(w, r)
		return
	}

	encodings := strings.Split(r.Header.Get("Accept-Encoding"), ",")
	for _, encoding = range encodings {
		encoding = strings.ToLower(encoding)

		if encoding == "gzip" {
			gzw = gzip.NewWriter(w)
			w.Header().Set("Content-Encoding", "gzip")
			break
		}

		if encoding == "deflate" {
			var err error
			gzw, err = flate.NewWriter(w, flate.DefaultCompression)
			if err != nil { // 若出错，不压缩，直接返回
				c.h.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Content-Encoding", "deflate")
			break
		}
	}

	cw := &compressWriter{
		w:  gzw,
		rw: w,
		hj: hj,
	}

	c.h.ServeHTTP(cw, r)
}
