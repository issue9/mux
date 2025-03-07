// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package tree

import (
	"slices"
	"strings"

	"github.com/issue9/mux/v9/internal/syntax"
	"github.com/issue9/mux/v9/types"
)

const (
	indexesSize  = 5 // Node.children 的数量只有达到此值时，才会为其建立 indexes 索引表。
	handlersSize = 3 // Node.handlers 的初始容量
)

type node[T any] struct {
	root    *Tree[T]
	parent  *node[T]
	segment *syntax.Segment
	pattern string

	methodIndex int // 在 methodIndexes 中的索引值
	handlers    map[string]T

	// 保存着 node 实例在 children 中的下标。
	//
	// 所有节点类型为字符串的子节点，其首字符必定是不同的（相同的都提升到父节点中），
	// 根据此特性，可以将所有字符串类型的首字符做个索引，这样字符串类型节点的比较，
	// 可以通过索引排除不必要的比较操作。
	indexes map[byte]int

	children []*node[T]
}

func (n *node[T]) size() int { return len(n.handlers) }

// 构建当前节点的索引表
func (n *node[T]) buildIndexes() {
	if len(n.children) < indexesSize {
		n.indexes = nil
		return
	}

	if n.indexes == nil {
		n.indexes = make(map[byte]int, indexesSize)
	}

	for index, node := range n.children {
		if node.segment.Type == syntax.String {
			n.indexes[node.segment.Value[0]] = index
		}
	}
}

func (n *node[T]) priority() int {
	ret := int(n.segment.Type) * 10 // 10 可以保证在当前类型的节点进行加权时，不会超过其它节点。

	if len(n.children) == 0 {
		ret++
	}
	if n.segment.Endpoint { // 同类型中权重最低
		ret++
	}

	return ret
}

// 获取指定路径下的节点，若节点不存在，则添加。
// segments 为被 [syntax.Split] 拆分之后的字符串数组。
func (n *node[T]) getNode(segments []*syntax.Segment) (*node[T], error) {
	child, err := n.addSegment(segments[0])
	if err != nil {
		return nil, err
	}

	if len(segments) == 1 { // 最后一个节点
		return child, nil
	}
	return child.getNode(segments[1:])
}

// 将 seg 添加到当前节点，并返回新节点，如果找到相同的节点，则直接返回该子节点。
func (n *node[T]) addSegment(seg *syntax.Segment) (*node[T], error) {
	var child *node[T] // 找到的最匹配节点
	var l int          // 最大的匹配字符数量
	for _, c := range n.children {
		l1 := c.segment.Similarity(seg)

		if l1 == -1 { // 找到完全相同的，则直接返回该节点
			return c, nil
		}

		if l1 > l { // 找到相似度更高的，保存该节点的信息
			l = l1
			child = c
		}
	}

	if l <= 0 { // 没有共同前缀，声明一个新的加入到当前节点
		nn := n.newChild(seg)
		n.sort()
		return nn, nil
	}

	parent, err := splitNode(child, l)
	if err != nil {
		return nil, err
	}

	// seg 与 parent 重叠
	if len(seg.Value) == l {
		return parent, nil
	}

	// seg.Value[:l] 与 child.segment.Value[:l] 暨 parent.Value 是相同的
	s, err := n.root.interceptors.NewSegment(seg.Value[l:])
	if err != nil {
		return nil, err
	}
	return parent.addSegment(s)
}

func (n *node[T]) Pattern() string { return n.pattern }

// 根据 s 内容为当前节点产生一个子节点，并返回该节点。
// 由调用方确保 s 的语法正确性，否则可能 panic。
func (n *node[T]) newChild(s *syntax.Segment) *node[T] {
	child := &node[T]{root: n.root, parent: n, segment: s, pattern: n.pattern + s.Value}
	n.children = append(n.children, child)
	return child
}

func (n *node[T]) sort() {
	slices.SortStableFunc(n.children, func(a, b *node[T]) int { return a.priority() - b.priority() })
	n.buildIndexes()
}

// 查找路由项，不存在返回 nil
func (n *node[T]) find(pattern string) *node[T] {
	for _, child := range n.children {
		if child.segment.Value == pattern {
			return child
		}

		if strings.HasPrefix(pattern, child.segment.Value) {
			if nn := child.find(pattern[len(child.segment.Value):]); nn != nil {
				return nn
			}
		}
	}

	return nil
}

// 清除路由项
func (n *node[T]) clean(prefix string) {
	if len(prefix) == 0 {
		n.children = n.children[:0]
		return
	}

	dels := make([]string, 0, len(n.children))
	for _, child := range n.children {
		if len(child.segment.Value) < len(prefix) {
			if strings.HasPrefix(prefix, child.segment.Value) {
				child.clean(prefix[len(child.segment.Value):])
			}
		}

		if strings.HasPrefix(child.segment.Value, prefix) {
			dels = append(dels, child.segment.Value)
		}
	}

	for _, del := range dels {
		n.children = removeNodes(n.children, del)
	}
	n.buildIndexes()
}

