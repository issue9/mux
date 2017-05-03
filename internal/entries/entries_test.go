// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entries

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

// 一些预定义的处理函数
var (
	f1 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
	}
	f2 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(2)
	}
	f3 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(3)
	}

	h1 = http.HandlerFunc(f1)
	h2 = http.HandlerFunc(f2)
	h3 = http.HandlerFunc(f3)
)

func TestEntries_Add_Remove_1(t *testing.T) {
	a := assert.New(t)
	es := New(false)
	a.NotNil(es)

	// 添加 delete /api/1
	a.NotError(es.Add("/api/1", h1, http.MethodDelete))
	a.Equal(es.entries.Len(), 1)

	// 添加 patch /api/1
	a.NotError(es.Add("/api/1", h1, http.MethodPatch))
	a.Equal(es.entries.Len(), 1) // 加在同一个 Entry 下，所以数量不变

	// 添加 post /api/2
	a.NotError(es.Add("/api/2", h1, http.MethodPost))
	a.Equal(es.entries.Len(), 2)

	// 删除 any /api/2
	es.Remove("/api/2")
	a.Equal(es.entries.Len(), 1)

	// 删除 delete /api/1
	es.Remove("/api/1", http.MethodDelete)
	a.Equal(es.entries.Len(), 1)

	// 删除 patch /api/1
	es.Remove("/api/1", http.MethodPatch)
	a.Equal(es.entries.Len(), 0)
}

func TestEntries_Clean(t *testing.T) {
	a := assert.New(t)
	es := New(false)
	a.NotNil(es)

	// 添加 delete /api/1
	a.NotError(es.Add("/api/1", h1, http.MethodDelete))
	a.NotError(es.Add("/api/1", h1, http.MethodPatch))
	a.Equal(es.entries.Len(), 1)

	es.Clean("")
	a.Equal(es.entries.Len(), 0)

	// 添加两条 entry
	a.NotError(es.Add("/api/1", h1, http.MethodDelete))
	a.NotError(es.Add("/api/1", h1, http.MethodPatch))
	a.NotError(es.Add("/api/2/1", h1, http.MethodPatch))
	a.NotError(es.Add("/api/2/1", h1, http.MethodDelete))
	a.Equal(es.entries.Len(), 2)
	es.Clean("/api/2") // 带路径参数的
	a.Equal(es.entries.Len(), 1)

	// 添加两条 entry
	a.NotError(es.Add("/api/2/1", h1, http.MethodDelete))
	a.NotError(es.Add("/1", h1, http.MethodDelete))
	es.Clean("/api/") // 带路径参数的
	a.Equal(es.entries.Len(), 1)
}
