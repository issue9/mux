// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"net/http"
	"testing"
)

// BenchmarkBasic_Match-4    	200000000	         6.75 ns/op		go1.8
func BenchmarkBasic_Match(b *testing.B) {
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
	e := New("/blog/post/1", hf)
	for i := 0; i < b.N; i++ {
		if 0 != e.Match("/blog/post/1") {
			b.Error("BenchmarkBasic_Match:error")
		}
	}
}

// BenchmarkStatic_Match-4   	200000000	         8.05 ns/op		go1.8
func BenchmarkStatic_Match(b *testing.B) {
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
	e := New("/blog/post/", hf)
	for i := 0; i < b.N; i++ {
		if e.Match("/blog/post/1") > 1 {
			b.Error("BenchmarkStatic_Match:error")
		}
	}
}

// BenchmarkRegexp_Match-4   	 5000000	       337 ns/op		go1.8
func BenchmarkRegexp_Match(b *testing.B) {
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
	e := New("/blog/post/{id:\\d+}", hf)
	for i := 0; i < b.N; i++ {
		if 0 != e.Match("/blog/post/1") {
			b.Error("BenchmarkRegexp_Match:error")
		}
	}
}
