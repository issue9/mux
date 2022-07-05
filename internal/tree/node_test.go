// SPDX-License-Identifier: MIT

package tree

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/mux/v7/internal/syntax"
	"github.com/issue9/mux/v7/types"
)

var _ types.Node = &node[http.Handler]{}

// 获取当前路由下所有处理函数的节点数量
func (n *node[T]) len() int {
	var cnt int
	for _, child := range n.children {
		cnt += child.len()
	}

	if len(n.handlers) > 0 {
		cnt++
	}

	return cnt
}

func TestRemoveNodes(t *testing.T) {
	a := assert.New(t, false)

	tree := NewTestTree(a, false, syntax.NewInterceptors())

	newNode := func(str string) *node[http.Handler] {
		s, err := tree.interceptors.NewSegment(str)
		a.NotError(err).NotNil(s)
		return &node[http.Handler]{segment: s, root: tree}
	}

	n1 := newNode("/1")
	n2 := newNode("/2")
	n21 := newNode("/2")
	n3 := newNode("/3")
	n4 := newNode("/4")

	nodes := []*node[http.Handler]{n1, n2, n21, n3, n4}

	// 不存在的元素
	nodes = removeNodes(nodes, "")
	a.Equal(len(nodes), 5)

	// 删除尾元素
	nodes = removeNodes(nodes, "/4")
	a.Equal(len(nodes), 4)

	// 删除中间元素
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
	a := assert.New(t, false)
	tree := NewTestTree(a, false, syntax.NewInterceptors())

	newNode := func(str string) *node[http.Handler] {
		s, err := tree.interceptors.NewSegment(str)
		a.NotError(err).NotNil(s)
		return &node[http.Handler]{segment: s, root: tree}
	}
	p := newNode("/blog")

	// 没有父节点
	a.Panic(func() {
		nn, _ := splitNode(p, 1)
		a.Nil(nn)
	})

	s, err := tree.interceptors.NewSegment("/posts/{id}/author")
	a.NotError(err).NotNil(s)
	node := p.newChild(s)
	a.NotNil(node).
		NotNil(node.root)

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
