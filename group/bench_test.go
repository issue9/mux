// SPDX-License-Identifier: MIT

package group

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func BenchmarkHost_Match(b *testing.B) {
	a := assert.New(b)
	h := NewHosts("caixw.io", "caixw.oi", "*.example.com")
	r := httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)

	for i := 0; i < b.N; i++ {
		_, ok := h.Match(r)
		a.True(ok)
	}
}

func BenchmarkHeaderVersion_Match(b *testing.B) {
	a := assert.New(b)
	h := &HeaderVersion{Versions: []string{"3.0", "4.0", "1.0", "2.0"}}
	r := httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	r.Header.Set("Accept", "application/json; version=1.0")

	for i := 0; i < b.N; i++ {
		_, ok := h.Match(r)
		a.True(ok)
	}
}

func BenchmarkPathVersion_Match(b *testing.B) {
	a := assert.New(b)
	r := httptest.NewRequest(http.MethodGet, "https://caixw.io/v1/test", nil)
	h := NewPathVersion("v4", "v3", "v1/", "/v2")

	for i := 0; i < b.N; i++ {
		r.URL.Path = "/v1/test" // 防止 r.URL.Path 被修改
		_, ok := h.Match(r)
		a.True(ok)
	}
}

func BenchmarkFindVersionNumberInHeader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = findVersionNumberInHeader("application/json;version=1.0;application/json")
	}
}
