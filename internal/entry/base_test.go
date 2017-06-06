// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux/internal/syntax"
)

var get = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
})

var options = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "options")
})

func TestBase_Add_Remove(t *testing.T) {
	a := assert.New(t)

	b := newBase(&syntax.Syntax{Pattern: "/", Wildcard: false})
	a.NotNil(b)

	a.NotError(b.Add(get, http.MethodGet, http.MethodPost))
	a.Error(b.Add(get, http.MethodPost)) // 存在相同的
	a.False(b.Remove(http.MethodPost))
	a.True(b.Remove(http.MethodGet))
	a.True(b.Remove(http.MethodGet))

	// OPTIONS
	a.NotError(b.Add(get, http.MethodOptions))
	a.Error(b.Add(get, http.MethodOptions))
	a.NotError(b.Remove(http.MethodOptions))
	a.NotError(b.Add(get, http.MethodOptions))

	// 删除不存在的内容
	b.Remove("not exists")
}

func TestBase_OptionsAllow(t *testing.T) {
	a := assert.New(t)

	b := newBase(&syntax.Syntax{Pattern: "/", Wildcard: false})
	a.NotNil(b)

	test := func(allow string) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("OPTIONS", "/empty", nil)
		b.Handler(http.MethodOptions).ServeHTTP(w, r)
		a.Equal(w.Header().Get("Allow"), allow)
	}

	// 默认
	a.Equal(b.optionsAllow, "OPTIONS")

	a.NotError(b.Add(get, http.MethodGet))
	test("GET, OPTIONS")

	a.NotError(b.Add(get, http.MethodPost))
	test("GET, OPTIONS, POST")

	// 显式调用 SetAllow() 之后，不再改变 optionsallow
	b.SetAllow("TEST,TEST1")
	test("TEST,TEST1")
	a.NotError(b.Add(get, http.MethodDelete))
	test("TEST,TEST1")

	// 显式使用 http.MehtodOptions 之后，所有输出都从 options 函数来获取。
	a.NotError(b.Add(options, http.MethodOptions))
	test("options")
	a.NotError(b.Add(get, http.MethodPatch))
	test("options")
	b.SetAllow("set allow") // SetAllow 无法改变其值
	test("options")
	// 强制删除
	a.False(b.Remove(http.MethodOptions))
	a.Nil(b.handlers[http.MethodOptions])
	b.SetAllow("set allow") // SetAllow 无法设置值
	a.Nil(b.handlers[http.MethodOptions])
	// 只能通过 add() 再次显示指定
	a.NotError(b.Add(options, http.MethodOptions))
	test("options")
}
