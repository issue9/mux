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

	fn := func(str, pattern string, hasParams bool) {
		strs := split(str)
		p, b := toPattern(strs)
		a.Equal(p, pattern).Equal(b, hasParams)
	}

	fn("/blog/post/{id}", "/blog/post/(?P<id>[^/]+)", true)
	fn("/blog/post/{id:\\d+}", "/blog/post/(?P<id>\\d+)", true)
	fn("/blog/{post}-{id}", "/blog/(?P<post>[^/]+)-(?P<id>[^/]+)", true)
	fn("/blog/{:\\w+}-{id}", "/blog/\\w+-(?P<id>[^/]+)", true)
}

func TestNewEntry(t *testing.T) {
	a := assert.New(t)
	h := func(w http.ResponseWriter, r *http.Request) {
	}
	hf := http.HandlerFunc(h)

	// 静态路由
	e := newEntry("/blog/post/1", hf)
	r := e.match("/blog/post/1")
	a.Equal(r, 0).Nil(e.getParams("/blog/post/1"))

	r = e.match("/blog/post/1/page/2")
	a.Equal(r, 7).Nil(e.getParams("/blog/post/1/page/2"))

	// 正则路由
	e = newEntry("/blog/post/{id}", hf)
	r = e.match("/blog/post/1")
	a.Equal(r, 0).Equal(e.getParams("/blog/post/1"), map[string]string{"id": "1"})

	r = e.match("/blog/post/2/page/1")
	a.Equal(r, 7).Equal(e.getParams("/blog/post/2/page/1"), map[string]string{"id": "2"})

	r = e.match("/plog/post/2") // 不匹配
	a.Equal(r, -1).Equal(0, len(e.getParams("/plog/post/2")))

	// 多个命名正则表达式
	e = newEntry("/blog/{action:\\w+}-{id:\\d+}/", hf)
	r = e.match("/blog/post-1/page-2")
	a.Equal(r, 6).Equal(e.getParams("/blog/post-1/page-2"), map[string]string{"action": "post", "id": "1"})

	r = e.match("/blog/post-1d/") // 不匹配
	a.Equal(r, -1).Equal(0, len(e.getParams("/blog/post-1d/")))
}
