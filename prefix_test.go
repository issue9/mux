// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func (t *tester) prefix(p string) *Prefix {
	return t.mux.Prefix(p)
}

func TestPrefix(t *testing.T) {
	a := assert.New(t)
	test := newTester(a, false, true, false)
	p := test.prefix("/p")

	p.Get("/h/1", buildHandler(1))
	test.matchTrue(http.MethodGet, "/p/h/1", 1)
	p.GetFunc("/f/1", buildFunc(1))
	test.matchTrue(http.MethodGet, "/p/f/1", 1)

	p.Post("/h/1", buildHandler(2))
	test.matchTrue(http.MethodPost, "/p/h/1", 2)
	p.PostFunc("/f/1", buildFunc(2))
	test.matchTrue(http.MethodPost, "/p/f/1", 2)

	p.Put("/h/1", buildHandler(3))
	test.matchTrue(http.MethodPut, "/p/h/1", 3)
	p.PutFunc("/f/1", buildFunc(3))
	test.matchTrue(http.MethodPut, "/p/f/1", 3)

	p.Patch("/h/1", buildHandler(4))
	test.matchTrue(http.MethodPatch, "/p/h/1", 4)
	p.PatchFunc("/f/1", buildFunc(4))
	test.matchTrue(http.MethodPatch, "/p/f/1", 4)

	p.Delete("/h/1", buildHandler(5))
	test.matchTrue(http.MethodDelete, "/p/h/1", 5)
	p.DeleteFunc("/f/1", buildFunc(5))
	test.matchTrue(http.MethodDelete, "/p/f/1", 5)

	// Any
	p.Any("/h/any", buildHandler(6))
	test.matchTrue(http.MethodGet, "/p/h/any", 6)
	test.matchTrue(http.MethodPost, "/p/h/any", 6)
	test.matchTrue(http.MethodPut, "/p/h/any", 6)
	test.matchTrue(http.MethodPatch, "/p/h/any", 6)
	test.matchTrue(http.MethodDelete, "/p/h/any", 6)
	test.matchTrue(http.MethodTrace, "/p/h/any", 6)

	p.AnyFunc("/f/any", buildFunc(6))
	test.matchTrue(http.MethodGet, "/p/f/any", 6)
	test.matchTrue(http.MethodPost, "/p/f/any", 6)
	test.matchTrue(http.MethodPut, "/p/f/any", 6)
	test.matchTrue(http.MethodPatch, "/p/f/any", 6)
	test.matchTrue(http.MethodDelete, "/p/f/any", 6)
	test.matchTrue(http.MethodTrace, "/p/f/any", 6)
}

func TestMux_Prefix(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, true, false, nil, nil)
	a.NotNil(srvmux)

	p := srvmux.Prefix("/abc")
	a.Equal(p.prefix, "/abc")
	a.Equal(p.Mux(), srvmux)

	p = srvmux.Prefix("")
	a.Equal(p.prefix, "")
}

func TestPrefix_Prefix(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, true, false, nil, nil)
	a.NotNil(srvmux)

	p := srvmux.Prefix("/abc")
	pp := p.Prefix("/def")
	a.Equal(pp.prefix, "/abc/def")
	a.Equal(p.Mux(), srvmux)

	p = srvmux.Prefix("")
	pp = p.Prefix("/abc")
	a.Equal(pp.prefix, "/abc")
}
