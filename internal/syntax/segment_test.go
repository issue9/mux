// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package syntax

import (
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/mux/v9/types"
)

func TestNewSegment(t *testing.T) {
	a := assert.New(t, false)
	i := newInterceptors(a)
	a.NotNil(i)

	seg, err := i.NewSegment("/post/1")
	a.NotError(err).Equal(seg.Type, String).Equal(seg.Value, "/post/1")

	seg, err = i.NewSegment("{id}/1")
	a.NotError(err).Equal(seg.Type, Named).
		Equal(seg.Value, "{id}/1").
		Empty(seg.rule).
		False(seg.Endpoint).
		Equal(seg.ambiguousLength, 4).
		Equal(seg.AmbiguousLen(), 6).
		Equal(seg.Suffix, "/1")

	seg, err = i.NewSegment("{id}")
	a.NotError(err).Equal(seg.Type, Named).
		Equal(seg.Value, "{id}").
		Empty(seg.rule).
		False(seg.ignoreName).
		True(seg.Endpoint).
		Equal(seg.ambiguousLength, 2).
		Empty(seg.Suffix)

	// ignore 的 regexp
	seg, err = i.NewSegment("{-id:\\d+}")
	a.NotError(err).Equal(seg.Type, Regexp).
		Equal(seg.Value, "{-id:\\d+}").
		Equal(seg.Name, "id").
		True(seg.ignoreName).
		Equal(seg.rule, "\\d+").
		False(seg.Endpoint).
		Equal(seg.ambiguousLength, 7).
		Empty(seg.Suffix)

	// ignore 的拦截
	seg, err = i.NewSegment("{-id:any}")
	a.NotError(err).Equal(seg.Type, Interceptor).
		Equal(seg.Value, "{-id:any}").
		Equal(seg.Name, "id").
		True(seg.ignoreName).
		Equal(seg.rule, "any").
		True(seg.Endpoint).
		Equal(seg.ambiguousLength, 7).
		Empty(seg.Suffix)

	// 没有名称的命名参数，匹配任意字符
	seg, err = i.NewSegment("{-id:}")
	a.NotError(err).Equal(seg.Type, Named).
		Equal(seg.Value, "{-id:}").
		True(seg.Endpoint).
		Empty(seg.rule).
		True(seg.ignoreName).
		Empty(seg.Suffix).
		Equal(seg.ambiguousLength, 3).
		Equal(seg.Name, "id")

	seg, err = i.NewSegment("{id:}")
	a.NotError(err).Equal(seg.Type, Named).
		Equal(seg.Value, "{id:}").
		True(seg.Endpoint).
		Empty(seg.rule).
		False(seg.ignoreName).
		Empty(seg.Suffix).
		Equal(seg.Name, "id")

	seg, err = i.NewSegment("{id}:")
	a.NotError(err).Equal(seg.Type, Named).
		Equal(seg.Value, "{id}:").
		False(seg.Endpoint).
		Empty(seg.rule).
		Equal(seg.Suffix, ":").
		Equal(seg.Name, "id")

	seg, err = i.NewSegment("{id:any}")
	a.NotError(err).Equal(seg.Type, Interceptor).
		Equal(seg.Value, "{id:any}").
		True(seg.Endpoint).
		Equal(seg.rule, "any").
		Equal(seg.ambiguousLength, 6).
		Empty(seg.Suffix)

	seg, err = i.NewSegment("{id:digit}/1")
	a.NotError(err).Equal(seg.Type, Interceptor).
		Equal(seg.Value, "{id:digit}/1").
		False(seg.Endpoint).
		Equal(seg.rule, "digit").
		Equal(seg.ambiguousLength, 10).
		Equal(seg.Suffix, "/1")

	seg, err = i.NewSegment("{id:\\d+}/1")
	a.NotError(err).Equal(seg.Type, Regexp).
		Equal(seg.Value, "{id:\\d+}/1").
		False(seg.Endpoint).
		Equal(seg.rule, "\\d+").
		Equal(seg.ambiguousLength, 8).
		Equal(seg.Suffix, "/1")

	seg, err = i.NewSegment("id:}{")
	a.Error(err).Nil(seg)

	seg, err = i.NewSegment("id:{}")
	a.Error(err).Nil(seg)

	seg, err = i.NewSegment("{path")
	a.NotError(err).NotNil(seg).
		Equal(seg.Type, String).
		Empty(seg.rule).
		Equal(seg.Value, "{path")

	seg, err = i.NewSegment("{:path")
	a.NotError(err).NotNil(seg).
		Equal(seg.Type, String).
		Equal(seg.Value, "{:path")
}

