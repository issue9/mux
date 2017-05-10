// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"fmt"
	"strings"
)

// 表示命名参数中的某个节点
type node struct {
	value    string // 当前节点的值。
	endByte  byte   // 下一个节点的起始字符
	isString bool   // 当前节点是否为一个字符串
}

// 命名参数是正则表达式的简化版本，
// 在一个参数中未指定正则表达式，默认使用此类型，
// 性能会比用正则好一些。
type named struct {
	*base
	nodes []*node
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

	names := make([]*node, 0, len(s.patterns))
	for index, str := range s.patterns {
		if str[0] == syntaxStart {
			endByte := byte('/')
			if index < len(s.patterns)-1 {
				endByte = s.patterns[index+1][0]
			}
			names = append(names, &node{
				value:    str[1 : len(str)-1],
				isString: false,
				endByte:  endByte,
			})
		} else {
			names = append(names, &node{
				value:    str,
				isString: true,
			})
		}
	}

	return &named{
		base:  newBase(pattern),
		nodes: names,
	}
}

func (n *named) priority() int {
	if n.wildcard {
		return typeNamed + 100
	}

	return typeNamed
}

// Entry.match
func (n *named) match(path string) (bool, map[string]string) {
	params := make(map[string]string, len(n.nodes))

	for i, name := range n.nodes {
		islast := (i == len(n.nodes)-1)

		if name.isString { // 普通字符串节点
			if !strings.HasPrefix(path, name.value) {
				return false, nil
			}

			if islast {
				return (path == name.value), params
			}

			path = path[len(name.value):]
		} else { // 带命名的节点
			index := strings.IndexByte(path, name.endByte)
			if !islast {
				if index == -1 {
					return false, nil
				}

				params[name.value] = path[:index]
				path = path[index:]
			} else { // 最后一个节点了
				if index == -1 {
					params[name.value] = path
					return !n.wildcard, params
				}

				params[name.value] = path[:index]
				return n.wildcard, params
			}
		} // end if
	} // end false
	return true, params
}

// URL
func (n *named) URL(params map[string]string, path string) (string, error) {
	ret := ""
	for _, name := range n.nodes {
		if name.isString {
			ret += name.value
			continue
		}

		if param, exists := params[name.value]; exists {
			ret += param
		} else {
			return "", fmt.Errorf("参数 %v 未指定", name.value)
		}
	}

	if n.wildcard {
		ret += "/" + path
	}

	return ret, nil
}
