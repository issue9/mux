// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"regexp/syntax"
	"sort"
	"strings"

	ts "github.com/issue9/mux/internal/tree/syntax"
)

// Noder 表示一个节点
type Noder interface {
	// 获取当前节点下指定请求方法对应的处理函数
	Handler(method string) http.Handler
}

type node struct {
	parent   *node
	nodeType ts.Type
	children []*node
	pattern  string
	handlers *handlers

	// 命名参数特有的参数
	name   string // 缓存着名称
	suffix string // 保存着命名之后的字符串内容

	// 正则特有的参数
	expr       *regexp.Regexp
	syntaxExpr *syntax.Regexp
}

// 当前节点的优先级，根据节点类型来判断，
// 若类型相同时，则有子节点的优先级低一些，但不会超过不同节点类型。
func (n *node) priority() int {
	if len(n.children) == 0 {
		return int(n.nodeType)
	}

	return int(n.nodeType) + 1
}

// 添加一条路由
func (n *node) add(segments []*ts.Segment, h http.Handler, methods ...string) error {
	current := segments[0]
	isLast := len(segments) == 1
	var child *node

	for _, c := range n.children {
		if c.nodeType != current.Type || c.pattern != current.Value {
			continue
		}

		child = c
		break
	}

	// 没有找到相关的子节点，新建一个 node 实例
	if child == nil {
		child = &node{
			parent:   n,
			pattern:  current.Value,
			nodeType: current.Type,
		}

		switch current.Type {
		case ts.TypeNamed:
			endIndex := strings.IndexByte(current.Value, ts.NameEnd)
			child.suffix = current.Value[endIndex+1:]
			child.name = current.Value[1:endIndex]
		case ts.TypeRegexp:
			reg := ts.Regexp(current.Value)
			expr, err := regexp.Compile(reg)
			if err != nil {
				return err
			}

			syntaxExpr, err := syntax.Parse(reg, syntax.Perl)
			if err != nil {
				return err
			}
			child.expr = expr
			child.syntaxExpr = syntaxExpr
		case ts.TypeWildcard:
			child.name = current.Value[1 : len(current.Value)-1]
		} // end switch

		n.children = append(n.children, child)
		sort.SliceStable(n.children, func(i, j int) bool {
			return n.children[i].priority() < n.children[j].priority()
		})
	}

	if isLast {
		if child.handlers == nil {
			child.handlers = newHandlers()
		}
		return child.handlers.add(h, methods...)
	}
	return child.add(segments[1:], h, methods...)
}

// 清除路由项
func (n *node) clean(prefix string) {
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

// remove
func (n *node) remove(segments []*ts.Segment, methods ...string) error {
	current := segments[0]
	isLast := len(segments) == 1
	var child *node

	for _, c := range n.children {
		if c.pattern == current.Value {
			child = c
			break
		}
	}

	if child == nil {
		return fmt.Errorf("不存在的节点 %v", current.Value)
	}

	if !isLast {
		return child.remove(segments[1:], methods...)
	}

	if child.handlers.remove(methods...) {
		n.children = removeNodes(n.children, current.Value)
	}
	return nil
}

// 从子节点中查找与当前路径匹配的节点，若找不到，则返回 nil
func (n *node) match(path string) *node {
	for _, node := range n.children {
		matched := false
		newPath := path

		switch node.nodeType {
		case ts.TypeBasic:
			matched = strings.HasPrefix(path, node.pattern)
			if matched {
				newPath = path[len(node.pattern):]
			}
		case ts.TypeNamed:
			index := strings.Index(path, node.suffix)
			if index > 0 { // 为零说明前面没有命名参数，肯定不正确
				matched = true
				newPath = path[index+len(node.suffix):]
			}
		case ts.TypeRegexp:
			loc := node.expr.FindStringIndex(path)
			if loc != nil && loc[0] == 0 {
				matched = true
				if loc[1] >= len(path) {
					newPath = path[:0]
				} else {
					newPath = path[loc[1]+1:]
				}
			}
		case ts.TypeWildcard:
			matched = true
			newPath = path[:0]
		default:
			// nodeType 错误，肯定是代码级别的错误，直接 panic
			panic("无效的 nodeType 值")
		}

		if matched {
			if len(newPath) == 0 { // 当前为最后节点
				return node
			}

			if nn := node.match(newPath); nn != nil {
				return nn
			}
		}
	} // end for

	return nil
}

// params 由调用方确保能正常匹配 path
func (n *node) params(path string) map[string]string {
	nodes := n.getParents()

	params := make(map[string]string, 10)

	for i := len(nodes) - 1; i >= 0; i-- {
		node := nodes[i]
		switch node.nodeType {
		case ts.TypeBasic:
			path = path[len(node.pattern):]
		case ts.TypeNamed:
			index := strings.Index(path, node.suffix)
			if index > 0 { // 为零说明前面没有命名参数，肯定不正确
				params[node.name] = path[:index]
				path = path[index+len(node.suffix):]
			}
		case ts.TypeRegexp:
			// 正确匹配正则表达式，则获相关的正则表达式命名变量。
			subexps := node.expr.SubexpNames()
			args := node.expr.FindStringSubmatch(path)
			for index, name := range subexps {
				if len(name) > 0 && index < len(args) {
					params[name] = args[index]
				}
			}

			path = path[len(args[0]):]
		case ts.TypeWildcard:
			params[node.name] = path
		}
	}

	return params
}

func (n *node) url(params map[string]string) (string, error) {
	nodes := n.getParents()
	buf := new(bytes.Buffer)

LOOP:
	for i := len(nodes) - 1; i >= 0; i-- {
		node := nodes[i]
		switch node.nodeType {
		case ts.TypeBasic:
			buf.WriteString(node.pattern)
		case ts.TypeNamed:
			buf.WriteString(params[node.name])
			buf.WriteString(node.suffix)
		case ts.TypeRegexp:
			if node.syntaxExpr == nil {
				continue LOOP
			}

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
		case ts.TypeWildcard:
			buf.WriteString(params[node.name])
		}
	}

	return buf.String(), nil
}

// 逐级向上获取父节点，包含当前节点。
func (n *node) getParents() []*node {
	nodes := make([]*node, 0, 10) // 从尾部向上开始获取节点

	for curr := n; curr != nil; curr = curr.parent {
		nodes = append(nodes, curr)
	}

	return nodes
}

// Handler 获取该节点下与参数相对应的处理函数
func (n *node) Handler(method string) http.Handler {
	if n.handlers == nil {
		return nil
	}

	return n.handlers.handler(method)
}

// 向客户端打印节点的树状结构
func (n *node) print(deep int) {
	fmt.Println(strings.Repeat(" ", deep*4), n.pattern)

	for _, child := range n.children {
		child.print(deep + 1)
	}
}

// 获取路由数量
func (n *node) len() int {
	var cnt int
	for _, child := range n.children {
		cnt += child.len()
	}

	if n.handlers != nil && len(n.handlers.handlers) > 0 {
		cnt++
	}

	return cnt
}

func removeNodes(nodes []*node, pattern string) []*node {
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
