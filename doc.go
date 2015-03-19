// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// mux提供了一系列http.Handler接口的实现。
// 以及一个根据http.Request切换的Context。
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

const Version = "0.5.9.150318"
