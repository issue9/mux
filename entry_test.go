// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestSplit(t *testing.T) {
	a := assert.New(t)

	a.Equal(split("/blog/post/1"), []string{"/blog/post/1"})
	a.Equal(split("/blog/post/{id}"), []string{"/blog/post/", "{id}"})
	a.Equal(split("/blog/post/{id:\\d}"), []string{"/blog/post/", "{id:\\d}"})
	a.Equal(split("/blog/{post}/{id}"), []string{"/blog/", "{post}", "/", "{id}"})
	a.Equal(split("/blog/{post}-{id}"), []string{"/blog/", "{post}", "-", "{id}"})

	a.Equal(split("{/blog/post/{id}"), []string{"{/blog/post/{id}"})
	a.Equal(split("}/blog/post/{id}"), []string{"}/blog/post/", "{id}"})
}

func TestToPattern(t *testing.T) {
	a := assert.New(t)

	fn := func(str []string, pattern string, hasParams bool) {
		p, b := toPattern(str)
		a.Equal(p, pattern).Equal(b, hasParams)
	}

	fn([]string{"/blog/post/1"}, "/blog/post/1", false)              // 静态
	fn([]string{"/blog/post/", "{:\\d+}"}, "/blog/post/\\d+", false) // 无命名路由参数

	fn([]string{"/blog/post/", "{id}"}, "/blog/post/(?P<id>[^/]+)", true)
	fn([]string{"/blog/post/", "{id:\\d+}"}, "/blog/post/(?P<id>\\d+)", true)
	fn([]string{"/blog/", "{post}", "-", "{id}"}, "/blog/(?P<post>[^/]+)-(?P<id>[^/]+)", true)
	fn([]string{"/blog/", "{:\\w+}", "-", "{id}"}, "/blog/\\w+-(?P<id>[^/]+)", true)
}

func TestEntry_match(t *testing.T) {
	a := assert.New(t)
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})

	// 静态路由-1
	e := newEntry("/blog/post/1", hf, nil)
	a.Equal(e.match("/blog/post/1"), 0)
	a.Equal(e.match("/blog/post/1/"), -1)
	a.Equal(e.match("/blog/post/1/page/2"), -1) // 非/结尾的，只能全路径匹配。
	a.Equal(e.match("/blog"), -1)               // 不匹配，长度太短

	// 静态路由-2
	e = newEntry("/blog/post/", hf, nil)
	a.Equal(e.match("/blog/post/1"), 1)
	a.Equal(e.match("/blog/post/1/page/2"), 8)
	a.Equal(e.match("/blog"), -1) // 不匹配，长度太短

	// 正则路由
	e = newEntry("/blog/post/{id}", hf, nil)
	a.Equal(e.match("/blog/post/1"), 0)
	a.Equal(e.match("/blog/post/2/page/1"), -1) // 正则没有部分匹配
	a.Equal(e.match("/plog/post/2"), -1)        // 不匹配

	// 多个命名正则表达式
	e = newEntry("/blog/{action:\\w+}-{id:\\d+}/", hf, nil)
	a.Equal(e.match("/blog/post-1/page-2"), -1) // 正则没有部分匹配功能
	a.Equal(e.match("/blog/post-1d/"), -1)      // 正则，不匹配

	// 通过group控制
	g := &Group{isRunning: false}
	e = newEntry("/blog/post/1", hf, g)
	a.Equal(e.match("/blog/post/1"), -1)        // 暂停状态下，均返回-1
	a.Equal(e.match("/blog/post/1/page/2"), -1) // 暂停状态下，均返回-1
	a.Equal(e.match("/blog"), -1)               // 暂停状态下，均返回-1
	g.isRunning = true
	a.Equal(e.match("/blog/post/1"), 0)
	a.Equal(e.match("/blog/post/1/page/2"), -1)

}

func TestEntry_getParams(t *testing.T) {
	a := assert.New(t)
	h := func(w http.ResponseWriter, r *http.Request) {
	}
	hf := http.HandlerFunc(h)

	// 静态路由
	e := newEntry("/blog/post/1", hf, nil)
	a.Nil(e.getParams("/blog/post/1"))
	a.Nil(e.getParams("/blog/post/1/page/2"))
	a.Nil(e.getParams("/blog"))

	// 正则路由
	e = newEntry("/blog/post/{id}", hf, nil)
	a.Equal(0, len(e.getParams("/plog/post/2")))             // 不匹配
	a.Equal(e.getParams("/blog/post/"), map[string]string{}) // 匹配，但未指定参数，默认为空
	a.Equal(e.getParams("/blog/post/1"), map[string]string{"id": "1"})
	a.Equal(e.getParams("/blog/post/2/page/1"), map[string]string{"id": "2"})

	// 多个命名正则表达式
	e = newEntry("/blog/{action:\\w+}-{id:\\d+}/", hf, nil)
	a.Equal(e.getParams("/blog/post-1/page-2"), map[string]string{"action": "post", "id": "1"})
	a.Equal(0, len(e.getParams("/blog/post-1d/"))) // 不匹配
}

// BenchmarkEntry_Match_Basic-4       	200000000	         6.57 ns/op    go1.6
func BenchmarkEntry_Match_Basic(b *testing.B) {
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
	e := newEntry("/blog/post/1", hf, nil)
	for i := 0; i < b.N; i++ {
		if 0 != e.match("/blog/post/1") {
			b.Error("BenchmarkEntry_Match_Basic:error")
		}
	}
}

// BenchmarkEntry_Match_Static-4      	100000000	        10.4 ns/op    go1.6
func BenchmarkEntry_Match_Static(b *testing.B) {
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
	e := newEntry("/blog/post/", hf, nil)
	for i := 0; i < b.N; i++ {
		if e.match("/blog/post/1") > 1 {
			b.Error("BenchmarkEntry_Match_Static:error")
		}
	}
}

// BenchmarkEntry_Match_Regexp-4      	 3000000	       436 ns/op    go1.6
func BenchmarkEntry_Match_Regexp(b *testing.B) {
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
	e := newEntry("/blog/post/{id:\\d+}", hf, nil)
	for i := 0; i < b.N; i++ {
		if 0 != e.match("/blog/post/1") {
			b.Error("BenchmarkEntry_Match_Regexp:error")
		}
	}
}
