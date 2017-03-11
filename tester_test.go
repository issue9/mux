// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/issue9/assert"
	"github.com/issue9/handlers"
)

// http.Handler 测试工具，测试 h 返回值是否与 response 相同。
type tester struct {
	// 以下为构建内容
	name    string // 该测试组的名称，方便定位
	pattern string // 地址匹配模式

	// 以下为请求及返回的内容
	url     string            // 访问测试所用的地址
	status  int               // 期望的返回结果
	params  map[string]string // 期望返回的地址参数
	headers map[string]string // 期望返回的报头
}

// 运行一组 tester 测试内容
func runTester(a *assert.Assertion, tests []*tester) {

	// 包含一个默认的错误处理函数，用于在出错时，输出 error 字符串.

	// 依次检测各个测试用例
	for _, test := range tests {
		srvmux := NewServeMux(false)
		a.NotNil(srvmux)
		a.NotError(srvmux.GetFunc(test.pattern, defaultHandler))

		srv := httptest.NewServer(handlers.Recovery(srvmux, errHandler))
		a.NotNil(srv)
		defer srv.Close()

		resp, err := http.Get(srv.URL + test.url)
		a.NotError(err).NotNil(resp)

		// 状态码是否相等
		errstr := "在执行[%v]时，其返回的状态码[%v]与预期值[%v]不相等;"
		a.Equal(resp.StatusCode, test.status, errstr, test.name, resp.StatusCode, test.status)

		// 报头
		errstr = "在执行[%v]时，其返回的报头[%v:%v]与预期值[%v]不相等;"
		for k, v := range test.headers {
			a.Equal(v, resp.Header.Get(k), errstr, test.name, k, resp.Header.Get(k), v)
		}

		// 返回参数
		if test.params != nil {
			errstr = "在执行[%v]时，其返回的参数[%v:%v]与预期值[%v]不相等;"
			data, err := ioutil.ReadAll(resp.Body)
			a.NotError(err).NotNil(data)

			params := map[string]string{}
			a.NotError(json.Unmarshal([]byte(data), &params))
			a.NotError(err)

			for k, v := range test.params {
				a.Equal(params[k], v, errstr, test.name, k, params[k], v)
			}
		}
	} // end for
}

func errHandler(w http.ResponseWriter, msg interface{}) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, msg)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	ps := r.Context().Value(ContextKeyParams)
	if ps == nil {
		return
	}

	mapped, ok := ps.(Params)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(mapped)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
