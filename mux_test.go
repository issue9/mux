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

func TestClearPath(t *testing.T) {
	a := assert.New(t)

	a.Equal(cleanPath(""), "/")
	a.Equal(cleanPath("/api//"), "/api/")

	a.Equal(cleanPath("/api/"), "/api/")
	a.Equal(cleanPath("/api/./"), "/api/")

	a.Equal(cleanPath("/api/.."), "/")
	a.Equal(cleanPath("/api/../"), "/")

	a.Equal(cleanPath("/api/../../"), "/")
	a.Equal(cleanPath("/api../"), "/api../")
}

// 测试各种匹配模式是否正常工作
func TestServeMux_ServeHTTP(t *testing.T) {
	a := assert.New(t)

	tests := []*tester{
		&tester{
			name:    "普通匹配",
			pattern: "/abc",
			method:  http.MethodGet,
			url:     "/abc",
			status:  http.StatusOK,
		},
		&tester{
			name:    "普通部分匹配",
			pattern: "/abc/",
			method:  http.MethodGet,
			url:     "/abc/d",
			status:  http.StatusOK,
		},
		&tester{
			name:    "普通不匹配-1",
			pattern: "/abc",
			method:  http.MethodGet,
			url:     "/abcd",
			status:  http.StatusNotFound,
		},
		&tester{
			name:    "普通不匹配-2",
			pattern: "/abc",
			method:  http.MethodGet,
			url:     "/cba",
			status:  http.StatusNotFound,
		},
		&tester{
			name:    "正则匹配数字",
			pattern: "/api/{version:\\d+}",
			method:  http.MethodPost,
			url:     "/api/2",
			status:  http.StatusOK,
			params:  map[string]string{"version": "2"},
		},
		&tester{
			name:    "不规则的路径-1",
			pattern: "/api/{version:\\d+}",
			method:  http.MethodDelete,
			url:     "/api//2",
			status:  http.StatusOK,
			params:  map[string]string{"version": "2"},
		},
		&tester{
			name:    "不规则的路径-2",
			pattern: "/api/{version:\\d+}",
			method:  http.MethodGet,
			url:     "/api/nest/../2", // 上一层路径
			status:  http.StatusOK,
			params:  map[string]string{"version": "2"},
		},
		&tester{
			name:    "不规则的路径-3",
			pattern: "/{version:\\d+}",
			method:  http.MethodGet,
			url:     "/api/../../../2", // 上 N 层路径，超过根路径
			status:  http.StatusOK,
			params:  map[string]string{"version": "2"},
		},
		&tester{
			name:    "不规则的路径-4",
			pattern: "/api/a../{version:\\d+}",
			method:  http.MethodGet,
			url:     "/api/a../2",
			status:  http.StatusOK,
			params:  map[string]string{"version": "2"},
		},
		&tester{
			name:    "正则匹配多个名称",
			pattern: "/api/{version:\\d+}/{name:\\w+}",
			method:  http.MethodGet,
			url:     "/api/2/login",
			status:  http.StatusOK,
			params:  map[string]string{"version": "2", "name": "login"},
		},
		&tester{
			name:    "正则不匹配多个名称",
			pattern: "/api/{version:\\d+}/{name:\\w+}",
			method:  http.MethodGet,
			url:     "/api/2.0/login",
			status:  http.StatusNotFound,
		},
		/*&tester{
			name:       "带域名的字符串不匹配", //无法匹配端口信息
			pattern:    "127.0.0.1/abc",
			method:http.MethodGet,
			url:      "/cba",
			status: 404,
		},
		&tester{
			name:       "带域名的正则匹配", //无法匹配端口信息
			pattern:    "127.0.0.1:{:\\d+}/abc",
			method:http.MethodGet,
			url:      "/abc",
			status: 200,
		},
		&tester{
			name:       "带域名的命名正则匹配", //无法匹配端口信息
			pattern:    "127.0.0.1:{:\\d+}/api/v{version:\\d+}/login",
			method:http.MethodGet,
			url:      "/api/v2/login",
			status: 200,
			params:     map[string]string{"version": "2"},
		},*/
	}

	runTester(a, tests)
}

func TestServeMux_Remove(t *testing.T) {
	a := assert.New(t)

	serveMux := NewServeMux(false)
	a.NotNil(serveMux)

	//
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
	/*f4 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(4)
	}
	f5 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(5)
	}
	f6 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(6)
	}*/

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

	serveMux := NewServeMux(false)
	a.NotNil(serveMux)
	serveMux.AddFunc("/post/", f1, "GET")          // f1
	serveMux.AddFunc("/post/{id:\\d+}", f2, "GET") // f2
	serveMux.AddFunc("/post/1", f3, "GET")         // f3
	//serveMux.AddFunc("127.0.0.1/post/", f4, "GET")          // f4
	//serveMux.AddFunc("127.0.0.1/post/{id:\\d+}", f5, "GET") // f5
	//serveMux.AddFunc("127.0.0.1/post/1", f6, "GET")         // f6

	test(serveMux, "GET", "", "/post/1", 3)   // f3 静态路由项完全匹配
	test(serveMux, "GET", "", "/post/2", 2)   // f2 正则完全匹配
	test(serveMux, "GET", "", "/post/abc", 1) // f1 匹配度最高
	//test(serveMux, "GET", "127.0.0.1", "/post/1", 6)   // f6 静态路由项完全匹配
	//test(serveMux, "GET", "127.0.0.1", "/post/2", 5)   // f5 正则完全匹配
	//test(serveMux, "GET", "127.0.0.1", "/post/abc", 4) // f4 匹配度最高
}

func TestMethodIsSupported(t *testing.T) {
	a := assert.New(t)

	a.True(MethodIsSupported("get"))
	a.True(MethodIsSupported("POST"))
	a.False(MethodIsSupported("not exists"))
}
