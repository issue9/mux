// SPDX-License-Identifier: MIT

package group

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v5"
	"github.com/issue9/mux/v5/params"
)

var _ Matcher = &Hosts{}

func TestHost_RegisterInterceptor(t *testing.T) {
	a := assert.New(t)

	h := NewHosts(true)
	a.NotNil(h)
	h.RegisterInterceptor(mux.InterceptorWord, "\\d+")
	h.Add("{sub:\\d+}.example.com")

	r := httptest.NewRequest(http.MethodGet, "http://sub--1.example.com/test", nil)
	rr, ok := h.Match(r)
	a.False(ok).Nil(rr)

	// 将 \\d+ 注册为任意非空字符
	r = httptest.NewRequest(http.MethodGet, "http://sub.example.com/test", nil)
	rr, ok = h.Match(r)
	a.True(ok).NotNil(rr) // rr 包含了参数信息
	ps := params.Get(rr)
	a.Equal(ps, params.Params{"sub": "sub"})
}

func TestHosts_Match(t *testing.T) {
	a := assert.New(t)

	h := NewHosts(true, "caixw.io", "caixw.oi", "{sub}.example.com")
	a.NotNil(h)

	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	rr, ok := h.Match(r)
	a.True(ok).Equal(rr, r)

	r = httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	rr, ok = h.Match(r)
	a.True(ok).Equal(rr, r)

	r = httptest.NewRequest(http.MethodGet, "https://CAIXW.io/test", nil)
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

func TestNewHosts(t *testing.T) {
	a := assert.New(t)

	h := NewHosts(false)
	a.NotNil(h)

	h = NewHosts(true, "{sub}.example.com")
	a.NotNil(h)

	// 相同的值
	a.Panic(func() {
		NewHosts(true, "{sub}.example.com", "{sub}.example.com")
	})
}

func TestHosts_Add_Delete(t *testing.T) {
	a := assert.New(t)

	h := NewHosts(true)
	a.NotNil(h)

	h.Add("xx.example.com")
	a.Panic(func() {
		h.Add("xx.example.com")
	})
	h.Add("{sub}.example.com")
	a.Panic(func() {
		h.Add("{sub}.example.com")
	})

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
