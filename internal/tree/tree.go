// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package tree 提供了以树形结构保存路由项的相关操作。
package tree

import (
	"fmt"
	"io"
	"net/http"

	"github.com/issue9/mux/internal/method"
	ts "github.com/issue9/mux/internal/tree/syntax"
)

// Tree 以树节点的形式保存的路由。
//
// 多段路由项，会提取其中的相同的内容组成树状结构的节点。
// 比如以下路由项：
//  /posts/{id}/author
//  /posts/{id}/author/emails
//  /posts/{id}/author/profile
//  /posts/1/author
// 会被转换成以下结构
//  /posts
//     |
//     +---- 1/author
//     |
//     +---- {id}/author
//               |
//               +---- profile
//               |
//               +---- emails
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
// methods 可以为空，表示添加除 OPTIONS 之外所有支持的请求方法。
func (tree *Tree) Add(pattern string, h http.Handler, methods ...string) error {
	ss, err := ts.Parse(pattern)
	if err != nil {
		return err
	}

	if len(methods) == 0 {
		methods = method.Default
	}

	return tree.add(ss, h, methods...)
}

// Match 查找与 path 匹配的节点
func (tree *Tree) Match(path string) *Node {
	return tree.match(path)
}

// Clean 清除路由项
func (tree *Tree) Clean(prefix string) {
	tree.clean(prefix)
}

// Remove 移除路由项
//
// methods 可以为空，表示删除所有内容。
func (tree *Tree) Remove(pattern string, methods ...string) error {
	if len(methods) == 0 {
		methods = method.Supported
	}

	return tree.remove(pattern, methods...)
}

// Print 向 w 输出树状结构
func (tree *Tree) Print(w io.Writer) {
	tree.print(w, 0)
}

// GetNode 查找路由项，不存在，则返回一个新建的实例。
func (tree *Tree) GetNode(pattern string) (*Node, error) {
	n := tree.find(pattern)
	if n != nil {
		return n, nil
	}

	// 没有找到，则尝试添加一个空的节点
	ss, err := ts.Parse(pattern)
	if err != nil {
		return nil, err
	}
	if err := tree.add(ss, nil); err != nil { // 不添加任何处理方法。
		return nil, err
	}

	// 再次查找刚才添加的节点
	n = tree.find(pattern)
	if n == nil {
		panic(fmt.Sprintf("添加了 %s 节点，却无法找到与 %s 相匹配的节点", pattern, pattern))
	}
	return n, nil
}
