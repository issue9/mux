// SPDX-License-Identifier: MIT

package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

var getHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	if _, err := w.Write([]byte("hello")); err != nil {
		panic(err)
	}
})

var optionsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "options")
})

func TestNew(t *testing.T) {
	a := assert.New(t)

	hs := New(true, true)
	a.NotNil(hs)
	a.NotError(hs.Add(getHandler, http.MethodGet))
	a.Equal(hs.Len(), 1) // 不包含自动生成的 OPTIONS

	hs.SetAllow("123")
	a.Equal(hs.Len(), 2). // 有 OPTIONS
				NotNil(hs.Handler(http.MethodGet)).
				NotNil(hs.Handler(http.MethodOptions))
}

func TestHandlers_Add(t *testing.T) {
	a := assert.New(t)

	hs := New(false, true)
	a.NotNil(hs)
	a.NotError(hs.Add(getHandler))
	a.Equal(hs.Len(), len(addAny)+1) // 包含自动生成的 OPTIONS

	hs = New(false, true)
	a.NotNil(hs)
	a.NotError(hs.Add(getHandler, http.MethodGet, http.MethodPut))
	a.Equal(hs.Len(), 3) // 包含自动生成的 OPTIONS
	a.Error(hs.Add(getHandler, "Not Exists"))

	// head
	hs = New(false, false)
	a.NotNil(hs)
	a.NotError(hs.Add(getHandler, http.MethodGet, http.MethodPut))
	a.Equal(hs.Len(), 4) // 包含自动生成的 OPTIONS 和 HEAD
	a.False(hs.disableHead)
	// 特意指定 head
	a.ErrorString(hs.Add(getHandler, http.MethodHead), "无法手动添加 head 请求方法")
	a.Equal(hs.Len(), 4)                         // 不会变多
	a.Error(hs.Add(getHandler, http.MethodHead)) // 多次添加
}

func TestHandlers_Add_Remove(t *testing.T) {
	a := assert.New(t)

	hs := New(false, false)
	a.NotNil(hs)

	a.NotError(hs.Add(getHandler, http.MethodDelete, http.MethodPost))
	a.Error(hs.Add(getHandler, http.MethodPost)) // 存在相同的
	a.False(hs.Remove(http.MethodPost))
	a.True(hs.Remove(http.MethodDelete))
	a.True(hs.Remove(http.MethodDelete))

	// OPTIONS
	a.NotError(hs.Add(getHandler, http.MethodOptions))
	a.Error(hs.Add(getHandler, http.MethodOptions))
	a.NotError(hs.Remove(http.MethodOptions))
	a.NotError(hs.Add(getHandler, http.MethodOptions))

	// HEAD 和 GET 一起删除
	a.NotError(hs.Add(getHandler, http.MethodGet))
	a.NotNil(hs.handlers[http.MethodHead]) // 自动添加 HEAD
	a.False(hs.Remove(http.MethodGet))
	a.Nil(hs.handlers[http.MethodHead]) // 自动删除 HEAD

	// 先删除 HEAD，再删除 GET
	a.NotError(hs.Add(getHandler, http.MethodGet))
	a.NotNil(hs.handlers[http.MethodHead]) // 自动添加 HEAD
	a.False(hs.Remove(http.MethodHead))
	a.Nil(hs.handlers[http.MethodHead])   // 自动删除 HEAD
	a.NotNil(hs.handlers[http.MethodGet]) // Get 还在

	// 删除不存在的内容
	a.False(hs.Remove("not exists"))

	a.True(hs.Remove())

	// 自动生成的 HEAD 和 OPTIONS，在 remove() 是会自动删除
	a.NotError(hs.Add(getHandler, http.MethodGet))
	a.NotError(hs.Add(getHandler, http.MethodPost))
	a.True(hs.Remove())
}

