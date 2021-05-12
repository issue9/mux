// SPDX-License-Identifier: MIT

package syntax

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v4/params"
)

func TestNewSegment(t *testing.T) {
	a := assert.New(t)

	seg, err := NewSegment("/post/1")
	a.NotError(err).Equal(seg.Type, String).Equal(seg.Value, "/post/1")

	seg, err = NewSegment("{id}/1")
	a.NotError(err).Equal(seg.Type, Named).
		Equal(seg.Value, "{id}/1").
		False(seg.Endpoint).
		Equal(seg.Suffix, "/1")

	seg, err = NewSegment("{id}")
	a.NotError(err).Equal(seg.Type, Named).
		Equal(seg.Value, "{id}").
		True(seg.Endpoint).
		Empty(seg.Suffix)

	seg, err = NewSegment("{id:}")
	a.NotError(err).Equal(seg.Type, Named).
		Equal(seg.Value, "{id:}").
		True(seg.Endpoint).
		Empty(seg.Suffix).
		Equal(seg.Name, "id")

	seg, err = NewSegment("{id}:")
	a.NotError(err).Equal(seg.Type, Named).
		Equal(seg.Value, "{id}:").
		False(seg.Endpoint).
		Equal(seg.Suffix, ":").
		Equal(seg.Name, "id")

	seg, err = NewSegment("{id:any}")
	a.NotError(err).Equal(seg.Type, Interceptor).
		Equal(seg.Value, "{id:any}").
		True(seg.Endpoint).
		Empty(seg.Suffix)

	seg, err = NewSegment("{id:digit}/1")
	a.NotError(err).Equal(seg.Type, Interceptor).
		Equal(seg.Value, "{id:digit}/1").
		False(seg.Endpoint).
		Equal(seg.Suffix, "/1")

	seg, err = NewSegment("{id:\\d+}/1")
	a.NotError(err).Equal(seg.Type, Regexp).
		Equal(seg.Value, "{id:\\d+}/1").
		False(seg.Endpoint).
		Equal(seg.Suffix, "/1")

	seg, err = NewSegment("id:}{")
	a.Error(err).Nil(seg)

	seg, err = NewSegment("id:{}")
	a.Error(err).Nil(seg)

	seg, err = NewSegment("{path")
	a.NotError(err).NotNil(seg).
		Equal(seg.Type, String).
		Equal(seg.Value, "{path")

	seg, err = NewSegment("{:path")
	a.NotError(err).NotNil(seg).
		Equal(seg.Type, String).
		Equal(seg.Value, "{:path")
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

	seg, err := NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)

	s1, err := NewSegment("{id}/author")
	a.NotError(err).Equal(-1, seg.Similarity(s1))

	s1, err = NewSegment("{id}/author/profile")
	a.NotError(err).Equal(11, seg.Similarity(s1))
}

func TestSegment_Split(t *testing.T) {
	a := assert.New(t)

	seg, err := NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)

	segs, err := seg.Split(4)
	a.NotError(err).
		Equal(segs[0].Value, "{id}").
		Equal(segs[0].Type, Named)
	a.Equal(segs[1].Value, "/author").
		Equal(segs[1].Type, String)

	segs, err = seg.Split(2)
	a.NotError(err)
	a.Equal(segs[0].Value, "{i").
		Equal(segs[0].Type, String)
	a.Equal(segs[1].Value, "d}/author").
		Equal(segs[1].Type, String)
}

