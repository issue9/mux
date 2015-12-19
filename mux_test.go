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

// 断言mux.list[method].Len() == len(mux.named[method]) == l
func assertLen(mux *ServeMux, a *assert.Assertion, l int, method string) {
	a.Equal(l, mux.list[method].Len())
	a.Equal(l, len(mux.named[method]))
}

func TestServeMux_Add(t *testing.T) {
	a := assert.New(t)
	m := NewServeMux()
	a.NotNil(m)

	fn := func(w http.ResponseWriter, req *http.Request) {}
	h := http.HandlerFunc(fn)

	a.NotPanic(func() { m.Get("h", h) })
	assertLen(m, a, 1, "GET")

	a.NotPanic(func() { m.Post("h", h) })
	assertLen(m, a, 1, "POST")

	a.NotPanic(func() { m.Put("h", h) })
	assertLen(m, a, 1, "PUT")

	a.NotPanic(func() { m.Delete("h", h) })
	assertLen(m, a, 1, "DELETE")

	a.NotPanic(func() { m.Patch("h", h) })
	assertLen(m, a, 1, "PATCH")

	a.NotPanic(func() { m.Any("anyH", h) })
	assertLen(m, a, 2, "PUT")
	assertLen(m, a, 2, "DELETE")

	a.NotPanic(func() { m.GetFunc("fn", fn) })
	assertLen(m, a, 3, "GET")

	a.NotPanic(func() { m.PostFunc("fn", fn) })
	assertLen(m, a, 3, "POST")

	a.NotPanic(func() { m.PutFunc("fn", fn) })
	assertLen(m, a, 3, "PUT")

	a.NotPanic(func() { m.DeleteFunc("fn", fn) })
	assertLen(m, a, 3, "DELETE")

	a.NotPanic(func() { m.PatchFunc("fn", fn) })
	assertLen(m, a, 3, "PATCH")

	a.NotPanic(func() { m.AnyFunc("anyFN", fn) })
	assertLen(m, a, 4, "DELETE")
	assertLen(m, a, 4, "GET")

	// 添加相同的pattern
	a.Panic(func() { m.Any("h", h) })

	// handler不能为空
	a.Panic(func() { m.Add("abc", nil, "GET") })
	// pattern不能为空
	//a.Panic(func() { m.Add("", h, "GET") })
	// 不支持的methods
	a.Panic(func() { m.Add("abc", h, "GET123") })

}

// 测试各种匹配模式是否正常工作
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
			h:          newServeMux("/abc/"),
			query:      "/abc/d",
			statusCode: 200,
		},
		&handlerTester{
			name:       "普通不匹配-1",
			h:          newServeMux("/abc"),
			query:      "/abcd",
			statusCode: 404,
		},
		&handlerTester{
			name:       "普通不匹配-2",
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
// BenchmarkServeMux_ServeHTTPStatic-4	10000000	       203 ns/op
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
// BenchmarkServeMux_ServeHTTPRegexp-4	 1000000	      1477 ns/op
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
// BenchmarkServeMux_ServeHTTPAll-4   	 2000000	       849 ns/op
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
