// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	ts "github.com/issue9/mux/internal/tree/syntax"
)

func newSegments(a *assert.Assertion, pattern string) []*ts.Segment {
	ss, err := ts.Parse(pattern)
	a.NotError(err).NotNil(ss)

	return ss
}

func TestNode_add_remove(t *testing.T) {
	a := assert.New(t)
	node := &node{}

	a.NotError(node.add(newSegments(a, "/"), h1, http.MethodGet))
	a.NotError(node.add(newSegments(a, "/posts/{id}"), h1, http.MethodGet))
	a.NotError(node.add(newSegments(a, "/posts/{id}/author"), h1, http.MethodGet))
	a.NotError(node.add(newSegments(a, "/posts/1/author"), h1, http.MethodGet))

	// / 和 /posts/ 以及 /posts/1/author
	a.Equal(len(node.children), 3)
	node.print(0)

	node.remove(newSegments(a, "/posts/1/author"), http.MethodGet)
	// / 和 /posts/
	a.Equal(len(node.children), 2)

	node.remove(newSegments(a, "/posts/{id}/author"), http.MethodGet)
	// / 和 /posts/
	a.Equal(len(node.children), 2)
}

func TestNode_match(t *testing.T) {
	a := assert.New(t)
	node := &node{}

	// 添加路由项
	a.NotError(node.add(newSegments(a, "/"), h1, http.MethodGet))
	a.NotError(node.add(newSegments(a, "/posts/{id}"), h2, http.MethodGet))
	a.NotError(node.add(newSegments(a, "/posts/{id}/author"), h3, http.MethodGet))
	a.NotError(node.add(newSegments(a, "/posts/1/author"), h4, http.MethodGet))

	test := func(path, method string, code int) {
		n := node.match(path)
		a.NotNil(n)
		a.NotNil(n.handlers)
		h := n.handlers.handler(method)
		a.NotNil(h)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(method, path, nil)
		h.ServeHTTP(w, r)
		a.Equal(w.Code, code)
	}

	test("/", http.MethodGet, 1)
	test("/posts/1", http.MethodGet, 2)
	test("/posts/2", http.MethodGet, 2)
	test("/posts/2/author", http.MethodGet, 3)
	test("/posts/1/author", http.MethodGet, 4)
}

func TestRemoveNoddes(t *testing.T) {
	a := assert.New(t)

	n1 := &node{pattern: "/1"}
	n2 := &node{pattern: "/2"}
	n3 := &node{pattern: "/3"}
	n4 := &node{pattern: "/4"}
	nodes := []*node{n1, n2, n3, n4}

	// 不存在的元素
	nodes = removeNodes(nodes, "")
	a.Equal(len(nodes), 4)

	// 删除尾元素
	nodes = removeNodes(nodes, "/4")
	a.Equal(len(nodes), 3)

	// 删除中间无素
	nodes = removeNodes(nodes, "/2")
	a.Equal(len(nodes), 2)

	// 已删除，不存在的元素
	nodes = removeNodes(nodes, "/2")
	a.Equal(len(nodes), 2)

	// 第一个元素
	nodes = removeNodes(nodes, "/1")
	a.Equal(len(nodes), 1)

	// 最后一个元素
	nodes = removeNodes(nodes, "/3")
	a.Equal(len(nodes), 0)
}
