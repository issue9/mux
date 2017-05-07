// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"fmt"
	stdregexp "regexp"
	stdsyntax "regexp/syntax"
	"strings"
)

type regexp struct {
	*base
	expr       *stdregexp.Regexp
	hasParams  bool
	syntaxExpr *stdsyntax.Regexp
}

func newRegexp(pattern string, s *syntax) (*regexp, error) {
	str := strings.Join(s.patterns, "")
	expr, err := stdregexp.Compile(str)
	if err != nil {
		return nil, err
	}

	syntaxExpr, err := stdsyntax.Parse(str, stdsyntax.Perl)
	if err != nil {
		return nil, err
	}

	return &regexp{
		base:       newItems(pattern),
		hasParams:  s.hasParams,
		expr:       expr,
		syntaxExpr: syntaxExpr,
	}, nil

}

// Entry.Type
func (r *regexp) Type() int {
	return TypeRegexp
}

// Entry.Match
func (r *regexp) Match(url string) int {
	loc := r.expr.FindStringIndex(url)

	if loc != nil &&
		loc[0] == 0 &&
		loc[1] == len(url) {
		return 0
	}
	return -1
}

// Entry.Params
func (r *regexp) Params(url string) map[string]string {
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

// fun
func (r *regexp) URL(params map[string]string) (string, error) {
	if r.syntaxExpr == nil {
		return r.pattern, nil
	}

	url := r.syntaxExpr.String()
	for _, sub := range r.syntaxExpr.Sub {
		if len(sub.Name) == 0 {
			continue
		}

		param, exists := params[sub.Name]
		if !exists {
			return "", fmt.Errorf("未找到参数 %v 的值", sub.Name)
		}
		url = strings.Replace(url, sub.String(), param, -1)
	}

	return url, nil
}