func TestSegment_Match(t *testing.T) {
	a := assert.New(t)

	// Named:any
	seg, err := NewSegment("{id:any}/author")
	ps := params.Params{}
	a.NotError(err).NotNil(seg)
	path := "1/author"
	index := seg.Match(path, ps)
	a.Empty(path[index:]).
		Equal(ps, params.Params{"id": "1"})

	// Named 完全匹配
	ps = params.Params{}
	seg, err = NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)
	path = "1/author"
	index = seg.Match(path, ps)
	a.Empty(path[index:]).
		Equal(ps, params.Params{"id": "1"})

	// Named 部分匹配
	ps = params.Params{}
	seg, err = NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)
	path = "1/author/email"
	index = seg.Match(path, ps)
	a.Equal(path[index:], "/email")

	// Named 不匹配
	ps = params.Params{}
	seg, err = NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)
	path = "1/aut"
	index = seg.Match(path, ps)
	a.Equal(index, -1)

	// Named 1/2 匹配 {id}
	ps = params.Params{}
	seg, err = NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)
	path = "1/2/author"
	index = seg.Match(path, ps)
	a.Equal(path[index:], "").
		Equal(ps, map[string]string{"id": "1/2"})

	// Interceptor 1/2 匹配 {id}
	ps = params.Params{}
	seg, err = NewSegment("{id:any}/author")
	a.NotError(err).NotNil(seg)
	path = "1/2/author"
	index = seg.Match(path, ps)
	a.Equal(path[index:], "").
		Equal(ps, map[string]string{"id": "1/2"})

	ps = params.Params{}
	seg, err = NewSegment("{any}/123")
	a.NotError(err).NotNil(seg)
	path = "123"
	index = seg.Match(path, ps)
	a.Equal(index, -1)

	ps = params.Params{}
	seg, err = NewSegment("{any:any}/123")
	a.NotError(err).NotNil(seg)
	path = "123"
	index = seg.Match(path, ps)
	a.Equal(index, -1)

	// 命名参数
	ps = params.Params{}
	seg, err = NewSegment("{any}123")
	a.NotError(err).NotNil(seg)
	a.Equal(seg.Type, Named)
	path = "123123"
	index = seg.Match(path, ps)
	a.Equal(index, 3).Equal(ps, map[string]string{"any": ""})

	// Interceptor
	ps = params.Params{}
	seg, err = NewSegment("{any:any}123")
	a.NotError(err).NotNil(seg)
	a.Equal(seg.Type, Interceptor)
	path = "123123"
	index = seg.Match(path, ps)
	a.Equal(index, 6).
		Equal(ps, map[string]string{"any": "123"})

	ps = params.Params{}
	seg, err = NewSegment("{any:any}123")
	a.NotError(err).NotNil(seg)
	path = "12345123"
	index = seg.Match(path, ps)
	a.Equal(index, 8).
		Equal(ps, map[string]string{"any": "12345"})

	ps = params.Params{}
	seg, err = NewSegment("{any:digit}123")
	a.NotError(err).NotNil(seg)
	path = "12345123"
	index = seg.Match(path, ps)
	a.Equal(index, 8).
		Equal(ps, map[string]string{"any": "12345"})

	// Named Endpoint 匹配
	ps = params.Params{}
	seg, err = NewSegment("{path}")
	a.NotError(err).NotNil(seg)
	path = "/posts/author"
	index = seg.Match(path, ps)
	a.Equal(path[index:], "")

	// Named:digit Endpoint 匹配
	ps = params.Params{}
	seg, err = NewSegment("{id:digit}")
	a.NotError(err).NotNil(seg)
	path = "123"
	index = seg.Match(path, ps)
	a.Equal(path[index:], "").
		Equal(ps, map[string]string{"id": "123"})

	// Named:digit Endpoint 不匹配
	ps = params.Params{}
	seg, err = NewSegment("{id:digit}")
	a.NotError(err).NotNil(seg)
	path = "one"
	index = seg.Match(path, ps)
	a.Equal(index, -1).Empty(ps)

	// Named:digit
	seg, err = NewSegment("{id:digit}/author")
	a.NotError(err).NotNil(seg)

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
	seg, err = NewSegment("/posts/author")
	a.NotError(err).NotNil(seg)

	// String 匹配
	path = "/posts/author"
	index = seg.Match(path, nil)
	a.Equal(path[index:], "")

	// String 不匹配
	path = "/posts/author/email"
	index = seg.Match(path, nil)
	a.Equal(path[index:], "/email")

	// Regexp
	seg, err = NewSegment("{id:\\d+}/author")
	a.NotError(err).NotNil(seg)

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
