// SPDX-License-Identifier: MIT

// Package tree 提供了以树形结构保存路由项的相关操作
package tree

import (
	"fmt"
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
	disableHead bool
	node        *Node

	// methods 保存着每个请求方法在所有子节点上的数量。
	// 在每次添加和删除请求方法之后，更新 methodIndex 的值。
	methods         map[string]int
	methodIndex     int
	optionsAsterisk bool
}

// New 声明一个 Tree 实例
//
// optionsAsterisk 表示是否支持 options * 操作，如果为 true，同时还将空路径重定向到 *。
func New(disableHead, optionsAsterisk bool) *Tree {
	s, err := syntax.NewSegment("")
	if err != nil {
		panic("发生了不该发生的错误，应该是 syntax.NewSegment 逻辑发生变化" + err.Error())
	}

	methods := make(map[string]int, len(Methods))
	for _, m := range Methods {
		methods[m] = 0
	}

	t := &Tree{
		disableHead: disableHead,
		node:        &Node{segment: s, handlers: make(map[string]http.Handler, 1)},

		methods:         methods,
		methodIndex:     methodIndexMap[http.MethodOptions],
		optionsAsterisk: optionsAsterisk,
	}
	t.node.root = t

	if optionsAsterisk {
		err := t.add(true, "*", http.HandlerFunc(t.optionsServeHTTP), http.MethodOptions)
		if err != nil {
			panic(err)
		}
	}

	return t
}

func (tree *Tree) optionsServeHTTP(w http.ResponseWriter, req *http.Request) {
	optionsHandler(w, methodIndexes[tree.methodIndex].options)
}

// Add 添加路由项
//
// methods 可以为空，表示添加除 OPTIONS 和 HEAD 之外所有支持的请求方法。
func (tree *Tree) Add(pattern string, h http.Handler, methods ...string) error {
	return tree.add(false, pattern, h, methods...)
}

func (tree *Tree) add(f bool, pattern string, h http.Handler, methods ...string) error {
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
	return n.addMethods(f, h, methods...)
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

	tree.buildMethods(-1, methods...)
}

// 获取指定的节点，若节点不存在，则在该位置生成一个新节点。
func (tree *Tree) getNode(pattern string) (*Node, error) {
	segs, err := syntax.Split(pattern)
	if err != nil {
		return nil, err
	}

	names := make(map[string]int, len(segs))
	for _, seg := range segs {
		if seg.Type == syntax.String {
			continue
		}
		if names[seg.Name] > 0 {
			return nil, fmt.Errorf("存在相同名称的路由参数：%s", seg.Name)
		}
		names[seg.Name]++
	}

	return tree.node.getNode(segs)
}

// Route 找到与当前内容匹配的 Node 实例
func (tree *Tree) Route(path string) (*Node, params.Params) {
	if path == "" && tree.optionsAsterisk {
		path = "*"
	}

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
