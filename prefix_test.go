// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"testing"

	"github.com/issue9/assert"
)

func TestPrefix_Clean1(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, nil, nil)
	a.NotNil(srvmux)

	// 添加 delete /api/1
	a.NotPanic(func() {
		srvmux.DeleteFunc("/api/1", f1).
			PatchFunc("/api/1", f1)
	})
	a.Equal(srvmux.entries.Len(), 1)

	// 添加 patch /api/2/1 和 delete /api/2/1
	prefix := srvmux.Prefix("/api/2")
	a.NotPanic(func() {
		prefix.PatchFunc("/1", f1).
			Delete("/1", h1)
	})
	a.Equal(srvmux.entries.Len(), 2)

	prefix.Clean()
	a.Equal(srvmux.entries.Len(), 1)
}

func TestPrefix_Clean2(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, nil, nil)
	a.NotNil(srvmux)

	p1 := srvmux.Prefix("/api")
	a.NotPanic(func() {
		p1.PatchFunc("/1", f1).
			Delete("/1", h1)
	})
	a.Equal(srvmux.entries.Len(), 1)

	p2 := srvmux.Prefix("/api")
	a.NotPanic(func() {
		p2.PatchFunc("/2", f1).
			Delete("/3", h1)
	})
	a.Equal(srvmux.entries.Len(), 3)

	p2.Clean()
	a.Equal(srvmux.entries.Len(), 0)
}

func TestMux_Prefix(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, nil, nil)
	a.NotNil(srvmux)

	p := srvmux.Prefix("/abc")
	a.Equal(p.prefix, "/abc")
	a.Equal(p.Mux(), srvmux)

	p = srvmux.Prefix("")
	a.Equal(p.prefix, "")
}

func TestPrefix_Prefix(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, nil, nil)
	a.NotNil(srvmux)

	p := srvmux.Prefix("/abc")
	pp := p.Prefix("/def")
	a.Equal(pp.prefix, "/abc/def")
	a.Equal(p.Mux(), srvmux)

	p = srvmux.Prefix("")
	pp = p.Prefix("/abc")
	a.Equal(pp.prefix, "/abc")
}
