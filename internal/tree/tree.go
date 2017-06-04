// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"net/http"

	ts "github.com/issue9/mux/internal/tree/syntax"
)

// Tree 以树节点的形式保存的路由
type Tree struct {
	*node
}

// New 声明一个 Tree 实例
func New() *Tree {
	return &Tree{
		node: &node{},
	}
}

// Clean 清除路由项
func (tree *Tree) Clean(prefix string) {
	tree.clean(prefix)
}

// Remove 移除路由项
func (tree *Tree) Remove(pattern string, methods ...string) error {
	ss, err := ts.Parse(pattern)
	if err != nil {
		return err
	}

	return tree.remove(ss, methods...)
}

// Add 添加路由项
func (tree *Tree) Add(pattern string, h http.Handler, methods ...string) error {
	ss, err := ts.Parse(pattern)
	if err != nil {
		return err
	}

	return tree.add(ss, h, methods...)
}

// Match 匹配路由项
func (tree *Tree) Match(path string) Noder {
	return tree.match(path)
}
