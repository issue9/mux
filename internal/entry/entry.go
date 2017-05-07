// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import "net/http"

// 表示 Entry 接口的类型
const (
	TypeBasic = iota + 1
	TypeStatic
	TypeRegexp
	TypeNamed
)

// Entry 表示一类资源的进入点，拥有统一的路由匹配模式。
type Entry interface {
	// 返回路由的匹配字符串
	Pattern() string

	// url 与当前的匹配程度：
	//  -1 表示完全不匹配；
	//  0  表示完全匹配；
	//  >0 表示部分匹配，值越小表示匹配程度越高。
	Match(url string) int

	// 获取路由中的参数，非正则匹配或是无参数返回 nil。
	Params(url string) map[string]string

	// 接口的实现类型
	Type() int

	// 获取指定请求方法对应的 http.Handler 实例，若不存在，则返回 nil。
	Handler(method string) http.Handler

	// 添加请求方法及其对应的处理函数。
	//
	// 若已经存在，则返回错误。
	// 若 method == http.MethodOptions，则可以去覆盖默认的处理方式。
	Add(handler http.Handler, methods ...string) error

	// 移除指定方法的处理函数。若 Entry 中已经没有任何 http.Handler，则返回 true
	//
	// 可以通过指定 http.MethodOptions 的方式，来强制删除 OPTIONS 请求方法的处理。
	Remove(method ...string) (empty bool)

	// 根据参数生成一条路径。
	URL(params map[string]string) (string, error)

	// 手动设置 OPTIONS 的 Allow 报头。不调用此函数，
	// 会自动根据当前的 Add 和 Remove 调整 Allow 报头，
	// 调用 SetAllow() 之后，这些自动设置不再启作用。
	SetAllow(string)
}

// New 根据内容，生成相应的 Entry 接口实例。
//
// pattern 匹配内容。
// h 对应的 http.Handler，外层调用者确保该值不能为 nil.
func New(pattern string, h http.Handler) (Entry, error) {
	s, err := parse(pattern)
	if err != nil {
		return nil, err

	}

	if s.nType == TypeRegexp {
		return newRegexp(pattern, s)
	} else if s.nType == TypeNamed {
		return newNamed(pattern, s), nil
	}

	pattern = s.patterns[0]
	return newBasic(pattern), nil
}
