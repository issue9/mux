// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

// 全静态匹配
// BenchmarkServeMux_ServeHTTPStatic-4	 5000000	       365 ns/op    go1.6
func BenchmarkServeMux_ServeHTTPStatic(b *testing.B) {
	a := assert.New(b)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler"))
	})
	srv := NewServeMux()
	srv.Get("/blog/post/1", h)
	srv.Get("/api/v2/login", h)

	r1, err := http.NewRequest("GET", "/blog/post/1", nil)
	a.NotError(err).NotNil(r1)
	r2, err := http.NewRequest("GET", "/api/v2/login", nil)
	a.NotError(err).NotNil(r2)
	r3, err := http.NewRequest("GET", "/api/v2x/login", nil)
	a.NotError(err).NotNil(r3)
	reqs := []*http.Request{r1, r2, r3}

	w := httptest.NewRecorder()

	srvfun := func(reqIndex int) {
		defer func() {
			_ = recover()
		}()

		srv.ServeHTTP(w, reqs[reqIndex])
	}
	for i := 0; i < b.N; i++ {
		srvfun(i % len(reqs))
	}
}

// 全正则匹配
// BenchmarkServeMux_ServeHTTPRegexp-4	 1000000	      1640 ns/op    go1.6
func BenchmarkServeMux_ServeHTTPRegexp(b *testing.B) {
	a := assert.New(b)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler"))
	})
	srv := NewServeMux()
	srv.Get("/blog/post/{id}", h)
	srv.Get("/api/v{version:\\d+}/login", h)

	r1, err := http.NewRequest("GET", "/blog/post/1", nil)
	a.NotError(err).NotNil(r1)
	r2, err := http.NewRequest("GET", "/api/v2/login", nil)
	a.NotError(err).NotNil(r2)
	r3, err := http.NewRequest("GET", "/api/v2x/login", nil)
	a.NotError(err).NotNil(r3)
	reqs := []*http.Request{r1, r2, r3}

	w := httptest.NewRecorder()

	srvfun := func(reqIndex int) {
		defer func() {
			_ = recover()
		}()

		srv.ServeHTTP(w, reqs[reqIndex])
	}
	for i := 0; i < b.N; i++ {
		srvfun(i % len(reqs))
	}
}

// 静态正则汇合匹配
// BenchmarkServeMux_ServeHTTPAll-4   	 1000000	      1024 ns/op    go1.6
func BenchmarkServeMux_ServeHTTPAll(b *testing.B) {
	a := assert.New(b)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler"))
	})
	srv := NewServeMux()
	srv.Get("/blog/post/1", h)
	srv.Get("/api/v{version:\\d+}/login", h)

	r1, err := http.NewRequest("GET", "/blog/post/1", nil)
	a.NotError(err).NotNil(r1)
	r2, err := http.NewRequest("GET", "/api/v2/login", nil)
	a.NotError(err).NotNil(r2)
	r3, err := http.NewRequest("GET", "/api/v2x/login", nil)
	a.NotError(err).NotNil(r3)
	reqs := []*http.Request{r1, r2, r3}

	w := httptest.NewRecorder()

	srvfun := func(reqIndex int) {
		defer func() {
			_ = recover()
		}()

		srv.ServeHTTP(w, reqs[reqIndex])
	}
	for i := 0; i < b.N; i++ {
		srvfun(i % len(reqs))
	}
}
