// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"fmt"
	"io"
	"strings"

	"github.com/issue9/mux/params"
)

// Print 向 w 输出树状结构
func (tree *Tree) Print(w io.Writer) {
	tree.print(w, 0)
}

// Trace 向 w 输出详细的节点匹配过程
func (tree *Tree) Trace(w io.Writer, path string) {
	params := make(params.Params, 10)
	tree.trace(w, 0, path, params)
}

// NOTE: 此函数与 node.match 是一样的，记得同步两边的代码。
func (n *node) trace(w io.Writer, deep int, path string, params params.Params) *node {
	if len(n.indexes) > 0 {
		node := n.children[n.indexes[path[0]]]
		fmt.Fprint(w, strings.Repeat(" ", deep*4), node.pattern, "---", node.nodeType, "---", path)

		if node == nil {
			fmt.Fprintln(w, "(!matched)")
			goto LOOP
		}

		matched, newPath := node.matchCurrent(path, params)
		if !matched {
			fmt.Fprintln(w, "(!matched)")
			goto LOOP
		}

		fmt.Fprintln(w, "(continue)")
		if nn := node.match(newPath, params); nn != nil {
			return nn
		}
	}

LOOP:
	// 即使 path 为空，也有可能子节点正好可以匹配空的内容。
	// 比如 /posts/{path:\\w*} 后面的 path 即为空节点。所以此处不判断 len(path)
	for i := len(n.indexes); i < len(n.children); i++ {
		node := n.children[i]
		fmt.Fprint(w, strings.Repeat(" ", deep*4), node.pattern, "---", node.nodeType, "---", path)
		matched, newPath := node.matchCurrent(path, params)
		if !matched {
			fmt.Fprintln(w, "(!matched)")
			continue
		}

		fmt.Fprintln(w, "(continue)")
		if nn := node.trace(w, deep+1, newPath, params); nn != nil {
			return nn
		}
	} // end for

	if len(path) == 0 {
		fmt.Fprintln(w, strings.Repeat(" ", (deep-1)*4), n.pattern, "---", n.nodeType, "---", path, "(matched)")
		return n
	}

	return nil
}

// 向 w 输出节点的树状结构
func (n *node) print(w io.Writer, deep int) {
	fmt.Fprintln(w, strings.Repeat(" ", deep*4), n.pattern)

	for _, child := range n.children {
		child.print(w, deep+1)
	}
}

// 获取当前路由下有处理函数的节点数量
func (n *node) len() int {
	var cnt int
	for _, child := range n.children {
		cnt += child.len()
	}

	if n.handlers != nil && n.handlers.Len() > 0 {
		cnt++
	}

	return cnt
}

// 仅上面的 trace 用到
func (t nodeType) String() string {
	switch t {
	case nodeTypeNamed:
		return "named"
	case nodeTypeRegexp:
		return "regexp"
	case nodeTypeString:
		return "string"
	default:
		return "<unknown>"
	}
}
