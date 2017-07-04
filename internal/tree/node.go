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
	"sort"
	"strings"

	"github.com/issue9/mux/internal/tree/handlers"
	"github.com/issue9/mux/internal/tree/segment"
	"github.com/issue9/mux/params"
)

// Node.children 的数量只有达到此值时，才会为其建立 indexes 索引表。
const indexesSize = 5

// Node 表示路由中的节点。
type Node struct {
	parent   *Node
	handlers *handlers.Handlers
	pattern  string
	nodeType segment.Type
	children []*Node

	// 用于表示当前是否为终点，仅对 Type 为 TypeRegexp 和 TypeNamed 有用。
	// 此值为 true，该节点的优先级会比同类型的节点低，以便优先对比其它非最终节点。
	endpoint bool

	// 参数名称，仅对 Type 为 TypeRegexp 和 TypeNamed 有用。
	name   string
	suffix string
	expr   *regexp.Regexp

	// 所有节点类型为字符串的子节点，其首字符必定是不同的（相同的都提升到父节点中），
	// 根据此特性，可以将所有字符串类型的首字符做个索引，这样字符串类型节点的比较，
	// 可以通过索引排除不必要的比较操作。
	// indexes 中保存着 *Node 实例在 children 中的下标。
	indexes map[byte]int
}

// 构建当前节点的索引表。
func (n *Node) buildIndexes() {
	if len(n.children) < indexesSize {
		n.indexes = nil
		return
	}

	if n.indexes == nil {
		n.indexes = make(map[byte]int, indexesSize)
	}

	for index, node := range n.children {
		if node.nodeType == segment.TypeString {
			n.indexes[node.pattern[0]] = index
		}
	}
}

// 当前节点的优先级。
func (n *Node) priority() int {
	// *10 可以保证在当前类型的节点进行加权时，不会超过其它节点。
	ret := int(n.nodeType) * 10

	// 有 children 的，Endpoit 必然为 false，两者不可能同时为 true
	if len(n.children) > 0 || n.endpoint {
		return ret + 1
	}

	return ret
}

// 获取指定路径下的节点，若节点不存在，则添加。
// segments 为被 segment.Split 拆分之后的字符串数组。
func (n *Node) getNode(segments []string) (*Node, error) {
	child, err := n.addSegment(segments[0])
	if err != nil {
		return nil, err
	}

	if len(segments) == 1 { // 最后一个节点
		return child, nil
	}

	return child.getNode(segments[1:])
}

// 将 seg 添加到当前节点，并返回新节点
func (n *Node) addSegment(seg string) (*Node, error) {
	var child *Node // 找到的最匹配节点
	var l int       // 最大的匹配字符数量
	for _, c := range n.children {
		if c.endpoint != segment.IsEndpoint(seg) {
			continue
		}

		if c.pattern == seg { // 有完全相同的节点
			return c, nil
		}

		if l1 := segment.LongestPrefix(c.pattern, seg); l1 > l {
			l = l1
			child = c
		}
	}

	if l <= 0 { // 没有共同前缀，声明一个新的加入到当前节点
		return n.newChild(seg)
	}

	parent, err := splitNode(child, l)
	if err != nil {
		return nil, err
	}

	if len(seg) == l {
		return parent, nil
	}

	return parent.addSegment(seg[l:])
}

// 将路由语法转换成正则表达式语法，比如：
//  {id:\\d+}/author => (?P<id>\\d+)
var repl = strings.NewReplacer(string("{"), "(?P<",
	string(":"), ">",
	string("}"), ")")

