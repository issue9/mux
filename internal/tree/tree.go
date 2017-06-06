// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"fmt"
	"net/http"

	ts "github.com/issue9/mux/internal/tree/syntax"
)

// Tree 以树节点的形式保存的路由
type Tree struct {
	*Node
}

// New 声明一个 Tree 实例
func New() *Tree {
	return &Tree{
		Node: &Node{},
	}
}

// Add 添加路由项。
//
// methods 可以为空，表示不为任何请求方法作设置。
func (tree *Tree) Add(pattern string, h http.Handler, methods ...string) error {
	ss, err := ts.Parse(pattern)
	if err != nil {
		return err
	}

	return tree.add(ss, h, methods...)
}

// Print 打印树状结构
func (tree *Tree) Print() {
	tree.print(0)
}

// GetNode 查找路由项，不存在，则返回一个新建的实例。
func (tree *Tree) GetNode(pattern string) (*Node, error) {
	n := tree.find(pattern)
	if n != nil {
		return n, nil
	}

	if err := tree.Add(pattern, nil); err != nil {
		return nil, err
	}

	n = tree.find(pattern)
	if n != nil {
		return n, nil
	}

	// 添加了，却找不到，肯定是代码问题，直接 panic
	panic(fmt.Sprintf("无法找到与 %s 相匹配的节点", pattern))
}
