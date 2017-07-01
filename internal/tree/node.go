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

var nodesPool = &sync.Pool{
	New: func() interface{} {
		return make([]*Node, 0, 10)
	},
}

// Node 表示路由中的节点。
type Node struct {
	parent   *Node
	children []*Node
	handlers *handlers.Handlers
	seg      segment.Segment
}

// 当前节点的优先级。
func (n *Node) priority() int {
	// 有 children 的，Endpoit 必然为 false，两者不可能同时为 true
	if len(n.children) > 0 || n.seg.Endpoint() {
		return int(n.seg.Type()) + 1
	}

	return int(n.seg.Type())
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

		if l1 := segment.PrefixLen(c.seg.Pattern(), s.Pattern()); l1 > l {
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

	if len(s.Pattern()) == l {
		return parent, nil
	}

	seg, err := segment.New(s.Pattern()[l:])
	if err != nil {
		return nil, err
	}
	return parent.addSegment(seg)
}

// 根据 seg 内容为当前节点产生一个子节点，并返回该新节点。
func (n *Node) newChild(seg segment.Segment) (*Node, error) {
	child := &Node{
		parent: n,
		seg:    seg,
	}

	n.children = append(n.children, child)
	sort.SliceStable(n.children, func(i, j int) bool {
		return n.children[i].priority() < n.children[j].priority()
	})

	return child, nil
}

// 查找路由项，不存在返回 nil
func (n *Node) find(pattern string) *Node {
	for _, child := range n.children {
		if len(child.seg.Pattern()) < len(pattern) {
			if !strings.HasPrefix(pattern, child.seg.Pattern()) {
				continue
			}

			nn := child.find(pattern[len(child.seg.Pattern()):])
			if nn != nil {
				return nn
			}
		}

		if child.seg.Pattern() == pattern {
			return child
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
		if len(child.seg.Pattern()) < len(prefix) {
			if strings.HasPrefix(prefix, child.seg.Pattern()) {
				child.clean(prefix[len(child.seg.Pattern()):])
			}
		}

		if strings.HasPrefix(child.seg.Pattern(), prefix) {
			dels = append(dels, child.seg.Pattern())
		}
	}

	for _, del := range dels {
		n.children = removeNodes(n.children, del)
	}
}

// 从子节点中查找与当前路径匹配的节点，若找不到，则返回 nil。
//
// NOTE: 此函数与 Node.trace 是一样的，记得同步两边的代码。
func (n *Node) match(path string) *Node {
	if len(n.children) == 0 && len(path) == 0 {
		return n
	}

	for _, node := range n.children {
		matched, newPath := node.seg.Match(path)
		if !matched {
			continue
		}

		// 即使 newPath 为空，也有可能子节点正好可以匹配空的内容。
		// 比如 /posts/{path:\\w*} 后面的 path 即为空节点。
		if nn := node.match(newPath); nn != nil {
			return nn
		}

		if len(newPath) == 0 { // 没有子节点匹配，才判断是否与当前节点匹配
			return node
		}
	} // end for

	return nil
}

// Params 获取 path 在当前路由节点下的参数。
//
// 由调用方确保能正常匹配 path
func (n *Node) Params(path string) params.Params {
	nodes := n.getParents()
	defer nodesPool.Put(nodes)

	params := make(params.Params, len(nodes))
	for i := len(nodes) - 1; i >= 0; i-- {
		node := nodes[i]
		if node.seg == nil {
			continue
		}

		path = node.seg.Params(path, params)
	} // end for LOOP

	return params
}

// URL 根据参数生成地址
func (n *Node) URL(params map[string]string) (string, error) {
	nodes := n.getParents()
	defer nodesPool.Put(nodes)

	buf := new(bytes.Buffer)
	for i := len(nodes) - 1; i >= 0; i-- {
		node := nodes[i]
		if node.seg == nil {
			continue
		}

		if err := node.seg.URL(buf, params); err != nil {
			return "", err
		}
	} // end for

	return buf.String(), nil
}

// 逐级向上获取父节点，包含当前节点。
//
// NOTE: 记得将 []*Node 放回对象池中。
func (n *Node) getParents() []*Node {
	nodes := nodesPool.Get().([]*Node)[:0]

	for curr := n; curr != nil; curr = curr.parent { // 从尾部向上开始获取节点
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
		if n.seg.Pattern() != pattern {
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
	if len(n.seg.Pattern()) <= pos { // 不需要拆分
		return n, nil
	}

	p := n.parent
	if p == nil {
		return nil, errors.New("splitNode:节点必须要有一个有效的父节点，才能进行拆分")
	}

	// 先从父节点中删除老的 n
	p.children = removeNodes(p.children, n.seg.Pattern())

	seg, err := segment.New(n.seg.Pattern()[:pos])
	if err != nil {
		return nil, err
	}
	ret, err := p.newChild(seg)
	if err != nil {
		return nil, err
	}

	seg, err = segment.New(n.seg.Pattern()[pos:])
	if err != nil {
		return nil, err
	}
	c, err := ret.newChild(seg)
	if err != nil {
		return nil, err
	}
	c.handlers = n.handlers
	c.children = n.children
	for _, item := range c.children {
		item.parent = c
	}

	return ret, nil
}
