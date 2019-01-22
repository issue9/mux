// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v2/internal/handlers"
	"github.com/issue9/mux/v2/internal/syntax"
)

func TestNode_find(t *testing.T) {
	a := assert.New(t)
	node := &node{}

	addNode := func(p string, code int, methods ...string) {
		segs, err := syntax.Split(p)
		a.NotError(err).NotNil(segs)
		nn := node.getNode(segs)
		a.NotNil(nn)

		if nn.handlers == nil {
			nn.handlers = handlers.New(false, false)
		}

		a.NotError(nn.handlers.Add(buildHandler(code), methods...))
	}

	addNode("/", 1, http.MethodGet)
	addNode("/posts/{id}", 1, http.MethodGet)
	addNode("/posts/{id}/author", 1, http.MethodGet)
	addNode("/posts/1/author", 1, http.MethodGet)
	addNode("/posts/{id}/{author:\\w+}/profile", 1, http.MethodGet)

	a.Equal(node.find("/").pattern, "/")
	a.Equal(node.find("/posts/{id}").pattern, "{id}")
	a.Equal(node.find("/posts/{id}/author").pattern, "author")
	a.Equal(node.find("/posts/{id}/{author:\\w+}/profile").pattern, "{author:\\w+}/profile")
}

func TestRemoveNodes(t *testing.T) {
	a := assert.New(t)
	newNode := func(str string) *node {
		return &node{pattern: str}
	}

	n1 := newNode("/1")
	n2 := newNode("/2")
	n21 := newNode("/2")
	n3 := newNode("/3")
	n4 := newNode("/4")

	nodes := []*node{n1, n2, n21, n3, n4}

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
	newNode := func(str string) *node {
		return &node{pattern: str}
	}
	p := newNode("/blog")

	// 没有父节点
	a.Panic(func() {
		nn := splitNode(p, 1)
		a.Nil(nn)
	})

	node := p.newChild("/posts/{id}/author")
	a.NotNil(node)

	nn := splitNode(node, 7) // 从 { 开始拆分
	a.NotNil(nn)
	a.Equal(len(nn.children), 1).
		Equal(nn.children[0].pattern, "{id}/author")
	a.Equal(nn.parent, p)

	nn = splitNode(node, 18) // 不需要拆分
	a.NotNil(nn)
	a.Equal(0, len(nn.children))

	a.Panic(func() { splitNode(node, 8) }) // 从 i 开始拆分，不可拆分
}
