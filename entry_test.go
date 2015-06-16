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

	// 静态路由
	e := newEntry("/blog/post/1", hf, nil)
	a.Equal(e.match("/blog/post/1"), 0)

	a.Equal(e.match("/blog/post/1/page/2"), 7)

	// 不匹配，长度太短
	a.Equal(e.match("/blog"), -1)

	// 正则路由
	e = newEntry("/blog/post/{id}", hf, nil)
	a.Equal(e.match("/blog/post/1"), 0)

	a.Equal(e.match("/blog/post/2/page/1"), 7)

	// 不匹配
	a.Equal(e.match("/plog/post/2"), -1)

	// 多个命名正则表达式
	e = newEntry("/blog/{action:\\w+}-{id:\\d+}/", hf, nil)
	a.Equal(e.match("/blog/post-1/page-2"), 6)

	// 正则，不匹配
	a.Equal(e.match("/blog/post-1d/"), -1)
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
