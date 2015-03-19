// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/issue9/assert"
	"github.com/issue9/context"
)

// http.Handler测试工具，测试h返回值是否与response相同。
type handlerTester struct {
	name  string       // 该测试组的名称，方便定位
	h     http.Handler // 用于测试的http.Handler实例
	query string       // 访问测试所用的查询字符串

	// 返回值部分
	response   string // 从h返回的response内容，仅包含主体部分，不包含头信息
	statusCode int    // 返回的状态码

	ctxName string            // h在context中设置的变量名称，若没有，则为空值。
	ctxMap  map[string]string // 以及该变量对应的值
}

type ctxHandler struct {
	a    *assert.Assertion
	test *handlerTester
	h    http.Handler
}

func (ch *ctxHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ch.h.ServeHTTP(w, req)

	/* 在正主执行ServeHTTP之后，才会有context存在 */

	ctx := context.Get(req)
	mapped, found := ctx.Get(ch.test.ctxName)
	ch.a.True(found)

	data, ok := mapped.(map[string]string)
	ch.a.True(ok, "在执行[%v]时，无法获取其Context相关参数", ch.test.name)
	errStr := "在执行[%v]时，context参数[%v]与预期值不相等"
	ch.a.Equal(data, ch.test.ctxMap, errStr, ch.test.name, data, ch.test.ctxMap)
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

		// 包含一个默认的错误处理函数，用于在出错时，输出error字符串.
		srv := httptest.NewServer(NewRecovery(test.h, errHandler))
		a.NotNil(srv)

		resp, err := http.Get(srv.URL + test.query)
		a.NotError(err).NotNil(resp)

		// 比较statusCode
		errStr := "在执行[%v]时，其返回的状态码不相等，预期值:[%v]，实际值:[%v]"
		a.Equal(resp.StatusCode, test.statusCode, errStr, test.name, test.statusCode, resp.StatusCode)

		// 比较response
		p, err := ioutil.ReadAll(resp.Body)
		a.NotError(err)
		errStr = "在执行[%v]时，其返回的内容不相等，预期值:[%v]，实际值:[%v]"
		a.Equal(p, []byte(test.response), errStr, test.name, test.response, string(p))

		srv.Close() // 在for的最后关闭当前的srv
	}
}

func errHandler(w http.ResponseWriter, msg interface{}) {
	w.WriteHeader(404)
	w.Write([]byte("error"))
}

// 默认的handler，向response输出ok。
func defaultHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}
