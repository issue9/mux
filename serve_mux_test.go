// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/issue9/assert"
)

// 断言mux下的method请求方法下有l条路由记录。
// 即mux.items[method].list.Len()的值与l相等。
func assertLen(mux *ServeMux, a *assert.Assertion, l int, method string) {
	info := func(v1, v2 int) string {
		_, file, line, ok := runtime.Caller(2)
		if !ok {
			return "<none>"
		}
		return fmt.Sprintf("v1:[%v] != v2:[%v]：@ %v:%v", v1, v2, file, line)
	}

	l1 := mux.items[method].list.Len()
	a.Equal(l, l1, info(l, l1))
	l1 = len(mux.items[method].named)
	a.Equal(l, l1, info(l, l1))
}

func TestServeMux_Add(t *testing.T) {
	a := assert.New(t)
	m := NewServeMux()
	a.NotNil(m)

	fn := func(w http.ResponseWriter, req *http.Request) {}
	h := http.HandlerFunc(fn)

	// handler不能为空
	a.Panic(func() { m.Add("abc", nil, "GET") })
	// pattern不能为空
	a.Panic(func() { m.Add("", h, "GET") })
	// 不支持的methods
	a.Panic(func() { m.Add("", h, "GET123") })

	// 向Get和Post添加一个路由abc
	a.NotPanic(func() { m.Add("abc", h, "GET", "POST") })
	assertLen(m, a, 1, "GET")
	assertLen(m, a, 1, "POST")
	assertLen(m, a, 0, "DELETE")
	// 再次向Get添加一条同名路由，会出错
	a.Panic(func() { m.Get("abc", h) })

	a.NotPanic(func() { m.Add("abcdefg", h) })
	assertLen(m, a, 2, "GET")
	assertLen(m, a, 2, "POST")
	assertLen(m, a, 1, "DELETE")

	a.NotPanic(func() { m.Get("def", h) })
	assertLen(m, a, 3, "GET")
	assertLen(m, a, 2, "POST")
	assertLen(m, a, 1, "DELETE")

	a.NotPanic(func() { m.Delete("abc", h) })
	assertLen(m, a, 3, "GET")
	assertLen(m, a, 2, "POST")
	assertLen(m, a, 2, "DELETE")

	a.NotPanic(func() { m.Any("abcd", h) })
	assertLen(m, a, 4, "GET")
	assertLen(m, a, 3, "POST")
	assertLen(m, a, 3, "DELETE")
}

func TestServeMux_Remove(t *testing.T) {
	a := assert.New(t)
	m := NewServeMux()
	a.NotNil(m)

	fn := func(w http.ResponseWriter, req *http.Request) {}
	h := http.HandlerFunc(fn)

	// 向Get和Post添加一个路由abc
	m.Add("abc", h, "GET", "POST", "DELETE")
	m.Add("abcd", h, "GET", "POST", "DELETE")
	assertLen(m, a, 2, "GET")
	assertLen(m, a, 2, "POST")
	assertLen(m, a, 2, "DELETE")

	// 删除Get,Post下的匹配项
	m.Remove("abc", "GET", "POST")
	assertLen(m, a, 1, "GET")
	assertLen(m, a, 1, "POST")
	assertLen(m, a, 2, "DELETE")

	// 删除GET下的匹配项。
	m.Remove("abcd", "GET")
	assertLen(m, a, 0, "GET")
	assertLen(m, a, 1, "POST")
	assertLen(m, a, 2, "DELETE")

	// 删除任意method下的匹配项。
	m.Remove("abcd")
	assertLen(m, a, 0, "GET")
	assertLen(m, a, 0, "POST")
	assertLen(m, a, 1, "DELETE")
}

