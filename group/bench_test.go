// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package group

import (
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/assert/v4/rest"

	"github.com/issue9/mux/v8/types"
)

func BenchmarkHost_Match(b *testing.B) {
	a := assert.New(b, false)
	h := NewHosts(true, "caixw.io", "caixw.oi", "*.example.com")
	a.NotNil(h)
	r := rest.Get(a, "https://caixw.io/test").Request()

	ps := types.NewContext()
	for i := 0; i < b.N; i++ {
		a.True(h.Match(r, ps))
	}
}

func BenchmarkHeaderVersionWithoutKey_Match(b *testing.B) {
	a := assert.New(b, false)
	h := NewHeaderVersion("", "version", nil, "3.0", "4.0", "1.0", "2.0")
	r := rest.Get(a, "https://caixw.io/test").
		Header("Accept", "application/json; version=1.0").
		Request()

	ps := types.NewContext()
	for i := 0; i < b.N; i++ {
		a.True(h.Match(r, ps))
	}
}

func BenchmarkHeaderVersionWithKey_Match(b *testing.B) {
	a := assert.New(b, false)
	h := NewHeaderVersion("version", "", nil, "3.0", "4.0", "1.0", "2.0")
	r := rest.Get(a, "https://caixw.io/test").
		Header("Accept", "application/json; version=1.0").
		Request()

	ps := types.NewContext()
	for i := 0; i < b.N; i++ {
		a.True(h.Match(r, ps))
	}
}

func BenchmarkPathVersionWithoutKey_Match(b *testing.B) {
	a := assert.New(b, false)
	h := NewPathVersion("", "v4", "v3", "v1/", "/v2")

	ps := types.NewContext()
	for i := 0; i < b.N; i++ {
		r := rest.Get(a, "https://caixw.io/v1/test").Request()
		a.True(h.Match(r, ps))
	}
}

func BenchmarkPathVersionWithKey_Match(b *testing.B) {
	a := assert.New(b, false)
	h := NewPathVersion("version", "v4", "v3", "v1/", "/v2")

	ps := types.NewContext()
	for i := 0; i < b.N; i++ {
		r := rest.Get(a, "https://caixw.io/v1/test").Request()
		a.True(h.Match(r, ps))
	}
}
