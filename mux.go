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
