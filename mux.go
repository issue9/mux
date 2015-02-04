// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// mux提供了一系列http.Handler接口的实现：
// 多域名匹配(Host)、根据请求方法匹配(Method)等操作。
//
// 一个多域名路由的示例：
//  var h1, h2, h3, h4 http.Handler
//
//  // 声明一个带method匹配的实例
//  m1 := mux.NewMethod(nil).
//            MustGet("api/logout", h1).
//            MustPost("api/login", h2)
//
//  // net/http包里的默认ServeMux实例
//  srv := http.NewServeMux()
//  srv.Handle(h3, "/")
//
//  // 添加到各自的域名下
//  h := mux.NewHost(nil)
//  h.Handle("api.example.com", m1)
//  h.Handle("?(\\w+).example.com", srv)
//
//  http.ListenAndServe("8080", h)
package mux

import (
	"net/http"
)

const Version = "0.4.6.150204"

// 错误状态处理函数。
//
// msg详细的错误信息；code错误状态码。
type ErrorHandler func(w http.ResponseWriter, msg string, code int)

// 默认的ErrorHandler实现，直接调用http.Error()实现。
func defaultErrorHandler(w http.ResponseWriter, msg string, code int) {
	http.Error(w, msg, code)
}
