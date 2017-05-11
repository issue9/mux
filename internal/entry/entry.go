// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import "net/http"

// Entry 表示一类资源的进入点，拥有统一的路由匹配模式。
type Entry interface {
	// 返回路由的匹配字符串
	pattern() string

	// 与当前是否匹配，若匹配的话，也将同时返回参数。
	// params 参数仅在 matched 为 true 时，才会有意义
	match(path string) (matched bool, params map[string]string)

	// 优先级，用于 entries 排定匹配优先级用，越小越靠前
	priority() int

	// 添加请求方法及其对应的处理函数。
	//
	// 若已经存在，则返回错误。
	// 若 method == http.MethodOptions，则可以去覆盖默认的处理方式。
	add(handler http.Handler, methods ...string) error

	// 移除指定方法的处理函数。若 Entry 中已经没有任何 http.Handler，则返回 true
	//
	// 可以通过指定 http.MethodOptions 的方式，来强制删除 OPTIONS 请求方法的处理。
	remove(method ...string) (empty bool)

	// 根据参数生成一条路径。
	// params 为替换匹配字符串的参数，
	// path 在有通配符的情况下，会替代通配符所代码的内容。
	URL(params map[string]string, path string) (string, error)

	// 获取指定请求方法对应的 http.Handler 实例，若不存在，则返回 nil。
	Handler(method string) http.Handler

	// 手动设置 OPTIONS 的 Allow 报头。不调用此函数，
	// 会自动根据当前的 Add 和 Remove 调整 Allow 报头，
	// 调用 SetAllow() 之后，这些自动设置不再启作用。
	SetAllow(string)
}

// newEntry 根据内容，生成相应的 Entry 接口实例。
//
// pattern 匹配内容。
func newEntry(pattern string) (Entry, error) {
	s, err := parse(pattern)
	if err != nil {
		return nil, err

	}

	if s.nType == typeRegexp {
		return newRegexp(pattern, s)
	} else if s.nType == typeNamed {
		return newNamed(pattern, s), nil
	}

	return newBasic(s.patterns[0]), nil
}
