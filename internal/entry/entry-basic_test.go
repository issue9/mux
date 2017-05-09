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
	b := newBasic("/basic/str")
	a.True(b.match("/basic/str"))
	a.False(b.match("/basic/"))

	// 无效的通配符
	b = newBasic("/basic*")
	a.False(b.match("/basic"))
	a.False(b.match("/basic/"))
	a.True(b.match("/basic*"))

	// 通配符
	b = newBasic("/basic/str/*")
	a.False(b.match("/basic/str"))
	a.False(b.match("/basic"))
	a.True(b.match("/basic/str/"))
	a.True(b.match("/basic/str/index.html"))
	a.True(b.match("/basic/str/abc/def"))
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
