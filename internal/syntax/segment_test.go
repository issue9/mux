// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package syntax

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v2/params"
)

func TestLongestPrefix(t *testing.T) {
	a := assert.New(t)

	test := func(s1, s2 string, len int) {
		a.Equal(longestPrefix(s1, s2), len)
	}

	test("", "", 0)
	test("/", "", 0)
	test("/test", "test", 0)
	test("/test", "/abc", 1)
	test("/test", "/test", 5)
	test("/te{st}", "/test", 3)
	test("/test", "/tes{t}", 4)
	test("/tes{t:\\d+}", "/tes{t:\\d+}/a", 4) // 不应该包含正则部分
	test("/tes{t:\\d+}/a", "/tes{t:\\d+}/", 12)
	test("{t}/a", "{t}/b", 4)
	test("{t}/abc", "{t}/bbc", 4)
	test("/tes{t:\\d+}", "/tes{t}", 4)
}

func TestSegment_Similarity(t *testing.T) {
	a := assert.New(t)

	seg := NewSegment("{id}/author")
	a.NotNil(seg)

	s1 := NewSegment("{id}/author")
	a.Equal(-1, seg.Similarity(s1))

	s1 = NewSegment("{id}/author/profile")
	a.Equal(11, seg.Similarity(s1))
}

func TestSegemnt_Split(t *testing.T) {
	a := assert.New(t)

	seg := NewSegment("{id}/author")
	a.NotNil(seg)

	segs := seg.Split(4)
	a.Equal(segs[0].Value, "{id}").
		Equal(segs[1].Value, "/author")
}

func TestSegment_Match(t *testing.T) {
	a := assert.New(t)

	// Named
	seg := NewSegment("/posts/{id}/author")
	a.NotNil(seg)

	// Named 完全匹配
	ps := params.Params{}
	ok, path := seg.Match("/posts/1/author", ps)
	a.True(ok).Equal(path, "")

	// Named 部分匹配
	ok, path = seg.Match("/posts/1/author/email", ps)
	a.True(ok).Equal(path, "/email")

	// Named 不匹配
	ok, path = seg.Match("/posts/1/aut", ps)
	a.False(ok).Equal(path, "/posts/1/aut")

	// Named 1/2 匹配 {id}
	ps = params.Params{}
	ok, path = seg.Match("/posts/1/2/author", ps)
	a.True(ok).Equal(path, "")

	// Named Endpoint 匹配
	seg = NewSegment("{path}")
	ok, path = seg.Match("/posts/author", ps)
	a.True(ok).Equal(path, "")

	// String
	seg = NewSegment("/posts/author")
	a.NotNil(seg)

	// String 匹配
	ok, path = seg.Match("/posts/author", nil)
	a.True(ok).Equal(path, "")

	// String 不匹配
	ok, path = seg.Match("/posts/author/email", nil)
	a.True(ok).Equal(path, "/email")

	// Regexp
	seg = NewSegment("/posts/{id:\\d+}/author")
	a.NotNil(seg)

	// Regexp 完全匹配
	ps = params.Params{}
	ok, path = seg.Match("/posts/1/author", ps)
	a.True(ok).Equal(path, "")

	// Regexp 不匹配
	ps = params.Params{}
	ok, path = seg.Match("/posts/xxx/author", ps)
	a.False(ok).Equal(path, "/posts/xxx/author")

	// Regexp 部分匹配
	ok, path = seg.Match("/posts/1/author/email", ps)
	a.True(ok).Equal(path, "/email")
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
