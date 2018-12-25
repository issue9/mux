// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"testing"

	"github.com/issue9/assert"
)

var _ error = &SyntaxError{}

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

func TestRegexp(t *testing.T) {
	a := assert.New(t)

	a.Equal(repl.Replace("{id:\\d+}"), "(?P<id>\\d+)")
	a.Equal(repl.Replace("{id:\\d+}/author"), "(?P<id>\\d+)/author")
}

func TestStringType(t *testing.T) {
	a := assert.New(t)

	a.Equal(stringType(""), nodeTypeString)
	a.Equal(stringType("/posts"), nodeTypeString)
	a.Equal(stringType("/posts/{id}"), nodeTypeNamed)
	a.Equal(stringType("/posts/{id}/author"), nodeTypeNamed)
	a.Equal(stringType("/posts/{id:\\d+}/author"), nodeTypeRegexp)
}

func TestSplit(t *testing.T) {
	a := assert.New(t)
	test := func(str string, isError bool, ss ...string) {
		s, err := split(str)
		if isError {
			a.Error(err)
			return
		}

		a.NotError(err).
			Equal(len(s), len(ss))
		for index, seg := range ss {
			a.Equal(seg, s[index])
		}
	}

	test("/", false, "/")

	test("/posts/1", false, "/posts/1")

	test("{action}/1", false, "{action}/1")

	// 以命名参数开头的
	test("/{action}", false, "/", "{action}")

	// 以通配符结尾
	test("/posts/{id}", false, "/posts/", "{id}")

	test("/posts/{id}/author/profile", false, "/posts/", "{id}/author/profile")

	// 以命名参数结尾的
	test("/posts/{id}/author", false, "/posts/", "{id}/author")

	// 命名参数及通配符
	test("/posts/{id}/page/{page}", false, "/posts/", "{id}/page/", "{page}")

	// 正则
	test("/posts/{id:\\d+}", false, "/posts/", "{id:\\d+}")

	// 正则，命名参数
	test("/posts/{id:\\d+}/page/{page}", false, "/posts/", "{id:\\d+}/page/", "{page}")

	test("", true)
	test("/posts/{id:}", true)
	test("/posts/{{id:\\d+}/author", true)
	test("/posts/{:\\d+}/author", true)
	test("/posts/{}/author", true)
	test("/posts/{id}{page}/", true)
	test("/posts/:id/author", true)
	test("/posts/{id}/{author", true)
	test("/posts/}/author", true)
}
