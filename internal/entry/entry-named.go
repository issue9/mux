// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"strings"
)

type name struct {
	name     string // 名称，或是值
	endByte  byte   // 结束后的第一个字符
	isString bool
}

type named struct {
	*items
	names []*name
}

// Entry.Type
func (n *named) Type() int {
	return TypeNamed
}

// Entry.Match
func (n *named) Match(path string) int {
	for _, name := range n.names {
		if name.isString {
			if !strings.HasPrefix(path, name.name) {
				return -1
			}
			path = path[len(n.names):]
		} else {
			if name.endByte == 0 { // 最后了
				return 0
			}

			index := strings.IndexByte(path, name.endByte)
			path = path[index:]
		}
	} // end false
	return 0
}

// Entry.Params
func (n *named) Params(url string) map[string]string {
	// TODO
	return nil
}
