// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package entry 路由项的相关操作。
package entry

import (
	"fmt"
	"net/http"

	"github.com/issue9/mux/internal/syntax"
)

// Entry 表示一类资源的进入点，拥有统一的路由匹配模式。
type Entry interface {
	// 返回路由的匹配字符串
	Pattern() string

	// 与当前路由项是否匹配，若匹配的话，也将同时返回参数。
	// params 参数仅在 matched 为 true 时，才会有意义
	Match(path string) (matched bool, params map[string]string)

	// 优先级，越小越靠前。
	// 当多个路由项对某个请求都匹配时，将根据优先级来确定哪条最终会获得匹配。
	Priority() int

	// 添加请求方法及其对应的处理函数。
	//
	// 若已经存在，则返回错误。
	// 若 method == http.MethodOptions，则可以去覆盖默认的处理方式。
	Add(handler http.Handler, methods ...string) error

	// 移除指定方法的处理函数。empty 表示当前路由项中已经没有任何处理方法。
	Remove(method ...string) (empty bool)

	// 根据参数生成一条路径。
	// params 为替换匹配字符串的参数；
	// path 在有通配符的情况下，会替代通配符所代码的内容。
	URL(params map[string]string, path string) (string, error)

	// 获取指定请求方法对应的处理函数，若不存在，则返回 nil。
	Handler(method string) http.Handler

	// 手动设置 OPTIONS 的 Allow 报头。不调用此函数，
	// 会自动根据当前的 Add 和 Remove 调整 Allow 报头，
	// 调用 SetAllow() 之后，这些自动设置不再启作用。
	SetAllow(string)
}

// New 声明一个 Entry 实例。
func New(pattern string) (Entry, error) {
	s, err := syntax.New(pattern)
	if err != nil {
		return nil, err

	}

	switch s.Type {
	case syntax.TypeRegexp:
		return newRegexp(s)
	case syntax.TypeNamed:
		return newNamed(s), nil
	case syntax.TypeBasic:
		return newBasic(s), nil
	default:
		return nil, fmt.Errorf("未知的类型：%v", s.Type)
	}
}
