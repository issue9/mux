// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux/internal/tree/handlers"
	"github.com/issue9/mux/internal/tree/segment"
)

func TestNode_find(t *testing.T) {
	a := assert.New(t)
	node := &Node{}

	addNode := func(p string, code int, methods ...string) {
		nn, err := node.getNode(split(a, p))
		a.NotError(err).NotNil(nn)

		if nn.handlers == nil {
			nn.handlers = handlers.New()
		}

		a.NotError(nn.handlers.Add(buildHandler(code), methods...))
	}

	addNode("/", 1, http.MethodGet)
	addNode("/posts/{id}", 1, http.MethodGet)
	addNode("/posts/{id}/author", 1, http.MethodGet)
	addNode("/posts/1/author", 1, http.MethodGet)
	addNode("/posts/{id}/{author:\\w+}/profile", 1, http.MethodGet)

	a.Equal(node.find("/").seg.Value(), "/")
	a.Equal(node.find("/posts/{id}").seg.Value(), "{id}")
	a.Equal(node.find("/posts/{id}/author").seg.Value(), "author")
	a.Equal(node.find("/posts/{id}/{author:\\w+}/profile").seg.Value(), "{author:\\w+}/profile")
}

func TestRemoveNoddes(t *testing.T) {
	a := assert.New(t)
	newNode := func(str string) *Node {
		seg, err := segment.New(str)
		a.NotError(err).NotNil(seg)
		return &Node{seg: seg}
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
		seg, err := segment.New(str)
		a.NotError(err).NotNil(seg)
		return &Node{seg: seg}
	}
	p := newNode("/blog")

	// 没有父节点
	nn, err := splitNode(p, 1)
	a.Error(err).Nil(nn)

	node, err := p.newChild("/posts/{id}/author")
	a.NotError(err).NotNil(node)

	nn, err = splitNode(node, 7) // 从 { 开始拆分
	a.NotError(err).NotNil(nn)
	a.Equal(len(nn.children), 1).
		Equal(nn.children[0].seg.Value(), "{id}/author")
	a.Equal(nn.parent, p)

	nn, err = splitNode(node, 18) // 不需要拆分
	a.NotError(err).NotNil(nn)
	a.Equal(0, len(nn.children))

	nn, err = splitNode(node, 8) // 从 i 开始拆分
	a.Error(err).Nil(nn)
}
