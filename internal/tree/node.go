// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"net/http"
	"regexp"
	"regexp/syntax"
	"sort"
	"strings"
)

type nodeType int8

const (
	nodeTypeUnknown nodeType = iota
	nodeTypeBasic
	nodeTypeNamed
	nodeTypeRegexp
	nodeTypeWildcard
)

type node struct {
	parent   *node
	nodeType nodeType
	children []*node
	pattern  string
	handlers *handlers

	expr       *regexp.Regexp
	syntaxExpr syntax.Regexp
}

func newNode(parent *node, pattern string) *node {
	nType := nodeTypeBasic
	if strings.IndexByte(pattern, start) > -1 {
		nType = nodeTypeNamed
	}
	return &node{
		parent:   parent,
		pattern:  pattern,
		nodeType: nType,
	}
}

// 添加一条路由
func (n *node) addToChild(pattern string, h http.Handler, methods ...string) error {
	var ll int
	var nn *node
	for _, child := range n.children {
		l := prefixLen(child.pattern, pattern)

		if l > ll {
			ll = l
			nn = child
		}
	}

	if nn == nil {
		if n.children == nil {
			n.children = make([]*node, 0, 10)
		}
		n.children = append(n.children, newNode(n, pattern))
		sort.SliceStable(n.children, func(i, j int) bool { return n.children[i].nodeType < n.children[j].nodeType })
		return nil
	}

	nn.addToChild(pattern[ll:], h, methods...)

	return nil
}

// remove
func (n *node) remove(pattern string, methods ...string) {

}

func (n *node) match(path string) *node {

	// TODO
	return nil
}

// 获取该节点下与参数相对应的处理函数
func (n *node) handler(method string) http.Handler {
	if n.handlers == nil {
		return nil
	}

	return n.handlers.handler(method)
}

func (n *node) url(params map[string]string, path string) (string, error) {
	//

	return "", nil
}
