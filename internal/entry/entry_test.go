// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestEntry_Match(t *testing.T) {
	a := assert.New(t)
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})

	// 静态路由-1
	e, err := New("/blog/post/1", hf)
	a.NotError(err)
	a.Equal(e.Match("/blog/post/1"), 0)
	a.Equal(e.Match("/blog/post/1/"), -1)
	a.Equal(e.Match("/blog/post/1/page/2"), -1) // 非 / 结尾的，只能全路径匹配。
	a.Equal(e.Match("/blog"), -1)               // 不匹配，长度太短

	// basic with wildcard
	e, err = New("/blog/post/*", hf)
	a.NotError(err)
	a.Equal(e.Match("/blog/post/1"), 0)
	a.Equal(e.Match("/blog/post/"), 0)
	a.Equal(e.Match("/blog/post/1/page/2"), 0)
	a.Equal(e.Match("/blog"), -1) // 不匹配，长度太短

	// 命名路由
	e, err = New("/blog/post/{id}", hf)
	a.NotError(err)
	a.Equal(e.Match("/blog/post/1"), 0)
	a.Equal(e.Match("/blog/post/2/page/1"), -1) // 不匹配
	a.Equal(e.Match("/plog/post/2"), -1)        // 不匹配

	// 多个命名正则表达式
	e, err = New("/blog/{action:\\w+}-{id:\\d+}/", hf)
	a.NotError(err)
	a.Equal(e.Match("/blog/post-1/"), 0)
	a.Equal(e.Match("/blog/post-1/page-2"), -1) // 正则没有部分匹配功能
	a.Equal(e.Match("/blog/post-1d/"), -1)      // 正则，不匹配

	// 多个命名正则表达式，带可选参数
	e, err = New("/blog/{action:\\w+}-{id:\\d*}/", hf)
	a.NotError(err)
	a.Equal(e.Match("/blog/post-/"), 0)
	a.Equal(e.Match("/blog/post-1/"), 0)
}

func TestEntry_Params(t *testing.T) {
	a := assert.New(t)
	h := func(w http.ResponseWriter, r *http.Request) {
	}
	hf := http.HandlerFunc(h)

	// 静态路由
	e, err := New("/blog/post/1", hf)
	a.NotError(err)
	a.Nil(e.Params("/blog/post/1"))
	a.Nil(e.Params("/blog/post/1/page/2"))
	a.Nil(e.Params("/blog"))

	// 命名路由
	e, err = New("/blog/post/{id}", hf)
	a.NotError(err)
	a.Equal(0, len(e.Params("/plog/post/2"))) // 不匹配
	a.Equal(e.Params("/blog/post/1"), map[string]string{"id": "1"})

	// 多个命名正则表达式
	e, err = New("/blog/{action:\\w+}-{id:\\d+}/", hf)
	a.NotError(err)
	a.Equal(e.Params("/blog/post-1/page-2"), map[string]string{"action": "post", "id": "1"})
	a.Equal(0, len(e.Params("/blog/post-1d/"))) // 不匹配
}
