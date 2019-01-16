// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package host

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func buildHandler(code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	})
}

func buildFunc(code int) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	})
}

func TestNew(t *testing.T) {
	a := assert.New(t)

	h := New(false, true, false)
	a.NotNil(h)
	a.False(h.disableOptions).
		False(h.skipCleanPath).
		True(h.disableHead).
		NotNil(h.tree)
}

func TestHosts_Add_URL_ClearAll(t *testing.T) {
	a := assert.New(t)
	hs := New(false, false, true)
	a.NotNil(hs)

	// 只有域名部分
	a.Panic(func() {
		hs.Add("*.example.com", buildHandler(1), http.MethodGet)
	})

	a.Panic(func() {
		hs.Add("", buildHandler(1), http.MethodGet)
	})

	hs.Add("/path", buildHandler(1), http.MethodGet)
	hs.Add("*.example.com/path", buildHandler(1), http.MethodGet)
	hs.Add("*.caixw.io/path", buildHandler(1), http.MethodGet)
	hs.Add("s1.example.com/path", buildHandler(1), http.MethodGet)
	hs.Add("s1.caixw.io/path", buildHandler(1), http.MethodGet)
	a.Equal(4, len(hs.hosts)).
		Equal(hs.hosts[0].domain, "s1.caixw.io"). // 顺序永远是泛域名在最后
		Equal(hs.hosts[1].domain, "s1.example.com").
		Equal(hs.hosts[2].domain, ".caixw.io")

	// URL
	url, err := hs.URL("/path", nil)
	a.NotError(err).Equal(url, "/path")

	url, err = hs.URL("*.example.com/path", nil)
	a.NotError(err).Equal(url, "*.example.com/path")

	url, err = hs.URL("not-exists.example.com/path", nil)
	a.Error(err).Empty(url)

	// 不能为空
	a.Panic(func() {
		url, err = hs.URL("", nil)
	})

	// CleanAll
	hs.CleanAll()
	a.Equal(0, len(hs.hosts))
}

func TestHosts_getTree_findTree(t *testing.T) {
	a := assert.New(t)
	hs := New(false, false, true)
	a.NotNil(hs)

	t1 := hs.getTree("/path")
	t2 := hs.getTree("*.example.com/path")
	t3 := hs.getTree("*.caixw.io/path")
	t4 := hs.getTree("s1.example.com/path")

	a.Equal(hs.getTree("/path"), t1)
	a.Equal(hs.getTree("*.example.com/path"), t2)
	a.Equal(hs.getTree("*.caixw.io/path"), t3)
	a.Equal(hs.getTree("s1.example.com/path"), t4)

	a.Equal(hs.findTree("/path"), t1)
	a.Equal(hs.findTree("*.example.com/path"), t2)
	a.Equal(hs.findTree("*.caixw.io/path"), t3)
	a.Equal(hs.findTree("s1.example.com/path"), t4)
	a.Equal(hs.findTree("/not-exists"), t1)
	a.Nil(hs.findTree("notexists.example.com/not-exists"))

	a.Panic(func() {
		hs.findTree("*.example.com")
	})
}

func TestNewHosts(t *testing.T) {
	a := assert.New(t)

	h := newHost("example.com", nil)
	a.NotNil(h).
		Equal(h.raw, "example.com").
		Equal(h.domain, "example.com").
		False(h.wildcard)

	h = newHost("*.example.com", nil)
	a.NotNil(h).
		Equal(h.raw, "*.example.com").
		Equal(h.domain, ".example.com").
		True(h.wildcard)
}
