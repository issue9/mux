// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux/internal/method"
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

type tester struct {
	a   *assert.Assertion
	mux *Mux
}

func newTester(a *assert.Assertion, disableOptions, skipClean bool) *tester {
	return &tester{
		a:   a,
		mux: New(disableOptions, skipClean, nil, nil),
	}
}

// 确保能正常匹配到指定的 URL
func (t *tester) matchTrue(method, url string, code int) {
	w := httptest.NewRecorder()
	t.a.NotNil(w)

	r, err := http.NewRequest(method, url, nil)
	t.a.NotError(err).NotNil(r)

	t.mux.ServeHTTP(w, r)
	t.a.Equal(w.Code, code)
}

// 确保能正确匹配地址，且拿到正确的 options 报头
func (t *tester) optionsTrue(url string, code int, allow string) {
	w := httptest.NewRecorder()
	t.a.NotNil(w)

	r, err := http.NewRequest(http.MethodOptions, url, nil)
	t.a.NotError(err).NotNil(r)

	t.mux.ServeHTTP(w, r)
	t.a.Equal(w.Code, code)
	t.a.Equal(w.Header().Get("Allow"), allow)
}

func TestMux_Add_Remove(t *testing.T) {
	a := assert.New(t)
	test := newTester(a, false, false)

	// 添加 GET /api/1
	// 添加 PUT /api/1
	// 添加 GET /api/2
	a.NotError(test.mux.HandleFunc("/api/1", buildFunc(1), http.MethodGet))
	a.NotPanic(func() {
		test.mux.PutFunc("/api/1", buildFunc(1))
	})
	a.NotPanic(func() {
		test.mux.GetFunc("/api/2", buildFunc(2))
	})

	test.matchTrue(http.MethodGet, "/api/1", 1)
	test.matchTrue(http.MethodPut, "/api/1", 1)
	test.matchTrue(http.MethodGet, "/api/2", 2)
	test.matchTrue(http.MethodDelete, "/api/1", http.StatusMethodNotAllowed) // 未实现

	// 删除 GET /api/1
	test.mux.Remove("/api/1", http.MethodGet)
	test.matchTrue(http.MethodGet, "/api/1", http.StatusMethodNotAllowed)
	test.matchTrue(http.MethodPut, "/api/1", 1) // 不影响 PUT
	test.matchTrue(http.MethodGet, "/api/2", 2)

	// 删除 GET /api/2，只有一个，所以相当于整个节点被删除
	test.mux.Remove("/api/2", http.MethodGet)
	test.matchTrue(http.MethodGet, "/api/1", http.StatusMethodNotAllowed)
	test.matchTrue(http.MethodPut, "/api/1", 1)                   // 不影响 PUT
	test.matchTrue(http.MethodGet, "/api/2", http.StatusNotFound) // 整个节点被删除

	// 添加 POST /api/1
	a.NotPanic(func() {
		test.mux.PostFunc("/api/1", buildFunc(1))
	})
	test.matchTrue(http.MethodPost, "/api/1", 1)

	// 删除 ANY /api/1
	test.mux.Remove("/api/1")
	test.matchTrue(http.MethodPost, "/api/1", http.StatusNotFound) // 404 表示整个节点都没了
}

func TestMux_Options(t *testing.T) {
	a := assert.New(t)
	test := newTester(a, false, false)

	// 添加 GET /api/1
	a.NotError(test.mux.Handle("/api/1", buildHandler(1), http.MethodGet))
	test.optionsTrue("/api/1", http.StatusOK, "GET, OPTIONS")

	// 添加 DELETE /api/1
	a.NotPanic(func() {
		test.mux.Delete("/api/1", buildHandler(1))
	})
	test.optionsTrue("/api/1", http.StatusOK, "DELETE, GET, OPTIONS")

	// 删除 DELETE /api/1
	test.mux.Remove("/api/1", http.MethodDelete)
	test.optionsTrue("/api/1", http.StatusOK, "GET, OPTIONS")

	// 通过 Options 自定义 Allow 报头
	test.mux.Options("/api/1", "CUSTOM OPTIONS1")
	test.optionsTrue("/api/1", http.StatusOK, "CUSTOM OPTIONS1")
	test.mux.Options("/api/1", "CUSTOM OPTIONS2")
	test.optionsTrue("/api/1", http.StatusOK, "CUSTOM OPTIONS2")

	test.mux.HandleFunc("/api/1", buildFunc(1), http.MethodOptions)
	test.optionsTrue("/api/1", 1, "")
}