func TestSegment_IsAmbiguous(t *testing.T) {
	a := assert.New(t, false)
	i := newInterceptors(a)

	data := []struct {
		s1, s2 string
		eq     bool
	}{
		{
			s1: "/{id}/1",
			s2: "/{id}/1",
		},
		{
			s1: "/{-id}/1",
			s2: "/{id}/1",
			eq: true,
		},
		{
			s1: "/{-id:any}",
			s2: "/{id:any}",
			eq: true,
		},
		{
			s1: "/{-id:}",
			s2: "/{id:any}",
		},
		{
			s1: "/{id:}",
			s2: "/{id:any}",
		},
		{
			s1: "/{id:digit}",
			s2: "/{id:any}",
		},
		{
			s1: "/{-id:digit}",
			s2: "/{-id:any}",
		},
		{
			s1: "/{-id:digit}",
			s2: "/{-id:digit}",
		},
		{
			s1: "/{-id:}",
			s2: "/{-id:}",
		},
		{
			s1: "/{id:}",
			s2: "/{-id:}",
			eq: true,
		},
		{
			s1: "/{id:\\d+}",
			s2: "/{-id:\\d+}",
			eq: true,
		},
	}

	for _, item := range data {
		ss1, err := i.NewSegment(item.s1)
		a.NotError(err).NotNil(ss1)

		ss2, err := i.NewSegment(item.s2)
		a.NotError(err).NotNil(ss2)

		if item.eq {
			a.True(ss1.IsAmbiguous(ss2), "false at %s", item.s1).
				True(ss2.IsAmbiguous(ss1), "false at %s", item.s1)
		} else {
			a.False(ss1.IsAmbiguous(ss2), "false at %s", item.s1).
				False(ss2.IsAmbiguous(ss1), "false at %s", item.s1)
		}
	}
}

func TestLongestPrefix(t *testing.T) {
	a := assert.New(t, false)

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
	test("{t}bc", "{t}abc", 0)
	test("aa{t}bc", "aa{t}abc", 2)
	test("/tes{t:\\d+}", "/tes{t}", 4)
}

func TestSegment_Similarity(t *testing.T) {
	a := assert.New(t, false)
	i := newInterceptors(a)
	a.NotNil(i)

	seg, err := i.NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)

	s1, err := i.NewSegment("{id}/author")
	a.NotError(err).Equal(-1, seg.Similarity(s1))

	s1, err = i.NewSegment("{id}/author/profile")
	a.NotError(err).Equal(11, seg.Similarity(s1))
}

func TestSegment_Split(t *testing.T) {
	a := assert.New(t, false)
	i := newInterceptors(a)
	a.NotNil(i)

	seg, err := i.NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)

	segs, err := seg.Split(i, 4)
	a.NotError(err).
		Equal(segs[0].Value, "{id}").
		Equal(segs[0].Type, Named)
	a.Equal(segs[1].Value, "/author").
		Equal(segs[1].Type, String)

	segs, err = seg.Split(i, 2)
	a.NotError(err)
	a.Equal(segs[0].Value, "{i").
		Equal(segs[0].Type, String)
	a.Equal(segs[1].Value, "d}/author").
		Equal(segs[1].Type, String)
}

