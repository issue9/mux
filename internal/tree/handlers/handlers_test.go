// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux/internal/method"
)

var getHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
})

var optionsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "options")
})

func TestHandlers_Add(t *testing.T) {
	a := assert.New(t)

	hs := New()
	a.NotNil(hs)

	a.NotError(hs.Add(getHandler))
	a.Equal(hs.Len(), 1) // 还有一个自动生成的 OPTIONS

	a.NotError(hs.Add(getHandler, http.MethodGet, http.MethodPut))
	a.Equal(hs.Len(), 3)

	a.Error(hs.Add(getHandler, "Not Exists"))
}

func TestHandlers_Add_Remove(t *testing.T) {
	a := assert.New(t)

	hs := New()
	a.NotNil(hs)

	a.NotError(hs.Add(getHandler, http.MethodGet, http.MethodPost))
	a.Error(hs.Add(getHandler, http.MethodPost)) // 存在相同的
	a.False(hs.Remove(http.MethodPost))
	a.True(hs.Remove(http.MethodGet))
	a.True(hs.Remove(http.MethodGet))

	// OPTIONS
	a.NotError(hs.Add(getHandler, http.MethodOptions))
	a.Error(hs.Add(getHandler, http.MethodOptions))
	a.NotError(hs.Remove(http.MethodOptions))
	a.NotError(hs.Add(getHandler, http.MethodOptions))

	// 删除不存在的内容
	hs.Remove("not exists")
}

func TestHandlers_optionsAllow(t *testing.T) {
	a := assert.New(t)

	hs := New()
	a.NotNil(hs)

	test := func(allow string) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("OPTIONS", "/empty", nil)
		hs.Handler(http.MethodOptions).ServeHTTP(w, r)
		a.Equal(w.Header().Get("Allow"), allow)
	}

	// 默认
	a.Equal(hs.optionsAllow, "OPTIONS")

	a.NotError(hs.Add(getHandler, http.MethodGet))
	test("GET, OPTIONS")

	a.NotError(hs.Add(getHandler, http.MethodPost))
	test("GET, OPTIONS, POST")

	// 显式调用 SetAllow() 之后，不再改变 optionsallow
	hs.SetAllow("TEST,TEST1")
	test("TEST,TEST1")
	a.NotError(hs.Add(getHandler, http.MethodDelete))
	test("TEST,TEST1")

	// 显式使用 http.MehtodOptions 之后，所有输出都从 options 函数来获取。
	a.NotError(hs.Add(optionsHandler, http.MethodOptions))
	test("options")
	a.NotError(hs.Add(getHandler, http.MethodPatch))
	test("options")
	hs.SetAllow("set allow") // SetAllow 无法改变其值
	test("options")
	// 强制删除
	a.False(hs.Remove(http.MethodOptions))
	a.Nil(hs.handlers[method.Options])
	hs.SetAllow("set allow") // SetAllow 无法设置值
	a.Nil(hs.handlers[method.Options])
	// 只能通过 Add() 再次显示指定
	a.NotError(hs.Add(optionsHandler, http.MethodOptions))
	test("options")
}