func TestServeMux_ServeHTTP(t *testing.T) {
	a := assert.New(t)

	newServeMux := func(pattern string) http.Handler {
		h := NewServeMux()
		a.NotError(h.AddFunc(pattern, defaultHandler, "GET"))
		return h
	}

	tests := []*handlerTester{
		&handlerTester{
			name:       "普通匹配",
			h:          newServeMux("/abc"),
			query:      "/abc",
			statusCode: 200,
		},
		&handlerTester{
			name:       "普通部分匹配",
			h:          newServeMux("/abc"),
			query:      "/abcd",
			statusCode: 200,
		},
		&handlerTester{
			name:       "普通不匹配",
			h:          newServeMux("/abc"),
			query:      "/cba",
			statusCode: 404,
		},
		&handlerTester{
			name:       "正则匹配数字",
			h:          newServeMux("/api/{version:\\d+}"),
			query:      "/api/2",
			statusCode: 200,
			ctxName:    "params",
			ctxMap:     map[string]string{"version": "2"},
		},
		&handlerTester{
			name:       "正则匹配多个名称",
			h:          newServeMux("/api/{version:\\d+}/{name:\\w+}"),
			query:      "/api/2/login",
			statusCode: 200,
			ctxName:    "params",
			ctxMap:     map[string]string{"version": "2", "name": "login"},
		},
		&handlerTester{
			name:       "正则不匹配多个名称",
			h:          newServeMux("/api/{version:\\d+}/{name:\\w+}"),
			query:      "/api/2.0/login",
			statusCode: 404,
		},
		&handlerTester{
			name:       "带域名的字符串不匹配", //无法匹配端口信息
			h:          newServeMux("127.0.0.1/abc"),
			query:      "/cba",
			statusCode: 404,
		},
		&handlerTester{
			name:       "带域名的正则匹配", //无法匹配端口信息
			h:          newServeMux("127.0.0.1:{:\\d+}/abc"),
			query:      "/abc",
			statusCode: 200,
		},
		&handlerTester{
			name:       "带域名的命名正则匹配", //无法匹配端口信息
			h:          newServeMux("127.0.0.1:{:\\d+}/api/v{version:\\d+}/login"),
			query:      "/api/v2/login",
			statusCode: 200,
			ctxName:    "params",
			ctxMap:     map[string]string{"version": "2"},
		},
	}

	runHandlerTester(a, tests)
}

// 全静态匹配
// BenchmarkServeMux_ServeHTTPStatic	 2000000	       669 ns/op
func BenchmarkServeMux_ServeHTTPStatic(b *testing.B) {
	a := assert.New(b)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler"))
	})
	srv := NewServeMux()
	srv.Get("/blog/post/1", h)
	srv.Get("/api/v2/login", h)

	r1, err := http.NewRequest("GET", "/blog/post/1", nil)
	a.NotError(err).NotNil(r1)
	r2, err := http.NewRequest("GET", "/api/v2/login", nil)
	a.NotError(err).NotNil(r2)
	r3, err := http.NewRequest("GET", "/api/v2x/login", nil)
	a.NotError(err).NotNil(r3)
	reqs := []*http.Request{r1, r2, r3}

	w := httptest.NewRecorder()

	srvfun := func(reqIndex int) {
		defer func() {
			_ = recover()
		}()

		srv.ServeHTTP(w, reqs[reqIndex])
	}
	for i := 0; i < b.N; i++ {
		srvfun(i % len(reqs))
	}
}

// 全正则匹配
// BenchmarkServeMux_ServeHTTPRegexp	  500000	      3596 ns/op
func BenchmarkServeMux_ServeHTTPRegexp(b *testing.B) {
	a := assert.New(b)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler"))
	})
	srv := NewServeMux()
	srv.Get("/blog/post/{id}", h)
	srv.Get("/api/v{version:\\d+}/login", h)

	r1, err := http.NewRequest("GET", "/blog/post/1", nil)
	a.NotError(err).NotNil(r1)
	r2, err := http.NewRequest("GET", "/api/v2/login", nil)
	a.NotError(err).NotNil(r2)
	r3, err := http.NewRequest("GET", "/api/v2x/login", nil)
	a.NotError(err).NotNil(r3)
	reqs := []*http.Request{r1, r2, r3}

	w := httptest.NewRecorder()

	srvfun := func(reqIndex int) {
		defer func() {
			_ = recover()
		}()

		srv.ServeHTTP(w, reqs[reqIndex])
	}
	for i := 0; i < b.N; i++ {
		srvfun(i % len(reqs))
	}
}

// 静态正则汇合匹配
// BenchmarkServeMux_ServeHTTPAll	  500000	      2302 ns/op
func BenchmarkServeMux_ServeHTTPAll(b *testing.B) {
	a := assert.New(b)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler"))
	})
	srv := NewServeMux()
	srv.Get("/blog/post/1", h)
	srv.Get("/api/v{version:\\d+}/login", h)

	r1, err := http.NewRequest("GET", "/blog/post/1", nil)
	a.NotError(err).NotNil(r1)
	r2, err := http.NewRequest("GET", "/api/v2/login", nil)
	a.NotError(err).NotNil(r2)
	r3, err := http.NewRequest("GET", "/api/v2x/login", nil)
	a.NotError(err).NotNil(r3)
	reqs := []*http.Request{r1, r2, r3}

	w := httptest.NewRecorder()

	srvfun := func(reqIndex int) {
		defer func() {
			_ = recover()
		}()

		srv.ServeHTTP(w, reqs[reqIndex])
	}
	for i := 0; i < b.N; i++ {
		srvfun(i % len(reqs))
	}
}
