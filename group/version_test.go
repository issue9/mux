// SPDX-License-Identifier: MIT

package group

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v5/params"
)

var (
	_ Matcher = &HeaderVersion{}
	_ Matcher = &PathVersion{}
)

func TestHeaderVersion_Match(t *testing.T) {
	a := assert.New(t)

	h := &HeaderVersion{
		Key:      "version",
		Versions: []string{"1.0", "2.0", "3.0"},
	}

	// 相同版本号
	r := httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(r)
	rr, ok := h.Match(r)
	a.True(ok).NotNil(rr).Equal(params.Get(rr).MustString("version", "not-exists"), "1.0")

	// 相同版本号，未指定 Key
	h = &HeaderVersion{
		Versions: []string{"1.0", "2.0", "3.0"},
	}
	r = httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(r)
	rr, ok = h.Match(r)
	a.True(ok).NotNil(rr).Equal(params.Get(rr).MustString("version", "not-exists"), "not-exists")

	// 空版本
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	r.Header.Set("Accept", "application/json; version=")
	a.NotNil(r)
	rr, ok = h.Match(r)
	a.False(ok).Nil(rr)

	// 不同版本
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	r.Header.Set("Accept", "application/json; version=2")
	a.NotNil(r)
	rr, ok = h.Match(r)
	a.False(ok).Nil(rr)

	// 空值，不匹配任何内容

	h = &HeaderVersion{}
	r = httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(r)
	rr, ok = h.Match(r)
	a.False(ok).Nil(rr)

	// 空版本
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	r.Header.Set("Accept", "application/json; version=")
	a.NotNil(r)
	rr, ok = h.Match(r)
	a.False(ok).Nil(rr)

	// 不同版本
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	r.Header.Set("Accept", "application/json; version=2")
	a.NotNil(r)
	rr, ok = h.Match(r)
	a.False(ok).Nil(rr)

	r = httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(r)
	rr, ok = h.Match(r)
	a.False(ok).Nil(rr)
}

func TestPathVersion_Match(t *testing.T) {
	a := assert.New(t)

	h := NewPathVersion("version", "v3", "/v2", "/v1")

	a.Panic(func() {
		NewPathVersion("version", "", "v3")
	})

	// 相同版本号
	r := httptest.NewRequest(http.MethodGet, "https://caixw.io/v1/test", nil)
	a.NotNil(r)
	rr, ok := h.Match(r)
	a.True(ok).NotNil(rr)
	a.Equal(rr.URL.Path, "/test").
		Equal(r.URL.Path, "/v1/test").
		Equal(params.Get(rr).MustString("version", "not-found"), "/v1")

		// 相同版本号，未指定 key
	h = NewPathVersion("", "v3", "/v2", "/v1")
	r = httptest.NewRequest(http.MethodGet, "https://caixw.io/v1/test", nil)
	a.NotNil(r)
	rr, ok = h.Match(r)
	a.True(ok).NotNil(rr)
	a.Equal(rr.URL.Path, "/test").
		Equal(r.URL.Path, "/v1/test").
		Equal(params.Get(rr).MustString("version", "not-found"), "not-found")

	// 空版本
	r = httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	a.NotNil(r)
	rr, ok = h.Match(r)
	a.False(ok).Nil(rr)
	a.Equal(r.URL.Path, "/test")

	// 不同版本
	r = httptest.NewRequest(http.MethodGet, "http://caixw.io/v111/test", nil)
	a.NotNil(r)
	rr, ok = h.Match(r)
	a.False(ok).Nil(rr)
	a.Equal(r.URL.Path, "/v111/test")

	// 空值，不匹配任何内容

	h = NewPathVersion("version")
	r = httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(r)
	rr, ok = h.Match(r)
	a.False(ok).Nil(rr)
}

func TestFindVersionNumberInHeader(t *testing.T) {
	a := assert.New(t)

	a.Equal(findVersionNumberInHeader(""), "")
	a.Equal(findVersionNumberInHeader("version="), "")
	a.Equal(findVersionNumberInHeader("Version="), "")
	a.Equal(findVersionNumberInHeader(";version="), "")
	a.Equal(findVersionNumberInHeader(";version=;"), "")
	a.Equal(findVersionNumberInHeader(";version=1.0"), "1.0")
	a.Equal(findVersionNumberInHeader(";version=1.0;"), "1.0")
	a.Equal(findVersionNumberInHeader(";version=1.0;application/json"), "1.0")
	a.Equal(findVersionNumberInHeader("application/json;version=1.0"), "1.0")
	a.Equal(findVersionNumberInHeader("application/json;version=1.0;application/json"), "1.0")
}
