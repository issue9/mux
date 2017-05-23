// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package list 提供了对 entry.Entry 元素的各种存储和匹配方式。
package list

import (
	"net/http"

	"github.com/issue9/mux/internal/entry"
	"github.com/issue9/mux/internal/syntax"
)

// 存储 entry.Entry 的容器需要实现的接口
type entries interface {
	// 清除所有的路由项，在 prefix 不为空的情况下，
	// 则为删除所有路径前缀为 prefix 的匹配项。
	clean(prefix string)

	// 移除指定的路由项。
	//
	// 当未指定 methods 时，将删除所有 method 匹配的项。
	// 指定错误的 methods 值，将自动忽略该值。
	remove(pattern string, methods ...string)

	// 添加一条路由数据。
	add(s *syntax.Syntax, h http.Handler, methods ...string) error

	// 查找指定匹配模式下的 entry.Entry，不存在，则声明新的
	entry(s *syntax.Syntax) (entry.Entry, error)

	// 查找与 path 最匹配的路由项以及对应的参数
	match(path string) (entry.Entry, map[string]string)

	// 返回当前列表的元素数量
	len() int

	printDeep(deep int)
}
