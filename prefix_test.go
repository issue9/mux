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
	test := newTester(t, false, true, false)
	p := test.prefix("/p")

	p.Get("/h/1", buildHandler(201))
	test.matchTrue(http.MethodGet, "/p/h/1", 201)
	p.GetFunc("/f/1", buildFunc(201))
	test.matchTrue(http.MethodGet, "/p/f/1", 201)

	p.Post("/h/1", buildHandler(202))
	test.matchTrue(http.MethodPost, "/p/h/1", 202)
	p.PostFunc("/f/1", buildFunc(202))
	test.matchTrue(http.MethodPost, "/p/f/1", 202)

	p.Put("/h/1", buildHandler(203))
	test.matchTrue(http.MethodPut, "/p/h/1", 203)
	p.PutFunc("/f/1", buildFunc(203))
	test.matchTrue(http.MethodPut, "/p/f/1", 203)

	p.Patch("/h/1", buildHandler(204))
	test.matchTrue(http.MethodPatch, "/p/h/1", 204)
	p.PatchFunc("/f/1", buildFunc(204))
	test.matchTrue(http.MethodPatch, "/p/f/1", 204)

	p.Delete("/h/1", buildHandler(205))
	test.matchTrue(http.MethodDelete, "/p/h/1", 205)
	p.DeleteFunc("/f/1", buildFunc(205))
	test.matchTrue(http.MethodDelete, "/p/f/1", 205)

	// Any
	p.Any("/h/any", buildHandler(206))
	test.matchTrue(http.MethodGet, "/p/h/any", 206)
	test.matchTrue(http.MethodPost, "/p/h/any", 206)
	test.matchTrue(http.MethodPut, "/p/h/any", 206)
	test.matchTrue(http.MethodPatch, "/p/h/any", 206)
	test.matchTrue(http.MethodDelete, "/p/h/any", 206)
	test.matchTrue(http.MethodTrace, "/p/h/any", 206)

	p.AnyFunc("/f/any", buildFunc(206))
	test.matchTrue(http.MethodGet, "/p/f/any", 206)
	test.matchTrue(http.MethodPost, "/p/f/any", 206)
	test.matchTrue(http.MethodPut, "/p/f/any", 206)
	test.matchTrue(http.MethodPatch, "/p/f/any", 206)
	test.matchTrue(http.MethodDelete, "/p/f/any", 206)
	test.matchTrue(http.MethodTrace, "/p/f/any", 206)
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
