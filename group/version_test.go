// SPDX-License-Identifier: MIT

package group

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v2"
)

var (
	_ Matcher = &HeaderVersion{}
	_ Matcher = &PathVersion{}
)

func TestHeaderVersion_Match(t *testing.T) {
	a := assert.New(t, false)

	h := NewHeaderVersion("version", "", nil, "1.0", "2.0", "3.0")

	// 相同版本号
	r, err := http.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(r)
	ps, ok := h.Match(r)
	a.True(ok).NotNil(ps).Equal(ps.MustString("version", "not-exists"), "1.0")

	// 相同版本号，未指定 paramName
	h = NewHeaderVersion("", "", nil, "1.0", "2.0", "3.0")
	r, err = http.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(r)
	ps, ok = h.Match(r)
	a.True(ok).Nil(ps)

	// 空版本
	r, err = http.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", "application/json; version=")
	a.NotNil(r)
	ps, ok = h.Match(r)
	a.False(ok).Nil(ps)

	// 不同版本
	r, err = http.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", "application/json; version = 2")
	a.NotNil(r)
	ps, ok = h.Match(r)
	a.False(ok).Nil(ps)

	// 空值，不匹配任何内容

	h = NewHeaderVersion("", "", nil)
	r, err = http.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(r)
	ps, ok = h.Match(r)
	a.False(ok).Nil(ps)

	// 空版本
	r, err = http.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", "application/json; version=")
	a.NotNil(r)
	ps, ok = h.Match(r)
	a.False(ok).Nil(ps)

	// 不同版本
	r, err = http.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", "application/json; version=2")
	a.NotNil(r)
	ps, ok = h.Match(r)
	a.False(ok).Nil(ps)

	r, err = http.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(r)
	ps, ok = h.Match(r)
	a.False(ok).Nil(ps)
}

func TestPathVersion_Match(t *testing.T) {
	a := assert.New(t, false)

	h := NewPathVersion("version", "v3", "/v2", "/v1")

	a.Panic(func() {
		NewPathVersion("version", "", "v3")
	})

	// 相同版本号
	r, err := http.NewRequest(http.MethodGet, "https://caixw.io/v1/test", nil)
	a.NotError(err).NotNil(r)
	a.NotNil(r)
	ps, ok := h.Match(r)
	a.True(ok).NotNil(ps)
	a.Equal(r.URL.Path, "/test").
		Equal(ps.MustString("version", "not-found"), "/v1")

	// 相同版本号，未指定 key
	h = NewPathVersion("", "v3", "/v2", "/v1")
	r, err = http.NewRequest(http.MethodGet, "https://caixw.io/v1/test", nil)
	a.NotError(err).NotNil(r)
	a.NotNil(r)
	ps, ok = h.Match(r)
	a.True(ok).Nil(ps)
	a.Equal(r.URL.Path, "/test")

	// 空版本
	r, err = http.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	a.NotError(err).NotNil(r)
	a.NotNil(r)
	ps, ok = h.Match(r)
	a.False(ok).Nil(ps)
	a.Equal(r.URL.Path, "/test")

	// 不同版本
	r, err = http.NewRequest(http.MethodGet, "https://caixw.io/v111/test", nil)
	a.NotError(err).NotNil(r)
	a.NotNil(r)
	ps, ok = h.Match(r)
	a.False(ok).Nil(ps)
	a.Equal(r.URL.Path, "/v111/test")

	// 空值，不匹配任何内容

	h = NewPathVersion("version")
	r, err = http.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(r)
	ps, ok = h.Match(r)
	a.False(ok).Nil(ps)
}
