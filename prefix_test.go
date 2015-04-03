// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"strings"
	"testing"

	"github.com/issue9/assert"
)

func TestPrefix_Add(t *testing.T) {
	a := assert.New(t)

	p := NewPrefix()
	a.NotNil(p)

	f := func(w http.ResponseWriter, r *http.Request) {}
	h := http.HandlerFunc(f)

	err := p.Add("", h)
	a.NotError(err)

	// 不能多次指定prefix为空的情况
	a.Error(p.Add("", h))

	// 返回h为空的错误信息
	err = p.Add("/admin", nil)
	a.True(strings.Index(err.Error(), "h") > -1)

	// 成功添加一个
	a.NotError(p.Add("/admin", h))
	a.Equal(1, len(p.items))

	// 返回重复prefix的错误信息
	a.Error(p.AddFunc("/admin", f))

	// 再次添加一条记录
	a.NotError(p.AddFunc("/ADMIN", f))
	a.Equal(2, len(p.items))
}

func TestPrefix_ServeHTTP(t *testing.T) {
	a := assert.New(t)

	h := NewPrefix()
	h.AddFunc("/admin", defaultHandler)
	h.AddFunc("/api", defaultHandler)
	// 指定了默认处理函数，不会触发errHandler
	h.AddFunc("", func(w http.ResponseWriter, req *http.Request) {
		// 可以返回任意值，只要能与errHandler和defaultHandler区分开就行
		w.WriteHeader(403)
		w.Write([]byte("no prefix"))
	})

	tests := []*handlerTester{
		&handlerTester{
			name:       "普通匹配1",
			h:          h,
			query:      "/admin/",
			response:   "OK",
			statusCode: 200,
		},
		&handlerTester{
			name:       "普通匹配2",
			h:          h,
			query:      "/admin",
			response:   "OK",
			statusCode: 200,
		},
		&handlerTester{
			name:       "普通匹配3",
			h:          h,
			query:      "/api",
			response:   "OK",
			statusCode: 200,
		},
		&handlerTester{
			name:       "普通匹配4",
			h:          h,
			query:      "/API",
			response:   "no prefix",
			statusCode: 403,
		},
		&handlerTester{
			name:       "普通匹配5",
			h:          h,
			query:      "/a/api",
			response:   "no prefix",
			statusCode: 403,
		},
	}

	runHandlerTester(a, tests)

	//////////////// 未指定默认处理函数，会触发panic，从而返回404错误
	h = NewPrefix()
	h.AddFunc("/admin", defaultHandler)

	tests = []*handlerTester{
		&handlerTester{
			name:       "普通匹配1",
			h:          h,
			query:      "/Admin/",
			response:   "error",
			statusCode: 404,
		},
	}
	runHandlerTester(a, tests)
}
