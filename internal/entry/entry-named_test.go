// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"testing"

	"github.com/issue9/assert"
)

var _ Entry = &named{}

func TestNewNammed(t *testing.T) {
	a := assert.New(t)

	pattern := "/posts/{id}"
	n := newNamed(pattern, &syntax{
		hasParams: true,
		nType:     typeNamed,
		patterns:  []string{"/posts/", "{id}"},
	})
	a.NotNil(n)
	a.Equal(n.patternString, pattern)
	a.Equal(len(n.nodes), 2)
	n0 := n.nodes[0]
	a.True(n0.isString).Equal(n0.value, "/posts/")
	n1 := n.nodes[1]
	a.False(n1.isString).
		Equal(n1.value, "id").
		Equal(n1.endByte, '/')

	pattern = "/posts/{id}/page/{page}"
	n = newNamed(pattern, &syntax{
		hasParams: true,
		nType:     typeNamed,
		patterns:  []string{"/posts/", "{id}", "/page/", "{page}"},
	})
	a.NotNil(n)
	a.Equal(n.patternString, pattern)
	a.Equal(len(n.nodes), 4)
	n0 = n.nodes[0]
	a.True(n0.isString).Equal(n0.value, "/posts/")
	n1 = n.nodes[1]
	a.False(n1.isString).
		Equal(n1.value, "id").
		Equal(n1.endByte, '/')
	n3 := n.nodes[3]
	a.False(n3.isString).
		Equal(n3.value, "page").
		Equal(n3.endByte, '/')
}

func TestNamed_match(t *testing.T) {
	a := assert.New(t)

	n, err := New("/posts/{id}", nil)
	a.NotError(err).NotNil(n)

	a.True(n.match("/posts/1"))
	a.True(n.match("/posts/2"))
	a.True(n.match("/posts/id"))
	a.True(n.match("/posts/id.html"))
	a.False(n.match("/posts/id.html/"))
	a.False(n.match("/posts/id.html/page"))
	a.False(n.match("/post/id"))

	n, err = New("/posts/{id}/page/{page}", nil)
	a.NotError(err).NotNil(n)
	a.True(n.match("/posts/1/page/1"))
	a.True(n.match("/posts/1.html/page/1"))
	a.False(n.match("/posts/id-1/page/1/"))
	a.False(n.match("/posts/id-1/page/1/size/1"))

	n, err = New("/posts/{id}-{page}", nil)
	a.NotError(err).NotNil(n)
	a.True(n.match("/posts/1-1"))
	a.True(n.match("/posts/1.html-1"))
	a.False(n.match("/posts/id-11/"))
	a.False(n.match("/posts/id-1/size/1"))
}

func TestNamed_match_wildcard(t *testing.T) {
	a := assert.New(t)

	n, err := New("/posts/{id}/*", nil)
	a.NotError(err).NotNil(n)
	a.False(n.match("/posts/1"))
	a.True(n.match("/posts/2/"))
	a.True(n.match("/posts/id/index.html"))
	a.True(n.match("/posts/id.html/index.html"))

	n, err = New("/posts/{id}/page/{page}/*", nil)
	a.NotError(err).NotNil(n)
	a.False(n.match("/posts/1/page/1"))
	a.True(n.match("/posts/1.html/page/1/"))
	a.True(n.match("/posts/id-1/page/1/index.html"))

	n, err = New("/posts/{id}/-{page}/*", nil)
	a.NotError(err).NotNil(n)
	a.False(n.match("/posts/1-1"))
	a.True(n.match("/posts/1.html-1/"))
	a.True(n.match("/posts/id-1/index.html"))
}

func TestNamed_Params(t *testing.T) {
	a := assert.New(t)
	n, err := New("/posts/{id}", nil)
	a.NotError(err).NotNil(n)
	a.Equal(n.Params("/posts/1"), map[string]string{"id": "1"})
	a.Equal(n.Params("/posts/1.html"), map[string]string{"id": "1.html"})
	a.Equal(len(n.Params("/posts/1.html/")), 0)

	n, err = New("/posts/{id}/page/{page}", nil)
	a.NotError(err).NotNil(n)
	a.Equal(n.Params("/posts/1/page/1"), map[string]string{"id": "1", "page": "1"})
	a.Equal(n.Params("/posts/1.html/page/1"), map[string]string{"id": "1.html", "page": "1"})
	a.Nil(n.Params("/posts/1.html/"))
}

func TestNamed_URL(t *testing.T) {
	a := assert.New(t)
	n, err := New("/posts/{id}", nil)
	a.NotError(err).NotNil(n)
	url, err := n.URL(map[string]string{"id": "5.html"}, "path")
	a.NotError(err).Equal(url, "/posts/5.html")
	url, err = n.URL(map[string]string{"id": "5.html/"}, "path")
	a.NotError(err).Equal(url, "/posts/5.html/")

	n, err = New("/posts/{id}/page/{page}", nil)
	a.NotError(err).NotNil(n)
	url, err = n.URL(map[string]string{"id": "5.html", "page": "1"}, "path")
	a.NotError(err).Equal(url, "/posts/5.html/page/1")

	// 少参数
	url, err = n.URL(map[string]string{"id": "5.html"}, "path")
	a.Error(err).Equal(url, "")

	// 带通配符
	n, err = New("/posts/{id}/page/{page}/*", nil)
	a.NotError(err).NotNil(n)
	url, err = n.URL(map[string]string{"id": "5.html", "page": "1"}, "path")
	a.NotError(err).Equal(url, "/posts/5.html/page/1/path")

	// 指定了空的 path 参数
	url, err = n.URL(map[string]string{"id": "5.html", "page": "1"}, "")
	a.NotError(err).Equal(url, "/posts/5.html/page/1/")
}
