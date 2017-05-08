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
	b := newBasic("/basic")
	a.True(b.match("/basic"))
	a.False(b.match("/basic/"))

	// 无效的通配符
	b = newBasic("/basic*")
	a.False(b.match("/basic"))
	a.False(b.match("/basic/"))
	a.True(b.match("/basic*"))

	// 通配符
	b = newBasic("/basic/*")
	a.False(b.match("/basic"))
	a.True(b.match("/basic/"))
	a.True(b.match("/basic/index.html"))
	a.True(b.match("/basic/abc/def"))
}
