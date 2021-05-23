// SPDX-License-Identifier: MIT

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

// Match 实现 Matcher 接口
func (f MatcherFunc) Match(r *http.Request) (*http.Request, bool) { return f(r) }

// Any 匹配任意内容
func Any(r *http.Request) (*http.Request, bool) { return r, true }

// And 按顺序符合每一个要求
//
// 前一个对象返回的实例将作为下一个对象的输入参数。
func And(m ...Matcher) Matcher {
	return MatcherFunc(func(r *http.Request) (rr *http.Request, ok bool) {
		rr = r
		for _, mm := range m {
			rr, ok = mm.Match(rr)
			if !ok {
				return nil, false
			}
		}
		return rr, true
	})
}

// Or 仅需符合一个要求
func Or(m ...Matcher) Matcher {
	return MatcherFunc(func(r *http.Request) (*http.Request, bool) {
		for _, mm := range m {
			if rr, ok := mm.Match(r); ok {
				return rr, true
			}
		}
		return nil, false
	})
}

// And 需同时符合每一个要求
func AndFunc(f ...func(*http.Request) (*http.Request, bool)) Matcher { return And(f2i(f...)...) }

// Or 仅需符合一个要求
func OrFunc(f ...func(*http.Request) (*http.Request, bool)) Matcher { return Or(f2i(f...)...) }

func f2i(f ...func(*http.Request) (*http.Request, bool)) []Matcher {
	ms := make([]Matcher, 0, len(f))
	for _, ff := range f {
		ms = append(ms, MatcherFunc(ff))
	}
	return ms
}
