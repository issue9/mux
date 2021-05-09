// SPDX-License-Identifier: MIT

package group

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

var (
	_ Matcher = &HeaderVersion{}
	_ Matcher = &PathVersion{}
)

func TestHeaderVersion_Match(t *testing.T) {
	a := assert.New(t)

	h := &HeaderVersion{
		Versions: []string{"1.0", "2.0", "3.0"},
	}

	// 相同版本号
	r := httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(r)
	a.True(h.Match(r))

	// 空版本
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	r.Header.Set("Accept", "application/json; version=")
	a.NotNil(r)
	a.False(h.Match(r))

	// 不同版本
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	r.Header.Set("Accept", "application/json; version=2")
	a.NotNil(r)
	a.False(h.Match(r))

	// 空值，不匹配任何内容

	h = &HeaderVersion{}
	r = httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(r)
	a.False(h.Match(r))

	// 空版本
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	r.Header.Set("Accept", "application/json; version=")
	a.NotNil(r)
	a.False(h.Match(r))

	// 不同版本
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	r.Header.Set("Accept", "application/json; version=2")
	a.NotNil(r)
	a.False(h.Match(r))

	r = httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(r)
	a.False(h.Match(r))
}

func TestPathVersion_Match(t *testing.T) {
	a := assert.New(t)

	h := NewPathVersion("v3", "/v2", "/v1")

	a.Panic(func() {
		NewPathVersion("", "v3")
	})

	// 相同版本号
	r := httptest.NewRequest(http.MethodGet, "https://caixw.io/v1/test", nil)
	a.NotNil(r)
	a.True(h.Match(r))
	a.Equal(r.URL.Path, "/test")

	// 空版本
	r = httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	a.NotNil(r)
	a.False(h.Match(r))
	a.Equal(r.URL.Path, "/test")

	// 不同版本
	r = httptest.NewRequest(http.MethodGet, "http://caixw.io/v111/test", nil)
	a.NotNil(r)
	a.False(h.Match(r))
	a.Equal(r.URL.Path, "/v111/test")

	// 空值，不匹配任何内容

	h = NewPathVersion()
	r = httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(r)
	a.False(h.Match(r))
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
