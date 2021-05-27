// SPDX-License-Identifier: MIT

// Package tree 提供了以树形结构保存路由项的相关操作
package tree

import (
	"net/http"

	"github.com/issue9/mux/v5/internal/syntax"
	"github.com/issue9/mux/v5/params"
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
	node        *Node
	disableHead bool
}

// New 声明一个 Tree 实例
func New(disableHead bool) *Tree {
	s, err := syntax.NewSegment("")
	if err != nil {
		panic("发生了不该发生的错误，应该是 syntax.NewSegment 逻辑发生变化" + err.Error())
	}

	return &Tree{
		node:        &Node{segment: s, handlers: make(map[string]http.Handler, handlersSize)},
		disableHead: disableHead,
	}
}

// Add 添加路由项
//
// methods 可以为空，表示添加除 OPTIONS 和 HEAD 之外所有支持的请求方法。
func (tree *Tree) Add(pattern string, h http.Handler, methods ...string) error {
	n, err := tree.getNode(pattern)
	if err != nil {
		return err
	}

	if n.handlers == nil {
		n.handlers = make(map[string]http.Handler, handlersSize)
	}

	if len(methods) == 0 {
		methods = addAny
	}
	return n.addMethods(tree.disableHead, h, methods...)
}

// Clean 清除路由项
func (tree *Tree) Clean(prefix string) { tree.node.clean(prefix) }

// Remove 移除路由项
//
// methods 可以为空，表示删除所有内容。单独删除 OPTIONS，将不会发生任何事情。
func (tree *Tree) Remove(pattern string, methods ...string) {
	child := tree.node.find(pattern)
	if child == nil || len(child.handlers) == 0 {
		return
	}

	if len(methods) == 0 {
		child.handlers = make(map[string]http.Handler, handlersSize)
	} else {
		for _, m := range methods {
			switch m {
			case http.MethodOptions: // OPTIONS 不作任何操作
			case http.MethodGet:
				delete(child.handlers, http.MethodHead)
				fallthrough
			default:
				delete(child.handlers, m)
			}
		}

		if _, found := child.handlers[http.MethodOptions]; found && len(child.handlers) == 1 { // 只有一个 OPTIONS 了
			delete(child.handlers, http.MethodOptions)
		}
	}
	child.buildMethods()

	if len(child.handlers) == 0 && len(child.children) == 0 {
		child.parent.children = removeNodes(child.parent.children, child.segment.Value)
		child.parent.buildIndexes()
	}
}

// 获取指定的节点，若节点不存在，则在该位置生成一个新节点。
func (tree *Tree) getNode(pattern string) (*Node, error) {
	segs, err := syntax.Split(pattern)
	if err != nil {
		return nil, err
	}
	return tree.node.getNode(segs)
}

// Route 找到与当前内容匹配的 Node 实例
func (tree *Tree) Route(path string) (*Node, params.Params) {
	ps := make(params.Params, 3)
	node := tree.node.match(path, ps)

	if node == nil || len(node.handlers) == 0 {
		return nil, nil
	}
	return node, ps
}

// Routes 获取当前的所有路由项
func (tree *Tree) Routes() map[string][]string {
	routes := make(map[string][]string, 100)
	tree.node.routes("", routes)
	return routes
}
