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
