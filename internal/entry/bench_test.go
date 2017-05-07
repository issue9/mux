// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

var benchHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
})

func BenchmarkBasic_Match(b *testing.B) {
	a := assert.New(b)
	e, err := New("/blog/post/1", benchHandler)
	a.NotError(err)

	for i := 0; i < b.N; i++ {
		if 0 != e.Match("/blog/post/1") {
			b.Error("BenchmarkBasic_Match:error")
		}
	}
}

func BenchmarkStatic_Match(b *testing.B) {
	a := assert.New(b)
	e, err := New("/blog/post/", benchHandler)
	a.NotError(err)

	for i := 0; i < b.N; i++ {
		if e.Match("/blog/post/1") > 1 {
			b.Error("BenchmarkStatic_Match:error")
		}
	}
}

func BenchmarkRegexp_Match(b *testing.B) {
	a := assert.New(b)
	e, err := New("/blog/post/{id:\\d+}", benchHandler)
	a.NotError(err)

	for i := 0; i < b.N; i++ {
		if 0 != e.Match("/blog/post/1") {
			b.Error("BenchmarkRegexp_Match:error")
		}
	}
}

func BenchmarkNamed_Match(b *testing.B) {
	a := assert.New(b)
	e, err := New("/blog/post/{id}/{id2}", benchHandler)
	a.NotError(err)

	for i := 0; i < b.N; i++ {
		if 0 != e.Match("/blog/post/1/2") {
			b.Error("BenchmarkNamed_Match:error")
		}
	}
}
