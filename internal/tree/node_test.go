// SPDX-License-Identifier: MIT

package tree

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v5/internal/syntax"
)

// 获取当前路由下所有处理函数的节点数量
func (n *Node) len() int {
	var cnt int
	for _, child := range n.children {
		cnt += child.len()
	}

	if len(n.handlers) > 0 {
		cnt++
	}

	return cnt
}

func TestNode_find(t *testing.T) {
	a := assert.New(t)
	tree := New(true)
	node := &Node{root: tree}

	addNode := func(p string, code int, methods ...string) {
		segs, err := syntax.Split(p)
		a.NotError(err).NotNil(segs)
		nn, err := node.getNode(segs)
		a.NotError(err).NotNil(nn)

		if nn.handlers == nil {
			nn.handlers = make(map[string]http.Handler, handlersSize)
		}

		a.NotError(nn.addMethods(buildHandler(code), methods...))
	}

	addNode("/", 1, http.MethodGet)
	addNode("/posts/{id}", 1, http.MethodGet)
	addNode("/posts/{id}/author", 1, http.MethodGet)
	addNode("/posts/1/author", 1, http.MethodGet)
	addNode("/posts/{id}/{author:\\w+}/profile", 1, http.MethodGet)

	a.Equal(node.find("/").segment.Value, "/")
	a.Equal(node.find("/posts/{id}").segment.Value, "{id}")
	a.Equal(node.find("/posts/{id}/author").segment.Value, "author")
	a.Equal(node.find("/posts/{id}/{author:\\w+}/profile").segment.Value, "{author:\\w+}/profile")
}

func TestRemoveNodes(t *testing.T) {
	a := assert.New(t)
	newNode := func(str string) *Node {
		s, err := syntax.NewSegment(str)
		a.NotError(err).NotNil(s)
		return &Node{segment: s}
	}

	n1 := newNode("/1")
	n2 := newNode("/2")
	n21 := newNode("/2")
	n3 := newNode("/3")
	n4 := newNode("/4")

	nodes := []*Node{n1, n2, n21, n3, n4}

	// 不存在的元素
	nodes = removeNodes(nodes, "")
	a.Equal(len(nodes), 5)

	// 删除尾元素
	nodes = removeNodes(nodes, "/4")
	a.Equal(len(nodes), 4)

	// 删除中间无素
	nodes = removeNodes(nodes, "/2")
	a.Equal(len(nodes), 3)

	// 删除另一个同名元素
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

func TestSplitNode(t *testing.T) {
	a := assert.New(t)
	newNode := func(str string) *Node {
		s, err := syntax.NewSegment(str)
		a.NotError(err).NotNil(s)
		return &Node{segment: s}
	}
	p := newNode("/blog")

	// 没有父节点
	a.Panic(func() {
		nn, _ := splitNode(p, 1)
		a.Nil(nn)
	})

	s, err := syntax.NewSegment("/posts/{id}/author")
	a.NotError(err).NotNil(s)
	node := p.newChild(s)
	a.NotNil(node)

	nn, err := splitNode(node, 7) // 从 { 开始拆分
	a.NotError(err).NotNil(nn)
	a.Equal(len(nn.children), 1).
		Equal(nn.children[0].segment.Value, "{id}/author")
	a.Equal(nn.parent, p)

	nn, err = splitNode(node, 18) // 不需要拆分
	a.NotError(err).NotNil(nn)
	a.Equal(0, len(nn.children))

	nn, err = splitNode(node, 8)
	a.NotError(err).NotNil(nn)
}
