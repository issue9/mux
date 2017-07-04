// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package segment

import (
	"testing"

	"github.com/issue9/assert"
)

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

func TestRegexp(t *testing.T) {
	a := assert.New(t)

	a.Equal(Regexp("{id:\\d+}"), "(?P<id>\\d+)")
	a.Equal(Regexp("{id:\\d+}/author"), "(?P<id>\\d+)/author")
}

func TestStringType(t *testing.T) {
	a := assert.New(t)

	a.Equal(StringType(""), TypeString)
	a.Equal(StringType("/posts"), TypeString)
	a.Equal(StringType("/posts/{id}"), TypeNamed)
	a.Equal(StringType("/posts/{id}/author"), TypeNamed)
	a.Equal(StringType("/posts/{id:\\d+}/author"), TypeRegexp)
}
