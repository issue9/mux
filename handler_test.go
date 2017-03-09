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
	name  string       // 该测试组的名称，方便定位
	h     http.Handler // 用于测试的 http.Handler 实例
	query string       // 访问测试所用的查询字符串

	statusCode int // 通过返回的状态码，判断是否是需要的值。

	ctxName string            // h 在 context 中设置的变量名称，若没有，则为空值。
	ctxMap  map[string]string // 以及该变量对应的值
}

type ctxHandler struct {
	a    *assert.Assertion
	test *handlerTester
	h    http.Handler
}

func (ch *ctxHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ch.h.ServeHTTP(w, req)

	/* 在正主执行 ServeHTTP 之后，才会有 context 存在 */

	params := req.Context().Value(ContextKeyParams).(Params)
	mapped, found := params[ch.test.ctxName]
	ch.a.True(found)

	errStr := "在执行[%v]时，context参数[%v]与预期值[%v]不相等"
	ch.a.Equal(mapped, ch.test.ctxMap, errStr, ch.test.name, mapped, ch.test.ctxMap)
}

// 运行一组handlerTester测试内容
func runHandlerTester(a *assert.Assertion, tests []*handlerTester) {
	for _, test := range tests {
		if len(test.ctxName) > 0 {
			test.h = &ctxHandler{
				a:    a,
				test: test,
				h:    test.h,
			}
		}

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

// 默认的handler，向response输出ok。
func defaultHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
}