func TestSegment_Match(t *testing.T) {
	a := assert.New(t, false)
	i := newInterceptors(a)
	a.NotNil(i)

	// Named:any
	seg, err := i.NewSegment("{id:any}/author")
	a.NotError(err).NotNil(seg)
	p := types.NewContext()
	p.Path = "1/author"
	a.True(seg.Match(p)).
		Empty(p.Path).
		Equal(1, p.Count()).Equal(p.MustString("id", "not-exists"), "1")

	// Named 完全匹配
	seg, err = i.NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)
	p = types.NewContext()
	p.Path = "1/author"
	a.True(seg.Match(p)).
		Empty(p.Path).
		Equal(1, p.Count()).
		Equal(1, p.Count()).Equal(p.MustString("id", "not-exists"), "1")

	// Named 部分匹配
	seg, err = i.NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)
	p = types.NewContext()
	p.Path = "1/author/email"
	a.True(seg.Match(p))
	a.Equal(p.Path, "/email").
		Equal(1, p.Count()).
		Equal(1, p.Count()).Equal(p.MustString("id", "not-exists"), "1")

	// Named 不匹配
	seg, err = i.NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)
	p = types.NewContext()
	p.Path = "1/aut"
	a.False(seg.Match(p))
	a.Equal(p.Path, "1/aut").Zero(p.Count())

	// Named 1/2 匹配 {id}
	seg, err = i.NewSegment("{id}/author")
	a.NotError(err).NotNil(seg)
	p = types.NewContext()
	p.Path = "1/2/author"
	a.True(seg.Match(p)).
		Equal(p.Path, "").
		Equal(1, p.Count()).
		Equal(1, p.Count()).Equal(p.MustString("id", "not-exists"), "1/2")

	// Interceptor 1/2 匹配 {id}
	seg, err = i.NewSegment("{id:any}/author")
	a.NotError(err).NotNil(seg)
	p = types.NewContext()
	p.Path = "1/2/author"
	a.True(seg.Match(p)).
		Equal(p.Path, "").
		Equal(1, p.Count()).
		Equal(p.MustString("id", "not-exists"), "1/2")

	seg, err = i.NewSegment("{any}/123")
	a.NotError(err).NotNil(seg)
	p = types.NewContext()
	p.Path = "123"
	a.False(seg.Match(p)).
		Equal(p.Path, "123").Zero(p.Count())

	seg, err = i.NewSegment("{any:any}/123")
	a.NotError(err).NotNil(seg)
	p = types.NewContext()
	p.Path = "123"
	a.False(seg.Match(p)).
		Equal(p.Path, "123").Zero(p.Count())

	// 命名参数，any 匹配到了空.
	seg, err = i.NewSegment("{any}123")
	a.NotError(err).NotNil(seg)
	a.Equal(seg.Type, Named)
	p = types.NewContext()
	p.Path = "123123"
	a.True(seg.Match(p)).
		Equal(p.Path, "123").
		Equal(1, p.Count()).
		Empty(p.MustString("any", "not-exists"))

	// Interceptor
	seg, err = i.NewSegment("{any:any}123")
	a.NotError(err).NotNil(seg)
	a.Equal(seg.Type, Interceptor)
	p = types.NewContext()
	p.Path = "123123"
	a.True(seg.Match(p))
	a.Empty(p.Path).
		Equal(1, p.Count()).
		Equal(p.MustString("any", "not-exists"), "123")

	seg, err = i.NewSegment("{any:any}123")
	a.NotError(err).NotNil(seg)
	p = types.NewContext()
	p.Path = "12345123"
	a.True(seg.Match(p)).
		Empty(p.Path).
		Equal(1, p.Count()).
		Equal(p.MustString("any", "not-exists"), "12345")

	seg, err = i.NewSegment("{any:digit}123")
	a.NotError(err).NotNil(seg)
	p = types.NewContext()
	p.Path = "12345123"
	a.True(seg.Match(p)).
		Empty(p.Path).
		Equal(1, p.Count()).
		Equal(p.MustString("any", "not-exists"), "12345")

	// Named Endpoint 匹配
	seg, err = i.NewSegment("{path}")
	a.NotError(err).NotNil(seg)
	p = types.NewContext()
	p.Path = "/posts/author"
	a.True(seg.Match(p)).
		Empty(p.Path).
		Equal(1, p.Count()).
		Equal(p.MustString("path", "not-exists"), "/posts/author")

	// Named:digit Endpoint 匹配
	seg, err = i.NewSegment("{id:digit}")
	a.NotError(err).NotNil(seg)
	p = types.NewContext()
	p.Path = "123"
	a.True(seg.Match(p))
	a.Empty(p.Path).
		Equal(1, p.Count()).
		Equal(p.MustString("id", "not-exists"), "123")

	// Named:digit Endpoint 不匹配，不会删除传入的参数
	seg, err = i.NewSegment("{id:digit}")
	a.NotError(err).NotNil(seg)
	p = types.NewContext()
	p.Path = "one"
	p.Set("p1", "v1")
	a.False(seg.Match(p)).
		Equal(p.Path, "one").
		Equal(1, p.Count()).
		Equal(p.MustString("p1", "not-exists"), "v1")

	// Named:digit
	seg, err = i.NewSegment("{id:digit}/author")
	a.NotError(err).NotNil(seg)

	// Named:digit 不匹配
	p = types.NewContext()
	p.Path = "1/aut"
	a.False(seg.Match(p)).
		Equal(p.Path, "1/aut").Zero(p.Count())

	// Named:digit 类型不匹配
	p = types.NewContext()
	p.Path = "xx/author"
	a.False(seg.Match(p)).
		Equal(p.Path, "xx/author").
		Zero(p.Count())

	// String
	seg, err = i.NewSegment("/posts/author")
	a.NotError(err).NotNil(seg)

	// String 匹配
	p = types.NewContext()
	p.Path = "/posts/author"
	p.Set("p1", "v1")
	a.True(seg.Match(p)).
		Empty(p.Path).
		Equal(1, p.Count()).
		Equal(p.MustString("p1", "not-exists"), "v1")

	// String 不匹配
	p = types.NewContext()
	p.Path = "/posts/author/email"
	a.True(seg.Match(p))
	a.Equal(p.Path, "/email").Zero(p.Count())

	// Regexp
	seg, err = i.NewSegment("{id:\\d+}/author")
	a.NotError(err).NotNil(seg)

	// Regexp 完全匹配
	p = types.NewContext()
	p.Path = "1/author"
	a.True(seg.Match(p)).
		Empty(p.Path).
		Equal(1, p.Count()).
		Equal(p.MustString("id", "not-exists"), "1")

	// Regexp 不匹配
	p = types.NewContext()
	p.Path = "xxx/author"
	a.False(seg.Match(p)).
		Equal(p.Path, "xxx/author").
		Zero(p.Count())

	// Regexp 部分匹配
	p = types.NewContext()
	p.Path = "1/author/email"
	a.True(seg.Match(p)).
		Equal(p.Path, "/email").
		Equal(1, p.Count()).
		Equal(p.MustString("id", "not-exists"), "1")
}

