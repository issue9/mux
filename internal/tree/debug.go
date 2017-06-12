// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"fmt"
	"io"
	"strings"
)

// Print 向 w 输出树状结构
func (tree *Tree) Print(w io.Writer) {
	tree.print(w, 0)
}

// Trace 向 w 输出详细的节点匹配过程
func (tree *Tree) Trace(w io.Writer, path string) {
	tree.trace(w, 0, path)
}

func (n *Node) trace(w io.Writer, deep int, path string) *Node {
	if len(n.children) == 0 && len(path) == 0 {
		fmt.Fprintln(w, strings.Repeat(" ", deep*4), n.pattern)
		return n
	}

	for _, node := range n.children {
		matched, newPath := node.matchCurrent(path)
		if !matched {
			continue
		}

		// 即使 newPath 为空，也有可能子节点正好可以匹配空的内容。
		// 比如 /posts/{path:\\w*} 后面的 path 即为空节点。
		if nn := node.trace(w, deep+1, newPath); nn != nil {
			fmt.Fprintln(w, strings.Repeat(" ", deep*4), nn.pattern)
			return nn
		}

		if len(newPath) == 0 { // 没有子节点匹配，才判断是否与当前节点匹配
			fmt.Fprintln(w, strings.Repeat(" ", deep*4), n.pattern)
			return node
		}
	} // end for

	return nil
}

// 向 w 输出节点的树状结构
func (n *Node) print(w io.Writer, deep int) {
	fmt.Fprintln(w, strings.Repeat(" ", deep*4), n.pattern)

	for _, child := range n.children {
		child.print(w, deep+1)
	}
}

// Len 获取当前路由下有处理函数的节点数量
func (n *Node) Len() int {
	var cnt int
	for _, child := range n.children {
		cnt += child.Len()
	}

	if n.handlers != nil && n.handlers.Len() > 0 {
		cnt++
	}

	return cnt
}
