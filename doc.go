// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package mux 提供了相对 http.ServeMux 更加复杂的路径功能。
//
//  m := mux.NewServerMux().
//          Get("/user/1", h). // 不限定域名，必须以/开头
//          Post("www.example/api/login", h). // 限定了域名
//          Get("/blog/post/{id:\\d+}", h). // 正则路由
//          Options("/user/1", "GET") // 手动指定 OPTIONS 请求的返回内容。
//
//  // 统一前缀名称的路由
//  p := m.Prefix("/api")
//  p.Get("/logout", h) // 相当于m.Get("/api/logout", h)
//  p.Post("/login", h) // 相当于m.Get("/api/login", h)
//
//  h := mux.NewReocvery(m, nil)
//  http.ListenAndServe("8080", h)
//
//
// OPTIONS:
//
// OPTIONS 请求是一个比较特殊的存在，默认情况下，用户无须显示地实现它，
// 系统会自动实现。当然用户也可以使用 ServeMux.Options() 函数指定特定的
// 的数据；或是直接使用 ServeMux.Add() 指定一个自定义的实现方式。
package mux