func TestMux_Params(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, false, nil, nil)
	a.NotNil(srvmux)
	params := map[string]string{}

	buildParamsHandler := func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ps := GetParams(r)
			a.NotNil(ps)
			params = ps
		})
	}

	requestParams := func(a *assert.Assertion, srvmux *Mux, method, url string, status int, ps map[string]string) {
		w := httptest.NewRecorder()
		a.NotNil(w)

		r, err := http.NewRequest(method, url, nil)
		a.NotError(err).NotNil(r)

		srvmux.ServeHTTP(w, r)

		a.Equal(w.Code, status)
		if ps != nil { // 由于 params 是公用数据，会保存上一次获取的值，所以只在有值时才比较
			a.Equal(params, ps)
		}
		params = nil // 清空全局的 params
	}

	// 添加 patch /api/{version:\\d+}
	a.NotError(srvmux.Patch("/api/{version:\\d+}", buildParamsHandler()))
	requestParams(a, srvmux, http.MethodPatch, "/api/2", http.StatusOK, map[string]string{"version": "2"})
	requestParams(a, srvmux, http.MethodPatch, "/api/256", http.StatusOK, map[string]string{"version": "256"})
	requestParams(a, srvmux, http.MethodGet, "/api/256", http.StatusMethodNotAllowed, nil) // 不存在的请求方法

	// 添加 patch /api/v2/{version:\\d*}
	a.NotError(srvmux.Patch("/api/v2/{version:\\d*}", buildParamsHandler()))
	requestParams(a, srvmux, http.MethodPatch, "/api/v2/2", http.StatusOK, map[string]string{"version": "2"})
	requestParams(a, srvmux, http.MethodPatch, "/api/v2/", http.StatusOK, map[string]string{"version": ""})

	// 添加 patch /api/v2/{version:\\d+}/test
	a.NotError(srvmux.Patch("/api/v2/{version:\\d*}/test", buildParamsHandler()))
	requestParams(a, srvmux, http.MethodPatch, "/api/v2/2/test", http.StatusOK, map[string]string{"version": "2"})
	requestParams(a, srvmux, http.MethodPatch, "/api/v2//test", http.StatusNotFound, nil) // 可选参数不能在路由中间
}

func TestMux_ServeHTTP(t *testing.T) {
	a := assert.New(t)
	test := newTester(a, false, false)

	test.mux.Handle("/posts/{path}.html", buildHandler(1))
	test.matchTrue(http.MethodGet, "/posts/2017/1.html", 1)

	test.mux.Handle("/posts/{path:.+}.html", buildHandler(2))
	test.matchTrue(http.MethodGet, "/posts/2017/1.html", 2)
}

// 测试匹配顺序是否正确
func TestMux_ServeHTTP_Order(t *testing.T) {
	a := assert.New(t)
	test := newTester(a, false, false)

	a.NotError(test.mux.GetFunc("/posts/{id}", buildFunc(3)))      // f3
	a.NotError(test.mux.GetFunc("/posts/{id:\\d+}", buildFunc(2))) // f2
	a.NotError(test.mux.GetFunc("/posts/1", buildFunc(1)))         // f1

	test.matchTrue(http.MethodGet, "/posts/1", 1)   // f1 普通路由项完全匹配
	test.matchTrue(http.MethodGet, "/posts/2", 2)   // f1 正则路由
	test.matchTrue(http.MethodGet, "/posts/abc", 3) // f3 命名路由

	test = newTester(a, false, false)

	a.NotError(test.mux.GetFunc("/p1/{p1}/p2/{p2:\\d+}", buildFunc(1))) // f1
	a.NotError(test.mux.GetFunc("/p1/{p1}/p2/{p2:\\w+}", buildFunc(2))) // f2

	test.matchTrue(http.MethodGet, "/p1/1/p2/1", 1) // f1
	test.matchTrue(http.MethodGet, "/p1/2/p2/s", 2) // f2

	test = newTester(a, false, false)

	a.NotError(test.mux.GetFunc("/posts/{id}/{page}", buildFunc(2))) // f2
	a.NotError(test.mux.GetFunc("/posts/{id}/1", buildFunc(1)))      // f1

	test.matchTrue(http.MethodGet, "/posts/1/1", 1) // f1 普通路由项完全匹配
	test.matchTrue(http.MethodGet, "/posts/2/5", 2) // f2 命名完全匹配
}

func TestClearPath(t *testing.T) {
	a := assert.New(t)

	a.Equal(cleanPath(""), "/")

	a.Equal(cleanPath("/api//"), "/api/")
	a.Equal(cleanPath("api//"), "/api/")
	a.Equal(cleanPath("//api//"), "/api/")

	a.Equal(cleanPath("/api/"), "/api/")
	a.Equal(cleanPath("/api/./"), "/api/./")

	a.Equal(cleanPath("/api/.."), "/api/..")
	a.Equal(cleanPath("/api/../"), "/api/../")
	a.Equal(cleanPath("/api/../../"), "/api/../../")
}

func TestSupportedMethods(t *testing.T) {
	a := assert.New(t)
	ms1 := SupportedMethods()
	a.Equal(method.Supported, ms1)

	// 不应该改变包内部的变量
	ms1[0] = "abc"
	a.NotEqual(method.Supported, ms1)
	a.False(MethodIsSupported("abc"))

	ms2 := SupportedMethods()
	a.Equal(method.Supported, ms2)
	ms2[0] = "def"
	a.NotEqual(ms1, ms2)
}

func BenchmarkCleanPath(b *testing.B) {
	a := assert.New(b)

	paths := []string{
		"",
		"/api//",
		"/api////users/1",
		"//api/users/1",
		"api///users////1",
		"api//",
		"/api/",
		"/api/./",
		"/api/..",
		"/api//../",
		"/api/..//../",
		"/api../",
		"api../",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ret := cleanPath(paths[i%len(paths)])
		a.True(len(ret) > 0)
	}
}
