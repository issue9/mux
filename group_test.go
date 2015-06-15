// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestGroup(t *testing.T) {
	a := assert.New(t)

	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewServeMux()
	a.NotNil(mux)

	g := mux.Group("g1")
	a.Equal(g.name, "g1").Equal(g.mux, mux).True(g.isRunning) // 保证初始化之后，isRunning为true

	g.Get("/abc", hf)
	a.Equal(1, mux.items["GET"].list.Len())
	// 通过ServeMux.Remove()可能删除从Group添加的内容。
	mux.Remove("/abc")
	a.Equal(0, mux.items["GET"].list.Len())
}