func TestSegment_Valid(t *testing.T) {
	a := assert.New(t, false)
	i := NewInterceptors()
	i.Add(MatchDigit, "digit")

	s, err := i.NewSegment("{id}")
	a.NotError(err).NotNil(s)
	a.True(s.Valid("5"))
	a.True(s.Valid("55xf"))

	s, err = i.NewSegment("{id:digit}")
	a.NotError(err).NotNil(s)
	a.True(s.Valid("5"))
	a.False(s.Valid("55xf"))

	s, err = i.NewSegment("{id:digit}/5xx")
	a.NotError(err).NotNil(s)
	a.True(s.Valid("5"))
	a.False(s.Valid("55xf"))

	s, err = i.NewSegment("{-id:digit}/5xx")
	a.NotError(err).NotNil(s)
	a.True(s.Valid("5"))
	a.False(s.Valid("55xf"))

	s, err = i.NewSegment("{id:\\d+}")
	a.NotError(err).NotNil(s)
	a.True(s.Valid("5"))
	a.False(s.Valid("55xfg"))

	s, err = i.NewSegment("{id:\\d+}/5xx")
	a.NotError(err).NotNil(s)
	a.True(s.Valid("5"))
	a.False(s.Valid("55xf"))

	s, err = i.NewSegment("{-id:\\d+}/5xx")
	a.NotError(err).NotNil(s)
	a.True(s.Valid("5"))
	a.False(s.Valid("55xf"))
}
