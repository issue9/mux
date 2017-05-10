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
	b := newBase(pattern)

	// 合并正则表达式
	str := strings.Join(s.patterns, "")
	if b.wildcard {
		str = str[:len(str)-1] // 去掉最后的星号
	}

	expr, err := stdregexp.Compile(str)
	if err != nil {
		return nil, err
	}

	syntaxExpr, err := stdsyntax.Parse(str, stdsyntax.Perl)
	if err != nil {
		return nil, err
	}

	return &regexp{
		base:       b,
		hasParams:  s.hasParams,
		expr:       expr,
		syntaxExpr: syntaxExpr,
	}, nil
}

func (r *regexp) priority() int {
	if r.wildcard {
		return typeRegexp + 100
	}

	return typeRegexp
}

// Entry.Match
func (r *regexp) match(url string) (bool, map[string]string) {
	loc := r.expr.FindStringIndex(url)

	if loc == nil || loc[0] != 0 {
		return false, nil
	}

	if loc[1] == len(url) {
		return true, r.params(url)
	}

	// 通配符的应该比较少，放最后比较
	if r.wildcard {
		if loc[1] < len(url) {
			return true, r.params(url)
		}
	}

	return false, nil
}

// Entry.Params
func (r *regexp) params(url string) map[string]string {
	if !r.hasParams {
		return nil
	}

	// 正确匹配正则表达式，则获相关的正则表达式命名变量。
	subexps := r.expr.SubexpNames()
	mapped := make(map[string]string, len(subexps))
	args := r.expr.FindStringSubmatch(url)
	for index, name := range subexps {
		if len(name) > 0 && index < len(args) {
			mapped[name] = args[index]
		}
	}
	return mapped
}

// fun
func (r *regexp) URL(params map[string]string, path string) (string, error) {
	if r.syntaxExpr == nil {
		return r.patternString, nil
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

	if r.wildcard {
		url += path
	}

	return url, nil
}