// 根据 s 内容为当前节点产生一个子节点，并返回该新节点。
func (n *Node) newChild(s string) (*Node, error) {
	child := &Node{
		parent:   n,
		pattern:  s,
		endpoint: segment.IsEndpoint(s),
		nodeType: segment.TypeString,
	}
	if strings.IndexByte(s, '{') > -1 {
		child.nodeType = segment.TypeNamed
		endIndex := strings.IndexByte(s, '}')
		if endIndex == -1 {
			return nil, fmt.Errorf("无效的路由语法：%s", s)
		}
		child.name = s[1:endIndex]
		child.suffix = s[endIndex+1:]
	}
	if strings.IndexByte(s, ':') > -1 {
		child.nodeType = segment.TypeRegexp
		index := strings.IndexByte(s, ':')

		r := repl.Replace(s)
		expr, err := regexp.Compile(r)
		if err != nil {
			return nil, err
		}
		child.name = s[1:index]
		child.expr = expr
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
		if child.pattern == pattern {
			return child
		}

		if strings.HasPrefix(pattern, child.pattern) {
			nn := child.find(pattern[len(child.pattern):])
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
	n.buildIndexes()
}

// 从子节点中查找与当前路径匹配的节点，若找不到，则返回 nil。
//
// NOTE: 此函数与 Node.trace 是一样的，记得同步两边的代码。
func (n *Node) match(path string, params params.Params) *Node {
	if len(n.indexes) > 0 {
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
		delete(params, n.name)
	} // end for

	// 没有子节点匹配，且 len(path)==0，可以判定与当前节点匹配
	if len(path) == 0 {
		return n
	}

	return nil
}

// fun
func (n *Node) matchCurrent(path string, params params.Params) (bool, string) {
	switch n.nodeType {
	case segment.TypeNamed:
		if n.endpoint {
			params[n.name] = path
			return true, path[:0]
		}

		// 为零说明前面没有命名参数，肯定不能与当前内容匹配
		if index := strings.Index(path, n.suffix); index > 0 {
			params[n.name] = path[:index]
			return true, path[index+len(n.suffix):]
		}
	case segment.TypeRegexp:
		locs := n.expr.FindStringSubmatchIndex(path)
		if locs == nil || locs[0] != 0 { // 不匹配
			return false, path
		}

		params[n.name] = path[:locs[3]]
		return true, path[locs[1]:]
	case segment.TypeString:
		if strings.HasPrefix(path, n.pattern) {
			return true, path[len(n.pattern):]
		}
	}

	return false, path
}

// URL 根据参数生成地址
func (n *Node) URL(params map[string]string) (string, error) {
	nodes := make([]*Node, 0, 5)
	for curr := n; curr.parent != nil; curr = curr.parent { // 从尾部向上开始获取节点
		nodes = append(nodes, curr)
	}

	buf := new(bytes.Buffer)
	for i := len(nodes) - 1; i >= 0; i-- {
		node := nodes[i]
		switch node.nodeType {
		case segment.TypeNamed:
			param, exists := params[node.name]
			if !exists {
				return "", fmt.Errorf("未找到参数 %s 的值", node.name)
			}
			buf.WriteString(param)
			buf.WriteString(node.suffix) // 如果是 endpoint suffix 肯定为空
		case segment.TypeRegexp:
			param, found := params[node.name]
			if !found {
				return "", fmt.Errorf("缺少参数 %s", n.name)
			}

			index := strings.IndexByte(node.pattern, '}')
			url := strings.Replace(node.pattern, node.pattern[:index+1], param, 1)

			if _, err := buf.WriteString(url); err != nil {
				return "", err
			}
		case segment.TypeString:
			buf.WriteString(string(node.pattern))
		}
	}

	return buf.String(), nil
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
// 若 pos 大于或等于 n.pattern 的长度，则直接返回 n 不会拆分，pos 处的字符作为子节点的内容。
//
// NOTE: 调用者需确保 pos 位置是可拆分的。
func splitNode(n *Node, pos int) (*Node, error) {
	if len(n.pattern) <= pos { // 不需要拆分
		return n, nil
	}

	p := n.parent
	if p == nil {
		return nil, errors.New("splitNode:节点必须要有一个有效的父节点，才能进行拆分")
	}

	// 先从父节点中删除老的 n
	p.children = removeNodes(p.children, n.pattern)
	p.buildIndexes()

	ret, err := p.newChild(n.pattern[:pos])
	if err != nil {
		return nil, err
	}

	c, err := ret.newChild(n.pattern[pos:])
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
