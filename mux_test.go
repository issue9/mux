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

// 断言mux.paths[method].Len() == l
func assertLen(mux *ServeMux, a *assert.Assertion, l int, method string) {
	a.Equal(l, mux.entries[method].Len())
}

func TestServeMux_Options(t *testing.T) {
	a := assert.New(t)
	m := NewServeMux()
	a.NotNil(m)

	fn := func(w http.ResponseWriter, req *http.Request) {}
	h := http.HandlerFunc(fn)

	a.NotPanic(func() { m.Get("/a", h) })
	a.Equal(get|options, m.options["/a"]) // 包含一个自动增加的OPTIONS

	a.NotPanic(func() { m.Post("/a", h) })
	a.Equal(get|options|post, m.options["/a"])

	// 手动调整，去掉了自动增加的OPTIONS
	a.NotPanic(func() { m.Options("/a", "GET", "POST", "PUT") })
	a.Equal(get|post|put, m.options["/a"])
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
	a.Panic(func() { m.Add("", h, "GET") })
	// 不支持的methods
	a.Panic(func() { m.Add("abc", h, "GET123") })
}

func TestServeMux_Clean(t *testing.T) {
	a := assert.New(t)
	m := NewServeMux()
	a.NotNil(m)

	fn := func(w http.ResponseWriter, req *http.Request) {}
	h := http.HandlerFunc(fn)

	m.Add("/", h)
	assertLen(m, a, 1, "GET")
	assertLen(m, a, 1, "DELETE")
	a.Equal(m.options["/"], get|post|put|patch|delete|options|trace|head)
	m.Clean()
	a.Equal(m.options["/"], 0)
	assertLen(m, a, 0, "GET")
	assertLen(m, a, 0, "DELETE")

	m.Add("/", h)
	m.Add("/index.html", h)
	m.Add("www.caixw.io/index.html", h)
	assertLen(m, a, 3, "GET")
	m.Clean()
	a.Equal(m.options["/"], 0)
	assertLen(m, a, 0, "GET")
	assertLen(m, a, 0, "DELETE")
}

func TestServeMux_Remove(t *testing.T) {
	a := assert.New(t)
	m := NewServeMux()
	a.NotNil(m)

	fn := func(w http.ResponseWriter, req *http.Request) {}
	h := http.HandlerFunc(fn)
	m.Add("/", h, "GET")
	a.Equal(m.options["/"], get|options)
	assertLen(m, a, 1, "GET")

	// method 不同
	m.Remove("/", "DELETE")
	a.Equal(m.options["/"], get|options)
	a.NotNil(m.base["GET"])
	assertLen(m, a, 1, "GET")
	assertLen(m, a, 0, "DELETE")

	m.Remove("/", "GET")
	a.Equal(m.options["/"], 0) // 若只有options，则直接清空
	assertLen(m, a, 0, "GET")

	m.Add("/", h)
	m.Remove("/", "DELETE")
	assertLen(m, a, 1, "GET")
	assertLen(m, a, 0, "DELETE")

	// 清除所有
	m.Remove("/")
	assertLen(m, a, 0, "GET")
	assertLen(m, a, 0, "POST")
	assertLen(m, a, 0, "DELETE")

	m.Add("www.caixw.io/index.html", h)
	m.Remove("www.caixw.io/index.html")
	a.Nil(m.base["GET"])
	assertLen(m, a, 0, "GET")
	assertLen(m, a, 0, "POST")
	assertLen(m, a, 0, "DELETE")
}

