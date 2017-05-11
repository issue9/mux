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

var options = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "options")
})

func TestItems_add_remove(t *testing.T) {
	a := assert.New(t)

	b := newBase("/")
	a.NotNil(b)

	a.NotError(b.add(get, http.MethodGet, http.MethodPost))
	a.Error(b.add(get, http.MethodPost)) // 存在相同的
	a.False(b.remove(http.MethodPost))
	a.True(b.remove(http.MethodGet))
	a.True(b.remove(http.MethodGet))

	// OPTIONS
	a.NotError(b.add(get, http.MethodOptions))
	a.Error(b.add(get, http.MethodOptions))
	a.NotError(b.remove(http.MethodOptions))
	a.NotError(b.add(get, http.MethodOptions))

	// 删除不存在的内容
	b.remove("not exists")
}

func TestItems_OptionsAllow(t *testing.T) {
	a := assert.New(t)

	b := newBase("/")
	a.NotNil(b)

	test := func(allow string) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("OPTIONS", "/empty", nil)
		b.Handler(http.MethodOptions).ServeHTTP(w, r)
		a.Equal(w.Header().Get("Allow"), allow)
	}

	// 默认
	a.Equal(b.optionsAllow, "OPTIONS")

	a.NotError(b.add(get, http.MethodGet))
	test("GET, OPTIONS")

	a.NotError(b.add(get, http.MethodPost))
	test("GET, OPTIONS, POST")

	// 显式调用 SetAllow() 之后，不再改变 optionsallow
	b.SetAllow("TEST,TEST1")
	test("TEST,TEST1")
	a.NotError(b.add(get, http.MethodDelete))
	test("TEST,TEST1")

	// 显式使用 http.MehtodOptions 之后，所有输出都从 options 函数来获取。
	a.NotError(b.add(options, http.MethodOptions))
	test("options")
	a.NotError(b.add(get, http.MethodPatch))
	test("options")
	b.SetAllow("set allow") // SetAllow 无法改变其值
	test("options")
	// 强制删除
	a.False(b.remove(http.MethodOptions))
	a.Nil(b.handlers[http.MethodOptions])
	b.SetAllow("set allow") // SetAllow 无法设置值
	a.Nil(b.handlers[http.MethodOptions])
	// 只能通过 add() 再次显示指定
	a.NotError(b.add(options, http.MethodOptions))
	test("options")
}
