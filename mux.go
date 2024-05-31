// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package mux 适用第三方框架实现可定制的路由
//
// # 语法
//
// 路由参数采用大括号包含，内部包含名称和规则两部分：`{name:rule}`，
// 其中的 name 表示参数的名称，rule 表示对参数的约束规则。
//
// name 可以包含 `-` 前缀，表示在实际执行过程中，不捕获该名称的对应的值，
// 可以在一定程序上提升性能。
//
// rule 表示对参数的约束，一般为正则或是空，为空表示匹配任意值，
// 拦截器一栏中有关 rule 的高级用法。以下是一些常见的示例。
//
//	/posts/{id}.html                  // 匹配 /posts/1.html
//	/posts-{id}-{page}.html           // 匹配 /posts-1-10.html
//	/posts/{path:\\w+}.html           // 匹配 /posts/2020/11/11/title.html
//	/tags/{tag:\\w+}/{path}           // 匹配 /tags/abc/title.html
package mux

import (
	"slices"

	"github.com/issue9/errwrap"

	"github.com/issue9/mux/v9/internal/syntax"
	"github.com/issue9/mux/v9/internal/tree"
)

var emptyInterceptors = syntax.NewInterceptors()

// CheckSyntax 检测路由项的语法格式
func CheckSyntax(pattern string) error {
	_, err := emptyInterceptors.Split(pattern)
	return err
}

// URL 根据参数生成地址
//
// pattern 为路由项的定义内容；
// params 为路由项中的参数，键名为参数名，键值为参数值。
//
// NOTE: 仅仅是将 params 填入到 pattern 中， 不会判断参数格式是否正确。
func URL(pattern string, params map[string]string) (string, error) {
	if len(params) == 0 {
		return pattern, nil
	}

	buf := errwrap.StringBuilder{}
	buf.Grow(len(pattern))
	if err := emptyInterceptors.URL(&buf, pattern, params); err != nil {
		return "", err
	}
	return buf.String(), buf.Err
}

// Methods 返回库支持的请求方法
func Methods() []string { return slices.Clone(tree.Methods) }

// AnyMethods 返回 [Router.Any] 中添加的请求方法
func AnyMethods() []string { return slices.Clone(tree.AnyMethods) }
