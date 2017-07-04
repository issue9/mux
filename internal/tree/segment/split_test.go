// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package segment

import (
	"testing"

	"github.com/issue9/assert"
)

func TestSplit(t *testing.T) {
	a := assert.New(t)
	test := func(str string, isError bool, ss ...string) {
		s, err := Split(str)
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
