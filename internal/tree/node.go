// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"bytes"
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/issue9/mux/internal/tree/handlers"
	"github.com/issue9/mux/internal/tree/segment"
	"github.com/issue9/mux/params"
)

// Node.children 的数量只有达到此值时，才会为其建立 indexes 索引表。
const indexesSize = 5

var nodesPool = &sync.Pool{
	New: func() interface{} {
		return make([]*Node, 0, 10)
	},
}

// Node 表示路由中的节点。
type Node struct {
	parent   *Node
	handlers *handlers.Handlers
	seg      segment.Segment
	children []*Node

	// 所有节点类型为字符串的子节点，其首字符必定是不同的（相同的都提升到父节点中），
	// 根据此特性，可以将所有字符串类型的首字符做个索引，这样字符串类型节点的比较，
	// 可以通过索引排除不必要的比较操作。
	// indexes 中的 *Node 也同时存在于 children 中。
	indexes map[byte]*Node

	// 从根节点以来，到当前节点为止，所有的参数数量，
	// 即 seg.Type() 不为 segment.TypeString 的数量。
	paramsSize int
}

// 构建当前节点的索引表。
func (n *Node) buildIndexes() {
	if len(n.children) < indexesSize {
		n.indexes = nil
		return
	}

	if n.indexes == nil {
		n.indexes = make(map[byte]*Node, indexesSize)
	}

	for _, node := range n.children {
		if node.seg.Type() == segment.TypeString {
			n.indexes[node.seg.Value()[0]] = node
		}
	}
}

// 当前节点的优先级。
func (n *Node) priority() int {
	// *10 可以保证在当前类型的节点进行加权时，不会超过其它节点。
	ret := int(n.seg.Type()) * 10

	// 有 children 的，Endpoit 必然为 false，两者不可能同时为 true
	if len(n.children) > 0 || n.seg.Endpoint() {
		return ret + 1
	}

	return ret
}

// 获取指定路径下的节点，若节点不存在，则添加
func (n *Node) getNode(segments []segment.Segment) (*Node, error) {
	child, err := n.addSegment(segments[0])
	if err != nil {
		return nil, err
	}

	if len(segments) == 1 { // 最后一个节点
		return child, nil
	}

	return child.getNode(segments[1:])
}

// 将 segment.Segment 添加到当前节点，并返回新节点
func (n *Node) addSegment(s segment.Segment) (*Node, error) {
	var child *Node // 找到的最匹配节点
	var l int       // 最大的匹配字符数量

	for _, c := range n.children {
		if c.seg.Endpoint() != s.Endpoint() ||
			c.seg.Type() != s.Type() {
			continue
		}

		if segment.Equal(c.seg, s) { // 有完全相同的节点
			return c, nil
		}

		if l1 := segment.PrefixLen(c.seg.Value(), s.Value()); l1 > l {
			l = l1
			child = c
		}
	}

	if l <= 0 { // 没有共同前缀，声明一个新的加入到当前节点
		return n.newChild(s)
	}

	parent, err := splitNode(child, l)
	if err != nil {
		return nil, err
	}

	if len(s.Value()) == l {
		return parent, nil
	}

	seg, err := segment.New(s.Value()[l:])
	if err != nil {
		return nil, err
	}
	return parent.addSegment(seg)
}

// 根据 seg 内容为当前节点产生一个子节点，并返回该新节点。
func (n *Node) newChild(seg segment.Segment) (*Node, error) {
	child := &Node{
		parent:     n,
		seg:        seg,
		paramsSize: n.paramsSize,
	}

	if seg.Type() != segment.TypeString {
		child.paramsSize++
	}

	n.children = append(n.children, child)
	sort.SliceStable(n.children, func(i, j int) bool {
		return n.children[i].priority() < n.children[j].priority()
	})
	n.buildIndexes()

	return child, nil
}

// 查找路由项，不存在返回 nil
func (n *Node) find(pattern string) *Node {
	for _, child := range n.children {
		if child.seg.Value() == pattern {
			return child
		}

		if strings.HasPrefix(pattern, child.seg.Value()) {
			nn := child.find(pattern[len(child.seg.Value()):])
			if nn != nil {
				return nn
			}
		}
	} // end for

	return nil
}

// 清除路由项
func (n *Node) clean(prefix string) {
	if len(prefix) == 0 {
		n.children = n.children[:0]
		return
	}

	dels := make([]string, 0, len(n.children))
	for _, child := range n.children {
		if len(child.seg.Value()) < len(prefix) {
			if strings.HasPrefix(prefix, child.seg.Value()) {
				child.clean(prefix[len(child.seg.Value()):])
			}
		}

		if strings.HasPrefix(child.seg.Value(), prefix) {
			dels = append(dels, child.seg.Value())
		}
	}

	for _, del := range dels {
		n.children = removeNodes(n.children, del)
	}
	n.buildIndexes()
}

