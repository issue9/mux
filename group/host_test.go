// SPDX-License-Identifier: MIT

package group

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v4/params"
)

var _ Matcher = &Hosts{}

func TestHosts_Match(t *testing.T) {
	a := assert.New(t)

	h, err := NewHosts("caixw.io", "caixw.oi", "{sub}.example.com")
	a.NotError(err).NotNil(h)

	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	rr, ok := h.Match(r)
	a.True(ok).Equal(rr, r)

	r = httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	rr, ok = h.Match(r)
	a.True(ok).Equal(rr, r)

	// 泛域名
	r = httptest.NewRequest(http.MethodGet, "https://xx.example.com/test", nil)
	rr, ok = h.Match(r)
	a.True(ok).NotEqual(rr, r) // 通过 context.WithValue 修改了 rr
	sub := params.Get(rr).MustString("sub", "yy")
	a.Equal(sub, "xx")

	// 泛域名
	r = httptest.NewRequest(http.MethodGet, "https://xx.yy.example.com/test", nil)
	rr, ok = h.Match(r)
	a.True(ok).NotEqual(rr, r) // 通过 context.WithValue 修改了 rr
	sub = params.Get(rr).MustString("sub", "yy")
	a.Equal(sub, "xx.yy")

	// 带端口
	r = httptest.NewRequest(http.MethodGet, "http://caixw.io:88/test", nil)
	rr, ok = h.Match(r)
	a.True(ok).Equal(rr, r)

	// 访问不允许的域名
	r = httptest.NewRequest(http.MethodGet, "http://sub.caixw.io/test", nil)
	rr, ok = h.Match(r)
	a.False(ok).Nil(rr)

	// 访问不允许的域名
	r = httptest.NewRequest(http.MethodGet, "http://sub.1example.com/test", nil)
	rr, ok = h.Match(r)
	a.False(ok).Nil(rr)
}

func TestHosts_Add_Delete(t *testing.T) {
	a := assert.New(t)

	h, err := NewHosts()
	a.NotError(err).NotNil(h)

	a.NotError(h.Add("xx.example.com"))
	a.Error(h.Add("xx.example.com"))
	a.NotError(h.Add("{sub}.example.com"))
	a.Error(h.Add("{sub}.example.com"))

	// delete xx.example.com
	r := httptest.NewRequest(http.MethodGet, "https://xx.example.com/api/path", nil)
	rr, ok := h.Match(r)
	a.True(ok).Equal(rr, r)

	// 删除 xx.example.com，则适配到 {sub}.example.com
	h.Delete("xx.example.com")
	r = httptest.NewRequest(http.MethodGet, "https://xx.example.com/api/path", nil)
	rr, ok = h.Match(r)
	a.True(ok).NotNil(rr).NotEqual(r, rr)

	// delete {sub}.example.com
	h.Delete("{sub}.example.com")
	r = httptest.NewRequest(http.MethodGet, "https://zzz.example.com/api/path", nil)
	rr, ok = h.Match(r)
	a.False(ok).Nil(rr)
}
