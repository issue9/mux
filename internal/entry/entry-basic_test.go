// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"testing"

	"github.com/issue9/assert"
)

func TestBasic_Type(t *testing.T) {
	a := assert.New(t)
	b := &basic{}
	a.Equal(b.Type(), TypeBasic)
}

func TestBasic_Match(t *testing.T) {
	a := assert.New(t)
	b := newBasic("/basic")
	a.Equal(b.Match("/basic"), 0)
	a.Equal(b.Match("/basic/"), -1)

	// 无效的通配符
	b = newBasic("/basic*")
	a.Equal(b.Match("/basic"), -1)
	a.Equal(b.Match("/basic/"), -1)
	a.Equal(b.Match("/basic*"), 0)

	// 通配符
	b = newBasic("/basic/*")
	a.Equal(b.Match("/basic"), -1)
	a.Equal(b.Match("/basic/"), 0)
	a.Equal(b.Match("/basic/index.html"), 0)
	a.Equal(b.Match("/basic/abc/def"), 0)
}