// 测试各种匹配模式是否正常工作
func TestServeMux_ServeHTTP(t *testing.T) {
	a := assert.New(t)

	tests := []*handlerTester{
		&handlerTester{
			name:       "普通匹配",
			pattern:    "/abc",
			query:      "/abc",
			statusCode: 200,
		},
		&handlerTester{
			name:       "普通部分匹配",
			pattern:    "/abc/",
			query:      "/abc/d",
			statusCode: 200,
		},
		&handlerTester{
			name:       "普通不匹配-1",
			pattern:    "/abc",
			query:      "/abcd",
			statusCode: 404,
		},
		&handlerTester{
			name:       "普通不匹配-2",
			pattern:    "/abc",
			query:      "/cba",
			statusCode: 404,
		},
		&handlerTester{
			name:       "正则匹配数字",
			pattern:    "/api/{version:\\d+}",
			query:      "/api/2",
			statusCode: 200,
			params:     map[string]string{"version": "2"},
		},
		&handlerTester{
			name:       "正则匹配多个名称",
			pattern:    "/api/{version:\\d+}/{name:\\w+}",
			query:      "/api/2/login",
			statusCode: 200,
			params:     map[string]string{"version": "2", "name": "login"},
		},
		&handlerTester{
			name:       "正则不匹配多个名称",
			pattern:    "/api/{version:\\d+}/{name:\\w+}",
			query:      "/api/2.0/login",
			statusCode: 404,
		},
		&handlerTester{
			name:       "带域名的字符串不匹配", //无法匹配端口信息
			pattern:    "127.0.0.1/abc",
			query:      "/cba",
			statusCode: 404,
		},
		&handlerTester{
			name:       "带域名的正则匹配", //无法匹配端口信息
			pattern:    "127.0.0.1:{:\\d+}/abc",
			query:      "/abc",
			statusCode: 200,
		},
		&handlerTester{
			name:       "带域名的命名正则匹配", //无法匹配端口信息
			pattern:    "127.0.0.1:{:\\d+}/api/v{version:\\d+}/login",
			query:      "/api/v2/login",
			statusCode: 200,
			params:     map[string]string{"version": "2"},
		},
	}

	runHandlerTester(a, tests)
}

// 测试匹配顺序是否正确
func TestServeMux_ServeHTTP_Order(t *testing.T) {
	a := assert.New(t)

	// 一些预定义的处理函数
	f1 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
	}
	f2 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(2)
	}
	f3 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(3)
	}
	f4 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(4)
	}
	f5 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(5)
	}
	f6 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(6)
	}

	test := func(m *ServeMux, method, host, path string, code int) {
		r, err := http.NewRequest(method, path, nil)
		if len(host) > 0 {
			r.Host = host
		}
		a.NotError(err).NotNil(r)
		w := httptest.NewRecorder()
		a.NotNil(w)
		m.ServeHTTP(w, r)
		a.Equal(w.Code, code)
	}

	serveMux := NewServeMux()
	a.NotNil(serveMux)
	serveMux.AddFunc("/post/", f1, "GET")                   // f1
	serveMux.AddFunc("/post/{id:\\d+}", f2, "GET")          // f2
	serveMux.AddFunc("/post/1", f3, "GET")                  // f3
	serveMux.AddFunc("127.0.0.1/post/", f4, "GET")          // f4
	serveMux.AddFunc("127.0.0.1/post/{id:\\d+}", f5, "GET") // f5
	serveMux.AddFunc("127.0.0.1/post/1", f6, "GET")         // f6

	test(serveMux, "GET", "", "/post/1", 3)            // f3 静态路由项完全匹配
	test(serveMux, "GET", "", "/post/2", 2)            // f2 正则完全匹配
	test(serveMux, "GET", "", "/post/abc", 1)          // f1 匹配度最高
	test(serveMux, "GET", "127.0.0.1", "/post/1", 6)   // f6 静态路由项完全匹配
	test(serveMux, "GET", "127.0.0.1", "/post/2", 5)   // f5 正则完全匹配
	test(serveMux, "GET", "127.0.0.1", "/post/abc", 4) // f4 匹配度最高
}

// 全静态匹配
// BenchmarkServeMux_ServeHTTPStatic-4	 5000000	       365 ns/op    go1.6
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
// BenchmarkServeMux_ServeHTTPRegexp-4	 1000000	      1640 ns/op    go1.6
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
// BenchmarkServeMux_ServeHTTPAll-4   	 1000000	      1024 ns/op    go1.6
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
