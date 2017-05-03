// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import "regexp"

type regexpr struct {
	*items
	expr      *regexp.Regexp
	hasParams bool
}

// Entry.Type
func (r *regexpr) Type() int {
	return TypeRegexp
}

// Entry.Match
func (r *regexpr) Match(url string) int {
	loc := r.expr.FindStringIndex(url)

	if loc != nil &&
		loc[0] == 0 &&
		loc[1] == len(url) {
		return 0
	}
	return -1
}

// Entry.Params
func (r *regexpr) Params(url string) map[string]string {
	if !r.hasParams {
		return nil
	}

	// 正确匹配正则表达式，则获相关的正则表达式命名变量。
	mapped := make(map[string]string, 3)
	subexps := r.expr.SubexpNames()
	args := r.expr.FindStringSubmatch(url)
	for index, name := range subexps {
		if len(name) > 0 && index < len(args) {
			mapped[name] = args[index]
		}
	}
	return mapped
}
