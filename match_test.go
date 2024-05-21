// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package mux

import (
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/assert/v4/rest"

	"github.com/issue9/mux/v9/header"
	"github.com/issue9/mux/v9/internal/syntax"
	"github.com/issue9/mux/v9/types"
)

var (
	_ Matcher = &Hosts{}
	_ Matcher = &headerVersion{}
	_ Matcher = &pathVersion{}
)

func TestAndMatcherFunc(t *testing.T) {
	a := assert.New(t, false)

	p1 := NewPathVersion("p1", "v1")
	p2 := NewPathVersion("p2", "v2")

	m := AndMatcherFunc(p1.Match, p2.Match)
	r := rest.Get(a, "/v1/v2/path").Request()
	ps := types.NewContext()
	ok := m.Match(r, ps)
	a.True(ok).Equal(r.URL.Path, "/path")

	m = AndMatcherFunc(p1.Match, p2.Match)
	r = rest.Get(a, "/v2/v1/path").Request()
	ps = types.NewContext()
	ok = m.Match(r, ps)
	a.False(ok)
}

func TestOrMatcherFunc(t *testing.T) {
	a := assert.New(t, false)

	p1 := NewPathVersion("p1", "v1")
	p2 := NewPathVersion("p2", "v2")

	m := OrMatcherFunc(p1.Match, p2.Match)
	r := rest.Get(a, "/v1/v2/path").Request()
	ps := types.NewContext()
	ok := m.Match(r, ps)
	a.True(ok).Equal(r.URL.Path, "/v2/path")

	m = OrMatcherFunc(p1.Match, p2.Match)
	r = rest.Get(a, "/v2/v1/path").Request()
	ps = types.NewContext()
	ok = m.Match(r, ps)
	a.True(ok).Equal(r.URL.Path, "/v1/path")

	m = OrMatcherFunc(p1.Match, p2.Match)
	r = rest.Get(a, "/v111/v2/v1/path").Request()
	ps = types.NewContext()
	ok = m.Match(r, ps)
	a.False(ok)
}

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

func TestHeaderVersion_Match(t *testing.T) {
	a := assert.New(t, false)

	h := NewHeaderVersion("version", "", nil, "1.0", "2.0", "3.0")

	// 相同版本号
	r := rest.Get(a, "https://caixw.io/test").
		Header(header.Accept, "application/json; version=1.0").
		Request()
	ps := types.NewContext()
	ok := h.Match(r, ps)
	a.True(ok).Equal(ps.MustString("version", "not-exists"), "1.0")

	// 相同版本号，未指定 paramName
	h = NewHeaderVersion("", "", nil, "1.0", "2.0", "3.0")
	r = rest.Get(a, "https://caixw.io/test").
		Header(header.Accept, "application/json; version=1.0").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.True(ok)

	// 空版本
	r = rest.Get(a, "https://not.exists/test").
		Header(header.Accept, "application/json; version=").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)

	// 不同版本
	r = rest.Get(a, "https://not.exists/test").
		Header(header.Accept, "application/json; version = 2").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)

	// 未指定版本
	r = rest.Get(a, "https://caixw.io/test").
		Header(header.Accept, "application/json").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)

	// 未指定 media type
	r = rest.Get(a, "https://caixw.io/test").
		Header(header.Accept, ";version=1.0").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)

	// 未指定 Accept
	r = rest.Get(a, "https://caixw.io/test").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)

	// 空值，不匹配任何内容

	h = NewHeaderVersion("", "", nil)
	r = rest.Get(a, "https://caixw.io/test").
		Header(header.Accept, "application/json; version=1.0").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)

	// 空版本
	r = rest.Get(a, "https://not.exists/test").
		Header(header.Accept, "application/json; version=").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)

	// 不同版本
	r = rest.Get(a, "https://not.exists/test").
		Header(header.Accept, "application/json; version=2").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)

	r = rest.Get(a, "https://caixw.io/test").
		Header(header.Accept, "application/json; version=1.0").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)
}

func TestPathVersion_Match(t *testing.T) {
	a := assert.New(t, false)

	h := NewPathVersion("version", "v3", "/v2", "/v1")

	a.Panic(func() {
		NewPathVersion("version", "", "v3")
	})

	// 相同版本号
	r := rest.Get(a, "https://caixw.io/v1/test").Request()
	ps := types.NewContext()
	ok := h.Match(r, ps)
	a.True(ok)
	a.Equal(r.URL.Path, "/test").
		Equal(ps.MustString("version", "not-found"), "/v1")

	// 相同版本号，未指定 key
	h = NewPathVersion("", "v3", "/v2", "/v1")
	r = rest.Get(a, "https://caixw.io/v1/test").Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.True(ok)
	a.Equal(r.URL.Path, "/test")

	// 空版本
	r = rest.Get(a, "https://caixw.io/test").Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)
	a.Equal(r.URL.Path, "/test")

	// 不同版本
	r = rest.Get(a, "https://caixw.io/v111/test").Request()
	a.NotNil(r)
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)
	a.Equal(r.URL.Path, "/v111/test")

	// 空值，不匹配任何内容

	h = NewPathVersion("version")
	r = rest.Get(a, "https://caixw.io/test").
		Header(header.Accept, "application/json; version=1.0").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)
}
