// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"testing"

	"github.com/issue9/assert"
)

func TestMux_Prefix(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, false, nil, nil)
	a.NotNil(srvmux)

	p := srvmux.Prefix("/abc")
	a.Equal(p.prefix, "/abc")
	a.Equal(p.Mux(), srvmux)

	p = srvmux.Prefix("")
	a.Equal(p.prefix, "")
}

func TestPrefix_Prefix(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, false, nil, nil)
	a.NotNil(srvmux)

	p := srvmux.Prefix("/abc")
	pp := p.Prefix("/def")
	a.Equal(pp.prefix, "/abc/def")
	a.Equal(p.Mux(), srvmux)

	p = srvmux.Prefix("")
	pp = p.Prefix("/abc")
	a.Equal(pp.prefix, "/abc")
}
