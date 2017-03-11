// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

var get = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
})

var post = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(201)
})

var options = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "options")
})

var put = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(203)
})

func TestItems_Add_Remove(t *testing.T) {
	a := assert.New(t)

	i := newItems()
	a.NotNil(i)

	a.NotError(i.Add(http.MethodGet, get))
	a.NotError(i.Add(http.MethodPost, get))
	a.Error(i.Add(http.MethodPost, get)) // 存在相同的
	a.False(i.Remove(http.MethodPost))
	a.True(i.Remove(http.MethodGet))
	a.True(i.Remove(http.MethodGet))

	// OPTIONS
	a.NotError(i.Add(http.MethodOptions, get))
	a.Error(i.Add(http.MethodOptions, get))
	i.Remove(http.MethodOptions)
	a.NotError(i.Add(http.MethodOptions, get))
}

func TestItems_OptionsAllow(t *testing.T) {
	a := assert.New(t)

	i := newItems()
	a.NotNil(i)

	test := func(allow string) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("OPTIONS", "/empty", nil)
		i.Handler(http.MethodOptions).ServeHTTP(w, r)
		a.Equal(w.Header().Get("Allow"), allow)
	}

	// 默认
	a.Equal(i.optionsAllow, "OPTIONS")

	a.NotError(i.Add(http.MethodGet, get))
	test("GET, OPTIONS")

	a.NotError(i.Add(http.MethodPost, get))
	test("GET, OPTIONS, POST")

	// 显式调用 SetAllow() 之后，不再改变 optionsallow
	i.SetAllow("TEST,TEST1")
	test("TEST,TEST1")
	a.NotError(i.Add(http.MethodDelete, get))
	test("TEST,TEST1")

	// 显式使用 http.MehtodOptions 之后，所有输出都以 options 为主。
	a.NotError(i.Add(http.MethodOptions, options))
	test("options")
	a.NotError(i.Add(http.MethodPatch, get))
	test("options")
}
