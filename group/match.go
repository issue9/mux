// SPDX-License-Identifier: MIT

// Package group 提供了按条件进行分组路由的功能
package group

import "net/http"

// Matcher 验证一个请求是否符合要求
//
// Matcher 常用于路由项的前置判断，用于对路由项进行归类，
// 符合同一个 Matcher 的路由项，再各自进行路由。比如按域名进行分组路由。
type Matcher interface {
	// Match 验证请求是否符合当前对象的要求
	//
	// 不应该直接对 r 作修改，而是将修改对象以返回值的形式返回。
	Match(r *http.Request) (*http.Request, bool)
}

// MatcherFunc 用于将 Match(*http.Request) (*http.Request, bool) 转换成 Matcher 接口
type MatcherFunc func(*http.Request) (*http.Request, bool)

func (f MatcherFunc) Match(r *http.Request) (*http.Request, bool) { return f(r) }

// Any 匹配任意内容
func Any(r *http.Request) (*http.Request, bool) { return r, true }
