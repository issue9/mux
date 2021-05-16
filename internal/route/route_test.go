// SPDX-License-Identifier: MIT

package route

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

func TestRoute_Add(t *testing.T) {
	a := assert.New(t)

	hs := New(true)
	a.NotNil(hs)
	a.NotError(hs.Add(getHandler))
	a.Equal(hs.Len(), len(addAny)+1) // 包含自动生成的 OPTIONS

	hs = New(true)
	a.NotNil(hs)
	a.NotError(hs.Add(getHandler, http.MethodGet, http.MethodPut))
	a.Equal(hs.Len(), 3) // 包含自动生成的 OPTIONS
	a.Error(hs.Add(getHandler, "Not Exists"))

	// head

	hs = New(false)
	a.NotNil(hs)
	a.NotError(hs.Add(getHandler, http.MethodGet, http.MethodPut))
	a.Equal(hs.Len(), 4) // 包含自动生成的 OPTIONS 和 HEAD
	a.False(hs.disableHead)

	// 特意指定 head
	a.ErrorString(hs.Add(getHandler, http.MethodHead), "无法手动添加 OPTIONS/HEAD 请求方法")
	a.Equal(hs.Len(), 4)                         // 不会变多
	a.Error(hs.Add(getHandler, http.MethodHead)) // 多次添加

	// options

	hs = New(false)
	a.NotNil(hs)
	a.NotError(hs.Add(getHandler, http.MethodGet, http.MethodPut))
	a.ErrorString(hs.Add(getHandler, http.MethodOptions), "无法手动添加 OPTIONS/HEAD 请求方法")
}

func TestRoute_Add_Remove(t *testing.T) {
	a := assert.New(t)

	r := New(false)
	a.NotNil(r)
	a.Empty(r.Methods()).Empty(r.Options())

	a.NotError(r.Add(getHandler, http.MethodDelete, http.MethodPost))
	a.Equal(r.Methods(), []string{http.MethodDelete, http.MethodOptions, http.MethodPost}).
		Equal(r.Options(), "DELETE, OPTIONS, POST")
	a.Error(r.Add(getHandler, http.MethodPost)) // 存在相同的

	empty, err := r.Remove(http.MethodPost)
	a.NotError(err).False(empty)
	a.Equal(r.Methods(), []string{http.MethodDelete, http.MethodOptions}).
		Equal(r.Options(), "DELETE, OPTIONS")

	empty, err = r.Remove(http.MethodDelete)
	a.NotError(err).True(empty)
	a.Empty(r.Methods()).Empty(r.Options())

	empty, err = r.Remove(http.MethodDelete)
	a.NotError(err).True(empty)
	a.Empty(r.Methods()).Empty(r.Options())

	// Get

	a.NotError(r.Add(getHandler, http.MethodGet))
	a.NotNil(r.handlers[http.MethodHead]) // 自动添加 HEAD
	a.Equal(r.Methods(), []string{http.MethodGet, http.MethodHead, http.MethodOptions}).
		Equal(r.Options(), "GET, HEAD, OPTIONS")

	empty, err = r.Remove(http.MethodGet)
	a.NotError(err).True(empty)
	a.Empty(r.Methods()).Empty(r.Options())

	// 先删除 HEAD，再删除 GET
	a.NotError(r.Add(getHandler, http.MethodGet))
	a.NotNil(r.handlers[http.MethodHead]) // 自动添加 HEAD
	a.Equal(r.Methods(), []string{http.MethodGet, http.MethodHead, http.MethodOptions}).
		Equal(r.Options(), "GET, HEAD, OPTIONS")

	empty, err = r.Remove(http.MethodHead)
	a.NotError(err).False(empty).Nil(r.handlers[http.MethodHead])
	a.Equal(r.Methods(), []string{http.MethodGet, http.MethodOptions}).
		Equal(r.Options(), "GET, OPTIONS")
	a.NotNil(r.handlers[http.MethodGet]) // Get 还在

	// 删除不存在的内容
	empty, err = r.Remove("not exists")
	a.False(empty).NotError(err)

	empty, err = r.Remove()
	a.True(empty).NotError(err)
	a.Empty(r.Methods()).Empty(r.Options())

	empty, err = r.Remove(http.MethodOptions)
	a.ErrorString(err, "不能手动删除 OPTIONS").False(empty)
}

func TestRoute_methods(t *testing.T) {
	a := assert.New(t)

	hs := New(true)
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
	a.Equal(hs.Options(), "")

	a.NotError(hs.Add(getHandler, http.MethodGet))
	test("GET, OPTIONS")
	a.Equal(hs.Options(), "GET, OPTIONS")

	a.NotError(hs.Add(getHandler, http.MethodPost))
	a.Equal(hs.Options(), "GET, OPTIONS, POST")
	test("GET, OPTIONS, POST")

	a.False(hs.Remove(http.MethodGet))
	test("OPTIONS, POST")
	a.Equal(hs.Options(), "OPTIONS, POST")
}
