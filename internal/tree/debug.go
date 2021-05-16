// SPDX-License-Identifier: MIT

package tree

import (
	"fmt"
	"io"
	"strings"
)

// Print 向 w 输出树状结构
func (tree *Tree) Print(w io.Writer) { tree.print(w, 0) }

func (n *node) print(w io.Writer, deep int) {
	fmt.Fprintln(w, strings.Repeat(" ", deep*4), n.segment.Value)

	for _, child := range n.children {
		child.print(w, deep+1)
	}
}
