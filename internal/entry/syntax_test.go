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

	// 首中尾都包含
	a.Equal(split("{action}/{post}/{id}"), []string{"{action}", "/", "{post}", "/", "{id}"})
	a.Equal(split("{action}/{post}-{id}"), []string{"{action}", "/", "{post}", "-", "{id}"})

	// 无法解析的内容
	a.Equal(split("{/blog/post/{id}"), []string{"{/blog/post/{id}"})
	a.Equal(split("}/blog/post/{id}"), []string{"}/blog/post/", "{id}"})
}
