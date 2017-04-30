// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package mux 提供了相对 http.ServeMux 更加复杂的路径匹配功能。
// 同时又兼容官方的 http.Handler 接口。
//
//  m := mux.New(false, nil, nil).
//          Get("/user/1", h).
//          Post("/api/login", h).
//          Get("/blog/post/{id:\\d+}", h). // 正则路由
//          Options("/user/1", "GET") // 手动指定 OPTIONS 请求的返回内容。
//
//  // 统一前缀路径的路由
//  p := m.Prefix("/api")
//  p.Get("/logout", h) // 相当于 m.Get("/api/logout", h)
//  p.Post("/login", h) // 相当于 m.Get("/api/login", h)
//
//  对同一资源的不同操作
//  res := p.Resource("/users/{id\\d+}")
//  res.Get(h)   // 相当于 m.Get("/api/users/{id}", h)
//  res.Post(h)  // 相当于 m.Post("/api/users/{id}", h)
//
//  http.ListenAndServe(":8080", m)
//
//
// 正则表达式：
//
// 路由中支持以正则表达式的方式进行匹配，表达式以大括号包含，内部以冒号分隔，
// 前半部分为变量的名称，后半部分为变量可匹配类型的正则表达式。比如：
//  /post/{id}     // 匹配 /post/ 开头的任意字符串，其后的字符串保存到 id 中；
//  /post/{id:\d+} // 同上，但 id 的值只能为 \d+，\d+ 为正则表达式；
//  /post/{:\d+}   // 同上，但是没有命名；
// 正则表达式可以使用 * 表示这是一个可选参数，但是不能出现在路径中间：
//  /post/{id:\d*}        // 匹配 /post/ 和 /post/1 等
//  /post/{id:\d*}/author // 只能匹配 /post/1/author 等，但不会匹配 /post/author
//
//
// 路由参数：
//
// 通过正则表达式匹配的路由，其中带命名的参数可通过 r.Context() 获取：
//  params := GetParams(r)
//  id, err := params.Int("id")
//  // 或是
//  id := params.MustInt("id", 0) // 0 表示在无法获取 id 参数的默认值
//
//
// 路径匹配规则：
//
// 可能会出现多条记录与同一请求都匹配的情况，这种情况下，
// 系统会找到一条认为最匹配的路由来处理，判断规则如下：
//  1. 静态路由优先于正则路由判断；
//  2. 完全匹配的路由项优先于部分匹配的路由项；
//  3. 只有以 / 结尾的静态路由才有部分匹配功能；
//  4. 同类的后插入先匹配。
//
//
// OPTIONS:
//
// OPTIONS 请求是一个比较特殊的存在，默认情况下，用户无须显示地实现它，
// 系统会自动实现。当然用户也可以使用 *.Options() 函数指定特定的数据；
// 或是直接使用 *.Add() 指定一个自定义的实现方式。
// 如果不需要的话，也可以在 New() 中将 disableOptions 设置为 true。
package mux
