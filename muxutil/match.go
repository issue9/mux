// SPDX-License-Identifier: MIT

package muxutil

import (
	"net/http"

	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/internal/syntax"
	"github.com/issue9/mux/v6/params"
)

// AndMatcher 按顺序符合每一个要求
//
// 前一个对象返回的实例将作为下一个对象的输入参数。
func AndMatcher(m ...mux.Matcher) mux.Matcher {
	return mux.MatcherFunc(func(r *http.Request) (ps params.Params, ok bool) {
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

// OrMatcher 仅需符合一个要求
func OrMatcher(m ...mux.Matcher) mux.Matcher {
	return mux.MatcherFunc(func(r *http.Request) (params.Params, bool) {
		for _, mm := range m {
			if rr, ok := mm.Match(r); ok {
				return rr, true
			}
		}
		return nil, false
	})
}

// AndMatcherFunc 需同时符合每一个要求
func AndMatcherFunc(f ...func(*http.Request) (params.Params, bool)) mux.Matcher {
	return AndMatcher(f2i(f...)...)
}

// OrMatcherFunc 仅需符合一个要求
func OrMatcherFunc(f ...func(*http.Request) (params.Params, bool)) mux.Matcher {
	return OrMatcher(f2i(f...)...)
}

func f2i(f ...func(*http.Request) (params.Params, bool)) []mux.Matcher {
	ms := make([]mux.Matcher, 0, len(f))
	for _, ff := range f {
		ms = append(ms, mux.MatcherFunc(ff))
	}
	return ms
}
