// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"testing"

	"github.com/issue9/assert"
)

var _ Entry = &basic{}

func TestBasic_match(t *testing.T) {
	a := assert.New(t)

	newMatcher(a, "/basic/str").
		True("/basic/str", nil).
		False("/basic/", nil)

	// 无效的通配符，不是放在 / 后面，只能当作普通字符处理
	newMatcher(a, "/basic*").
		False("/basic", nil).
		False("/basic/", nil).
		True("/basic*", nil).
		False("/basic/*", nil)

	// 通配符
	newMatcher(a, "/basic/str/*").
		False("/basic/str", nil).
		False("/basic", nil).
		True("/basic/str/", nil).
		True("/basic/str/index.html", nil).
		True("/basic/str/abc/def", nil)
}

func TestBasic_URL(t *testing.T) {
	a := assert.New(t)
	b := newBasic("/basic")

	url, err := b.URL(map[string]string{"id": "1"}, "/abc")
	a.Error(err).Equal(url, "")

	b = newBasic("/basic/*")
	url, err = b.URL(map[string]string{"id": "1"}, "abc")
	a.NotError(err).Equal(url, "/basic/abc")

	// 指定了空的  path
	url, err = b.URL(map[string]string{"id": "1"}, "")
	a.NotError(err).Equal(url, "/basic/")
}
