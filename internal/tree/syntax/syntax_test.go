// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package syntax

import (
	"testing"

	"github.com/issue9/assert"
)

func TestParse(t *testing.T) {
	a := assert.New(t)
	test := func(str string, isError bool, ss ...*Segment) {
		s, err := Parse(str)
		if isError {
			a.Error(err)
			return
		}

		a.NotError(err).
			Equal(len(s), len(ss))
		for index, seg := range ss {
			a.Equal(seg, s[index])
		}
	}

	test("/", false, &Segment{Value: "/", Type: TypeString})

	test("/posts/1", false, &Segment{Value: "/posts/1", Type: TypeString})

	// 以命名参数开头的
	test("/{action}", false, &Segment{Value: "/", Type: TypeString},
		&Segment{Value: "{action}", Type: TypeNamed, Endpoint: true})

	// 以通配符结尾
	test("/posts/{id}", false, &Segment{Value: "/posts/", Type: TypeString},
		&Segment{Value: "{id}", Type: TypeNamed, Endpoint: true})

	test("/posts/{id}/author/profile", false, &Segment{Value: "/posts/", Type: TypeString},
		&Segment{Value: "{id}/author/profile", Type: TypeNamed})

	// 以命名参数结尾的
	test("/posts/{id}/author", false, &Segment{Value: "/posts/", Type: TypeString},
		&Segment{Value: "{id}/author", Type: TypeNamed})

	// 命名参数及通配符
	test("/posts/{id}/page/{page}", false, &Segment{Value: "/posts/", Type: TypeString},
		&Segment{Value: "{id}/page/", Type: TypeNamed},
		&Segment{Value: "{page}", Type: TypeNamed, Endpoint: true})

	// 正则
	test("/posts/{id:\\d+}", false, &Segment{Value: "/posts/", Type: TypeString},
		&Segment{Value: "{id:\\d+}", Type: TypeRegexp, Endpoint: true})

	// 正则，命名参数
	test("/posts/{id:\\d+}/page/{page}", false, &Segment{Value: "/posts/", Type: TypeString},
		&Segment{Value: "{id:\\d+}/page/", Type: TypeRegexp},
		&Segment{Value: "{page}", Type: TypeNamed, Endpoint: true})

	test("/posts/{id:}", true, nil)
	test("/posts/{{id:\\d+}/author", true, nil)
	test("/posts/{:\\d+}/author", true, nil)
	test("/posts/{}/author", true, nil)
	test("/posts/{id}{page}/", true, nil)
	test("/posts/:id/author", true, nil)
	test("/posts/{id}/{author", true, nil)
	test("/posts/}/author", true, nil)
}

func TestRegexp(t *testing.T) {
	a := assert.New(t)

	a.Equal(Regexp("{id:\\d+}"), "(?P<id>\\d+)")
	a.Equal(Regexp("{id:\\d+}/author"), "(?P<id>\\d+)/author")
}

func TestPrefixLen(t *testing.T) {
	a := assert.New(t)

	a.Equal(PrefixLen("", ""), 0)
	a.Equal(PrefixLen("/", ""), 0)
	a.Equal(PrefixLen("/test", "test"), 0)
	a.Equal(PrefixLen("/test", "/abc"), 1)
	a.Equal(PrefixLen("/test", "/test"), 5)
	a.Equal(PrefixLen("/te{st", "/test"), 3)
	a.Equal(PrefixLen("/test", "/tes{t"), 4)
	a.Equal(PrefixLen("/tes{t:\\d+}", "/tes{t:\\d+}/a"), 4) // 不应该包含正则部分
	a.Equal(PrefixLen("/tes{t:\\d+}/a", "/tes{t:\\d+}/"), 12)
	a.Equal(PrefixLen("{t}/a", "{t}/b"), 4)
	a.Equal(PrefixLen("{t}/abc", "{t}/bbc"), 4)
	a.Equal(PrefixLen("/tes{t:\\d+}", "/tes{t}"), 4)
}

func BenchmarkParse(b *testing.B) {
	patterns := []string{
		"/",
		"/posts/1",
		"/posts/{id}",
		"/posts/{id}/author/profile",
		"/posts/{id}/author",
	}

	for i := 0; i < b.N; i++ {
		for index, pattern := range patterns {
			v, _ := Parse(pattern)
			if v == nil {
				b.Errorf("BenchmarkParse: %d", index)
			}
		}
	}
}
