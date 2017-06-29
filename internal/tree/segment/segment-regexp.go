// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package segment

import (
	"bytes"
	"fmt"
	"regexp"
	"regexp/syntax"
	"strings"
)

// 将路由语法转换成正则表达式语法，比如：
//  {id:\\d+}/author => (?P<id>\\d+)
var repl = strings.NewReplacer(string(NameStart), "(?P<",
	string(RegexpSeparator), ">",
	string(NameEnd), ")")

type reg struct {
	pattern    string
	endpoint   bool
	expr       *regexp.Regexp
	syntaxExpr *syntax.Regexp
}

func newReg(str string) (Segment, error) {
	r := repl.Replace(str)
	expr, err := regexp.Compile(r)
	if err != nil {
		return nil, err
	}

	syntaxExpr, err := syntax.Parse(r, syntax.Perl)
	if err != nil {
		return nil, err
	}

	return &reg{
		pattern:    str,
		expr:       expr,
		syntaxExpr: syntaxExpr,
		endpoint:   str[len(str)-1] == NameEnd,
	}, nil
}

func (r *reg) Type() Type {
	return TypeRegexp
}

func (r *reg) Pattern() string {
	return r.pattern
}

func (r *reg) Endpoint() bool {
	return r.endpoint
}

func (r *reg) Match(path string) (bool, string) {
	loc := r.expr.FindStringIndex(path)
	if loc == nil || loc[0] != 0 { // 不匹配
		return false, path
	}

	if loc[1] == len(path) {
		return true, path[:0]
	}
	return true, path[loc[1]+1:]
}

func (r *reg) Params(path string, params map[string]string) string {
	subexps := r.expr.SubexpNames()
	args := r.expr.FindStringSubmatch(path)
	for index, name := range subexps {
		if len(name) > 0 && index < len(args) {
			params[name] = args[index]
		}
	}

	return path[len(args[0]):]
}

func (r *reg) URL(buf *bytes.Buffer, params map[string]string) error {
	url := r.syntaxExpr.String()
	subs := append(r.syntaxExpr.Sub, r.syntaxExpr)
	for _, sub := range subs {
		if len(sub.Name) == 0 {
			continue
		}

		param, exists := params[sub.Name]
		if !exists {
			return fmt.Errorf("未找到参数 %v 的值", sub.Name)
		}
		url = strings.Replace(url, sub.String(), param, -1)
	}

	_, err := buf.WriteString(url)
	return err
}
