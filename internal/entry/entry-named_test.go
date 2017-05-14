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
		nType:     TypeNamed,
		patterns:  []string{"/posts/", "{id}"},
	})
	a.NotNil(n)
	a.Equal(n.pattern, pattern)
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
		nType:     TypeNamed,
		patterns:  []string{"/posts/", "{id}", "/page/", "{page}"},
	})
	a.NotNil(n)
	a.Equal(n.pattern, pattern)
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

	newMatcher(a, "/posts/{id}").
		True("/posts/1", map[string]string{"id": "1"}).
		False("/posts", nil).
		True("/posts/id.html", map[string]string{"id": "id.html"}).
		False("/posts/id.html/", nil).
		False("/posts/id.html/page", nil)

	newMatcher(a, "/posts/{id}/page/{page}").
		True("/posts/1/page/1", map[string]string{"id": "1", "page": "1"}).
		False("/posts/1", nil).
		True("/posts/1.html/page/1", map[string]string{"id": "1.html", "page": "1"}).
		False("/posts/id-1/page/1/", nil).
		False("/posts/id-1/page/1/size/1", nil)

	newMatcher(a, "/posts/{id}-{page}").
		True("/posts/1-1", map[string]string{"id": "1", "page": "1"}).
		True("/posts/1.html-1", map[string]string{"id": "1.html", "page": "1"}).
		False("/posts/id-11/", nil).
		False("/posts/id-1/size/1", nil)

	newMatcher(a, "/users/{user}/{repos}/pulls").
		False("/users/user/repos/pulls/number", nil).
		False("/users/user/repos/pullsnumber", nil)

	newMatcher(a, "/users/{user}/repos/{pulls}").
		False("/users/user/repos/pulls/number", nil)
}

func TestNamed_match_wildcard(t *testing.T) {
	a := assert.New(t)

	newMatcher(a, "/posts/{id}/*").
		False("/posts/1", nil).
		False("/posts", nil).
		True("/posts/2/", map[string]string{"id": "2"}).
		True("/posts/id.html/index.html", map[string]string{"id": "id.html"})

	newMatcher(a, "/posts/{id}/page/{page}/*").
		False("/posts/1/page/1", nil).
		True("/posts/1.html/page/1/", map[string]string{"id": "1.html", "page": "1"}).
		True("/posts/id-1/page/1/index.html", map[string]string{"id": "id-1", "page": "1"})

	newMatcher(a, "/posts/{id}-{page}/*").
		False("/posts/1-1", nil).
		True("/posts/1.html-1/", map[string]string{"id": "1.html", "page": "1"}).
		True("/posts/id-1/index.html", map[string]string{"id": "id", "page": "1"})
}

func TestNamed_URL(t *testing.T) {
	a := assert.New(t)
	n, err := New("/posts/{id}")
	a.NotError(err).NotNil(n)
	url, err := n.URL(map[string]string{"id": "5.html"}, "path")
	a.NotError(err).Equal(url, "/posts/5.html")
	url, err = n.URL(map[string]string{"id": "5.html/"}, "path")
	a.NotError(err).Equal(url, "/posts/5.html/")

	n, err = New("/posts/{id}/page/{page}")
	a.NotError(err).NotNil(n)
	url, err = n.URL(map[string]string{"id": "5.html", "page": "1"}, "path")
	a.NotError(err).Equal(url, "/posts/5.html/page/1")

	// 少参数
	url, err = n.URL(map[string]string{"id": "5.html"}, "path")
	a.Error(err).Equal(url, "")

	// 带通配符
	n, err = New("/posts/{id}/page/{page}/*")
	a.NotError(err).NotNil(n)
	url, err = n.URL(map[string]string{"id": "5.html", "page": "1"}, "path")
	a.NotError(err).Equal(url, "/posts/5.html/page/1/path")

	// 指定了空的 path 参数
	url, err = n.URL(map[string]string{"id": "5.html", "page": "1"}, "")
	a.NotError(err).Equal(url, "/posts/5.html/page/1/")
}
