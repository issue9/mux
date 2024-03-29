// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package group

import (
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/assert/v4/rest"

	"github.com/issue9/mux/v7/internal/syntax"
	"github.com/issue9/mux/v7/types"
)

var _ Matcher = &Hosts{}

func TestHost_RegisterInterceptor(t *testing.T) {
	a := assert.New(t, false)

	h := NewHosts(true)
	a.NotNil(h)
	h.RegisterInterceptor(syntax.MatchWord, "\\d+")
	h.Add("{sub:\\d+}.example.com")

	r := rest.Get(a, "http://sub--1.example.com/test").Request()
	ps := types.NewContext()
	a.False(h.Match(r, ps))

	// 将 \\d+ 注册为任意非空字符
	r = rest.Get(a, "http://sub.example.com/test").Request()
	ps = types.NewContext()
	ok := h.Match(r, ps)
	a.True(ok).Equal(ps.MustString("sub", "def"), "sub")
}

func TestHosts_Match(t *testing.T) {
	a := assert.New(t, false)

	h := NewHosts(true, "caixw.io", "caixw.oi", "{sub}.example.com")
	a.NotNil(h)

	r := rest.Get(a, "http://caixw.io/test").Request()
	ps := types.NewContext()
	a.True(h.Match(r, ps)).Zero(ps.Count())

	r = rest.Get(a, "https://caixw.io/test").Request()
	ps = types.NewContext()
	a.True(h.Match(r, ps)).Zero(ps.Count())

	r = rest.Get(a, "https://CAIXW.io/test").Request()
	ps = types.NewContext()
	a.True(h.Match(r, ps)).Zero(ps.Count())

	// 泛域名
	r = rest.Get(a, "https://xx.example.com/test").Request()
	ps = types.NewContext()
	a.True(h.Match(r, ps)).Equal(ps.MustString("sub", "yy"), "xx")

	// 泛域名
	r = rest.Get(a, "https://xx.yy.example.com/test").Request()
	ps = types.NewContext()
	a.True(h.Match(r, ps)).Equal(ps.MustString("sub", "yy"), "xx.yy")

	// 带端口
	r = rest.Get(a, "http://caixw.io:88/test").Request()
	ps = types.NewContext()
	a.True(h.Match(r, ps)).Zero(ps.Count())

	// 访问不允许的域名
	r = rest.Get(a, "http://sub.caixw.io/test").Request()
	ps = types.NewContext()
	a.False(h.Match(r, ps))

	// 访问不允许的域名
	r = rest.Get(a, "http://sub.1eample.com/test").Request()
	ps = types.NewContext()
	a.False(h.Match(r, ps))
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
	r := rest.Get(a, "https://xx.example.com/api/path").Request()
	ps := types.NewContext()
	a.True(h.Match(r, ps)).Zero(ps.Count())

	// 删除 xx.example.com，则适配到 {sub}.example.com
	h.Delete("xx.example.com")
	r = rest.Get(a, "https://xx.example.com/api/path").Request()
	ps = types.NewContext()
	a.True(h.Match(r, ps))

	// delete {sub}.example.com
	h.Delete("{sub}.example.com")
	r = rest.Get(a, "https://zzz.example.com/api/path").Request()
	ps = types.NewContext()
	a.False(h.Match(r, ps))
}
