// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"fmt"
	"strings"
)

type name struct {
	name     string // 名称，或是值
	endByte  byte   // 结束后的第一个字符
	isString bool
}

type named struct {
	*base
	names []*name
}

// 声明一个新的 named 实例。
// pattern 并不实际参与 synatax 的计算。
func newNamed(pattern string, s *syntax) *named {
	names := make([]*name, 0, len(s.patterns))
	for index, str := range s.patterns {
		if str[0] == syntaxStart {
			var endByte byte
			if index < len(s.patterns)-1 {
				endByte = s.patterns[index+1][0]
			}
			names = append(names, &name{
				name:     str[1 : len(str)-1],
				isString: false,
				endByte:  endByte,
			})
		} else {
			names = append(names, &name{
				name:     str,
				isString: true,
			})
		}
	}
	return &named{
		base:  newBase(pattern),
		names: names,
	}
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
			path = path[len(name.name):]
		} else {
			if name.endByte == 0 { // 最后了
				if strings.IndexByte(path, '/') >= 0 {
					return -1
				}
				return 0
			}

			index := strings.IndexByte(path, name.endByte)
			path = path[index:]
		}
	} // end false
	return 0
}

// Entry.Params
func (n *named) Params(path string) map[string]string {
	params := make(map[string]string, len(n.names))

	for _, name := range n.names {
		if name.isString {
			if !strings.HasPrefix(path, name.name) {
				return nil
			}
			path = path[len(name.name):]
		} else {
			if name.endByte == 0 { // 最后了
				if strings.IndexByte(path, '/') >= 0 {
					return nil
				}
				params[name.name] = path
				break
			}

			index := strings.IndexByte(path, name.endByte)
			params[name.name] = path[:index]
			path = path[index:]
		}
	}
	return params
}

// URL
func (n *named) URL(params map[string]string) (string, error) {
	ret := ""
	for _, name := range n.names {
		if name.isString {
			ret += name.name
			continue
		}

		if param, exists := params[name.name]; exists {
			ret += param
		} else {
			return "", fmt.Errorf("参数 %v 未指定", name.name)
		}
	}

	return ret, nil
}
