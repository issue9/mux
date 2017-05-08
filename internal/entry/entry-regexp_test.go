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

var _ Entry = &regexp{}

func TestNewRegexp(t *testing.T) {
	a := assert.New(t)

	pattern := "/posts/{id:\\d+}"
	r, err := newRegexp(pattern, &syntax{
		hasParams: true,
		nType:     typeRegexp,
		patterns:  []string{"/posts/", "(?P<id>\\d+)"},
	})
	a.NotError(err).NotNil(r)
	a.Equal(r.patternString, pattern)
	a.Equal(r.expr.String(), "/posts/(?P<id>\\d+)")

	pattern = "/posts/{id}/page/{page:\\d+}/size/{:\\d+}"
	r, err = newRegexp(pattern, &syntax{
		hasParams: true,
		nType:     typeRegexp,
		patterns:  []string{"/posts/", "(?P<id>[^/]+)", "/page/", "(?P<page>\\d+)", "/size/", "(\\d+)"},
	})
	a.NotError(err).NotNil(r)
	a.Equal(r.patternString, pattern)
	a.Equal(r.expr.String(), "/posts/(?P<id>[^/]+)/page/(?P<page>\\d+)/size/(\\d+)")
}

func TestRegexp_match(t *testing.T) {
	a := assert.New(t)

	n, err := New("/posts/{id:\\d+}", nil)
	a.NotError(err).NotNil(n)

	a.True(n.match("/posts/1"))
	a.True(n.match("/posts/2"))
	a.False(n.match("/posts/id"))
	a.False(n.match("/posts/id.html"))
	a.False(n.match("/posts/id.html/"))
	a.False(n.match("/posts/id.html/page"))
	a.False(n.match("/post/id"))

	n, err = New("/posts/{id}/page/{page:\\d+}", nil)
	a.True(n.match("/posts/1/page/1"))
	a.True(n.match("/posts/1.html/page/1"))

	a.False(n.match("/posts/1.html/page/x"))
	a.False(n.match("/posts/id-1/page/1/"))
	a.False(n.match("/posts/id-1/page/1/size/1"))

	n, err = New("/posts/{id}/page/{page:\\d+}/size/{:\\d+}", nil)
	a.True(n.match("/posts/1.html/page/1/size/1"))
	a.True(n.match("/posts/1.html/page/1/size/11"))
	a.False(n.match("/posts/1.html/page/x/size/11"))
}

func TestRegexp_match_wildcard(t *testing.T) {
	a := assert.New(t)

	n, err := New("/posts/{id:\\d+}/*", nil)
	a.NotError(err).NotNil(n)

	a.False(n.match("/posts/1"))
	a.True(n.match("/posts/1/"))
	a.True(n.match("/posts/1/index.html"))
	a.False(n.match("/posts/id.html/page"))

	n, err = New("/posts/{id}/page/{page:\\d+}/*", nil)
	a.False(n.match("/posts/1/page/1"))
	a.True(n.match("/posts/1.html/page/1/"))
	a.True(n.match("/posts/1.html/page/1/index.html"))
	a.False(n.match("/posts/1.html/page/x/index.html"))

	n, err = New("/posts/{id}/page/{page:\\d+}/size/{:\\d+}/*", nil)
	a.False(n.match("/posts/1.html/page/1/size/1"))
	a.True(n.match("/posts/1.html/page/1/size/1/index.html"))
}

func TestRegexp_Params(t *testing.T) {
	a := assert.New(t)
	n, err := New("/posts/{id:[^/]+}", nil)
	a.NotError(err).NotNil(n)
	a.Equal(n.Params("/posts/1"), map[string]string{"id": "1"})
	a.Equal(n.Params("/posts/1.html"), map[string]string{"id": "1.html"})
	a.Equal(n.Params("/posts/1.html/"), map[string]string{"id": "1.html"})

	n, err = New("/posts/{id}/page/{page:\\d+}", nil)
	a.Equal(n.Params("/posts/1/page/1"), map[string]string{"id": "1", "page": "1"})
	a.Equal(n.Params("/posts/1.html/page/1"), map[string]string{"id": "1.html", "page": "1"})

	// 带有未命名参数
	n, err = New("/posts/{id}/page/{page:\\d+}/size/{:\\d+}", nil)
	a.Equal(n.Params("/posts/1.html/page/1/size/1"), map[string]string{"id": "1.html", "page": "1"})
}

func TestRegexp_URL(t *testing.T) {
	a := assert.New(t)
	n, err := New("/posts/{id:[^/]+}", nil)
	a.NotError(err).NotNil(n)
	url, err := n.URL(map[string]string{"id": "5.html"}, "path")
	a.NotError(err).Equal(url, "/posts/5.html")
	url, err = n.URL(map[string]string{"id": "5.html/"}, "path")
	a.NotError(err).Equal(url, "/posts/5.html/")

	n, err = New("/posts/{id:[^/]+}/page/{page}", nil)
	url, err = n.URL(map[string]string{"id": "5.html", "page": "1"}, "path")
	a.NotError(err).Equal(url, "/posts/5.html/page/1")

	// 少参数
	url, err = n.URL(map[string]string{"id": "5.html"}, "path")
	a.Error(err).Equal(url, "")

	// 带有未命名参数
	n, err = New("/posts/{id}/page/{page:\\d+}/size/{:\\d+}", nil)
	url, err = n.URL(map[string]string{"id": "5.html", "page": "1"}, "path")
	a.NotError(err).Equal(url, "/posts/5.html/page/1/size/[0-9]+")

	// 带通配符
	n, err = New("/posts/{id:[^/]+}/page/{page}/*", nil)
	url, err = n.URL(map[string]string{"id": "5.html", "page": "1"}, "path")
	a.NotError(err).Equal(url, "/posts/5.html/page/1/path")

	// 指定了空的 path
	url, err = n.URL(map[string]string{"id": "5.html", "page": "1"}, "")
	a.NotError(err).Equal(url, "/posts/5.html/page/1/")
}

///////////////////////////////////////////////////////////////
// 以下为一个性能测试用，用于验证将一个正则表达式折分成多个
// 和不折分，哪个性能下高一点

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
