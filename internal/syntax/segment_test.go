// SPDX-License-Identifier: MIT

package syntax

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux/v5/params"
)

func TestNewSegment(t *testing.T) {
	a := assert.New(t)

	seg, err := NewSegment("/post/1")
	a.NotError(err).Equal(seg.Type, String).Equal(seg.Value, "/post/1")

	seg, err = NewSegment("{id}/1")
	a.NotError(err).Equal(seg.Type, Named).
		Equal(seg.Value, "{id}/1").
		Empty(seg.Rule).
		False(seg.Endpoint).
		Equal(seg.Suffix, "/1")

	seg, err = NewSegment("{id}")
	a.NotError(err).Equal(seg.Type, Named).
		Equal(seg.Value, "{id}").
		Empty(seg.Rule).
		True(seg.Endpoint).
		Empty(seg.Suffix)

	seg, err = NewSegment("{id:}")
	a.NotError(err).Equal(seg.Type, Named).
		Equal(seg.Value, "{id:}").
		True(seg.Endpoint).
		Empty(seg.Rule).
		Empty(seg.Suffix).
		Equal(seg.Name, "id")

	seg, err = NewSegment("{id}:")
	a.NotError(err).Equal(seg.Type, Named).
		Equal(seg.Value, "{id}:").
		False(seg.Endpoint).
		Empty(seg.Rule).
		Equal(seg.Suffix, ":").
		Equal(seg.Name, "id")

	seg, err = NewSegment("{id:any}")
	a.NotError(err).Equal(seg.Type, Interceptor).
		Equal(seg.Value, "{id:any}").
		True(seg.Endpoint).
		Equal(seg.Rule, "any").
		Empty(seg.Suffix)

	seg, err = NewSegment("{id:digit}/1")
	a.NotError(err).Equal(seg.Type, Interceptor).
		Equal(seg.Value, "{id:digit}/1").
		False(seg.Endpoint).
		Equal(seg.Rule, "digit").
		Equal(seg.Suffix, "/1")

	seg, err = NewSegment("{id:\\d+}/1")
	a.NotError(err).Equal(seg.Type, Regexp).
		Equal(seg.Value, "{id:\\d+}/1").
		False(seg.Endpoint).
		Equal(seg.Rule, "\\d+").
		Equal(seg.Suffix, "/1")

	seg, err = NewSegment("id:}{")
	a.Error(err).Nil(seg)

	seg, err = NewSegment("id:{}")
	a.Error(err).Nil(seg)

	seg, err = NewSegment("{path")
	a.NotError(err).NotNil(seg).
		Equal(seg.Type, String).
		Empty(seg.Rule).
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
	a.NotError(err).NotNil(seg)
	p := &MatchParam{Path: "1/author"}
	a.True(seg.Match(p)).
		Empty(p.Path).
		Equal(p.Params, params.Params{"id": "1"})

	// Named 完全匹配
	seg, err = NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)
	p = &MatchParam{Path: "1/author"}
	a.True(seg.Match(p)).
		Empty(p.Path).
		Equal(p.Params, params.Params{"id": "1"})

	// Named 部分匹配
	seg, err = NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)
	p = &MatchParam{Path: "1/author/email"}
	a.True(seg.Match(p))
	a.Equal(p.Path, "/email").Equal(p.Params, params.Params{"id": "1"})

	// Named 不匹配
	seg, err = NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)
	p = &MatchParam{Path: "1/aut"}
	a.False(seg.Match(p))
	a.Equal(p.Path, "1/aut").Empty(p.Params)

	// Named 1/2 匹配 {id}
	seg, err = NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)
	p = &MatchParam{Path: "1/2/author"}
	a.True(seg.Match(p)).
		Equal(p.Path, "").
		Equal(p.Params, map[string]string{"id": "1/2"})

	// Interceptor 1/2 匹配 {id}
	seg, err = NewSegment("{id:any}/author")
	a.NotError(err).NotNil(seg)
	p = &MatchParam{Path: "1/2/author"}
	a.True(seg.Match(p)).
		Equal(p.Path, "").
		Equal(p.Params, map[string]string{"id": "1/2"})

	seg, err = NewSegment("{any}/123")
	a.NotError(err).NotNil(seg)
	p = &MatchParam{Path: "123"}
	a.False(seg.Match(p)).
		Equal(p.Path, "123").Empty(p.Params)

	seg, err = NewSegment("{any:any}/123")
	a.NotError(err).NotNil(seg)
	p = &MatchParam{Path: "123"}
	a.False(seg.Match(p)).
		Equal(p.Path, "123").Empty(p.Params)

	// 命名参数，any 匹配到了空.
	seg, err = NewSegment("{any}123")
	a.NotError(err).NotNil(seg)
	a.Equal(seg.Type, Named)
	p = &MatchParam{Path: "123123"}
	a.True(seg.Match(p)).
		Equal(p.Path, "123").
		Equal(p.Params, map[string]string{"any": ""})

	// Interceptor
	seg, err = NewSegment("{any:any}123")
	a.NotError(err).NotNil(seg)
	a.Equal(seg.Type, Interceptor)
	p = &MatchParam{Path: "123123"}
	a.True(seg.Match(p))
	a.Empty(p.Path).
		Equal(p.Params, map[string]string{"any": "123"})

	seg, err = NewSegment("{any:any}123")
	a.NotError(err).NotNil(seg)
	p = &MatchParam{Path: "12345123"}
	a.True(seg.Match(p)).
		Empty(p.Path).
		Equal(p.Params, map[string]string{"any": "12345"})

	seg, err = NewSegment("{any:digit}123")
	a.NotError(err).NotNil(seg)
	p = &MatchParam{Path: "12345123"}
	a.True(seg.Match(p)).
		Empty(p.Path).
		Equal(p.Params, map[string]string{"any": "12345"})

	// Named Endpoint 匹配
	seg, err = NewSegment("{path}")
	a.NotError(err).NotNil(seg)
	p = &MatchParam{Path: "/posts/author"}
	a.True(seg.Match(p)).
		Empty(p.Path).
		Equal(p.Params, params.Params{"path": "/posts/author"})

	// Named:digit Endpoint 匹配
	seg, err = NewSegment("{id:digit}")
	a.NotError(err).NotNil(seg)
	p = &MatchParam{Path: "123"}
	a.True(seg.Match(p))
	a.Empty(p.Path).
		Equal(p.Params, map[string]string{"id": "123"})

	// Named:digit Endpoint 不匹配，不会删除传入的参数
	seg, err = NewSegment("{id:digit}")
	a.NotError(err).NotNil(seg)
	p = &MatchParam{Path: "one", Params: params.Params{"p1": "v1"}}
	a.False(seg.Match(p)).
		Equal(p.Path, "one").
		Equal(p.Params, params.Params{"p1": "v1"})

	// Named:digit
	seg, err = NewSegment("{id:digit}/author")
	a.NotError(err).NotNil(seg)

	// Named:digit 不匹配
	p = &MatchParam{Path: "1/aut"}
	a.False(seg.Match(p)).
		Equal(p.Path, "1/aut").Empty(p.Params)

	// Named:digit 类型不匹配
	p = &MatchParam{Path: "xx/author", Params: params.Params{}}
	a.False(seg.Match(p)).
		Equal(p.Path, "xx/author").
		Empty(p.Params)

	// String
	seg, err = NewSegment("/posts/author")
	a.NotError(err).NotNil(seg)

	// String 匹配
	p = &MatchParam{Path: "/posts/author", Params: params.Params{"p1": "v1"}}
	a.True(seg.Match(p)).
		Empty(p.Path).
		Equal(p.Params, params.Params{"p1": "v1"})

	// String 不匹配
	p = &MatchParam{Path: "/posts/author/email"}
	a.True(seg.Match(p))
	a.Equal(p.Path, "/email").
		Empty(p.Params)

	// Regexp
	seg, err = NewSegment("{id:\\d+}/author")
	a.NotError(err).NotNil(seg)

	// Regexp 完全匹配
	p = &MatchParam{Path: "1/author"}
	a.True(seg.Match(p)).
		Empty(p.Path).
		Equal(p.Params, params.Params{"id": "1"})

	// Regexp 不匹配
	p = &MatchParam{Path: "xxx/author"}
	a.False(seg.Match(p)).
		Equal(p.Path, "xxx/author").
		Empty(p.Params)

	// Regexp 部分匹配
	p = &MatchParam{Path: "1/author/email"}
	a.True(seg.Match(p)).
		Equal(p.Path, "/email").
		Equal(p.Params, params.Params{"id": "1"})
}
