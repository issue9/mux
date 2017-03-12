// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"testing"

	"github.com/issue9/assert"
)

func TestPrefix_Clean(t *testing.T) {
	a := assert.New(t)
	srvmux := NewServeMux(false)
	a.NotNil(srvmux)

	// 添加 delete /api/1
	a.NotPanic(func() {
		srvmux.DeleteFunc("/api/1", f1).
			PatchFunc("/api/1", f1)
	})
	a.Equal(srvmux.entries.Len(), 1)

	// 添加 patch /api/1 和 delete /api/1
	prefix := srvmux.Prefix("/api/2")
	a.NotPanic(func() {
		prefix.PatchFunc("/1", f1).
			Delete("/1", h1)
	})
	a.Equal(srvmux.entries.Len(), 2)

	prefix.Clean()
	a.Equal(srvmux.entries.Len(), 1)
}

// https://github.com/issue9/mux/issues/4
func TestIssue4(t *testing.T) {
	a := assert.New(t)
	srvmux := NewServeMux(false)
	a.NotNil(srvmux)

	// 添加 delete /api1
	a.NotPanic(func() {
		srvmux.DeleteFunc("/api1", f1)
	})
	a.Equal(srvmux.entries.Len(), 1)

	// 添加 patch /api/1 和 delete /api/1
	prefix := srvmux.Prefix("/api")
	a.NotPanic(func() {
		prefix.PatchFunc("/1", f1).
			Delete("/1", h1)
	})
	a.Equal(srvmux.entries.Len(), 2)

	prefix.Clean()
	a.Equal(srvmux.entries.Len(), 1)
}
