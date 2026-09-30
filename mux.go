// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package mux 适用第三方框架实现可定制的路由
package mux

import (
	"iter"
	"net/http"

	"github.com/issue9/errwrap"

	"github.com/issue9/mux/v10/internal/syntax"
	"github.com/issue9/mux/v10/internal/trace"
	"github.com/issue9/mux/v10/internal/tree"
)

var emptyInterceptors = syntax.NewInterceptors()

// CheckSyntax 检测路由项的语法格式
func CheckSyntax(pattern string) error {
	_, err := emptyInterceptors.Split(pattern)
	return err
}

// Trace 一种简单的处理 TRACE 请求的方法
//
// 可以传递给 [NewRouter] 或是 [NewGroup] 的 trace 参数。
func Trace(w http.ResponseWriter, r *http.Request, body bool) error { return trace.Trace(w, r, body) }

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

// Methods 遍历支持当前库支持的方法
func Methods() iter.Seq[string] {
	return func(yield func(v string) bool) {
		for _, m := range tree.Methods {
			if !yield(m) {
				break
			}
		}
	}
}

// AnyMethods 遍历支持 [Router.Any] 的方法
func AnyMethods() iter.Seq[string] {
	return func(yield func(v string) bool) {
		for _, m := range tree.AnyMethods {
			if !yield(m) {
				break
			}
		}
	}
}
