// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestPrefix(t *testing.T) {
	a := assert.New(t)

	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewServeMux()
	p := mux.Prefix("p1")
	a.Equal(p.prefix, "p1").Equal(p.mux, mux)

	p.Post("/abc", hf)
	p.Get("/abc", hf)
	p.Delete("/abc", hf)

	assertLen(mux, a, 1, "GET")
	assertLen(mux, a, 1, "POST")
	assertLen(mux, a, 1, "DELETE")
	p.Remove("/abc", "GET") // 从Prefix.Remove()删除
	assertLen(mux, a, 0, "GET")
	assertLen(mux, a, 1, "POST")
	assertLen(mux, a, 1, "DELETE")
	mux.Remove("/abc", "POST") // 从ServeMux.Remove()删除，未带上p1前缀，无法删除
	assertLen(mux, a, 0, "GET")
	assertLen(mux, a, 1, "POST")
	assertLen(mux, a, 1, "DELETE")
	mux.Remove("p1/abc", "POST") // 从ServeMux.Remove()删除
	assertLen(mux, a, 0, "GET")
	assertLen(mux, a, 0, "POST")
	assertLen(mux, a, 1, "DELETE")
}

func TestGroup_Prefix(t *testing.T) {
	a := assert.New(t)
	mux := NewServeMux()
	a.NotNil(mux)

	g := mux.Group()
	a.NotNil(g)

	p := g.Prefix("/p")
	a.Equal(p.group, g)
}

func TestPrefix_Prefix(t *testing.T) {
	a := assert.New(t)

	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewServeMux()
	p1 := mux.Prefix("/p1")
	a.Equal(p1.prefix, "/p1").Equal(p1.mux, mux)

	p2 := p1.Prefix("/p2")
	a.Equal(p2.prefix, "/p1/p2")

	p1.Get("/abc", hf)
	p2.Get("/abc", hf)
	assertLen(mux, a, 2, "GET")
}

func TestPrefix_Add(t *testing.T) {
	a := assert.New(t)
	m := NewServeMux()
	a.NotNil(m)

	p := m.Prefix("prefix")
	a.NotNil(p)

	fn := func(w http.ResponseWriter, req *http.Request) {}
	h := http.HandlerFunc(fn)

	a.NotPanic(func() { p.Get("h", h) })
	assertLen(m, a, 1, "GET")

	a.NotPanic(func() { p.Post("h", h) })
	assertLen(m, a, 1, "POST")

	a.NotPanic(func() { p.Put("h", h) })
	assertLen(m, a, 1, "PUT")

	a.NotPanic(func() { p.Delete("h", h) })
	assertLen(m, a, 1, "DELETE")

	a.NotPanic(func() { p.Patch("h", h) })
	assertLen(m, a, 1, "PATCH")

	a.NotPanic(func() { p.Any("anyH", h) })
	assertLen(m, a, 2, "PUT")
	assertLen(m, a, 2, "DELETE")

	a.NotPanic(func() { p.GetFunc("fn", fn) })
	assertLen(m, a, 3, "GET")

	a.NotPanic(func() { p.PostFunc("fn", fn) })
	assertLen(m, a, 3, "POST")

	a.NotPanic(func() { p.PutFunc("fn", fn) })
	assertLen(m, a, 3, "PUT")

	a.NotPanic(func() { p.DeleteFunc("fn", fn) })
	assertLen(m, a, 3, "DELETE")

	a.NotPanic(func() { p.PatchFunc("fn", fn) })
	assertLen(m, a, 3, "PATCH")

	a.NotPanic(func() { p.AnyFunc("anyFN", fn) })
	assertLen(m, a, 4, "DELETE")
	assertLen(m, a, 4, "GET")

	// Prefix中的pattern可以为空
	a.NotPanic(func() { p.Add("", h, "GET") })

	// 添加相同的pattern
	a.Panic(func() { p.Any("h", h) })
	// handler不能为空
	a.Panic(func() { p.Add("abc", nil, "GET") })
	// 不支持的methods
	a.Panic(func() { p.Add("abc", h, "GET123") })

}
