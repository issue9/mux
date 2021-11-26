// SPDX-License-Identifier: MIT

package group

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/mux/v5"
	"github.com/issue9/mux/v5/internal/syntax"
)

var _ Matcher = &Hosts{}

func TestHost_RegisterInterceptor(t *testing.T) {
	a := assert.New(t, false)

	h := NewHosts(true)
	a.NotNil(h)
	h.RegisterInterceptor(mux.InterceptorWord, "\\d+")
	h.Add("{sub:\\d+}.example.com")

	r, err := http.NewRequest(http.MethodGet, "http://sub--1.example.com/test", nil)
	a.NotError(err).NotNil(r)
	rr, ok := h.Match(r)
	a.False(ok).Nil(rr)

	// 将 \\d+ 注册为任意非空字符
	r, err = http.NewRequest(http.MethodGet, "http://sub.example.com/test", nil)
	a.NotError(err).NotNil(r)
	rr, ok = h.Match(r)
	a.True(ok).NotNil(rr) // rr 包含了参数信息
	ps := syntax.GetParams(rr)
	a.Equal(ps, &syntax.Params{Params: []syntax.Param{{K: "sub", V: "sub"}}})
}

func TestHosts_Match(t *testing.T) {
	a := assert.New(t, false)

	h := NewHosts(true, "caixw.io", "caixw.oi", "{sub}.example.com")
	a.NotNil(h)

	r, err := http.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	a.NotError(err).NotNil(r)
	rr, ok := h.Match(r)
	a.True(ok).Equal(rr, r)

	r, err = http.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	a.NotError(err).NotNil(r)
	rr, ok = h.Match(r)
	a.True(ok).Equal(rr, r)

	r, err = http.NewRequest(http.MethodGet, "https://CAIXW.io/test", nil)
	a.NotError(err).NotNil(r)
	rr, ok = h.Match(r)
	a.True(ok).Equal(rr, r)

	// 泛域名
	r, err = http.NewRequest(http.MethodGet, "https://xx.example.com/test", nil)
	a.NotError(err).NotNil(r)
	rr, ok = h.Match(r)
	a.True(ok).NotEqual(rr, r) // 通过 context.WithValue 修改了 rr
	sub := syntax.GetParams(rr).MustString("sub", "yy")
	a.Equal(sub, "xx")

	// 泛域名
	r, err = http.NewRequest(http.MethodGet, "https://xx.yy.example.com/test", nil)
	a.NotError(err).NotNil(r)
	rr, ok = h.Match(r)
	a.True(ok).NotEqual(rr, r) // 通过 context.WithValue 修改了 rr
	sub = syntax.GetParams(rr).MustString("sub", "yy")
	a.Equal(sub, "xx.yy")

	// 带端口
	r, err = http.NewRequest(http.MethodGet, "http://caixw.io:88/test", nil)
	a.NotError(err).NotNil(r)
	rr, ok = h.Match(r)
	a.True(ok).Equal(rr, r)

	// 访问不允许的域名
	r, err = http.NewRequest(http.MethodGet, "http://sub.caixw.io/test", nil)
	a.NotError(err).NotNil(r)
	rr, ok = h.Match(r)
	a.False(ok).Nil(rr)

	// 访问不允许的域名
	r, err = http.NewRequest(http.MethodGet, "http://sub.1example.com/test", nil)
	a.NotError(err).NotNil(r)
	rr, ok = h.Match(r)
	a.False(ok).Nil(rr)
}

func TestNewHosts(t *testing.T) {
	a := assert.New(t, false)

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
	a := assert.New(t, false)

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
	r, err := http.NewRequest(http.MethodGet, "https://xx.example.com/api/path", nil)
	a.NotError(err).NotNil(r)
	rr, ok := h.Match(r)
	a.True(ok).Equal(rr, r)

	// 删除 xx.example.com，则适配到 {sub}.example.com
	h.Delete("xx.example.com")
	r, err = http.NewRequest(http.MethodGet, "https://xx.example.com/api/path", nil)
	a.NotError(err).NotNil(r)
	rr, ok = h.Match(r)
	a.True(ok).NotNil(rr).NotEqual(r, rr)

	// delete {sub}.example.com
	h.Delete("{sub}.example.com")
	r, err = http.NewRequest(http.MethodGet, "https://zzz.example.com/api/path", nil)
	a.NotError(err).NotNil(r)
	rr, ok = h.Match(r)
	a.False(ok).Nil(rr)
}
