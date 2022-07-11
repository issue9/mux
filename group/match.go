// SPDX-License-Identifier: MIT

package group

import (
	"net/http"

	"github.com/issue9/mux/v7/types"
)

// Matcher 验证一个请求是否符合要求
//
// Matcher 用于路由项的前置判断，用于对路由项进行归类，
// 符合同一个 Matcher 的路由项，再各自进行路由。比如按域名进行分组路由。
type Matcher interface {
	// Match 验证请求是否符合当前对象的要求
	//
	// 返回值表示是否匹配成功；
	Match(*http.Request, *types.Context) bool
}

type MatcherFunc func(*http.Request, *types.Context) bool

func (f MatcherFunc) Match(r *http.Request, p *types.Context) bool { return f(r, p) }

func anyRouter(*http.Request, *types.Context) bool { return true }

// AndMatcher 按顺序符合每一个要求
//
// 前一个对象返回的实例将作为下一个对象的输入参数。
func AndMatcher(m ...Matcher) Matcher {
	return MatcherFunc(func(r *http.Request, ctx *types.Context) bool {
		for _, mm := range m {
			if !mm.Match(r, ctx) {
				return false
			}
		}
		return true
	})
}

// OrMatcher 仅需符合一个要求
func OrMatcher(m ...Matcher) Matcher {
	return MatcherFunc(func(r *http.Request, ctx *types.Context) bool {
		for _, mm := range m {
			if ok := mm.Match(r, ctx); ok {
				return true
			}
		}
		return false
	})
}

// AndMatcherFunc 需同时符合每一个要求
func AndMatcherFunc(f ...func(*http.Request, *types.Context) bool) Matcher {
	return AndMatcher(f2i(f...)...)
}

// OrMatcherFunc 仅需符合一个要求
func OrMatcherFunc(f ...func(*http.Request, *types.Context) bool) Matcher {
	return OrMatcher(f2i(f...)...)
}

func f2i(f ...func(*http.Request, *types.Context) bool) []Matcher {
	ms := make([]Matcher, 0, len(f))
	for _, ff := range f {
		ms = append(ms, MatcherFunc(ff))
	}
	return ms
}
