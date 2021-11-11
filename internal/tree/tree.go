// SPDX-License-Identifier: MIT

// Package tree 提供了以树形结构保存路由项的相关操作
package tree

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/issue9/mux/v5/internal/syntax"
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
	// OPTIONS * 用到的请求方法列表
	// node.methodIndex 保存自身的请求方法 CORS 的预检机制用到 node.methodIndex。
	methodIndex int

	// methods 保存着每个请求方法在所有子节点上的数量。
	methods map[string]int

	node *Node

	interceptors *syntax.Interceptors

	locker *sync.RWMutex
}

// New 声明一个 Tree 实例
func New(lock bool, i *syntax.Interceptors) *Tree {
	s, err := i.NewSegment("")
	if err != nil {
		panic("发生了不该发生的错误，应该是 syntax.NewSegment 逻辑发生变化" + err.Error())
	}

	t := &Tree{
		methods:      make(map[string]int, len(Methods)),
		node:         &Node{segment: s, methodIndex: methodIndexMap[http.MethodOptions]},
		interceptors: i,
	}
	t.node.root = t
	t.node.handlers = map[string]http.Handler{http.MethodOptions: http.HandlerFunc(t.optionsServeHTTP)}

	if lock {
		t.locker = &sync.RWMutex{}
	}

	return t
}

func (tree *Tree) optionsServeHTTP(w http.ResponseWriter, _ *http.Request) {
	optionsHandle(w, methodIndexes[tree.methodIndex].options)
}

// Add 添加路由项
//
// methods 可以为空，表示添加除 OPTIONS 和 HEAD 之外所有支持的请求方法。
func (tree *Tree) Add(pattern string, h http.Handler, methods ...string) error {
	if err := tree.checkAmbiguous(pattern); err != nil {
		return err
	}

	if tree.locker != nil {
		tree.locker.Lock()
		defer tree.locker.Unlock()
	}

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
	return n.addMethods(h, methods...)
}

func (tree *Tree) checkAmbiguous(pattern string) error {
	n, has, err := tree.node.checkAmbiguous(pattern, false)
	if err != nil {
		return err
	}

	if n == nil || !has {
		return nil
	}

	var s string
	for n != nil {
		s = n.segment.Value + s
		n = n.parent
	}

	return fmt.Errorf("存在有歧义的节点：%s", s)
}

// Clean 清除路由项
func (tree *Tree) Clean(prefix string) {
	if tree.locker != nil {
		tree.locker.Lock()
		defer tree.locker.Unlock()
	}

	tree.node.clean(prefix)
}

// Remove 移除路由项
//
// methods 可以为空，表示删除所有内容。单独删除 OPTIONS，将不会发生任何事情。
func (tree *Tree) Remove(pattern string, methods ...string) {
	if tree.locker != nil {
		tree.locker.Lock()
		defer tree.locker.Unlock()
	}

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

	for len(child.handlers) == 0 && len(child.children) == 0 {
		child.parent.children = removeNodes(child.parent.children, child.segment.Value)
		child.parent.buildIndexes()
		child = child.parent
	}

	tree.buildMethods(-1, methods...)
}

// 获取指定的节点，若节点不存在，则在该位置生成一个新节点。
func (tree *Tree) getNode(pattern string) (*Node, error) {
	segs, err := tree.interceptors.Split(pattern)
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
//
// NOTE: 调用方需要调用 syntax.Params.Destroy 销毁对象
func (tree *Tree) Route(path string) (*Node, *syntax.Params) {
	if tree.locker != nil {
		tree.locker.RLock()
		defer tree.locker.RUnlock()
	}

	if path == "*" || path == "" {
		return tree.node, nil
	}

	p := syntax.NewParams(path)
	node := tree.node.matchChildren(p)
	if node == nil || len(node.handlers) == 0 {
		p.Destroy()
		return nil, nil
	}
	return node, p
}

// Routes 获取当前的所有路由项以及对应的请求方法
func (tree *Tree) Routes() map[string][]string {
	if tree.locker != nil {
		tree.locker.RLock()
		defer tree.locker.RUnlock()
	}

	routes := make(map[string][]string, 100)
	routes["*"] = []string{http.MethodOptions}

	for _, v := range tree.node.children {
		v.routes("", routes)
	}

	return routes
}

func (tree *Tree) URL(pattern string, ps map[string]string) (string, error) {
	return tree.interceptors.URL(pattern, ps)
}
