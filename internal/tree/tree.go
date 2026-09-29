// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package tree 提供了以树形结构保存路由项的相关操作
package tree

import (
	"fmt"
	"iter"
	"net/http"
	"slices"
	"sync"

	"github.com/issue9/errwrap"

	"github.com/issue9/mux/v10/internal/syntax"
	"github.com/issue9/mux/v10/types"
)

// Tree 以树节点的形式保存的路由
//
// 多段路由项，会提取其中的相同的内容组成树状结构的节点。比如以下路由项：
//
//	/posts/{id}/author
//	/posts/{id}/author/emails
//	/posts/{id}/author/profile
//	/posts/1/author
//
// 会被转换成以下结构
//
//	/posts
//	   |
//	   +---- 1/author
//	   |
//	   +---- {id}/author/
//	             |
//	             +---- profile
//	             |
//	             +---- emails
type Tree[T any] struct {
	methods map[string]int // 保存着每个请求方法在所有子节点上的数量。
	node    *node[T]       // 空节点，正好用于处理 OPTIONS * 请求。

	// 由 New 负责初始化的内容

	locker                                  *sync.RWMutex
	interceptors                            *syntax.Interceptors
	name                                    string
	notFound, trace                         T
	hasTrace                                bool
	optionsBuilder, methodNotAllowedBuilder types.BuildNodeHandler[T]
}

func New[T any](
	name string,
	lock bool,
	i *syntax.Interceptors,
	notFound T,
	traceBuilder, methodNotAllowedBuilder, optionsBuilder types.BuildNodeHandler[T],
) *Tree[T] {
	s, err := i.NewSegment("")
	if err != nil {
		panic("发生了不该发生的错误，应该是 syntax.NewSegment 逻辑发生变化：" + err.Error())
	}

	if name == "" {
		panic("参数 name 不能为空")
	}

	hasTrace := traceBuilder != nil

	tree := &Tree[T]{
		methods: make(map[string]int, len(Methods)),
		node:    &node[T]{segment: s},

		interceptors:            i,
		name:                    name,
		hasTrace:                hasTrace,
		notFound:                notFound,
		optionsBuilder:          optionsBuilder,
		methodNotAllowedBuilder: methodNotAllowedBuilder,
	}
	tree.node.root = tree
	tree.node.handlers = map[string]T{
		http.MethodOptions: tree.optionsBuilder(tree.node),
	}
	if tree.hasTrace {
		tree.trace = traceBuilder(tree.node)
	}

	if lock {
		tree.locker = &sync.RWMutex{}
	}

	tree.node.buildMethods()

	return tree
}

func (tree *Tree[T]) Name() string { return tree.name }

// Add 添加路由项
//
// methods 可以为空，表示采用 [AnyMethods] 中的值。
func (tree *Tree[T]) Add(pattern string, h T, ms []types.Middleware[T], methods ...string) error {
	if tree.locker != nil {
		tree.locker.Lock()
		defer tree.locker.Unlock()
	}

	// checkAmbiguous 会遍历整棵树，必须置于锁内，否则会与其它 goroutine 的注册操作产生数据竞争。
	if err := tree.checkAmbiguous(pattern); err != nil {
		return err
	}

	n, err := tree.getNode(pattern)
	if err != nil {
		return err
	}

	if n.handlers == nil {
		n.handlers = make(map[string]T, handlersSize)
	}

	if len(methods) == 0 {
		methods = AnyMethods
	}
	return n.addMethods(h, pattern, ms, methods...)
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
	return fmt.Errorf("%s 与已有的节点 %s 存在歧义", pattern, s)
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

	child, _ := tree.find(pattern)
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
	segments, err := tree.interceptors.Split(pattern)
	if err != nil {
		return nil, err
	}
	return tree.node.getNode(segments)
}

// Handler 查找与参数匹配的处理对象
//
// 如果未找到，也会返回相应在的处理对象，比如 tree.notFound 或是相应的 methodNotAllowed 方法。
func (tree *Tree[T]) Handler(route *types.Route, method string) (n types.Node, handler T, found bool) {
	route.SetRouterName(tree.Name())

	if tree.hasTrace && method == http.MethodTrace {
		return tree.node, tree.trace, true
	}

	// 节点的匹配与 node.handlers 的读取必须在同一把读锁之内完成，否则注册阶段对 node.handlers 的修改，会与此处的读取产生数据竞争。
	if tree.locker != nil {
		tree.locker.RLock()
		defer tree.locker.RUnlock()
	}

	var node *node[T]
	if route.Path == "*" || route.Path == "" {
		node = tree.node
	} else {
		if node, found = tree.node.matchChildren(route); !found || node.size() == 0 {
			return nil, tree.notFound, false
		}
	}

	if h, exists := node.handlers[method]; exists {
		return node, h, true
	}
	return node, node.handlers[methodNotAllowed], false
}

// Routes 获取当前的所有路由项以及对应的请求方法
func (tree *Tree[T]) Routes() iter.Seq2[string, []string] {
	return func(yield func(string, []string) bool) {
		if tree.locker != nil {
			tree.locker.RLock()
			defer tree.locker.RUnlock()
		}

		ms := []string{http.MethodOptions}
		if tree.hasTrace {
			ms = append(ms, http.MethodTrace)
		}

		if !yield("*", ms) {
			return
		}

		for _, v := range tree.node.children {
			if !v.routes(yield) {
				return
			}
		}
	}
}

// find 查找匹配的节点
func (tree *Tree[T]) find(pattern string) (*node[T], int) { return tree.node.find(pattern, 0) }

// URL 将 ps 填入 pattern 生成 URL
//
// NOTE: 会检测 pattern 是否存在于 tree 中。
func (tree *Tree[T]) URL(buf *errwrap.StringBuilder, pattern string, ps map[string]string) error {
	n, size := tree.find(pattern)
	if n == nil {
		return fmt.Errorf("%s 并不是一条有效的注册路由项", pattern)
	}

	nodes := make([]*node[T], 0, size)
	for curr := n; curr.parent != nil; curr = curr.parent { // 从尾部向上开始获取节点
		nodes = append(nodes, curr)
	}

	for _, node := range slices.Backward(nodes) {
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

// ApplyMiddleware 为已有的路由项添加中间件
func (tree *Tree[T]) ApplyMiddleware(ms ...types.Middleware[T]) {
	tree.notFound = ApplyMiddleware(tree.notFound, "", "", tree.Name(), ms...)
	if tree.hasTrace {
		tree.trace = ApplyMiddleware(tree.trace, http.MethodTrace, "", tree.Name(), ms...)
	}
	tree.node.applyMiddleware(ms...)
}
