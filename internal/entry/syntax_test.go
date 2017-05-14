// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"testing"

	"github.com/issue9/assert"
)

func TestIsSyntax(t *testing.T) {
	a := assert.New(t)
	a.True(isSyntax("{abc}"))
	a.True(isSyntax("{abc:\\w+}"))

	a.False(isSyntax("w{abc}"))
	a.False(isSyntax("{abc}w"))
	a.False(isSyntax("w{abc}w"))
	a.False(isSyntax("{abc"))
	a.False(isSyntax("abc}"))
}

func TestParse(t *testing.T) {
	a := assert.New(t)

	fn := func(pattern string, isErr bool, s *syntax) {
		ret, err := parse(pattern)
		if isErr {
			a.Error(err)
			return
		}

		a.Equal(ret.nType, s.nType).
			Equal(ret.hasParams, s.hasParams).
			Equal(ret.patterns, s.patterns)
	}

	fn("", true, &syntax{})
	fn(" ", true, &syntax{})
	fn("/", false, &syntax{
		hasParams: false,
		nType:     TypeBasic,
		patterns:  []string{"/"},
	})
	fn("/posts/1", false, &syntax{
		hasParams: false,
		nType:     TypeBasic,
		patterns:  []string{"/posts/1"},
	})
	fn("/posts/{id", false, &syntax{
		hasParams: false,
		nType:     TypeBasic,
		patterns:  []string{"/posts/{id"},
	})

	// Named
	fn("/posts/{id}", false, &syntax{
		hasParams: true,
		nType:     TypeNamed,
		patterns:  []string{"/posts/", "{id}"},
	})
	fn("/posts/{id}/page/{page}", false, &syntax{
		hasParams: true,
		nType:     TypeNamed,
		patterns:  []string{"/posts/", "{id}", "/page/", "{page}"},
	})

	fn("/posts/{id}-{id}", true, nil) // 相同参数名

	// regexp
	fn("/posts/{id:\\d+}", false, &syntax{
		hasParams: true,
		nType:     TypeRegexp,
		patterns:  []string{"/posts/", "(?P<id>\\d+)"},
	})

	fn("/posts/{id:\\d+}-{id}", true, nil) // 相同参数名

	fn("/posts/{id:\\d+}/page/{page:\\d+}", false, &syntax{
		hasParams: true,
		nType:     TypeRegexp,
		patterns:  []string{"/posts/", "(?P<id>\\d+)", "/page/", "(?P<page>\\d+)"},
	})
	// 未命名参数
	fn("/posts/{:\\d+}", false, &syntax{
		hasParams: false,
		nType:     TypeRegexp,
		patterns:  []string{"/posts/", "\\d+"},
	})
	// 有一个未命名参数
	fn("/posts/{:\\d+}/page/{page:\\d+}", false, &syntax{
		hasParams: true,
		nType:     TypeRegexp,
		patterns:  []string{"/posts/", "\\d+", "/page/", "(?P<page>\\d+)"},
	})
	// 多个未命名参数
	fn("/posts/{:\\d+}/page/{:\\d+}", false, &syntax{
		hasParams: false,
		nType:     TypeRegexp,
		patterns:  []string{"/posts/", "\\d+", "/page/", "\\d+"},
	})

	// 命名与未命名混合
	fn("/posts/{id}/page/{:\\d+}", false, &syntax{
		hasParams: true,
		nType:     TypeRegexp,
		patterns:  []string{"/posts/", "(?P<id>[^/]+)", "/page/", "\\d+"},
	})

	// 命名与正则名混合
	fn("/posts/{id}/page/{page:\\d+}", false, &syntax{
		hasParams: true,
		nType:     TypeRegexp,
		patterns:  []string{"/posts/", "(?P<id>[^/]+)", "/page/", "(?P<page>\\d+)"},
	})

	// 命名与正则、未命名名混合
	fn("/posts/{id}/page/{page:\\d+}/size/{:\\d+}", false, &syntax{
		hasParams: true,
		nType:     TypeRegexp,
		patterns:  []string{"/posts/", "(?P<id>[^/]+)", "/page/", "(?P<page>\\d+)", "/size/", "\\d+"},
	})
}

func TestDuplicateName(t *testing.T) {
	a := assert.New(t)

	names := []string{
		"name1", "name2", "name3", "name1",
	}

	a.True(duplicateName(names) > -1)
}

func TestSplit(t *testing.T) {
	a := assert.New(t)

	// 为空
	a.Equal(split(""), []string{})

	// 不存在 {}
	a.Equal(split("/blog/post/1"), []string{"/blog/post/1"})

	// 开头包含 {}
	a.Equal(split("{action}/post/1"), []string{"{action}", "/post/1"})

	// 结尾包含 {}
	a.Equal(split("/blog/post/{id}"), []string{"/blog/post/", "{id}"})
	a.Equal(split("/blog/post/{id:\\d}"), []string{"/blog/post/", "{id:\\d}"})

	// 中间包含
	a.Equal(split("/blog/post/{id}/author"), []string{"/blog/post/", "{id}", "/author"})
	a.Equal(split("/blog/{post}-{id}/author"), []string{"/blog/", "{post}", "-", "{id}", "/author"})

	// 首中尾都包含
	a.Equal(split("{action}/{post}/{id}"), []string{"{action}", "/", "{post}", "/", "{id}"})
	a.Equal(split("{action}/{post}-{id}"), []string{"{action}", "/", "{post}", "-", "{id}"})

	// 无法解析的内容
	a.Equal(split("{/blog/post/{id}"), []string{"{/blog/post/{id}"})
	a.Equal(split("}/blog/post/{id}"), []string{"}/blog/post/", "{id}"})
}
