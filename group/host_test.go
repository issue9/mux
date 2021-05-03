// SPDX-License-Identifier: MIT

package group

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

var _ Matcher = &Hosts{}

func TestHosts_Match(t *testing.T) {
	a := assert.New(t)

	h := NewHosts("caixw.io", "caixw.oi", "*.example.com")
	a.NotNil(h)
	a.Equal(len(h.domains), 2).
		Equal(len(h.wildcards), 1)

	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	a.True(h.Match(r))

	r = httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	a.True(h.Match(r))

	// 泛域名
	r = httptest.NewRequest(http.MethodGet, "https://xx.example.com/test", nil)
	a.True(h.Match(r))

	// 带端口
	r = httptest.NewRequest(http.MethodGet, "http://caixw.io:88/test", nil)
	a.True(h.Match(r))

	// 访问不允许的域名
	r = httptest.NewRequest(http.MethodGet, "http://sub.caixw.io/test", nil)
	a.False(h.Match(r))

	// 访问不允许的域名
	r = httptest.NewRequest(http.MethodGet, "http://sub.1example.com/test", nil)
	a.False(h.Match(r))
}

func TestHosts_Add_Delete(t *testing.T) {
	a := assert.New(t)

	h := NewHosts()

	h.Add("xx.example.com")
	h.Add("xx.example.com")
	h.Add("xx.example.com")
	h.Add("*.example.com")
	h.Add("*.example.com")
	h.Add("*.example.com")
	a.Equal(1, len(h.domains)).
		Equal(1, len(h.wildcards))

	h.Delete("*.example.com")
	a.Equal(1, len(h.domains)).
		Equal(0, len(h.wildcards))

	h.Delete("*.example.com")
	a.Equal(1, len(h.domains)).
		Equal(0, len(h.wildcards))

	h.Delete("xx.example.com")
	a.Equal(0, len(h.domains)).
		Equal(0, len(h.wildcards))
}
