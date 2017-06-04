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
func (n *nodeTest) add(method, pattern string, code int) {
	segs := newSegments(n.a, pattern)
	n.a.NotError(n.n.add(segs, buildHandler(code), method))
}

// 验证指定的路径是否匹配正确的路由项，通过 code 来确定，并返回该节点的实例。
func (n *nodeTest) matchTrue(method, path string, code int) *node {
	nn := n.n.match(path)
	n.a.NotNil(nn)

	h := nn.handler(method)
	n.a.NotNil(h)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, path, nil)
	h.ServeHTTP(w, r)
	n.a.Equal(w.Code, code)

	return nn
}

// 验证指定的路径返回的参数是否正确
func (n *nodeTest) paramsTrue(method, path string, code int, params map[string]string) {
	nn := n.matchTrue(method, path, code)

	ps := nn.params(path)
	n.a.Equal(ps, params)
}

// 验证 node.url 的正确性
// method+path 用于获取指定的节点
func (n *nodeTest) urlTrue(method, path string, code int, params map[string]string, url string) {
	nn := n.matchTrue(method, path, code)

	u, err := nn.url(params)
	n.a.NotError(err)
	n.a.Equal(u, url)
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
	//node.print(0)

	node.remove(newSegments(a, "/posts/1/author"), http.MethodGet)
	// / 和 /posts/
	a.Equal(len(node.children), 2)

	node.remove(newSegments(a, "/posts/{id}/author"), http.MethodGet)
	// / 和 /posts/
	a.Equal(len(node.children), 2)
}

func TestNode_clean(t *testing.T) {
	a := assert.New(t)
	node := &node{}

	a.NotError(node.add(newSegments(a, "/"), buildHandler(1), http.MethodGet))
	a.NotError(node.add(newSegments(a, "/posts/1/author"), buildHandler(1), http.MethodGet))
	a.NotError(node.add(newSegments(a, "/posts/{id}"), buildHandler(1), http.MethodGet))
	a.NotError(node.add(newSegments(a, "/posts/{id}/author"), buildHandler(1), http.MethodGet))
	a.NotError(node.add(newSegments(a, "/posts/{id}/{author:\\w+}/profile"), buildHandler(1), http.MethodGet))

	a.Equal(node.len(), 5)

	node.clean("/posts/{id")
	a.Equal(node.len(), 2)

	node.clean("")
	a.Equal(node.len(), 0)
}

func TestNode_match(t *testing.T) {
	a := assert.New(t)
	test := newNodeTest(a)

	// 添加路由项
	test.add(http.MethodGet, "/", 1)
	test.add(http.MethodGet, "/posts/{id}", 2)
	test.add(http.MethodGet, "/posts/{id}/author", 3)
	test.add(http.MethodGet, "/posts/1/author", 4)
	test.add(http.MethodGet, "/posts/{id:\\d+}", 5)
	test.add(http.MethodGet, "/posts/{id:\\d+}/author", 6)
	//test.n.print(0)

	test.matchTrue(http.MethodGet, "/", 1)
	test.matchTrue(http.MethodGet, "/posts/1", 5)             // 正则
	test.matchTrue(http.MethodGet, "/posts/2", 5)             // 正则
	test.matchTrue(http.MethodGet, "/posts/2/author", 6)      // 正则
	test.matchTrue(http.MethodGet, "/posts/1/author", 4)      // 字符串
	test.matchTrue(http.MethodGet, "/posts/1.html", 2)        // 命名参数
	test.matchTrue(http.MethodGet, "/posts/2.html/author", 3) // 命名参数
}

func TestNode_params(t *testing.T) {
	a := assert.New(t)
	test := newNodeTest(a)

	// 添加路由项
	test.add(http.MethodGet, "/posts/{id}", 1)                       // 命名
	test.add(http.MethodGet, "/posts/{id}/author/{action}/", 2)      // 命名
	test.add(http.MethodGet, "/posts/{id:\\d+}", 3)                  // 正则
	test.add(http.MethodGet, "/posts/{id:\\d+}/author/{action}/", 4) // 正则

	// 正则
	test.paramsTrue(http.MethodGet, "/posts/1", 3, map[string]string{"id": "1"})
	test.paramsTrue(http.MethodGet, "/posts/1/author/profile/", 4, map[string]string{"id": "1", "action": "profile"})

	// 命名
	test.paramsTrue(http.MethodGet, "/posts/1.html", 1, map[string]string{"id": "1.html"})
	test.paramsTrue(http.MethodGet, "/posts/1.html/author/profile/", 2, map[string]string{"id": "1.html", "action": "profile"})
}

func TestNode_url(t *testing.T) {
	a := assert.New(t)
	test := newNodeTest(a)

	// 添加路由项
	test.add(http.MethodGet, "/posts/{id}", 1)                       // 命名
	test.add(http.MethodGet, "/posts/{id}/author/{action}/", 2)      // 命名
	test.add(http.MethodGet, "/posts/{id:\\d+}", 3)                  // 正则
	test.add(http.MethodGet, "/posts/{id:\\d+}/author/{action}/", 4) // 正则

	test.urlTrue(http.MethodGet, "/posts/1", 3, map[string]string{"id": "100"}, "/posts/100")
	test.urlTrue(http.MethodGet, "/posts/1/author/profile/", 4, map[string]string{"id": "100", "action": "p"}, "/posts/100/author/p/")
	test.urlTrue(http.MethodGet, "/posts/1.html", 1, map[string]string{"id": "100.htm"}, "/posts/100.htm")
	test.urlTrue(http.MethodGet, "/posts/1.html/author/profile/", 2, map[string]string{"id": "100.htm", "action": "p"}, "/posts/100.htm/author/p/")
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
