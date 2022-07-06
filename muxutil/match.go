// SPDX-License-Identifier: MIT

package muxutil

import (
	"net/http"

	"github.com/issue9/mux/v7"
	"github.com/issue9/mux/v7/types"
)

// AndMatcher 按顺序符合每一个要求
//
// 前一个对象返回的实例将作为下一个对象的输入参数。
func AndMatcher(m ...mux.Matcher) mux.Matcher {
	return mux.MatcherFunc(func(r *http.Request, ctx *types.Context) (ok bool) {
		for _, mm := range m {
			if !mm.Match(r, ctx) {
				return false
			}
		}
		return true
	})
}

// OrMatcher 仅需符合一个要求
func OrMatcher(m ...mux.Matcher) mux.Matcher {
	return mux.MatcherFunc(func(r *http.Request, ctx *types.Context) bool {
		for _, mm := range m {
			if ok := mm.Match(r, ctx); ok {
				return true
			}
		}
		return false
	})
}

// AndMatcherFunc 需同时符合每一个要求
func AndMatcherFunc(f ...func(*http.Request, *types.Context) bool) mux.Matcher {
	return AndMatcher(f2i(f...)...)
}

// OrMatcherFunc 仅需符合一个要求
func OrMatcherFunc(f ...func(*http.Request, *types.Context) bool) mux.Matcher {
	return OrMatcher(f2i(f...)...)
}

func f2i(f ...func(*http.Request, *types.Context) bool) []mux.Matcher {
	ms := make([]mux.Matcher, 0, len(f))
	for _, ff := range f {
		ms = append(ms, mux.MatcherFunc(ff))
	}
	return ms
}
