// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestResource(t *testing.T) {
	a := assert.New(t)

	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewServeMux()
	r := mux.Resource("p1")
	a.Equal(r.pattern, "p1").Equal(r.mux, mux)

	r.Post(hf)
	r.Get(hf)
	r.Delete(hf)

	assertLen(mux, a, 1, "GET")
	assertLen(mux, a, 1, "POST")
	assertLen(mux, a, 1, "DELETE")
}

func TestResource_Clean(t *testing.T) {
	a := assert.New(t)
	m := NewServeMux()
	a.NotNil(m)
	r := m.Resource("/123")
	a.NotNil(r)
	fn := func(w http.ResponseWriter, req *http.Request) {}
	h := http.HandlerFunc(fn)

	r.Add(h)
	assertLen(m, a, 1, "GET")
	r.Clean()
	assertLen(m, a, 0, "GET")
	assertLen(m, a, 0, "DELETE")
}

func TestResource_Add(t *testing.T) {
	a := assert.New(t)
	m := NewServeMux()
	a.NotNil(m)

	r := m.Resource("resource")
	a.NotNil(r)

	fn := func(w http.ResponseWriter, req *http.Request) {}
	h := http.HandlerFunc(fn)

	a.NotPanic(func() { r.Get(h) })
	assertLen(m, a, 1, "GET")

	a.NotPanic(func() { r.Post(h) })
	assertLen(m, a, 1, "POST")

	a.NotPanic(func() { r.Put(h) })
	assertLen(m, a, 1, "PUT")

	a.NotPanic(func() { r.Delete(h) })
	assertLen(m, a, 1, "DELETE")

	a.NotPanic(func() { r.Patch(h) })
	assertLen(m, a, 1, "PATCH")

	r = m.Resource("resource1")
	a.NotNil(r)
	a.NotPanic(func() { r.GetFunc(fn) })
	assertLen(m, a, 2, "GET")

	a.NotPanic(func() { r.PostFunc(fn) })
	assertLen(m, a, 2, "POST")

	a.NotPanic(func() { r.PutFunc(fn) })
	assertLen(m, a, 2, "PUT")

	a.NotPanic(func() { r.DeleteFunc(fn) })
	assertLen(m, a, 2, "DELETE")

	a.NotPanic(func() { r.PatchFunc(fn) })
	assertLen(m, a, 2, "PATCH")
}
