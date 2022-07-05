// SPDX-License-Identifier: MIT

// Package tree 提供了以树形结构保存路由项的相关操作
package tree

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/issue9/errwrap"

	"github.com/issue9/mux/v7/internal/syntax"
	"github.com/issue9/mux/v7/types"
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
//     +---- {id}/author/
//               |
//               +---- profile
//               |
//               +---- emails
type Tree[T any] struct {
	methods map[string]int // 保存着每个请求方法在所有子节点上的数量。
	node    *node[T]       // 空节点，正好用于处理 OPTIONS *。

	// 由 New 负责初始化的内容
	locker       *sync.RWMutex
	interceptors *syntax.Interceptors

	notFound T
	optionsBuilder,
	methodNotAllowedBuilder types.BuildNodeHandleOf[T]
}

func New[T any](lock bool, i *syntax.Interceptors, notFound T, methodNotAllowedBuilder, optionsBuilder types.BuildNodeHandleOf[T]) *Tree[T] {
	s, err := i.NewSegment("")
	if err != nil {
		panic("发生了不该发生的错误，应该是 syntax.NewSegment 逻辑发生变化" + err.Error())
	}

	t := &Tree[T]{
		methods: make(map[string]int, len(Methods)),
		node:    &node[T]{segment: s, methodIndex: methodIndexMap[http.MethodOptions]},

		interceptors:            i,
		notFound:                notFound,
		optionsBuilder:          optionsBuilder,
		methodNotAllowedBuilder: methodNotAllowedBuilder,
	}
	t.node.root = t
	t.node.handlers = map[string]T{
		http.MethodOptions: t.optionsBuilder(t.node),
	}

	if lock {
		t.locker = &sync.RWMutex{}
	}

	return t
}

// Add 添加路由项
//
// methods 可以为空，表示添加除 OPTIONS 和 HEAD 之外所有支持的请求方法。
func (tree *Tree[T]) Add(pattern string, h T, methods ...string) error {
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
		n.handlers = make(map[string]T, handlersSize)
	}

	if len(methods) == 0 {
		methods = addAny
	}
	return n.addMethods(h, methods...)
}

func (tree *Tree[T]) checkAmbiguous(pattern string) error {
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
func (tree *Tree[T]) Clean(prefix string) {
	if tree.locker != nil {
		tree.locker.Lock()
		defer tree.locker.Unlock()
	}

	tree.node.clean(prefix)
}

// Remove 移除路由项
//
// methods 可以为空，表示删除所有内容。单独删除 OPTIONS，将不会发生任何事情。
func (tree *Tree[T]) Remove(pattern string, methods ...string) {
	if tree.locker != nil {
		tree.locker.Lock()
		defer tree.locker.Unlock()
	}

	child := tree.Find(pattern)
	if child == nil {
		return
	}

	if len(methods) == 0 {
		child.handlers = nil
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

		if child.size() == 2 { // 只有一个 OPTIONS 和 method not allowed 了
			_, e1 := child.handlers[http.MethodOptions]
			_, e2 := child.handlers[methodNotAllowed]
			if e1 && e2 {
				delete(child.handlers, http.MethodOptions)
				delete(child.handlers, methodNotAllowed)
			}
		}
	}

	child.buildMethods()

	for child.size() == 0 && len(child.children) == 0 {
		child.parent.children = removeNodes(child.parent.children, child.segment.Value)
		child.parent.buildIndexes()
		child = child.parent
	}

	tree.buildMethods(-1, methods...)
}

// 获取指定的节点，若节点不存在，则在该位置生成一个新节点。
func (tree *Tree[T]) getNode(pattern string) (*node[T], error) {
	segs, err := tree.interceptors.Split(pattern)
	if err != nil {
		return nil, err
	}
	return tree.node.getNode(segs)
}

// match 找到与路径 path 匹配的 node 实例
func (tree *Tree[T]) match(p *syntax.Params) *node[T] {
	if tree.locker != nil {
		tree.locker.RLock()
		defer tree.locker.RUnlock()
	}

	if p.Path == "*" || p.Path == "" {
		return tree.node
	}

	node := tree.node.matchChildren(p)
	if node == nil || node.size() == 0 {
		return nil
	}
	return node
}

// Handler 查找与参数匹配的处理对象
//
// 如果未找到，也会返回相应在的处理对象，比如 tree.notFound 或是相应的 methodNotAllowed 方法。
func (tree *Tree[T]) Handler(p *syntax.Params, method string) (types.Node, T, bool) {
	node := tree.match(p)
	if node == nil {
		return nil, tree.notFound, false
	}

	if h, exists := node.handlers[method]; exists {
		return node, h, true
	}
	return node, node.handlers[methodNotAllowed], false
}

// Routes 获取当前的所有路由项以及对应的请求方法
func (tree *Tree[T]) Routes() map[string][]string {
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

// Find 查找匹配的节点
func (tree *Tree[T]) Find(pattern string) *node[T] { return tree.node.find(pattern) }

// URL 将 ps 填入 pattern 生成 URL
//
// NOTE: 会检测 pattern 是否存在于 tree 中。
func (tree *Tree[T]) URL(buf *errwrap.StringBuilder, pattern string, ps map[string]string) error {
	n := tree.Find(pattern)
	if n == nil {
		return fmt.Errorf("%s 并不是一条有效的注册路由项", pattern)
	}

	nodes := make([]*node[T], 0, 5)
	for curr := n; curr.parent != nil; curr = curr.parent { // 从尾部向上开始获取节点
		nodes = append(nodes, curr)
	}
	l := len(nodes)
	for i, j := 0, l-1; i < j; i, j = i+1, j-1 {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	}

	for _, node := range nodes {
		s := node.segment
		switch s.Type {
		case syntax.String:
			buf.WString(s.Value)
		case syntax.Named, syntax.Regexp:
			param, exists := ps[s.Name]
			if !exists {
				return fmt.Errorf("未找到参数 %s 的值", s.Name)
			}
			if !s.Valid(param) {
				return fmt.Errorf("参数 %s 格式不匹配", s.Name)
			}

			buf.WString(param).WString(s.Suffix)
		}
	}

	return nil
}

func (tree *Tree[T]) ApplyMiddleware(ms ...types.MiddlewareOf[T]) {
	tree.notFound = ApplyMiddlewares(tree.notFound, ms...)
	tree.node.applyMiddlewares(ms...)
}