func TestHandlers_optionsAllow(t *testing.T) {
	a := assert.New(t)

	hs := New(false, true)
	a.NotNil(hs)

	test := func(allow string) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("OPTIONS", "/empty", nil)
		h := hs.Handler(http.MethodOptions)
		a.NotNil(h)
		h.ServeHTTP(w, r)
		a.Equal(w.Header().Get("Allow"), allow)
	}

	// 默认
	a.Equal(hs.Options(), "OPTIONS")

	a.NotError(hs.Add(getHandler, http.MethodGet))
	test("GET, OPTIONS")
	a.Equal(hs.Options(), "GET, OPTIONS")

	a.NotError(hs.Add(getHandler, http.MethodPost))
	test("GET, OPTIONS, POST")

	// 显式调用 SetAllow() 之后，不再改变 optionsAllow
	hs.SetAllow("TEST,TEST1")
	test("TEST,TEST1")
	a.NotError(hs.Add(getHandler, http.MethodDelete))
	test("TEST,TEST1")

	// 显式使用 http.MethodOptions 之后，所有输出都从 options 函数来获取。
	a.NotError(hs.Add(optionsHandler, http.MethodOptions))
	test("options")
	a.NotError(hs.Add(getHandler, http.MethodPatch))
	test("options")
	hs.SetAllow("set allow") // SetAllow 无法改变其值
	test("options")
	// 强制删除
	a.False(hs.Remove(http.MethodOptions))
	a.Nil(hs.handlers[http.MethodOptions])
	hs.SetAllow("set allow") // SetAllow 无法设置值
	a.NotNil(hs.handlers[http.MethodOptions])
	a.NotError(hs.Add(optionsHandler, http.MethodOptions)) // 通过 Add() 再次显示指定
	test("options")
}

func TestHandlers_Methods(t *testing.T) {
	a := assert.New(t)

	hs := New(false, false)
	a.NotNil(hs)
	a.NotError(hs.Add(getHandler, http.MethodPut))
	a.Equal(hs.Methods(true, true), []string{http.MethodPut})

	hs = New(false, false)
	a.NotNil(hs)
	a.NotError(hs.Add(getHandler, http.MethodPut, http.MethodTrace))
	a.Equal(hs.Methods(true, true), []string{http.MethodPut, http.MethodTrace})

	hs = New(false, false)
	a.NotNil(hs)
	a.NotError(hs.Add(getHandler, http.MethodPut, http.MethodTrace, http.MethodGet))
	a.Equal(hs.Methods(true, true), []string{http.MethodGet, http.MethodPut, http.MethodTrace})

	hs = New(false, false)
	a.NotNil(hs)
	a.NotError(hs.Add(getHandler, http.MethodPut, http.MethodTrace, http.MethodGet))
	a.Equal(hs.Methods(false, true), []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodTrace})

	hs = New(false, false)
	a.NotNil(hs)
	a.NotError(hs.Add(getHandler, http.MethodPut, http.MethodGet))
	a.Equal(hs.Methods(false, false), []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut})

	hs = New(true, true)
	a.NotNil(hs)
	a.NotError(hs.Add(getHandler, http.MethodPut, http.MethodGet))
	a.Equal(hs.Methods(false, false), []string{http.MethodGet, http.MethodPut})

	// 动态插入删除操作

	hs = New(false, false)
	a.NotNil(hs)
	a.NotError(hs.Add(getHandler, http.MethodPut, http.MethodGet))
	a.Equal(hs.Methods(false, false), []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut})
	a.Equal(hs.Methods(true, true), []string{http.MethodGet, http.MethodPut})

	// 删除 GET
	hs.Remove(http.MethodGet)
	a.Equal(hs.Methods(false, false), []string{http.MethodOptions, http.MethodPut})
	a.Equal(hs.Methods(true, true), []string{http.MethodPut})
	a.NotError(hs.Add(getHandler, http.MethodGet))

	// 强制指定 OPTIONS，不影响 methods() 的返回
	hs.SetAllow("xx")
	a.Equal(hs.Methods(false, false), []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut})
	a.Equal(hs.Methods(true, true), []string{http.MethodGet, http.MethodOptions, http.MethodPut})

	// 删除除 OPTIONS 之外的其它请求方法
	hs.Remove(http.MethodGet, http.MethodPut)
	a.Equal(hs.Methods(false, false), []string{http.MethodOptions})
	a.Equal(hs.Methods(true, true), []string{http.MethodOptions})

	// 删除所有
	hs.Remove()
	a.Empty(hs.Methods(false, false))
	a.Empty(hs.Methods(true, true))
}
