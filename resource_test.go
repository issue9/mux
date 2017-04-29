// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"testing"

	"github.com/issue9/assert"
)

func TestResource_Clean(t *testing.T) {
	a := assert.New(t)
	srvmux := NewServeMux(false, nil, nil)
	a.NotNil(srvmux)

	// 添加 delete /api/1
	a.NotPanic(func() {
		srvmux.DeleteFunc("/api/1", f1).
			PatchFunc("/api/1", f1)
	})
	a.Equal(srvmux.entries.Len(), 1)

	// 添加 patch /api/2 和 delete /api/2
	res := srvmux.Resource("/api/2")
	a.NotPanic(func() {
		res.PatchFunc(f1).
			Delete(h1)
	})
	a.Equal(srvmux.entries.Len(), 2)

	res.Clean()
	a.Equal(srvmux.entries.Len(), 1)
}

func TestServeMux_Resource(t *testing.T) {
	a := assert.New(t)
	srvmux := NewServeMux(false, nil, nil)
	a.NotNil(srvmux)

	res := srvmux.Resource("/abc/1")
	a.Equal(res.Mux(), srvmux)
	a.Equal(res.pattern, "/abc/1")
}