// 从子节点中查找与当前路径匹配的节点，若找不到，则返回 nil。
//
// NOTE: 此函数与 Node.trace 是一样的，记得同步两边的代码。
func (n *Node) match(path string) *Node {
	if len(n.indexes) > 0 {
		node := n.indexes[path[0]]
		if node == nil {
			goto LOOP
		}

		matched, newPath := node.seg.Match(path)
		if !matched {
			goto LOOP
		}

		if nn := node.match(newPath); nn != nil {
			return nn
		}
	}

LOOP:
	// 即使 path 为空，也有可能子节点正好可以匹配空的内容。
	// 比如 /posts/{path:\\w*} 后面的 path 即为空节点。所以此处不判断 len(path)
	for i := len(n.indexes); i < len(n.children); i++ {
		node := n.children[i]

		matched, newPath := node.seg.Match(path)
		if !matched {
			continue
		}

		if nn := node.match(newPath); nn != nil {
			return nn
		}
	} // end for

	// 没有子节点匹配，且 len(path)==0，可以判定与当前节点匹配
	if len(path) == 0 {
		return n
	}

	return nil
}

// Params 获取 path 在当前路由节点下的参数。
//
// 由调用方确保能正常匹配 path
func (n *Node) Params(path string) params.Params {
	if n.paramsSize == 0 { // 没有参数
		return nil
	}

	params := make(params.Params, n.paramsSize)

	nodes := n.parents()
	defer nodesPool.Put(nodes)
	for i := len(nodes) - 1; i >= 0; i-- {
		path = nodes[i].seg.Params(path, params)
	}

	return params
}

// URL 根据参数生成地址
func (n *Node) URL(params map[string]string) (string, error) {
	buf := new(bytes.Buffer)

	nodes := n.parents()
	defer nodesPool.Put(nodes)
	for i := len(nodes) - 1; i >= 0; i-- {
		if err := nodes[i].seg.URL(buf, params); err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}

// 逐级向上获取父节点，包含当前节点。
//
// NOTE: 记得将 []*Node 放回对象池中。
func (n *Node) parents() []*Node {
	nodes := nodesPool.Get().([]*Node)[:0]

	for curr := n; curr.parent != nil; curr = curr.parent { // 从尾部向上开始获取节点
		nodes = append(nodes, curr)
	}

	return nodes
}

// Handler 获取该节点下与参数相对应的处理函数
func (n *Node) Handler(method string) http.Handler {
	if n.handlers == nil {
		return nil
	}

	return n.handlers.Handler(method)
}

// 从 nodes 中删除一个 pattern 字段为指定值的元素，
// 若存在多个同名的，则只删除第一个匹配的元素。
//
// NOTE: 实际应该中，理论上不会出现多个相同的元素，
// 所以此处不作多余的判断。
func removeNodes(nodes []*Node, pattern string) []*Node {
	lastIndex := len(nodes) - 1
	for index, n := range nodes {
		if n.seg.Value() != pattern {
			continue
		}

		switch {
		case len(nodes) == 1: // 只有一个元素
			return nodes[:0]
		case index == lastIndex: // 最后一个元素
			return nodes[:lastIndex]
		default:
			return append(nodes[:index], nodes[index+1:]...)
		}
	} // end for

	return nodes
}

// 将节点 n 从 pos 位置进行拆分。后一段作为当前段的子节点，并返回当前节点。
// 若 pos 大于或等于 n.pattern 的长度，则直接返回 n 不会拆分，pos 处的字符作为子节点的内容。
//
// NOTE: 调用者需确保 pos 位置是可拆分的。
func splitNode(n *Node, pos int) (*Node, error) {
	if len(n.seg.Value()) <= pos { // 不需要拆分
		return n, nil
	}

	p := n.parent
	if p == nil {
		return nil, errors.New("splitNode:节点必须要有一个有效的父节点，才能进行拆分")
	}

	// 先从父节点中删除老的 n
	p.children = removeNodes(p.children, n.seg.Value())
	p.buildIndexes()

	seg, err := segment.New(n.seg.Value()[:pos])
	if err != nil {
		return nil, err
	}
	ret, err := p.newChild(seg)
	if err != nil {
		return nil, err
	}

	seg, err = segment.New(n.seg.Value()[pos:])
	if err != nil {
		return nil, err
	}
	c, err := ret.newChild(seg)
	if err != nil {
		return nil, err
	}
	c.handlers = n.handlers
	c.children = n.children
	c.indexes = n.indexes
	for _, item := range c.children {
		item.parent = c
	}

	return ret, nil
}
