// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux/internal/syntax"
)

func BenchmarkBasic_Match(b *testing.B) {
	a := assert.New(b)
	e, err := New(&syntax.Syntax{Pattern: "/blog/post/1", Type: syntax.TypeBasic})
	a.NotError(err)

	for i := 0; i < b.N; i++ {

		if ok, _ := e.Match("/blog/post/1"); !ok {
			b.Error("BenchmarkBasic_match:error")
		}
	}
}

func BenchmarkRegexp_Match(b *testing.B) {
	a := assert.New(b)
	e, err := New(&syntax.Syntax{Pattern: "/blog/post/{id:\\d+}", Type: syntax.TypeRegexp})
	a.NotError(err)

	for i := 0; i < b.N; i++ {
		if ok, _ := e.Match("/blog/post/1"); !ok {
			b.Error("BenchmarkRegexp_match:error")
		}
	}
}

func BenchmarkNamed_Match(b *testing.B) {
	a := assert.New(b)
	e, err := New(&syntax.Syntax{Pattern: "/blog/post/{id}/{id2}", Type: syntax.TypeNamed})
	a.NotError(err)

	for i := 0; i < b.N; i++ {
		if ok, _ := e.Match("/blog/post/1/2"); !ok {
			b.Error("BenchmarkNamed_match:error")
		}
	}
}
