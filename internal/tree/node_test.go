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

// node 的测试工具
type nodeTest struct {
	n *node
	a *assert.Assertion
}

func newNodeTest(a *assert.Assertion) *nodeTest {
	return &nodeTest{
		n: &node{},
		a: a,
	}
}

// 添加一条路由项。code 表示该路由项返回的报头，
// 测试路由项的 code 需要唯一，之后也是通过此值来判断其命中的路由项。
func (n *nodeTest) add(pattern string, code int, method string) {
	segs := newSegments(n.a, pattern)
	n.a.NotError(n.n.add(segs, buildHandler(code), method))
}

// 验证指定的路径是否匹配正确的路由项，通过 code 来确定
func (n *nodeTest) matchTrue(path string, code int, method string) {
	nn := n.n.match(path)
	n.a.NotNil(nn)
	n.a.NotNil(nn.handlers)

	h := nn.handlers.handler(method)
	n.a.NotNil(h)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, path, nil)
	h.ServeHTTP(w, r)
	n.a.Equal(w.Code, code)
}

func newSegments(a *assert.Assertion, pattern string) []*ts.Segment {
	ss, err := ts.Parse(pattern)
	a.NotError(err).NotNil(ss)

	return ss
}

func TestNode_add_remove(t *testing.T) {
	a := assert.New(t)
	node := &node{}

	a.NotError(node.add(newSegments(a, "/"), buildHandler(1), http.MethodGet))
	a.NotError(node.add(newSegments(a, "/posts/{id}"), buildHandler(1), http.MethodGet))
	a.NotError(node.add(newSegments(a, "/posts/{id}/author"), buildHandler(1), http.MethodGet))
	a.NotError(node.add(newSegments(a, "/posts/1/author"), buildHandler(1), http.MethodGet))
	a.NotError(node.add(newSegments(a, "/posts/{id}/{author:\\w+}/profile"), buildHandler(1), http.MethodGet))

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
	test := newNodeTest(a)

	// 添加路由项
	test.add("/", 1, http.MethodGet)
	test.add("/posts/{id}", 2, http.MethodGet)
	test.add("/posts/{id}/author", 3, http.MethodGet)
	test.add("/posts/1/author", 4, http.MethodGet)

	test.matchTrue("/", 1, http.MethodGet)
	test.matchTrue("/posts/1", 2, http.MethodGet)
	test.matchTrue("/posts/2", 2, http.MethodGet)
	test.matchTrue("/posts/2/author", 3, http.MethodGet)
	test.matchTrue("/posts/1/author", 4, http.MethodGet)
}

func TestNode_getParents(t *testing.T) {
	a := assert.New(t)

	n1 := &node{
		children: make([]*node, 0, 1),
	}
	n2 := &node{
		children: make([]*node, 0, 1),
		parent:   n1,
	}
	n3 := &node{
		parent: n2,
	}

	a.Equal(n3.getParents(), []*node{n3, n2, n1})
	a.Equal(n2.getParents(), []*node{n2, n1})
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
