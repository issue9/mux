// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package list

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux/internal/entry"
	"github.com/issue9/mux/internal/method"
	"github.com/issue9/mux/internal/syntax"
)

var (
	f1 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
	}
	h1 = http.HandlerFunc(f1)
)

func TestEntries_add_remove(t *testing.T) {
	a := assert.New(t)
	es := newEntries(false)
	a.NotNil(es)

	// 添加 delete /api/1
	a.NotError(es.add("/api/1", h1, http.MethodDelete))
	a.Equal(len(es.entries), 1)

	// 添加 patch /api/1
	a.NotError(es.add("/api/1", h1, http.MethodPatch))
	a.Equal(len(es.entries), 1) // 加在同一个 Entry 下，所以数量不变

	// 添加 post /api/2
	a.NotError(es.add("/api/2", h1, http.MethodPost))
	a.Equal(len(es.entries), 2)

	// 删除 any /api/2
	es.remove("/api/2", method.Supported...)
	a.Equal(len(es.entries), 1)

	// 删除 delete /api/1
	es.remove("/api/1", http.MethodDelete)
	a.Equal(len(es.entries), 1)

	// 删除 patch /api/1
	es.remove("/api/1", http.MethodPatch)
	a.Equal(len(es.entries), 0)
}

func TestEntries_clean(t *testing.T) {
	a := assert.New(t)
	es := newEntries(false)
	a.NotNil(es)

	// 添加 delete /api/1
	a.NotError(es.add("/api/1", h1, http.MethodDelete))
	a.NotError(es.add("/api/1", h1, http.MethodPatch))
	a.Equal(len(es.entries), 1)

	es.clean("")
	a.Equal(len(es.entries), 0)

	// 添加两条 entry
	a.NotError(es.add("/api/1", h1, http.MethodDelete))
	a.NotError(es.add("/api/1", h1, http.MethodPatch))
	a.NotError(es.add("/api/2/1", h1, http.MethodPatch))
	a.NotError(es.add("/api/2/1", h1, http.MethodDelete))
	a.Equal(len(es.entries), 2)
	es.clean("/api/2") // 带路径参数的
	a.Equal(len(es.entries), 1)

	// 添加两条 entry
	a.NotError(es.add("/api/2/1", h1, http.MethodDelete))
	a.NotError(es.add("/1", h1, http.MethodDelete))
	es.clean("/api/") // 带路径参数的
	a.Equal(len(es.entries), 1)
}

func TestRemoveEntries(t *testing.T) {
	a := assert.New(t)

	newSyntax := func(pattern string) *syntax.Syntax {
		s, err := syntax.New(pattern)
		a.NotError(err).NotNil(s)

		return s
	}

	n1, err := entry.New(newSyntax("/1"))
	n2, err := entry.New(newSyntax("/2"))
	n3, err := entry.New(newSyntax("/3"))
	n4, err := entry.New(newSyntax("/4"))
	a.NotError(err)
	es := []entry.Entry{n1, n2, n3, n4}

	// 不存在的元素
	es = removeEntries(es, "")
	a.Equal(len(es), 4)

	// 删除尾元素
	es = removeEntries(es, "/4")
	a.Equal(len(es), 3)

	// 删除中间无素
	es = removeEntries(es, "/2")
	a.Equal(len(es), 2)

	// 已删除，不存在的元素
	es = removeEntries(es, "/2")
	a.Equal(len(es), 2)

	// 第一个元素
	es = removeEntries(es, "/1")
	a.Equal(len(es), 1)

	// 最后一个元素
	es = removeEntries(es, "/3")
	a.Equal(len(es), 0)
}
