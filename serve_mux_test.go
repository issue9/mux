// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestServeMux_Add(t *testing.T) {
	a := assert.New(t)
	m := NewServeMux()
	a.NotNil(m)

	// handler不能为空
	a.Error(m.Add("abc", nil, "GET"))

	fn := func(w http.ResponseWriter, req *http.Request) {}
	h := http.HandlerFunc(fn)

	a.Error(m.Add("", h, "GET"))

	// 向Get和Post添加一个路由abc
	a.NotError(m.Add("abc", h, "GET", "POST"))
	entries, found := m.items["GET"]
	a.True(found).Equal(1, len(entries.named))
	entries, found = m.items["POST"]
	a.True(found).Equal(1, len(entries.named))
	entries, found = m.items["DELETE"]
	a.True(found).Equal(0, len(entries.named))

	// 再次向Get添加一条同名路由，会出错
	a.Error(m.Get("abc", h))

	a.NotError(m.Get("def", h))
	es, found := m.items["GET"]
	a.True(found).Equal(2, len(es.statics))

	// Delete
	a.NotError(m.Delete("abc", h))
	es, found = m.items["DELETE"]
	a.True(found).Equal(1, len(es.statics))
	a.NotError(m.DeleteFunc("abcd", fn))
	a.True(found).Equal(2, len(es.statics))

	//Put
	a.NotError(m.Put("abc", h))
	es, found = m.items["PUT"]
	a.True(found).Equal(1, len(es.statics))
	a.NotError(m.PutFunc("abcd", fn))
	a.True(found).Equal(2, len(es.statics))
}

func TestServeMux_Remove(t *testing.T) {
	a := assert.New(t)
	m := NewServeMux()
	a.NotNil(m)

	fn := func(w http.ResponseWriter, req *http.Request) {}
	h := http.HandlerFunc(fn)

	// 向Get和Post添加一个路由abc
	a.NotError(m.Add("abc", h, "GET", "POST", "DELETE"))
	a.NotError(m.Add("abcd", h, "GET", "POST", "DELETE"))
	a.Equal(2, len(m.items["GET"].statics))
	a.Equal(2, len(m.items["POST"].statics))
	a.Equal(2, len(m.items["DELETE"].statics))

	// 删除Get,Post下的匹配项
	m.Remove("abc", "GET", "POST")
	a.Equal(1, len(m.items["GET"].statics))
	a.Equal(1, len(m.items["POST"].statics))
	a.Equal(2, len(m.items["DELETE"].statics))

	// 删除GET下的匹配项。
	m.Remove("abcd", "GET")
	a.Equal(0, len(m.items["GET"].statics))
	a.Equal(1, len(m.items["POST"].statics))
	a.Equal(2, len(m.items["DELETE"].statics))

	// 删除任意method下的匹配项。
	m.Remove("abcd")
	a.Equal(0, len(m.items["GET"].statics))
	a.Equal(0, len(m.items["POST"].statics))
	a.Equal(1, len(m.items["DELETE"].statics))
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
// BenchmarkServeMux_ServeHTTPStatic	 2000000	       843 ns/op
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
// BenchmarkServeMux_ServeHTTPRegexp	  300000	      4082 ns/op
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
// BenchmarkServeMux_ServeHTTPAll	  500000	      2506 ns/op
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
