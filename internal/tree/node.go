// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/issue9/mux/v2/internal/handlers"
	"github.com/issue9/mux/v2/internal/syntax"
	"github.com/issue9/mux/v2/params"
)

// node.children 的数量只有达到此值时，才会为其建立 indexes 索引表。
const indexesSize = 5

// 表示路由中的节点。
type node struct {
	parent   *node
	handlers *handlers.Handlers
	children []*node
	segment  *syntax.Segment

	// 保存着 *node 实例在 children 中的下标。
	//
	// 所有节点类型为字符串的子节点，其首字符必定是不同的（相同的都提升到父节点中），
	// 根据此特性，可以将所有字符串类型的首字符做个索引，这样字符串类型节点的比较，
	// 可以通过索引排除不必要的比较操作。
	indexes map[byte]int
}

// 构建当前节点的索引表。
func (n *node) buildIndexes() {
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

// 当前节点的优先级。
//
// parent.children 根据此值进行排序。
// 不同的节点类型拥有不同的优先级，相同类型的，则有子节点的优先级低。
func (n *node) priority() int {
	// 目前节点类型只有 3 种，10
	// 可以保证在当前类型的节点进行加权时，不会超过其它节点。
	ret := int(n.segment.Type) * 10

	// 有 children 的，endpoint 必然为 false，两者不可能同时为 true
	if len(n.children) > 0 || n.segment.Endpoint {
		return ret + 1
	}

	return ret
}

// 获取指定路径下的节点，若节点不存在，则添加。
// segments 为被 syntax.Split 拆分之后的字符串数组。
func (n *node) getNode(segments []*syntax.Segment) *node {
	child := n.addSegment(segments[0])

	if len(segments) == 1 { // 最后一个节点
		return child
	}

	return child.getNode(segments[1:])
}

// 将 seg 添加到当前节点，并返回新节点，如果找到相同的节点，则直接返回该子节点。
func (n *node) addSegment(seg *syntax.Segment) *node {
	var child *node // 找到的最匹配节点
	var l int       // 最大的匹配字符数量
	for _, c := range n.children {
		if c.segment.Endpoint != seg.Endpoint { // 完全不同的节点
			continue
		}

		if c.segment.Value == seg.Value { // 有完全相同的节点
			return c
		}

		if l1 := syntax.LongestPrefix(c.segment.Value, seg.Value); l1 > l {
			l = l1
			child = c
		}
	}

	if l <= 0 { // 没有共同前缀，声明一个新的加入到当前节点
		return n.newChild(seg)
	}

	parent := splitNode(child, l)

	// seg.Value 与 parent 重叠
	if len(seg.Value) == l {
		return parent
	}

	// seg.Value[:l] 与 child.segment.Value[:l] 是相同的
	return parent.addSegment(syntax.NewSegment(seg.Value[l:]))
}

// 根据 s 内容为当前节点产生一个子节点，并返回该新节点。
// 由调用方确保 s 的语法正确性，否则可能 panic。
func (n *node) newChild(s *syntax.Segment) *node {
	child := &node{
		parent:  n,
		segment: s,
	}

	n.children = append(n.children, child)
	sort.SliceStable(n.children, func(i, j int) bool {
		return n.children[i].priority() < n.children[j].priority()
	})
	n.buildIndexes()

	return child
}

// 查找路由项，不存在返回 nil
func (n *node) find(pattern string) *node {
	for _, child := range n.children {
		if child.segment.Value == pattern {
			return child
		}

		if strings.HasPrefix(pattern, child.segment.Value) {
			nn := child.find(pattern[len(child.segment.Value):])
			if nn != nil {
				return nn
			}
		}
	} // end for

	return nil
}

// 清除路由项
func (n *node) clean(prefix string) {
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
//
// NOTE: 此函数与 node.trace 是一样的，记得同步两边的代码。
func (n *node) match(path string, params params.Params) *node {
	if len(n.indexes) > 0 && len(path) > 0 {
		node := n.children[n.indexes[path[0]]]
		if node == nil {
			goto LOOP
		}

		matched, newPath := node.matchCurrent(path, params)
		if !matched {
			goto LOOP
		}

		if nn := node.match(newPath, params); nn != nil {
			return nn
		}
	}

LOOP:
	// 即使 path 为空，也有可能子节点正好可以匹配空的内容。
	// 比如 /posts/{path:\\w*} 后面的 path 即为空节点。所以此处不判断 len(path)
	for i := len(n.indexes); i < len(n.children); i++ {
		node := n.children[i]
		matched, newPath := node.matchCurrent(path, params)
		if !matched {
			continue
		}

		if nn := node.match(newPath, params); nn != nil {
			return nn
		}

		// 不匹配，则删除写入的参数
		delete(params, n.segment.Name)
	} // end for

	// 没有子节点匹配，且 len(path)==0，可以判定与当前节点匹配
	if len(path) == 0 {
		// 虽然与当前节点匹配，但是当前节点下无任何处理函数，
		// 理论上这个节点不会匹配任何实际路径，所以不作为最终匹配节点返回。
		if n.handlers == nil || n.handlers.Len() == 0 {
			return nil
		}
		return n
	}

	return nil
}

func (n *node) matchCurrent(path string, params params.Params) (bool, string) {
	switch n.segment.Type {
	case syntax.String:
		if strings.HasPrefix(path, n.segment.Value) {
			return true, path[len(n.segment.Value):]
		}
	case syntax.Named:
		if n.segment.Endpoint {
			params[n.segment.Name] = path
			return true, path[:0]
		}

		// 为零说明前面没有命名参数，肯定不能与当前内容匹配
		if index := strings.Index(path, n.segment.Suffix); index > 0 {
			params[n.segment.Name] = path[:index]
			return true, path[index+len(n.segment.Suffix):]
		}
	case syntax.Regexp:
		locs := n.segment.Expr.FindStringSubmatchIndex(path)
		if locs == nil || locs[0] != 0 { // 不匹配
			return false, path
		}

		params[n.segment.Name] = path[:locs[3]]
		return true, path[locs[1]:]
	}

	return false, path
}

// URL 根据参数生成地址
func (n *node) url(params map[string]string) (string, error) {
	nodes := make([]*node, 0, 5)
	for curr := n; curr.parent != nil; curr = curr.parent { // 从尾部向上开始获取节点
		nodes = append(nodes, curr)
	}

	buf := new(bytes.Buffer)
	for i := len(nodes) - 1; i >= 0; i-- {
		node := nodes[i]
		switch node.segment.Type {
		case syntax.String:
			buf.WriteString(node.segment.Value)
		case syntax.Named, syntax.Regexp:
			param, exists := params[node.segment.Name]
			if !exists {
				return "", fmt.Errorf("未找到参数 %s 的值", node.segment.Name)
			}
			buf.WriteString(param)
			buf.WriteString(node.segment.Suffix) // 如果是 endpoint suffix 肯定为空
		} // end switch
	} // end for

	return buf.String(), nil
}

// 从 nodes 中删除一个 pattern 字段为指定值的元素，
//
// NOTE: 实际应该中，理论上不会出现多个相同的元素，
// 所以此处不作多余的判断。
func removeNodes(nodes []*node, pattern string) []*node {
	for index, n := range nodes {
		if n.segment.Value == pattern {
			return append(nodes[:index], nodes[index+1:]...)
		}
	}

	return nodes
}

// 将节点 n 从 pos 位置进行拆分。后一段作为当前段的子节点，并返回当前节点。
// 若 pos 大于或等于 n.pattern 的长度，则直接返回 n 不会拆分，pos 处的字符作为子节点的内容。
//
// 若 pos 位置是不可拆分的，或是 n.parent 为 nil，都将触发 panic
func splitNode(n *node, pos int) *node {
	if len(n.segment.Value) <= pos { // 不需要拆分
		return n
	}

	p := n.parent
	if p == nil {
		panic("节点必须要有一个有效的父节点，才能进行拆分")
	}

	// 先从父节点中删除老的 n
	p.children = removeNodes(p.children, n.segment.Value)
	p.buildIndexes()

	segs := n.segment.Split(pos)
	ret := p.newChild(segs[0])
	c := ret.newChild(segs[1])
	c.handlers = n.handlers
	c.children = n.children
	c.indexes = n.indexes
	for _, item := range c.children {
		item.parent = c
	}

	return ret
}
