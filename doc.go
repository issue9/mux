// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// mux提供了一系列实现了http.Handler接口的中间件。
//
//  m := mux.NewServerMux().
//          Get("/user/logout", h). // 不限定域名，必须以/开头
//          Post("www.example/api/login", h) // 限定了域名
//          Get("/blog/post/{id:\\d+}", h) // 正则路由
//
//  // 统一前缀名称的路由
//  p := m.Prefix("/api")
//  p.Get("/logout", h) // 相当于m.Get("/api/logout", h)
//  p.Post("/login", h) // 相当于m.Get("/api/login", h)
//
//  // 分组路由，该分组可以在运行过程中控制是否暂停
//  g := m.Group("admin")
//  g.Get("/admin", h).
//      Get("/api/admin/logout").
//      Post("/api/admin/login")
//
//  h := mux.NewReocvery(m, nil)
//  http.ListenAndServe("8080", h)
package mux
