// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package syntax

import (
	"testing"

	"github.com/issue9/assert"
)

func TestRegexp(t *testing.T) {
	a := assert.New(t)

	a.Equal(repl.Replace("{id:\\d+}"), "(?P<id>\\d+)")
	a.Equal(repl.Replace("{id:\\d+}/author"), "(?P<id>\\d+)/author")
}

func TestGetType(t *testing.T) {
	a := assert.New(t)

	a.Equal(getType(""), String)
	a.Equal(getType("/posts"), String)
	a.Equal(getType("/posts/{id}"), Named)
	a.Equal(getType("/posts/{id}/author"), Named)
	a.Equal(getType("/posts/{id:\\d+}/author"), Regexp)
}

func TestType_String(t *testing.T) {
	a := assert.New(t)

	a.Equal(Named.String(), "named")
	a.Equal(Regexp.String(), "regexp")
	a.Equal(String.String(), "string")
	a.Panic(func() {
		_ = (Type(5)).String()
	})
}

func TestSplit(t *testing.T) {
	a := assert.New(t)
	test := func(str string, isError bool, ss ...*Segment) {
		if isError {
			a.Panic(func() {
				Split(str)
			})
			return
		}

		s := Split(str)
		a.Equal(len(s), len(ss))
		for index, seg := range ss {
			a.Equal(seg, s[index])
		}
	}

	test("/", false, NewSegment("/"))

	test("/posts/1", false, NewSegment("/posts/1"))

	test("{action}/1", false, NewSegment("{action}/1"))

	// 以命名参数开头的
	test("/{action}", false, NewSegment("/"), NewSegment("{action}"))

	// 以通配符结尾
	test("/posts/{id}", false, NewSegment("/posts/"), NewSegment("{id}"))

	test("/posts/{id}/author/profile", false, NewSegment("/posts/"), NewSegment("{id}/author/profile"))

	// 以命名参数结尾的
	test("/posts/{id}/author", false, NewSegment("/posts/"), NewSegment("{id}/author"))

	// 命名参数及通配符
	test("/posts/{id}/page/{page}", false, NewSegment("/posts/"), NewSegment("{id}/page/"), NewSegment("{page}"))

	// 正则
	test("/posts/{id:\\d+}", false, NewSegment("/posts/"), NewSegment("{id:\\d+}"))

	// 正则，命名参数
	test("/posts/{id:\\d+}/page/{page}", false, NewSegment("/posts/"), NewSegment("{id:\\d+}/page/"), NewSegment("{page}"))

	test("/posts/{id:}", true)
	test("/posts/{{id:\\d+}/author", true)
	test("/posts/{:\\d+}/author", true)
	test("/posts/{}/author", true)
	test("/posts/{id}{page}/", true)
	test("/posts/:id/author", true)
	test("/posts/{id}/{author", true)
	test("/posts/}/author", true)

	a.Panic(func() {
		Split("")
	})
}
