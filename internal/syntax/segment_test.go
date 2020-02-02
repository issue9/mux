// SPDX-License-Identifier: MIT

package syntax

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v2/params"
)

func BenchmarkSegment_Match_Named(b *testing.B) {
	a := assert.New(b)

	seg := NewSegment("/posts/{id}/author")
	a.NotNil(seg)

	ps := params.Params{}
	path := "/posts/1/author"
	for i := 0; i < b.N; i++ {
		index := seg.Match(path, ps)
		a.Equal(index, len(path))
	}
}

func BenchmarkSegment_Match_String(b *testing.B) {
	a := assert.New(b)

	seg := NewSegment("/posts/author")
	a.NotNil(seg)

	path := "/posts/author"
	for i := 0; i < b.N; i++ {
		index := seg.Match(path, nil)
		a.Equal(index, len(path))
	}
}

func BenchmarkSegment_Match_Regexp(b *testing.B) {
	a := assert.New(b)

	seg := NewSegment("/posts/{id:\\d+}/author")
	a.NotNil(seg)

	path := "/posts/1/author"
	ps := params.Params{}
	for i := 0; i < b.N; i++ {
		index := seg.Match(path, ps)
		a.Equal(index, len(path))
	}
}

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
	path := "/posts/1/author"
	index := seg.Match(path, ps)
	a.Empty(path[index:])

	// Named 部分匹配
	path = "/posts/1/author/email"
	index = seg.Match(path, ps)
	a.Equal(path[index:], "/email")

	// Named 不匹配
	path = "/posts/1/aut"
	index = seg.Match(path, ps)
	a.Equal(index, -1)

	// Named 1/2 匹配 {id}
	ps = params.Params{}
	path = "/posts/1/2/author"
	index = seg.Match(path, ps)
	a.Equal(path[index:], "")

	// Named Endpoint 匹配
	seg = NewSegment("{path}")
	path = "/posts/author"
	index = seg.Match(path, ps)
	a.Equal(path[index:], "")

	// String
	seg = NewSegment("/posts/author")
	a.NotNil(seg)

	// String 匹配
	path = "/posts/author"
	index = seg.Match(path, nil)
	a.Equal(path[index:], "")

	// String 不匹配
	path = "/posts/author/email"
	index = seg.Match(path, nil)
	a.Equal(path[index:], "/email")

	// Regexp
	seg = NewSegment("/posts/{id:\\d+}/author")
	a.NotNil(seg)

	// Regexp 完全匹配
	ps = params.Params{}
	path = "/posts/1/author"
	index = seg.Match(path, ps)
	a.Equal(path[index:], "")

	// Regexp 不匹配
	ps = params.Params{}
	path = "/posts/xxx/author"
	index = seg.Match(path, ps)
	a.Equal(index, -1)

	// Regexp 部分匹配
	path = "/posts/1/author/email"
	index = seg.Match(path, ps)
	a.Equal(path[index:], "/email")
}
