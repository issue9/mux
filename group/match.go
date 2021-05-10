// SPDX-License-Identifier: MIT

// Package group 提供了按条件进行分组路由的功能
package group

import "net/http"

// Matcher 验证一个请求是否符合要求
//
// Matcher 常用于路由项的前置判断，用于对路由项进行归类，
// 符合同一个 Matcher 的路由项，再各自进行路由。 比如按域名进行分组路由。
type Matcher interface {
	// Match 验证请求是否符合当前对象的要求
	//
	// 可能会对参数做出修改，比如通过 context.WithValue 等。
	Match(*http.Request) (*http.Request, bool)
}

// MatcherFunc 用于转一个 Match(http.Request) bool 转换成 Matcher 接口
type MatcherFunc func(*http.Request) (*http.Request, bool)

func (f MatcherFunc) Match(r *http.Request) (*http.Request, bool) { return f(r) }

// Any 匹配任意内容
func Any(r *http.Request) (*http.Request, bool) { return r, true }
