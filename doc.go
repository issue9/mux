// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// mux提供了一系列http.Handler接口的实现，方便用户进行正则路由匹
// 配(Path)、多域名匹配(Host)、根据请求方法匹配(Method)等操作。
//
// 一个多域名路由的示例：
//  var h1, h2, h3, h4 http.Handler
//
//  // 声明一个带method匹配的实例
//  m1 := mux.NewMethod().
//            Get(mux.NewPath(h1, "api/logout")).
//            Post(mux.NewPath(h2, "api/login"))
//
//  // net/http包里的默认ServeMux实例
//  srv := http.NewServeMux()
//  srv.Handle(h3, "/")
//
//  // 将srv和一个正则路由压入到m2中
//  m2 := mux.NewMethod().
//            Get(mux.NewPath(h4, "/")).
//            Any(mux.Handler2Matcher(srv))
//
//  // 添加到各自的域名下
//  h1 := mux.NewHost(m1, "api.example.com")
//  h2 := mux.NewHost(m2, "(\\w+).example.com")
//
//  http.ListenAndServe("8080", NewMatches(h1, h2))
package mux

const Version = "0.1.7.140922"
