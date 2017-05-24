// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

var get = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
})

var options = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "options")
})

func TestHandlers_add_remove(t *testing.T) {
	a := assert.New(t)

	hs := newHandlers()
	a.NotNil(hs)

	a.NotError(hs.add(get, http.MethodGet, http.MethodPost))
	a.Error(hs.add(get, http.MethodPost)) // 存在相同的
	a.False(hs.remove(http.MethodPost))
	a.True(hs.remove(http.MethodGet))
	a.True(hs.remove(http.MethodGet))

	// OPTIONS
	a.NotError(hs.add(get, http.MethodOptions))
	a.Error(hs.add(get, http.MethodOptions))
	a.NotError(hs.remove(http.MethodOptions))
	a.NotError(hs.add(get, http.MethodOptions))

	// 删除不存在的内容
	hs.remove("not exists")
}

func TestHandlers_optionsAllow(t *testing.T) {
	a := assert.New(t)

	hs := newHandlers()
	a.NotNil(hs)

	test := func(allow string) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("OPTIONS", "/empty", nil)
		hs.handler(http.MethodOptions).ServeHTTP(w, r)
		a.Equal(w.Header().Get("Allow"), allow)
	}

	// 默认
	a.Equal(hs.optionsAllow, "OPTIONS")

	a.NotError(hs.add(get, http.MethodGet))
	test("GET, OPTIONS")

	a.NotError(hs.add(get, http.MethodPost))
	test("GET, OPTIONS, POST")

	// 显式调用 SetAllow() 之后，不再改变 optionsallow
	hs.setAllow("TEST,TEST1")
	test("TEST,TEST1")
	a.NotError(hs.add(get, http.MethodDelete))
	test("TEST,TEST1")

	// 显式使用 http.MehtodOptions 之后，所有输出都从 options 函数来获取。
	a.NotError(hs.add(options, http.MethodOptions))
	test("options")
	a.NotError(hs.add(get, http.MethodPatch))
	test("options")
	hs.setAllow("set allow") // SetAllow 无法改变其值
	test("options")
	// 强制删除
	a.False(hs.remove(http.MethodOptions))
	a.Nil(hs.handlers[http.MethodOptions])
	hs.setAllow("set allow") // SetAllow 无法设置值
	a.Nil(hs.handlers[http.MethodOptions])
	// 只能通过 add() 再次显示指定
	a.NotError(hs.add(options, http.MethodOptions))
	test("options")
}
