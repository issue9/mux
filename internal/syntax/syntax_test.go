// SPDX-License-Identifier: MIT

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
			item := s[index]
			a.Equal(seg.Value, item.Value).
				Equal(seg.Name, item.Name).
				Equal(seg.Endpoint, item.Endpoint).
				Equal(seg.Suffix, item.Suffix)
		}
	}

	test("/", false, NewSegment("/"))

	test("/posts/1", false, NewSegment("/posts/1"))

	test("{action}/1", false, NewSegment("{action}/1"))
	test("{act/ion}/1", false, NewSegment("{act/ion}/1")) // 名称中包含非常规则字符
	test("{中文}/1", false, NewSegment("{中文}/1"))           // 名称中包含中文

	// 以命名参数开头的
	test("/{action}", false, NewSegment("/"), NewSegment("{action}"))

	// 以通配符结尾
	test("/posts/{id}", false, NewSegment("/posts/"), NewSegment("{id}"))

	test("/posts/{id}/author/profile", false, NewSegment("/posts/"), NewSegment("{id}/author/profile"))

	// 以命名参数结尾的
	test("/posts/{id}/author", false, NewSegment("/posts/"), NewSegment("{id}/author"))

	test("/posts/{id:digit}/author", false, NewSegment("/posts/"), NewSegment("{id:digit}/author"))

	// 命名参数及通配符
	test("/posts/{id}/page/{page}", false, NewSegment("/posts/"), NewSegment("{id}/page/"), NewSegment("{page}"))
	test("/posts/{id}/page/{page:digit}", false, NewSegment("/posts/"), NewSegment("{id}/page/"), NewSegment("{page:digit}"))

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

func TestIsWell(t *testing.T) {
	a := assert.New(t)

	a.NotError(IsWell("/posts/"))
	a.NotError(IsWell("/posts/{id}"))
	a.NotError(IsWell("/posts/{id:\\d+}"))

	a.Error(IsWell("/posts/{id:\\d+"))
	a.Error(IsWell("/posts/id:\\d+}"))
}
