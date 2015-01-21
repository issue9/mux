// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
)

// Matches是多个Matcher组成的数组。
//  p1 := NewPath(h1, "/api/")
//  p2 := NewPath(h2, "/login")
//  actions := NewMatches(p1, p2)
//  http.ListenAndServe("8080", actions)
type Matches []Matcher

func NewMatches(matches ...Matcher) Matches {
	return matches
}

func (m Matches) Add(matches ...Matcher) Matches {
	return append(m, matches...)
}

// 遍历子元素，当其中的一个元素返回true时，立即中断后续的执行。
func (m Matches) ServeHTTP2(w http.ResponseWriter, r *http.Request) bool {
	for _, matcher := range m {
		if matcher.ServeHTTP2(w, r) {
			return true
		}
	}
	return false
}

func (m Matches) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.ServeHTTP2(w, r)
}
