// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// mux提供了一系列实现了http.Handler接口的中间件。
//
// 一个多域名路由的示例：
//  var h1, h2 http.Handler
//
//  // 声明一个带method匹配的实例
//  m := mux.NewServerMux2().
//            Get("api/logout", h1).
//            Post("www.example/api/login", h2)
//
//  http.ListenAndServe("8080", m)
package mux

const Version = "0.6.14.150405"
