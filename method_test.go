// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestMethod_Add(t *testing.T) {
	a := assert.New(t)
	m := NewMethod()
	a.NotNil(m)

	// handler不能为空
	a.Error(m.Add("abc", nil, "GET"))

	fn := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})

	// methods为空
	a.Error(m.Add("abc", fn))

	a.NotError(m.Add("abc", fn, "GET", "POST"))
	_, found := m.entries["GET"]
	a.True(found)
	_, found = m.entries["POST"]
	a.True(found)
	_, found = m.entries["DELETE"]
	a.False(found)

	a.NotError(m.Get("def", fn))
	es, found := m.entries["GET"]
	a.True(found).Equal(2, len(es.list))
}

func TestMethod_ServeHTTP(t *testing.T) {
	a := assert.New(t)

	// 默认的handler，向response写入method1Handler或是错误信息。
	method1Handler := func(w http.ResponseWriter, req *http.Request) {
		ctx := GetContext(req)
		params, found := ctx.Get("params")
		a.True(found)
		if params != nil {
			_, ok := params.(map[string]string)
			a.True(ok)
			w.Write([]byte("method1Handler"))
		} else {
			w.Write([]byte("错误，未捕获端口信息"))
		}
	}

	m := NewMethod()
	a.NotNil(m)

	a.NotError(m.Add("/abc", http.HandlerFunc(method1Handler), "GET"))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		//t.Error(req.URL.Path)
	}))
	defer srv.Close()
	http.Get(srv.URL + "/abc")
}
