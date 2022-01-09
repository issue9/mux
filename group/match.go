// SPDX-License-Identifier: MIT

package group

import (
	"net/http"

	"github.com/issue9/mux/v6/internal/syntax"
	"github.com/issue9/mux/v6/params"
)

// Matcher 验证一个请求是否符合要求
//
// Matcher 用于路由项的前置判断，用于对路由项进行归类，
// 符合同一个 Matcher 的路由项，再各自进行路由。比如按域名进行分组路由。
type Matcher interface {
	// Match 验证请求是否符合当前对象的要求
	//
	// ps 为匹配过程中生成的参数信息，可以返回 nil；
	// ok 表示是否匹配成功；
	Match(r *http.Request) (ps params.Params, ok bool)
}

// MatcherFunc 用于将 Match(*http.Request) (*http.Request, bool) 转换成 Matcher 接口
type MatcherFunc func(*http.Request) (params.Params, bool)

// Match 实现 Matcher 接口
func (f MatcherFunc) Match(r *http.Request) (params.Params, bool) { return f(r) }

// Any 匹配任意内容
func Any(*http.Request) (params.Params, bool) { return nil, true }

// And 按顺序符合每一个要求
//
// 前一个对象返回的实例将作为下一个对象的输入参数。
func And(m ...Matcher) Matcher {
	return MatcherFunc(func(r *http.Request) (ps params.Params, ok bool) {
		ps = syntax.NewParams("")

		for _, mm := range m {
			ps2, ok := mm.Match(r)
			if !ok {
				return nil, false
			}
			if ps2.Count() == 0 {
				continue
			}

			ps2.Range(func(k, v string) {
				ps.Set(k, v)
			})
		}
		return ps, true
	})
}

// Or 仅需符合一个要求
func Or(m ...Matcher) Matcher {
	return MatcherFunc(func(r *http.Request) (params.Params, bool) {
		for _, mm := range m {
			if rr, ok := mm.Match(r); ok {
				return rr, true
			}
		}
		return nil, false
	})
}

// AndFunc 需同时符合每一个要求
func AndFunc(f ...func(*http.Request) (params.Params, bool)) Matcher { return And(f2i(f...)...) }

// OrFunc 仅需符合一个要求
func OrFunc(f ...func(*http.Request) (params.Params, bool)) Matcher { return Or(f2i(f...)...) }

func f2i(f ...func(*http.Request) (params.Params, bool)) []Matcher {
	ms := make([]Matcher, 0, len(f))
	for _, ff := range f {
		ms = append(ms, MatcherFunc(ff))
	}
	return ms
}
