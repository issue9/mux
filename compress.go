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
	gzw io.Writer
	rw  http.ResponseWriter
	hj  http.Hijacker
}

func (cw *compressWriter) Write(bs []byte) (int, error) {
	return cw.gzw.Write(bs)
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

// 支持gzip或是deflate功能的handler，根据客户端请求内容自动匹配相应的压缩算法。
func NewCompress(h http.Handler) http.Handler {
	return &compress{h: h}
}

func (c *compress) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		c.h.ServeHTTP(w, r)
		return
	}

	var gzw io.WriteCloser
	defer func() {
		if gzw != nil {
			gzw.Close()
		}
	}()

	var encoding string
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

	w.Header().Add("Vary", "Accept-Encoding")
	cw := &compressWriter{
		gzw: gzw,
		rw:  w,
		hj:  hj,
	}

	c.h.ServeHTTP(cw, r)
}
