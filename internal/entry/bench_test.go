// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"testing"

	"github.com/issue9/assert"
)

func BenchmarkBasic_match(b *testing.B) {
	a := assert.New(b)
	e, err := newEntry("/blog/post/1")
	a.NotError(err)

	for i := 0; i < b.N; i++ {

		if ok, _ := e.match("/blog/post/1"); !ok {
			b.Error("BenchmarkBasic_match:error")
		}
	}
}

func BenchmarkRegexp_match(b *testing.B) {
	a := assert.New(b)
	e, err := newEntry("/blog/post/{id:\\d+}")
	a.NotError(err)

	for i := 0; i < b.N; i++ {
		if ok, _ := e.match("/blog/post/1"); !ok {
			b.Error("BenchmarkRegexp_match:error")
		}
	}
}

func BenchmarkNamed_match(b *testing.B) {
	a := assert.New(b)
	e, err := newEntry("/blog/post/{id}/{id2}")
	a.NotError(err)

	for i := 0; i < b.N; i++ {
		if ok, _ := e.match("/blog/post/1/2"); !ok {
			b.Error("BenchmarkNamed_match:error")
		}
	}
}
