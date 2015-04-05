// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestEntry_New(t *testing.T) {
	a := assert.New(t)

	fn := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})

	// 普通情况
	e, err := newEntry("abc", fn)
	a.NotError(err).Equal(e.pattern, "abc")

	// 正则匹配
	e, err = newEntry("?abc", fn)
	a.NotError(err).Equal("abc", e.pattern)

	// pattern为空
	e, err = newEntry("", fn)
	a.Error(err).Nil(e)

	// pattern正则内容为空
	e, err = newEntry("?", fn)
	a.Error(err).Nil(e)
}
