// SPDX-License-Identifier: MIT

// Package mux 功能完备的路由中间件
//
// 语法
//
// 路由支持以 {} 的形式包含参数，比如：/posts/{id}.html，id 在解析时会解析任意字符。
// 也可以在 {} 中约束参数的范围，比如 /posts/{id:\\d+}.html，表示 id 只能匹配数字。
// 路由地址可以是 ascii 字符，但是参数名称如果是非 ascii，在正则表达式中无法使用。
package mux

import (
	"net/http"
	"strings"

	"github.com/issue9/mux/v5/internal/syntax"
	"github.com/issue9/mux/v5/internal/tree"
	"github.com/issue9/mux/v5/params"
)

// Params 获取路由的参数集合
func Params(r *http.Request) params.Params { return params.Get(r) }

// IsWell 语法格式是否正确
//
// 如果出错，则会返回具体的错误信息。
func IsWell(pattern string) error {
	_, err := syntax.Split(pattern)
	return err
}

// Methods 返回所有支持的请求方法
func Methods() []string {
	methods := make([]string, len(tree.Methods))
	copy(methods, tree.Methods)
	return methods
}

// 清除路径中的重复的 / 字符
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}

	var b strings.Builder
	b.Grow(len(p) + 1)

	if p[0] != '/' {
		b.WriteByte('/')
	}

	index := strings.Index(p, "//")
	if index == -1 {
		b.WriteString(p)
		return b.String()
	}

	b.WriteString(p[:index+1])

	slash := true
	for i := index + 2; i < len(p); i++ {
		if p[i] == '/' {
			if slash {
				continue
			}
			slash = true
		} else {
			slash = false
		}
		b.WriteByte(p[i])
	}

	return b.String()
}
