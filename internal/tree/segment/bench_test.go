// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package segment

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux/params"
)

func BenchmarkReg_Match(b *testing.B) {
	a := assert.New(b)

	r, err := newReg("{id:\\d+}/author")
	a.NotError(err).NotNil(r)

	ps := make(params.Params, 1)

	for i := 0; i < b.N; i++ {
		if ok, _ := r.Match("5/author/profile", ps); !ok {
			b.Error("BenchmarkReg_Match")
		}
	}
}

func BenchmarkNamed_Match(b *testing.B) {
	a := assert.New(b)

	r, err := newNamed("{id}/author")
	a.NotError(err).NotNil(r)

	ps := make(params.Params, 1)

	for i := 0; i < b.N; i++ {
		if ok, _ := r.Match("5/author/profile", ps); !ok {
			b.Error("BenchmarkNamed_Match")
		}
	}
}

func BenchmarkStr_Match(b *testing.B) {
	r := str("1/author")
	ps := make(params.Params, 1)

	for i := 0; i < b.N; i++ {
		if ok, _ := r.Match("1/author/profile", ps); !ok {
			b.Error("BenchmarkStr_Match")
		}
	}
}
