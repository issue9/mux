// SPDX-License-Identifier: MIT

// Package group 提供了按条件进行分组路由的功能
//
// 比如按指定的域名进行分类路由等。
package group

import "net/http"

// Matcher 验证一个请求是否符合要求
type Matcher interface {
	Match(*http.Request) bool
}

// MatcherFunc 用于转一个 Match(http.Request) bool 转换成 Matcher 接口
type MatcherFunc func(*http.Request) bool

func (f MatcherFunc) Match(r *http.Request) bool {
	return f(r)
}

// Any 匹配任意内容
func Any(*http.Request) bool { return true }
