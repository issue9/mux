// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"testing"

	"github.com/issue9/assert"
)

var _ Entry = &basic{}

func TestBasic_Type(t *testing.T) {
	a := assert.New(t)
	b := &basic{}
	a.Equal(b.Type(), TypeBasic)
}

func TestBasic_Match(t *testing.T) {
	a := assert.New(t)
	b := newBasic("/basic")
	a.True(b.Match("/basic"))
	a.False(b.Match("/basic/"))

	// 无效的通配符
	b = newBasic("/basic*")
	a.False(b.Match("/basic"))
	a.False(b.Match("/basic/"))
	a.True(b.Match("/basic*"))

	// 通配符
	b = newBasic("/basic/*")
	a.False(b.Match("/basic"))
	a.True(b.Match("/basic/"))
	a.True(b.Match("/basic/index.html"))
	a.True(b.Match("/basic/abc/def"))
}
