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

// 一些预定义的处理函数
var (
	f1 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
	}
	f2 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(2)
	}
	f3 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(3)
	}

	h1 = http.HandlerFunc(f1)
	h2 = http.HandlerFunc(f2)
	h3 = http.HandlerFunc(f3)
)

func request(a *assert.Assertion, srvmux *Mux, method, url string, status int) {
	w := httptest.NewRecorder()
	a.NotNil(w)

	r, err := http.NewRequest(method, url, nil)
	a.NotError(err).NotNil(r)

	srvmux.ServeHTTP(w, r)
	a.Equal(w.Code, status)
}

func requestOptions(a *assert.Assertion, srvmux *Mux, url string, status int, allow string) {
	w := httptest.NewRecorder()
	a.NotNil(w)

	r, err := http.NewRequest(http.MethodOptions, url, nil)
	a.NotError(err).NotNil(r)

	srvmux.ServeHTTP(w, r)
	a.Equal(w.Code, status)
	a.Equal(w.Header().Get("Allow"), allow)
}

func TestMux_Add_Remove(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, false, nil, nil)
	a.NotNil(srvmux)

	// 添加 GET /api/1
	// 添加 PUT /api/1
	// 添加 GET /api/2
	a.NotError(srvmux.HandleFunc("/api/1", f1, http.MethodGet))
	a.NotPanic(func() {
		srvmux.PutFunc("/api/1", f1)
	})
	a.NotPanic(func() {
		srvmux.GetFunc("/api/2", f2)
	})
	request(a, srvmux, http.MethodGet, "/api/1", 1)
	request(a, srvmux, http.MethodPut, "/api/1", 1)
	request(a, srvmux, http.MethodGet, "/api/2", 2)
	request(a, srvmux, http.MethodDelete, "/api/1", http.StatusMethodNotAllowed) // 未实现

	// 删除 GET /api/1
	srvmux.Remove("/api/1", http.MethodGet)
	request(a, srvmux, http.MethodGet, "/api/1", http.StatusMethodNotAllowed)
	request(a, srvmux, http.MethodPut, "/api/1", 1) // 不影响 PUT
	request(a, srvmux, http.MethodGet, "/api/2", 2)

	// 删除 GET /api/2，只有一个，所以相当于整个 Entry 被删除
	srvmux.Remove("/api/2", http.MethodGet)
	request(a, srvmux, http.MethodGet, "/api/1", http.StatusMethodNotAllowed)
	request(a, srvmux, http.MethodPut, "/api/1", 1)                   // 不影响 PUT
	request(a, srvmux, http.MethodGet, "/api/2", http.StatusNotFound) // 整个 entry 被删除

	// 添加 POST /api/1
	a.NotPanic(func() {
		srvmux.PostFunc("/api/1", f1)
	})
	request(a, srvmux, http.MethodPost, "/api/1", 1)

	// 删除 ANY /api/1
	srvmux.Remove("/api/1")
	request(a, srvmux, http.MethodPost, "/api/1", http.StatusNotFound) // 404 表示整个 entry 都没了
}

func TestMux_Options(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, false, nil, nil)
	a.NotNil(srvmux)

	// 添加 GET /api/1
	a.NotError(srvmux.Handle("/api/1", h1, http.MethodGet))
	requestOptions(a, srvmux, "/api/1", http.StatusOK, "GET, OPTIONS")

	// 添加 DELETE /api/1
	a.NotPanic(func() {
		srvmux.Delete("/api/1", h1)
	})
	requestOptions(a, srvmux, "/api/1", http.StatusOK, "DELETE, GET, OPTIONS")

	// 删除 DELETE /api/1
	srvmux.Remove("/api/1", http.MethodDelete)
	requestOptions(a, srvmux, "/api/1", http.StatusOK, "GET, OPTIONS")

	// 通过 Options 自定义 Allow 报头
	srvmux.Options("/api/1", "CUSTOM OPTIONS1")
	requestOptions(a, srvmux, "/api/1", http.StatusOK, "CUSTOM OPTIONS1")
	srvmux.Options("/api/1", "CUSTOM OPTIONS2")
	requestOptions(a, srvmux, "/api/1", http.StatusOK, "CUSTOM OPTIONS2")

	srvmux.HandleFunc("/api/1", f1, http.MethodOptions)
	requestOptions(a, srvmux, "/api/1", 1, "")
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

	// 添加 patch /api/v2/{version:\\d+}
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
	srvmux := New(false, false, nil, nil)
	a.NotNil(srvmux)

	srvmux.Handle("/posts/{path}.html", h1) // 命名参数不能带 /
	request(a, srvmux, http.MethodGet, "/posts/2017/1.html", 404)

	srvmux.Handle("/posts/{path:.+}.html", h2)
	request(a, srvmux, http.MethodGet, "/posts/2017/1.html", 2)
}

// 测试匹配顺序是否正确
func TestMux_ServeHTTP_Order(t *testing.T) {
	a := assert.New(t)

	serveMux := New(false, false, nil, nil)
	a.NotNil(serveMux)

	a.NotError(serveMux.GetFunc("/posts/{id}", f3))    // f3
	a.NotError(serveMux.GetFunc("/posts/{id:.+}", f2)) // f2
	a.NotError(serveMux.GetFunc("/posts/1", f1))       // f1

	request(a, serveMux, http.MethodGet, "/posts/1", 1)        // f1 普通路由项完全匹配
	request(a, serveMux, http.MethodGet, "/posts/2", 3)        // f1 命名路由
	request(a, serveMux, http.MethodGet, "/posts/2/author", 2) // f2 正则正则匹配
	request(a, serveMux, http.MethodGet, "/posts/abc", 3)      // f3 命名路由

	serveMux = New(false, false, nil, nil)
	a.NotNil(serveMux)

	a.NotError(serveMux.GetFunc("/posts/*", f3))         // f3
	a.NotError(serveMux.GetFunc("/posts/{id:\\d+}", f2)) // f2
	a.NotError(serveMux.GetFunc("/posts/1", f1))         // f1

	request(a, serveMux, http.MethodGet, "/posts/1", 1)   // f1 普通路由项完全匹配
	request(a, serveMux, http.MethodGet, "/posts/2", 2)   // f2 正则完全匹配
	request(a, serveMux, http.MethodGet, "/posts/abc", 3) // f3 通配符完全匹配

	serveMux = New(false, false, nil, nil)
	a.NotNil(serveMux)

	a.NotError(serveMux.GetFunc("/p1/{p1}/p2/{p2:\\d+}", f1)) // f1
	a.NotError(serveMux.GetFunc("/p1/{p1}/p2/{p2:\\w+}", f2)) // f2

	request(a, serveMux, http.MethodGet, "/p1/1/p2/1", 1) // f1
	request(a, serveMux, http.MethodGet, "/p1/2/p2/s", 2) // f2

	serveMux = New(false, false, nil, nil)
	a.NotNil(serveMux)

	a.NotError(serveMux.GetFunc("/posts/{id}/{page}", f2))
	a.NotError(serveMux.GetFunc("/posts/{id}/1", f1))

	request(a, serveMux, http.MethodGet, "/posts/1/1", 1) // f1 普通路由项完全匹配
	request(a, serveMux, http.MethodGet, "/posts/2/5", 2) // f2 命名完全匹配
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
