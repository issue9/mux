// SPDX-License-Identifier: MIT

// Package mux 功能完备的路由中间件
package mux

import (
	"github.com/issue9/errwrap"

	"github.com/issue9/mux/v6/internal/syntax"
	"github.com/issue9/mux/v6/internal/tree"
	"github.com/issue9/mux/v6/params"
)

type (
	// MiddlewareFuncOf 中间件处理函数
	MiddlewareFuncOf[T any] func(T) T

	// InterceptorFunc 拦截器的函数原型
	InterceptorFunc = syntax.InterceptorFunc

	// Params 路由参数
	Params = params.Params
)

func applyMiddlewares[T any](h T, f ...MiddlewareFuncOf[T]) T {
	for _, ff := range f {
		h = ff(h)
	}
	return h
}

var emptyInterceptors = syntax.NewInterceptors()

// CheckSyntax 检测路由项的语法格式
//
// 路由中可通过 {} 指定参数名称，如果参数名中带 :，则 : 之后的为参数的约束条件，
// 比如 /posts/{id}.html 表示匹配任意任意字符的参数 id。/posts/{id:\d+}.html，
// 表示匹配正则表达式 \d+ 的参数 id。
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

// Methods 返回所有支持的请求方法
func Methods() []string {
	methods := make([]string, len(tree.Methods))
	copy(methods, tree.Methods)
	return methods
}

func NewParams() Params { return syntax.NewParams("") }
