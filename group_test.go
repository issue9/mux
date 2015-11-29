// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestGroups(t *testing.T) {
	a := assert.New(t)
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	mux := NewServeMux()
	a.NotNil(mux)

	// 应该获取的是同一个group
	g1 := mux.Group()
	g2 := mux.Group()
	a.Equal(g1, g2)

	g1.Get("/abc", hf)
	assertLen(mux, a, 1, "GET")
}

// Group的各类状态，比如name,isrunning等
func TestGroup_Status(t *testing.T) {
	a := assert.New(t)
	mux := NewServeMux()
	a.NotNil(mux)

	g := mux.Group()
	a.Equal(g.isRunning, true).
		Equal(g.IsRunning(), g.isRunning).
		Equal(g.mux, mux)

	g.Stop()
	a.False(g.IsRunning())

	g.Start()
	a.True(g.IsRunning())
}

func TestGroup_Add(t *testing.T) {
	a := assert.New(t)
	m := NewServeMux()
	a.NotNil(m)

	g := m.Group()
	a.NotNil(g)

	fn := func(w http.ResponseWriter, req *http.Request) {}
	h := http.HandlerFunc(fn)

	a.NotPanic(func() { g.Get("h", h) })
	assertLen(m, a, 1, "GET")

	a.NotPanic(func() { g.Post("h", h) })
	assertLen(m, a, 1, "POST")

	a.NotPanic(func() { g.Put("h", h) })
	assertLen(m, a, 1, "PUT")

	a.NotPanic(func() { g.Delete("h", h) })
	assertLen(m, a, 1, "DELETE")

	a.NotPanic(func() { g.Patch("h", h) })
	assertLen(m, a, 1, "PATCH")

	a.NotPanic(func() { g.Any("anyH", h) })
	assertLen(m, a, 2, "PUT")
	assertLen(m, a, 2, "DELETE")

	a.NotPanic(func() { g.GetFunc("fn", fn) })
	assertLen(m, a, 3, "GET")

	a.NotPanic(func() { g.PostFunc("fn", fn) })
	assertLen(m, a, 3, "POST")

	a.NotPanic(func() { g.PutFunc("fn", fn) })
	assertLen(m, a, 3, "PUT")

	a.NotPanic(func() { g.DeleteFunc("fn", fn) })
	assertLen(m, a, 3, "DELETE")

	a.NotPanic(func() { g.PatchFunc("fn", fn) })
	assertLen(m, a, 3, "PATCH")

	a.NotPanic(func() { g.AnyFunc("anyFN", fn) })
	assertLen(m, a, 4, "DELETE")
	assertLen(m, a, 4, "GET")

	// 添加相同的pattern
	a.Panic(func() { g.Any("h", h) })

	// handler不能为空
	a.Panic(func() { g.Add("abc", nil, "GET") })
	// pattern不能为空
	//a.Panic(func() { g.Add("", h, "GET") })
	// 不支持的methods
	a.Panic(func() { g.Add("abc", h, "GET123") })

}
