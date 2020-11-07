// SPDX-License-Identifier: MIT

package syntax

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v3/params"
)

func TestNewSegment(t *testing.T) {
	a := assert.New(t)

	seg := NewSegment("/post/1")
	a.Equal(seg.Type, String).Equal(seg.Value, "/post/1")

	seg = NewSegment("{id}/1")
	a.Equal(seg.Type, Named).
		Equal(seg.Value, "{id}/1").
		False(seg.Endpoint).
		Equal(seg.Suffix, "/1")

	seg = NewSegment("{id}")
	a.Equal(seg.Type, Named).
		Equal(seg.Value, "{id}").
		True(seg.Endpoint).
		Empty(seg.Suffix)

	seg = NewSegment("{id:digit}/1")
	a.Equal(seg.Type, Named).
		Equal(seg.Value, "{id:digit}/1").
		False(seg.Endpoint).
		Equal(seg.Suffix, "/1")

	seg = NewSegment("{id:\\d+}/1")
	a.Equal(seg.Type, Regexp).
		Equal(seg.Value, "{id:\\d+}/1").
		False(seg.Endpoint).
		Equal(seg.Suffix, "/1")
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

func TestSegment_Split(t *testing.T) {
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
	seg := NewSegment("{id}/author")
	a.NotNil(seg)

	// Named 完全匹配
	ps := params.Params{}
	path := "1/author"
	index := seg.Match(path, ps)
	a.Empty(path[index:]).
		Equal(ps, params.Params{"id": "1"})

	// Named 部分匹配
	ps = params.Params{}
	path = "1/author/email"
	index = seg.Match(path, ps)
	a.Equal(path[index:], "/email")

	// Named 不匹配
	ps = params.Params{}
	path = "1/aut"
	index = seg.Match(path, ps)
	a.Equal(index, -1)

	// Named 1/2 匹配 {id}
	ps = params.Params{}
	path = "1/2/author"
	index = seg.Match(path, ps)
	a.Equal(path[index:], "")

	// Named Endpoint 匹配
	ps = params.Params{}
	seg = NewSegment("{path}")
	path = "/posts/author"
	index = seg.Match(path, ps)
	a.Equal(path[index:], "")

	// Named:digit
	seg = NewSegment("{id:digit}/author")
	a.NotNil(seg)

	// Named:digit 不匹配
	ps = params.Params{}
	path = "1/aut"
	index = seg.Match(path, ps)
	a.Equal(index, -1)

	// Named:digit 类型不匹配
	ps = params.Params{}
	path = "xx/author"
	index = seg.Match(path, ps)
	a.Equal(index, -1)

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
	seg = NewSegment("{id:\\d+}/author")
	a.NotNil(seg)

	// Regexp 完全匹配
	ps = params.Params{}
	path = "1/author"
	index = seg.Match(path, ps)
	a.Equal(path[index:], "")

	// Regexp 不匹配
	ps = params.Params{}
	path = "xxx/author"
	index = seg.Match(path, ps)
	a.Equal(index, -1)

	// Regexp 部分匹配
	path = "1/author/email"
	index = seg.Match(path, ps)
	a.Equal(path[index:], "/email")
}
