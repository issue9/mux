// SPDX-License-Identifier: MIT

// Package mux 功能完备的路由中间件
package mux

import (
	"net/http"

	"github.com/issue9/mux/v5/internal/syntax"
	"github.com/issue9/mux/v5/internal/tree"
	"github.com/issue9/mux/v5/params"
)

// Params 获取路由中的参数集合
func Params(r *http.Request) params.Params { return params.Get(r) }

// CheckSyntax 检测路由项的语法格式
//
// 路由中可通过 {} 指定参数名称，如果参数名中带 :，则 : 之后的为参数的约束条件，
// 比如 /posts/{id}.html 表示匹配任意任意字符的参数 id。/posts/{id:\d+}.html，
// 表示匹配正则表达式 \d+ 的参数 id。；
func CheckSyntax(pattern string) error {
	_, err := syntax.Split(pattern)
	return err
}

// Methods 返回所有支持的请求方法
func Methods() []string {
	methods := make([]string, len(tree.Methods))
	copy(methods, tree.Methods)
	return methods
}
