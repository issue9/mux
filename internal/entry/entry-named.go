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
	str := s.patterns[len(s.patterns)-1]
	if strings.HasSuffix(str, "/*") {
		str = str[:len(str)-2]
		if len(str) == 0 {
			s.patterns = s.patterns[:len(s.patterns)-1]
		} else {
			s.patterns[len(s.patterns)-1] = str
		}
	}

	names := make([]*name, 0, len(s.patterns))
	for index, str := range s.patterns {
		if str[0] == syntaxStart {
			endByte := byte('/')
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

func (n *named) priority() int {
	if n.wildcard {
		return typeNamed + 100
	}

	return typeNamed
}

// Entry.Match
func (n *named) match(path string) bool {
	for i, name := range n.names {
		islast := (i == len(n.names)-1)

		if name.isString {
			if !strings.HasPrefix(path, name.name) {
				return false
			}
			path = path[len(name.name):]
		} else {
			index := strings.IndexByte(path, name.endByte)
			if !islast {
				path = path[index:]
			} else {
				if index < 0 { // 没有 / 符号了
					if n.wildcard { // 通配符，但是没有后续内容
						return false
					}
					return true
				}

				if n.wildcard { // 通配符，但是没有后续内容
					return true
				}

				return false
			}
		} // end if
	} // end false
	return true
}

// Entry.Params
func (n *named) Params(path string) map[string]string {
	params := make(map[string]string, len(n.names))

	for i, name := range n.names {
		islast := (i == len(n.names)-1)

		if name.isString {
			if !strings.HasPrefix(path, name.name) {
				return nil
			}
			path = path[len(name.name):]
		} else {
			index := strings.IndexByte(path, name.endByte)

			if !islast {
				params[name.name] = path[:index]
				path = path[index:]
			} else {
				if index < 0 { // 没有 / 符号了
					params[name.name] = path
					break
				}

				if n.wildcard {
					params[name.name] = path[:index]
					break
				}
			}
		}
	} // end for
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
