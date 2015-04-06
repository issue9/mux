// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
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
	a.Error(m.Add("?", h, "GET"))

	// methods为空
	a.Error(m.Add("abc", h))

	// 向Get和Post添加一个路由abc
	a.NotError(m.Add("abc", h, "GET", "POST"))
	_, found := m.methods["GET"]
	a.True(found)
	_, found = m.methods["POST"]
	a.True(found)
	_, found = m.methods["DELETE"]
	a.False(found)

	// 再次向Get添加一条同名路由，会出错
	a.Error(m.Get("abc", h))

	a.NotError(m.Get("def", h))
	es, found := m.methods["GET"]
	a.True(found).Equal(2, len(es.list))

	// Delete
	a.NotError(m.Delete("abc", h))
	es, found = m.methods["DELETE"]
	a.True(found).Equal(1, len(es.list))
	a.NotError(m.DeleteFunc("abcd", fn))
	a.True(found).Equal(2, len(es.list))

	//Put
	a.NotError(m.Put("abc", h))
	es, found = m.methods["PUT"]
	a.True(found).Equal(1, len(es.list))
	a.NotError(m.PutFunc("abcd", fn))
	a.True(found).Equal(2, len(es.list))
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
			name:       "普通不匹配",
			h:          newServeMux("/abc"),
			query:      "/abcd",
			statusCode: 404,
		},
		&handlerTester{
			name:       "正则匹配数字",
			h:          newServeMux("?/api/(?P<version>\\d+)"),
			query:      "/api/2",
			statusCode: 200,
			ctxName:    "params",
			ctxMap:     map[string]string{"version": "2"},
		},
		&handlerTester{
			name:       "正则匹配多个名称",
			h:          newServeMux("?/api/(?P<version>\\d+)/(?P<name>\\w+)"),
			query:      "/api/2/login",
			statusCode: 200,
			ctxName:    "params",
			ctxMap:     map[string]string{"version": "2", "name": "login"},
		},
		&handlerTester{
			name:       "正则不匹配多个名称",
			h:          newServeMux("?/api/(?P<version>\\d+)/(?P<name>\\w+)"),
			query:      "/api/2.0/login",
			statusCode: 404,
		},
		&handlerTester{
			name:       "带域名的字符串不匹配", //无法匹配端口信息
			h:          newServeMux("127.0.0.1/abc"),
			query:      "/abc",
			statusCode: 404,
		},
		&handlerTester{
			name:       "带域名的正则匹配", //无法匹配端口信息
			h:          newServeMux("?127.0.0.1:\\d+/abc"),
			query:      "/abc",
			statusCode: 200,
		},
		&handlerTester{
			name:       "带域名的命名正则匹配", //无法匹配端口信息
			h:          newServeMux("?127.0.0.1:\\d+/api/v(?P<version>\\d+)/login"),
			query:      "/api/v2/login",
			statusCode: 200,
			ctxName:    "params",
			ctxMap:     map[string]string{"version": "2"},
		},
	}

	runHandlerTester(a, tests)
}
