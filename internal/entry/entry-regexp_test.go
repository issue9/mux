// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	stdregexp "regexp"
	"strings"
	"testing"

	"github.com/issue9/assert"
)

// 测试用内容，键名为正则，键值为或匹配的值
var regexpStrs = map[string]string{
	"/blog/posts/":   "/blog/posts/",
	"(?P<id>\\d+)":   "100",
	"/page/":         "/page/",
	"(?P<page>\\d+)": "100",
	"/size/":         "/size/",
	"(?P<size>\\d+)": "100",
}

// 将所有的内容当作一条正则进行处理
func BenchmarkRegexp_One(b *testing.B) {
	a := assert.New(b)

	regstr := ""
	match := ""
	for k, v := range regexpStrs {
		regstr += k
		match += v
	}

	expr, err := stdregexp.Compile(regstr)
	a.NotError(err).NotNil(expr)

	for i := 0; i < b.N; i++ {
		loc := expr.FindStringIndex(match)
		if loc == nil || loc[0] != 0 {
			b.Error("BenchmarkBasic_Match:error")
		}
	}
}

// 将内容细分，仅将其中的正则部分处理成正则表达式，其它的仍然以字符串作比较
//
// 目前看来，仅在只有一条正则夹在其中的时候，才有一占点优势，否则可能更慢。
func BenchmarkRegexp_Mult(b *testing.B) {
	type item struct {
		pattern string
		expr    *stdregexp.Regexp
	}

	items := make([]*item, 0, len(regexpStrs))

	match := ""
	for k, v := range regexpStrs {
		if strings.IndexByte(k, '?') >= 0 {
			items = append(items, &item{expr: stdregexp.MustCompile(k)})
		} else {
			items = append(items, &item{pattern: k})
		}
		match += v
	}

	test := func(path string) bool {
		for _, i := range items {
			if i.expr == nil {
				if !strings.HasPrefix(path, i.pattern) {
					return false
				}
				path = path[len(i.pattern):]
			} else {
				loc := i.expr.FindStringIndex(path)
				if loc == nil || loc[0] != 0 {
					return false
				}
				path = path[loc[1]:]
			}
		}

		return true
	}

	for i := 0; i < b.N; i++ {
		if !test(match) {
			b.Error("er")
		}
	}
}
