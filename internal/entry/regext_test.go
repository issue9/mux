// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"testing"

	"github.com/issue9/assert"
)

func TestSplit(t *testing.T) {
	a := assert.New(t)

	// 为空
	a.Equal(split(""), []string{})

	// 不存在 {}
	a.Equal(split("/blog/post/1"), []string{"/blog/post/1"})

	// 开头包含 {}
	a.Equal(split("{action}/post/1"), []string{"{action}", "/post/1"})

	// 结尾包含 {}
	a.Equal(split("/blog/post/{id}"), []string{"/blog/post/", "{id}"})
	a.Equal(split("/blog/post/{id:\\d}"), []string{"/blog/post/", "{id:\\d}"})

	// 中间包含
	a.Equal(split("/blog/post/{id}/author"), []string{"/blog/post/", "{id}", "/author"})
	a.Equal(split("/blog/{post}-{id}/author"), []string{"/blog/", "{post}", "-", "{id}", "/author"})

	a.Equal(split("{action}/{post}/{id}"), []string{"{action}", "/", "{post}", "/", "{id}"})
	a.Equal(split("{action}/{post}-{id}"), []string{"{action}", "/", "{post}", "-", "{id}"})

	// 无法解析的内容
	a.Equal(split("{/blog/post/{id}"), []string{"{/blog/post/{id}"})
	a.Equal(split("}/blog/post/{id}"), []string{"}/blog/post/", "{id}"})
}

func TestToPattern(t *testing.T) {
	a := assert.New(t)

	fn := func(str []string, pattern string, hasParams bool, hasError bool) {
		p, b, err := toPattern(str)
		if hasError {
			a.Error(err)
		} else {
			a.NotError(err).Equal(p, pattern).Equal(b, hasParams)
		}
	}

	fn([]string{"/blog/post/1"}, "/blog/post/1", false, false)              // 静态
	fn([]string{"/blog/post/", "{:\\d+}"}, "/blog/post/\\d+", false, false) // 无命名路由参数

	fn([]string{"/blog/post/", "{id}"}, "/blog/post/(?P<id>[^/]+)", true, false)
	fn([]string{"/blog/post/", "{id}", "/", "{id:\\d+}"}, "", true, true) // 重复的参数名
	fn([]string{"/blog/post/", "{id:\\d+}"}, "/blog/post/(?P<id>\\d+)", true, false)
	fn([]string{"/blog/", "{post}", "-", "{id}"}, "/blog/(?P<post>[^/]+)-(?P<id>[^/]+)", true, false) // 两个参数
	fn([]string{"/blog/", "{:\\w+}", "-", "{id}"}, "/blog/\\w+-(?P<id>[^/]+)", true, false)
}
