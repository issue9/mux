// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package segment

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux/params"
)

var _ Segment = str("")
var _ Segment = &named{}
var _ Segment = &reg{}

func TestStringType(t *testing.T) {
	a := assert.New(t)

	a.Equal(stringType(""), TypeString)
	a.Equal(stringType("/posts"), TypeString)
	a.Equal(stringType("/posts/{id}"), TypeNamed)
	a.Equal(stringType("/posts/{id}/author"), TypeNamed)
	a.Equal(stringType("/posts/{id:\\d+}/author"), TypeRegexp)
}

func TestEqual(t *testing.T) {
	a := assert.New(t)

	a.True(Equal(&named{}, &named{}))

	s1, err := newNamed("{action}")
	a.NotError(err).NotNil(s1)
	s2, err := newNamed("{action}")
	a.NotError(err).NotNil(s2)
	a.True(Equal(s1, s2))

	s2, err = newNamed("{action}/1")
	a.NotError(err).NotNil(s2)
	a.False(Equal(s1, s2))
}

func TestLongestPrefix(t *testing.T) {
	a := assert.New(t)

	test := func(s1, s2 string, len int) {
		a.Equal(LongestPrefix(s1, s2), len)
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

func TestRepl(t *testing.T) {
	a := assert.New(t)

	a.Equal(repl.Replace("{id:\\d+}"), "(?P<id>\\d+)")
	a.Equal(repl.Replace("{id:\\d+}/author"), "(?P<id>\\d+)/author")
}

func BenchmarkStr_Match(b *testing.B) {
	a := assert.New(b)

	r, err := newReg("{id:\\d+}/author")
	a.NotError(err).NotNil(r)

	ps := make(params.Params, 1)

	for i := 0; i < b.N; i++ {
		if ok, _ := r.Match("5/author", ps); !ok {
			b.Error("BenchmarkStr_Match")
		}
	}
}
