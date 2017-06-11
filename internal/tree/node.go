// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"regexp/syntax"
	"sort"
	"strings"
	"sync"

	"github.com/issue9/mux/internal/tree/handlers"
	ts "github.com/issue9/mux/internal/tree/syntax"
)

var nodesPool = &sync.Pool{
	New: func() interface{} {
		return make([]*Node, 0, 10)
	},
}

// Node 表示路由中的节点。
type Node struct {
	parent   *Node
	nodeType ts.Type
	children []*Node
	pattern  string
	handlers *handlers.Handlers

	// 用于表示当前节点是否为终点，仅对 nodeType 为 TypeRegexp 和 TypeNamed 有用。
	// 此值为 true，该节点的优先级会比同类型的节点低，以便优先对比其它非最终节点，
	// 且该节点之后不会有子节点。
	endpoint bool

	// 命名参数特有的参数
	name   string
	suffix string

	// 正则参数特有的参数
	expr       *regexp.Regexp
	syntaxExpr *syntax.Regexp
}

// 当前节点的优先级，根据节点类型来判断，
func (n *Node) priority() int {
	// 有 children 的，endpoit 必须为 false
	if len(n.children) > 0 || n.endpoint {
		return int(n.nodeType) + 1
	}

	return int(n.nodeType)
}

// 添加一条路由。当 methods 为空时，表示仅添加节点，而不添加任何处理函数。
func (n *Node) add(segments []*ts.Segment, h http.Handler, methods ...string) error {
	child, err := n.addSegment(segments[0])
	if err != nil {
		return err
	}

	if len(segments) > 1 {
		return child.add(segments[1:], h, methods...)
	}

	// 最后一个节点
	if child.handlers == nil {
		child.handlers = handlers.New()
	}
	return child.handlers.Add(h, methods...)
}