// 从子节点中查找与当前路径匹配的节点，若找不到，则返回 nil。
func (n *node[T]) matchChildren(ctx *types.Context) *node[T] {
	if len(n.indexes) > 0 && len(ctx.Path) > 0 { // 普通字符串的匹配
		child := n.children[n.indexes[ctx.Path[0]]]
		if child == nil {
			goto LOOP
		}

		path := ctx.Path

		if !child.segment.Match(ctx) { // 这会修改 ctx.Path 的值
			goto LOOP
		}
		if nn := child.matchChildren(ctx); nn != nil {
			return nn
		}

		ctx.Path = path
	}

LOOP:
	// 即使 p.Path 为空，也有可能子节点正好可以匹配空的内容。
	// 比如 /posts/{path:\\w*} 后面的 path 即为空节点。所以此处不判断 len(p.Path)
	for i := len(n.indexes); i < len(n.children); i++ {
		child := n.children[i]
		path := ctx.Path

		if !child.segment.Match(ctx) { // 不匹配
			continue
		}
		if nn := child.matchChildren(ctx); nn != nil {
			return nn
		}

		// 不匹配子元素，则恢复原有数据
		ctx.Path = path
		ctx.Delete(n.segment.Name)
	}

	// 没有子节点匹配，len(p.Path)==0，且子节点不为空，可以判定与当前节点匹配。
	if len(ctx.Path) == 0 && n.size() > 0 {
		return n
	}
	return nil
}

// 从 nodes 中删除一个 pattern 字段为指定值的元素，
//
// NOTE: 实际应该中，理论上不会出现多个相同的元素，
// 所以此处不作多余的判断。
func removeNodes[T any](nodes []*node[T], pattern string) []*node[T] {
	for index, n := range nodes {
		if n.segment.Value == pattern {
			return slices.Delete(nodes, index, index+1)
		}
	}
	return nodes
}

// 将节点 n 从 pos 位置进行拆分。后一段作为当前段的子节点，并返回当前节点。
// 若 pos 大于或等于 n.pattern 的长度，则直接返回 n 不会拆分，pos 处的字符作为子节点的内容。
//
// 若 n.parent 为 nil，都将触发 panic
func splitNode[T any](n *node[T], pos int) (*node[T], error) {
	if len(n.segment.Value) <= pos { // 不需要拆分
		return n, nil
	}

	p := n.parent
	if p == nil {
		panic("节点必须要有一个有效的父节点，才能进行拆分")
	}
	p.children = removeNodes(p.children, n.segment.Value) // 先从父节点中删除老的 n

	segs, err := n.segment.Split(n.root.interceptors, pos)
	if err != nil {
		return nil, err
	}
	ret := p.newChild(segs[0])
	c := ret.newChild(segs[1])
	c.handlers = n.handlers
	c.methodIndex = n.methodIndex
	c.children = n.children
	c.indexes = n.indexes
	for _, item := range c.children {
		item.parent = c
	}

	// ret 和 c 的内容在 newChild 之后被修改，所以需要对其子元素重新排序。
	ret.sort()
	p.sort()

	return ret, nil
}

// 将所有的路由地址列表写入 routes
func (n *node[T]) routes(routes map[string][]string) {
	if n.methodIndex > 0 {
		routes[n.Pattern()] = n.Methods()
	}

	for _, v := range n.children {
		v.routes(routes)
	}
}

func (n *node[T]) checkAmbiguous(pattern string, hasNonString bool) (*node[T], bool, error) {
	if pattern == "" {
		if n.size() > 0 {
			return n, hasNonString, nil
		}
		return nil, false, nil
	}

	for _, c := range n.children {
		seg := c.segment

		if strings.HasPrefix(pattern, seg.Value) {
			node, hasNonString, err := c.checkAmbiguous(pattern[len(seg.Value):], hasNonString)
			if err != nil {
				return nil, false, err
			}

			if node != nil {
				return node, hasNonString, nil
			}
			continue
		}

		segs, err := n.root.interceptors.Split(pattern)
		if err != nil {
			return nil, false, err
		}
		s0 := segs[0]

		if seg.IsAmbiguous(s0) {
			node, hasNonString, err := c.checkAmbiguous(pattern[s0.AmbiguousLen():], true)
			if err != nil {
				return nil, false, err
			}
			if node != nil {
				return node, hasNonString, nil
			}
		}
	}

	return nil, false, nil
}

func (n *node[T]) applyMiddleware(ms ...types.Middleware[T]) {
	for m, h := range n.handlers {
		n.handlers[m] = ApplyMiddleware(h, m, n.Pattern(), n.root.Name(), ms...)
	}

	for _, c := range n.children {
		c.applyMiddleware(ms...)
	}
}

func ApplyMiddleware[T any](h T, method, pattern, router string, f ...types.Middleware[T]) T {
	for _, ff := range f {
		h = ff.Middleware(h, method, pattern, router)
	}
	return h
}
