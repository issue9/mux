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
	e, err := newEntry("/blog/post/1")
	a.NotError(err).NotNil(e)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if ok, _ := e.Match("/blog/post/1"); !ok {
			b.Error("BenchmarkBasic_Match:error")
		}
	}
}

func BenchmarkRegexp_Match(b *testing.B) {
	a := assert.New(b)
	e, err := newEntry("/blog/post/{id:\\d+}")
	a.NotError(err).NotNil(e)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if ok, _ := e.Match("/blog/post/1"); !ok {
			b.Error("BenchmarkRegexp_Match:error")
		}
	}
}

func BenchmarkNamed_Match(b *testing.B) {
	a := assert.New(b)
	e, err := newEntry("/blog/post/{id}.html/{id2}")
	a.NotError(err).NotNil(e)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if ok, _ := e.Match("/blog/post/1.html/2"); !ok {
			b.Error("BenchmarkNamed_Match:error")
		}
	}
}

func BenchmarkBasic_URL(b *testing.B) {
	a := assert.New(b)
	e, err := newEntry("/blog/post/1/*")
	a.NotError(err).NotNil(e)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := e.URL(nil, "abc"); err != nil {
			b.Errorf("BenchmarkBasic_URL:%v", err)
		}
	}
}

func BenchmarkRegexp_URL(b *testing.B) {
	a := assert.New(b)
	e, err := newEntry("/blog/post/{id:\\d+}/*")
	a.NotError(err).NotNil(e)
	params := map[string]string{"id": "1"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := e.URL(params, "author/profile"); err != nil {
			b.Errorf("BenchmarkRegexp_URL:%v", err)
		}
	}
}

func BenchmarkNamed_URL(b *testing.B) {
	a := assert.New(b)
	e, err := newEntry("/blog/post/{id}/{id2}/*")
	a.NotError(err).NotNil(e)
	params := map[string]string{"id": "1", "id2": "2"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := e.URL(params, "named"); err != nil {
			b.Errorf("BenchmarkNamed_URL:%v", err)
		}
	}
}
