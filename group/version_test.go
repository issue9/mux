// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package group

import (
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/assert/v4/rest"

	"github.com/issue9/mux/v8/types"
)

var (
	_ Matcher = &HeaderVersion{}
	_ Matcher = &PathVersion{}
)

func TestHeaderVersion_Match(t *testing.T) {
	a := assert.New(t, false)

	h := NewHeaderVersion("version", "", nil, "1.0", "2.0", "3.0")

	// 相同版本号
	r := rest.Get(a, "https://caixw.io/test").
		Header("Accept", "application/json; version=1.0").
		Request()
	ps := types.NewContext()
	ok := h.Match(r, ps)
	a.True(ok).Equal(ps.MustString("version", "not-exists"), "1.0")

	// 相同版本号，未指定 paramName
	h = NewHeaderVersion("", "", nil, "1.0", "2.0", "3.0")
	r = rest.Get(a, "https://caixw.io/test").
		Header("Accept", "application/json; version=1.0").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.True(ok)

	// 空版本
	r = rest.Get(a, "https://not.exists/test").
		Header("Accept", "application/json; version=").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)

	// 不同版本
	r = rest.Get(a, "https://not.exists/test").
		Header("Accept", "application/json; version = 2").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)

	// 未指定版本
	r = rest.Get(a, "https://caixw.io/test").
		Header("Accept", "application/json").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)

	// 未指定 media type
	r = rest.Get(a, "https://caixw.io/test").
		Header("Accept", ";version=1.0").
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
		Header("Accept", "application/json; version=1.0").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)

	// 空版本
	r = rest.Get(a, "https://not.exists/test").
		Header("Accept", "application/json; version=").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)

	// 不同版本
	r = rest.Get(a, "https://not.exists/test").
		Header("Accept", "application/json; version=2").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)

	r = rest.Get(a, "https://caixw.io/test").
		Header("Accept", "application/json; version=1.0").
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
		Header("Accept", "application/json; version=1.0").
		Request()
	ps = types.NewContext()
	ok = h.Match(r, ps)
	a.False(ok)
}
