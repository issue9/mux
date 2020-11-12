// SPDX-License-Identifier: MIT

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

// 向 w 输出节点的树状结构
func (n *node) print(w io.Writer, deep int) {
	fmt.Fprintln(w, strings.Repeat(" ", deep*4), n.segment.Value)

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