// 将 ts.Segment 添加到当前节点，并返回新节点
func (n *Node) addSegment(s *ts.Segment) (*Node, error) {
	var child *Node // 找到的最匹配节点
	var l int       // 最大的匹配字符数量

	for _, c := range n.children {
		if c.endpoint != s.Endpoint ||
			c.nodeType != s.Type {
			continue
		}

		if c.endpoint == s.Endpoint && // 有完全相同的节点
			c.pattern == s.Value &&
			c.nodeType == s.Type {
			return c, nil
		}

		if l1 := ts.PrefixLen(c.pattern, s.Value); l1 > l {
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

	if len(s.Value) == l {
		return parent, nil
	}
	return parent.addSegment(ts.NewSegment(s.Value[l:]))
}

// 根据 seg 内容为当前节点产生一个子节点，并返回该新节点。
func (n *Node) newChild(seg *ts.Segment) (*Node, error) {
	child := &Node{
		parent:   n,
		pattern:  seg.Value,
		nodeType: seg.Type,
	}

	switch seg.Type {
	case ts.TypeNamed:
		endIndex := strings.IndexByte(seg.Value, ts.NameEnd)
		if endIndex == -1 { // TODO 由 ts.Segment 保证语法是有效的，是否更佳？
			return nil, fmt.Errorf("无效的路由语法：%s", seg.Value)
		}
		child.suffix = seg.Value[endIndex+1:]
		child.name = seg.Value[1:endIndex]
		child.endpoint = seg.Endpoint
	case ts.TypeRegexp:
		reg := ts.Regexp(seg.Value)
		// TODO: 如果能保证 seg.Value 是正确的，使用 regexp.MustCompile 更好
		expr, err := regexp.Compile(reg)
		if err != nil {
			return nil, err
		}

		syntaxExpr, err := syntax.Parse(reg, syntax.Perl)
		if err != nil {
			return nil, err
		}
		child.expr = expr
		child.syntaxExpr = syntaxExpr
		child.endpoint = seg.Endpoint
	} // end switch

	n.children = append(n.children, child)
	sort.SliceStable(n.children, func(i, j int) bool {
		return n.children[i].priority() < n.children[j].priority()
	})

	return child, nil
}

// 查找路由项，不存在返回 nil
func (n *Node) find(pattern string) *Node {
	for _, child := range n.children {
		if len(child.pattern) < len(pattern) {
			if !strings.HasPrefix(pattern, child.pattern) {
				continue
			}

			nn := child.find(pattern[len(child.pattern):])
			if nn != nil {
				return nn
			}
		}

		if child.pattern == pattern {
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
		if len(child.pattern) < len(prefix) {
			if strings.HasPrefix(prefix, child.pattern) {
				child.clean(prefix[len(child.pattern):])
			}
		}

		if strings.HasPrefix(child.pattern, prefix) {
			dels = append(dels, child.pattern)
		}
	}

	for _, del := range dels {
		n.children = removeNodes(n.children, del)
	}
}

// 移除路由项
func (n *Node) remove(pattern string, methods ...string) error {
	child := n.find(pattern)

	if child == nil {
		return fmt.Errorf("不存在的节点 %v", pattern)
	}

	if child.handlers == nil {
		if len(child.children) == 0 {
			child.parent.children = removeNodes(child.parent.children, child.pattern)
		}
		return nil
	}

	if child.handlers.Remove(methods...) && len(child.children) == 0 {
		child.parent.children = removeNodes(child.parent.children, child.pattern)
	}
	return nil
}

// 从子节点中查找与当前路径匹配的节点，若找不到，则返回 nil
func (n *Node) match(path string) *Node {
	if len(n.children) == 0 && len(path) == 0 {
		return n
	}

	for _, node := range n.children {
		matched, newPath := node.matchCurrent(path)
		if !matched {
			continue
		}

		// 即使 newPath 为空，也有可能子节点正好可以匹配空的内容。
		// 比如 /posts/{path:\\w*} 后面的 path 即为空节点。
		if nn := node.match(newPath); nn != nil {
			return nn
		}

		if len(newPath) == 0 {
			return node
		}
	} // end for

	return nil
}

// 确定当前节点是否与 path 匹配。
func (n *Node) matchCurrent(path string) (bool, string) {
	switch n.nodeType {
	case ts.TypeString:
		if strings.HasPrefix(path, n.pattern) {
			return true, path[len(n.pattern):]
		}
	case ts.TypeNamed:
		if n.endpoint {
			return true, path[:0]
		}

		index := strings.Index(path, n.suffix)
		if index > 0 { // 为零说明前面没有命名参数，肯定不正确
			return true, path[index+len(n.suffix):]
		}
	case ts.TypeRegexp:
		loc := n.expr.FindStringIndex(path)
		if loc == nil || loc[0] != 0 { // 不匹配
			break
		}

		if loc[1] == len(path) {
			return true, path[:0]
		}
		return true, path[loc[1]+1:]
	default: // nodeType 错误，肯定是代码级别的错误，直接 panic
		panic("无效的 nodeType 值")
	} // end switch

	return false, path
}

// Params 获取 path 在当前路由节点下的参数。
//
// 由调用方确保能正常匹配 path
func (n *Node) Params(path string) map[string]string {
	nodes := n.getParents()
	defer nodesPool.Put(nodes)

	params := make(map[string]string, len(nodes))

	for i := len(nodes) - 1; i >= 0; i-- {
		node := nodes[i]
		switch node.nodeType {
		case ts.TypeString:
			path = path[len(node.pattern):]
		case ts.TypeNamed:
			if node.endpoint {
				params[node.name] = path
				return params
			}

			index := strings.Index(path, node.suffix)
			params[node.name] = path[:index]
			path = path[index+len(node.suffix):]
		case ts.TypeRegexp:
			subexps := node.expr.SubexpNames()
			args := node.expr.FindStringSubmatch(path)
			for index, name := range subexps {
				if len(name) > 0 && index < len(args) {
					params[name] = args[index]
				}
			}

			path = path[len(args[0]):]
		}
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
		switch node.nodeType {
		case ts.TypeString:
			buf.WriteString(node.pattern)
		case ts.TypeNamed:
			param, exists := params[node.name]
			if !exists {
				return "", fmt.Errorf("未找到参数 %s 的值", node.name)
			}
			buf.WriteString(param)
			buf.WriteString(node.suffix) // 如果是 endpoint suffix 肯定为空
		case ts.TypeRegexp:
			url := node.syntaxExpr.String()
			subs := append(node.syntaxExpr.Sub, node.syntaxExpr)
			for _, sub := range subs {
				if len(sub.Name) == 0 {
					continue
				}

				param, exists := params[sub.Name]
				if !exists {
					return "", fmt.Errorf("未找到参数 %v 的值", sub.Name)
				}
				url = strings.Replace(url, sub.String(), param, -1)
			}

			buf.WriteString(url)
		}
	} // end for

	return buf.String(), nil
}

// 逐级向上获取父节点，包含当前节点。
// NOTE: 记得将 []*Node 放回对象池中。
func (n *Node) getParents() []*Node {
	nodes := nodesPool.Get().([]*Node)[:0]

	for curr := n; curr != nil; curr = curr.parent { // 从尾部向上开始获取节点
		nodes = append(nodes, curr)
	}

	return nodes
}

// SetAllow 设置当前节点的 allow 报头
func (n *Node) SetAllow(allow string) {
	if n.handlers == nil {
		n.handlers = handlers.New()
	}

	n.handlers.SetAllow(allow)
}

// Handler 获取该节点下与参数相对应的处理函数
func (n *Node) Handler(method string) http.Handler {
	if n.handlers == nil {
		return nil
	}

	return n.handlers.Handler(method)
}

// 向客户端打印节点的树状结构
func (n *Node) print(deep int) {
	fmt.Println(strings.Repeat(" ", deep*4), n.pattern)

	for _, child := range n.children {
		child.print(deep + 1)
	}
}

// 获取路由数量
func (n *Node) len() int {
	var cnt int
	for _, child := range n.children {
		cnt += child.len()
	}

	if n.handlers != nil && n.handlers.Len() > 0 {
		cnt++
	}

	return cnt
}

func removeNodes(nodes []*Node, pattern string) []*Node {
	lastIndex := len(nodes) - 1
	for index, n := range nodes {
		if n.pattern != pattern {
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
// 若 pos 大于或等于 n.pattern 的长度，则直接返回 n 不会拆分。
//
// NOTE: 调用者需确保 pos 位置是可拆分的。
func splitNode(n *Node, pos int) (*Node, error) {
	if len(n.pattern) <= pos { // 不需要拆分
		return n, nil
	}

	p := n.parent
	if p == nil {
		return nil, errors.New("split:节点必须要有一个有效的父节点，才能进行拆分")
	}

	// 先从父节点中删除老的 n
	p.children = removeNodes(p.children, n.pattern)

	ret, err := p.newChild(ts.NewSegment(n.pattern[:pos]))
	if err != nil {
		return nil, err
	}

	c, err := ret.newChild(ts.NewSegment(n.pattern[pos:]))
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
