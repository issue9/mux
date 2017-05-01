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

func benchHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("handler"))
}

// go1.8 BenchmarkMux_ServeHTTPBasic-4    	 5000000	       281 ns/op
func BenchmarkMux_ServeHTTPBasic(b *testing.B) {
	a := assert.New(b)
	srv := New(false, nil, nil)

	srv.GetFunc("/blog/post/1", benchHandler)
	srv.GetFunc("/api/v2/login", benchHandler)

	r1, err := http.NewRequest("GET", "/blog/post/1", nil)
	a.NotError(err).NotNil(r1)
	r2, err := http.NewRequest("GET", "/api/v2/login", nil)
	a.NotError(err).NotNil(r2)
	r3, err := http.NewRequest("GET", "/api/v2x/login", nil)
	a.NotError(err).NotNil(r3)
	reqs := []*http.Request{r1, r2, r3}

	w := httptest.NewRecorder()

	srvfun := func(reqIndex int) {
		srv.ServeHTTP(w, reqs[reqIndex])
	}
	for i := 0; i < b.N; i++ {
		srvfun(i % len(reqs))
	}
}

// go1.8 BenchmarkMux_ServeHTTPStatic-4   	 5000000	       297 ns/op
func BenchmarkMux_ServeHTTPStatic(b *testing.B) {
	a := assert.New(b)
	srv := New(false, nil, nil)

	srv.GetFunc("/blog/post/", benchHandler)
	srv.GetFunc("/api/v2/", benchHandler)

	r1, err := http.NewRequest("GET", "/blog/post/1", nil)
	a.NotError(err).NotNil(r1)
	r2, err := http.NewRequest("GET", "/api/v2/login", nil)
	a.NotError(err).NotNil(r2)
	r3, err := http.NewRequest("GET", "/api/v2x/login", nil)
	a.NotError(err).NotNil(r3)
	reqs := []*http.Request{r1, r2, r3}

	w := httptest.NewRecorder()

	srvfun := func(reqIndex int) {
		srv.ServeHTTP(w, reqs[reqIndex])
	}
	for i := 0; i < b.N; i++ {
		srvfun(i % len(reqs))
	}
}

// go1.8 BenchmarkMux_ServeHTTPRegexp-4   	 1000000	      1350 ns/op
func BenchmarkMux_ServeHTTPRegexp(b *testing.B) {
	a := assert.New(b)
	srv := New(false, nil, nil)

	srv.GetFunc("/blog/post/{id}", benchHandler)
	srv.GetFunc("/api/v{version:\\d+}/login", benchHandler)

	r1, err := http.NewRequest("GET", "/blog/post/1", nil)
	a.NotError(err).NotNil(r1)
	r2, err := http.NewRequest("GET", "/api/v2/login", nil)
	a.NotError(err).NotNil(r2)
	r3, err := http.NewRequest("GET", "/api/v2x/login", nil)
	a.NotError(err).NotNil(r3)
	reqs := []*http.Request{r1, r2, r3}

	w := httptest.NewRecorder()

	srvfun := func(reqIndex int) {
		srv.ServeHTTP(w, reqs[reqIndex])
	}
	for i := 0; i < b.N; i++ {
		srvfun(i % len(reqs))
	}
}

// go1.8 BenchmarkMux_ServeHTTPAll-4      	 2000000	       737 ns/op
func BenchmarkMux_ServeHTTPAll(b *testing.B) {
	a := assert.New(b)
	srv := New(false, nil, nil)

	srv.GetFunc("/blog/basic/1", benchHandler)
	srv.GetFunc("/blog/static/", benchHandler)
	srv.GetFunc("/api/v{version:\\d+}/login", benchHandler)

	r1, err := http.NewRequest("GET", "/blog/static/1", nil)
	a.NotError(err).NotNil(r1)
	r2, err := http.NewRequest("GET", "/blog/basic/1", nil)
	a.NotError(err).NotNil(r2)
	r3, err := http.NewRequest("GET", "/api/v2/login", nil)
	a.NotError(err).NotNil(r3)
	r4, err := http.NewRequest("GET", "/api/v2x/login", nil)
	a.NotError(err).NotNil(r4)
	reqs := []*http.Request{r1, r2, r3, r4}

	w := httptest.NewRecorder()

	srvfun := func(reqIndex int) {
		srv.ServeHTTP(w, reqs[reqIndex])
	}
	for i := 0; i < b.N; i++ {
		srvfun(i % len(reqs))
	}
}
