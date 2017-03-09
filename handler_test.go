// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/issue9/assert"
	"github.com/issue9/handlers"
)

// http.Handler 测试工具，测试h返回值是否与 response 相同。
type handlerTester struct {
	name    string       // 该测试组的名称，方便定位
	h       http.Handler // 用于测试的 http.Handler 实例
	query   string       // 访问测试所用的查询字符串
	pattern string       // 地址匹配模式

	statusCode int // 通过返回的状态码，判断是否是需要的值。

	params map[string]string // 路径中的参数列表
}

// 运行一组 handlerTester 测试内容
func runHandlerTester(a *assert.Assertion, tests []*handlerTester) {
	newServeMux := func(t *handlerTester) http.Handler {
		h := NewServeMux()
		a.NotError(h.Add(t.pattern, buildDefaultHandler(a, t), "GET"))
		return h
	}

	for _, test := range tests {
		test.h = newServeMux(test)

		// 包含一个默认的错误处理函数，用于在出错时，输出 error 字符串.
		srv := httptest.NewServer(handlers.Recovery(test.h, errHandler))
		a.NotNil(srv)

		resp, err := http.Get(srv.URL + test.query)
		a.NotError(err).NotNil(resp)

		msg, err := ioutil.ReadAll(resp.Body)
		a.NotError(err)
		errStr := "在执行[%v]时，其返回的状态码[%v]与预期值[%v]不相等;提示信息为：[%v]"
		a.Equal(resp.StatusCode, test.statusCode, errStr, test.name, resp.StatusCode, test.statusCode, string(msg))

		srv.Close() // 在for的最后关闭当前的srv
	}
}

func errHandler(w http.ResponseWriter, msg interface{}) {
	w.WriteHeader(404)
	fmt.Fprint(w, msg)
}

func buildDefaultHandler(a *assert.Assertion, t *handlerTester) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if t.params != nil {
			ps := r.Context().Value(ContextKeyParams)
			a.NotNil(ps, "在执行[%v]时，无法读取 ContextKeyParms 的值", t.name)

			mapped, ok := ps.(Params)
			a.True(ok, "在执行[%v]时，无法将值转换成 Params 类型", t.name).NotNil(mapped)

			errStr := "在执行[%v]时，context参数[%v]与预期值[%v]不相等"
			a.Equal(mapped, t.params, errStr, t.name, mapped, t.params)
		}
		w.WriteHeader(http.StatusOK)
	})
}
