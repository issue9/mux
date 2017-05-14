// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"testing"

	"github.com/issue9/assert"
)

func BenchmarkBasic_Match(b *testing.B) {
	a := assert.New(b)
	e, err := New("/blog/post/1")
	a.NotError(err).NotNil(e)

	for i := 0; i < b.N; i++ {

		if ok, _ := e.Match("/blog/post/1"); !ok {
			b.Error("BenchmarkBasic_match:error")
		}
	}
}

func BenchmarkRegexp_Match(b *testing.B) {
	a := assert.New(b)
	e, err := New("/blog/post/{id:\\d+}")
	a.NotError(err).NotNil(e)

	for i := 0; i < b.N; i++ {
		if ok, _ := e.Match("/blog/post/1"); !ok {
			b.Error("BenchmarkRegexp_match:error")
		}
	}
}

func BenchmarkNamed_Match(b *testing.B) {
	a := assert.New(b)
	e, err := New("/blog/post/{id}/{id2}")
	a.NotError(err).NotNil(e)

	for i := 0; i < b.N; i++ {
		if ok, _ := e.Match("/blog/post/1/2"); !ok {
			b.Error("BenchmarkNamed_match:error")
		}
	}
}
