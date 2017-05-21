// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"fmt"
	"strings"

	"github.com/issue9/mux/internal/syntax"
)

// 表示命名参数中的某个节点
type node struct {
	value    string // 当前节点的值。
	endByte  byte   // 下一个节点的起始字符
	isString bool   // 当前节点是否为一个字符串
	isLast   bool   // 是否为最后的节点
}

// 命名参数是正则表达式的简化版本，
// 在一个参数中未指定正则表达式，默认使用此类型，
// 性能会比用正则好一些。
type named struct {
	*base
	nodes []*node
}

// 声明一个新的 named 实例。
func newNamed(s *syntax.Syntax) *named {
	b := newBase(s)

	str := s.Patterns[len(s.Patterns)-1]
	if b.wildcard {
		str = str[:len(str)-2]
		if len(str) == 0 {
			s.Patterns = s.Patterns[:len(s.Patterns)-1]
		} else {
			s.Patterns[len(s.Patterns)-1] = str
		}
	}

	nodes := make([]*node, 0, len(s.Patterns))
	for index, str := range s.Patterns {
		last := index >= len(s.Patterns)-1
		if str[0] == syntax.Start {
			endByte := byte('/')
			if !last {
				endByte = s.Patterns[index+1][0]
			}
			nodes = append(nodes, &node{
				value:    str[1 : len(str)-1],
				isString: false,
				endByte:  endByte,
				isLast:   last,
			})
		} else {
			nodes = append(nodes, &node{
				value:    str,
				isString: true,
				isLast:   last,
			})
		}
	}

	return &named{
		base:  b,
		nodes: nodes,
	}
}

func (n *named) Priority() int {
	return syntax.TypeNamed
}

func (n *named) Match(path string) (bool, map[string]string) {
	rawPath := path
	for i, node := range n.nodes {
		if node.isString { // 普通字符串节点
			if !strings.HasPrefix(path, node.value) {
				return false, nil
			}
			path = path[len(node.value):]

			if node.isLast {
				if len(path) == 0 {
					return true, n.params(rawPath)
				}
				return false, nil
			}
		} else { // 带命名的节点
			if !node.isLast {
				index := strings.Index(path, n.nodes[i+1].value) // 下一个必定是字符串节点
				if index == -1 {
					return false, nil
				}

				path = path[index:]
			} else { // 最后一个节点了
				index := strings.IndexByte(path, node.endByte)
				if index == -1 {
					if n.wildcard {
						return false, nil
					}
					return true, n.params(rawPath)
				}

				if !n.wildcard {
					return false, nil
				}
				return true, n.params(rawPath)
			}
		} // end if
	} // end for

	return true, n.params(rawPath)
}

func (n *named) params(path string) map[string]string {
	// 由调用者 n.match() 保证 path 参数始终是与当前路由项匹配的。
	// 所以以下代码不再作是否匹配的检测工作。

	params := make(map[string]string, len(n.nodes))
	for i, node := range n.nodes {
		if node.isString { // 普通字符串节点
			if node.isLast {
				return params
			}

			path = path[len(node.value):]
		} else { // 带命名的节点
			if !node.isLast {
				index := strings.Index(path, n.nodes[i+1].value) // 下一个必定是字符串节点
				params[node.value] = path[:index]
				path = path[index:]
			} else { // 最后一个节点了
				index := strings.IndexByte(path, node.endByte)
				if index == -1 {
					params[node.value] = path
					return params
				}

				params[node.value] = path[:index]
				return params
			}
		} // end if
	} // end for

	return params
}

func (n *named) URL(params map[string]string, path string) (string, error) {
	ret := make([]byte, 0, 500)

	for _, node := range n.nodes {
		if node.isString {
			ret = append(ret, node.value...)
			continue
		}

		if param, exists := params[node.value]; exists {
			ret = append(ret, param...)
		} else {
			return "", fmt.Errorf("参数 %v 未指定", node.value)
		}
	}

	if n.wildcard {
		ret = append(ret, '/')
		ret = append(ret, path...)
	}

	return string(ret), nil
}
