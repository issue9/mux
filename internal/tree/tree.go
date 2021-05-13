// SPDX-License-Identifier: MIT

// Package tree 提供了以树形结构保存路由项的相关操作
package tree

import (
	"net/http"

	"github.com/issue9/mux/v4/internal/handlers"
	"github.com/issue9/mux/v4/internal/syntax"
	"github.com/issue9/mux/v4/params"
)

// Tree 以树节点的形式保存的路由
//
// 多段路由项，会提取其中的相同的内容组成树状结构的节点。
// 比如以下路由项：
//  /posts/{id}/author
//  /posts/{id}/author/emails
//  /posts/{id}/author/profile
//  /posts/1/author
// 会被转换成以下结构
//  /posts
//     |
//     +---- 1/author
//     |
//     +---- {id}/author
//               |
//               +---- /profile
//               |
//               +---- /emails
type Tree struct {
	node
	disableHead bool
}

// New 声明一个 Tree 实例
func New(disableHead bool) *Tree {
	s, err := syntax.NewSegment("")
	if err != nil {
		panic("发生了不该发生的错误，应该是 syntax.NewSegment 逻辑发生变化" + err.Error())
	}

	return &Tree{
		node:        node{segment: s},
		disableHead: disableHead,
	}
}

// Add 添加路由项
//
// methods 可以为空，表示添加除 OPTIONS 之外所有支持的请求方法。
func (tree *Tree) Add(pattern string, h http.Handler, methods ...string) error {
	n, err := tree.getNode(pattern)
	if err != nil {
		return err
	}

	if n.handlers == nil {
		n.handlers = handlers.New(tree.disableHead)
	}

	return n.handlers.Add(h, methods...)
}

// Clean 清除路由项
func (tree *Tree) Clean(prefix string) { tree.clean(prefix) }

// Remove 移除路由项
//
// methods 可以为空，表示删除所有内容。
func (tree *Tree) Remove(pattern string, methods ...string) error {
	child := tree.find(pattern)
	if child == nil || child.handlers == nil {
		return nil
	}

	empty, err := child.handlers.Remove(methods...)
	if err != nil {
		return err
	}
	if empty && len(child.children) == 0 {
		child.parent.children = removeNodes(child.parent.children, child.segment.Value)
		child.parent.buildIndexes()
	}
	return nil
}

// 获取指定的节点，若节点不存在，则在该位置生成一个新节点。
func (tree *Tree) getNode(pattern string) (*node, error) {
	segs, err := syntax.Split(pattern)
	if err != nil {
		return nil, err
	}

	return tree.node.getNode(segs)
}

// Handler 找到与当前内容匹配的 handlers.Handlers 实例
func (tree *Tree) Handler(path string) (*handlers.Handlers, params.Params) {
	ps := make(params.Params, 3)
	node := tree.match(path, ps)

	if node == nil || node.handlers == nil || node.handlers.Len() == 0 {
		return nil, nil
	}
	return node.handlers, ps
}

// All 获取当前的所有路径项
func (tree *Tree) All() map[string][]string {
	routes := make(map[string][]string, 100)
	tree.all("", routes)
	return routes
}
