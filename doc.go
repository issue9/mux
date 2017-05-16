// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package mux 是一个实现了 http.Handler 的中间件，为用户提供了路由匹配等功能。
//
//  m := mux.New(false, false, nil, nil).
//          Get("/users/1", h).
//          Post("/login", h).
//          Get("/posts/{id:\\d+}", h).  // 正则路由
//          Options("/users/1", "GET")   // 手动指定 OPTIONS 请求的返回内容。
//
//  // 统一前缀路径的路由
//  p := m.Prefix("/api")
//  p.Get("/logout", h) // 相当于 m.Get("/api/logout", h)
//  p.Post("/login", h) // 相当于 m.Get("/api/login", h)
//
//  // 对同一资源的不同操作
//  res := p.Resource("/users/{id\\d+}")
//  res.Get(h)   // 相当于 m.Get("/api/users/{id}", h)
//  res.Post(h)  // 相当于 m.Post("/api/users/{id}", h)
//  res.URL(map[string]string{"id": "5"}, "") // /users/5
//
//  http.ListenAndServe(":8080", m)
//
//
//
// 正则表达式
//
// 路由中支持以正则表达式的方式进行匹配，表达式以大括号包含，内部以冒号分隔，
// 前半部分为变量的名称，后半部分为变量可匹配类型的正则表达式。比如：
//  /post/{id:\\d+} // id 的值只能为 \d+，\d+ 为正则表达式；
//  /post/{:\\d+}   // 同上，但是没有命名；
//
//
//
// 命名参数
//
// 若路由字符串中，所有的正则表达式都只有名称部分（没有冒号及之后的内容），
// 则会被转换成命名参数，因为不需要作正则验证，性能会比较正则稍微好上一些。
//  /posts/{id}                // 命名参数
//  /blog/{action}/{page:\\d+} // 正则，page 使用了正则匹配
//
//
//
// 通配符
//
// 在路由字符串最后加上 /* 可以启到通配符的作用，但是不可用于中间。
//  /blog/assets/*
//  /blog/{posts}/*
//  /blog/{tags:\\w+}/*
//  /blog/*/assets      // 不正确
//  /blog/assets*       // 不正确
//
//
//
// 路由参数：
//
// 通过正则表达式匹配的路由，其中带命名的参数可通过 GetParams() 获取：
//  params := GetParams(r)
//
//  id, err := params.Int("id")
//  // 或是
//  id := params.MustInt("id", 0) // 0 表示在无法获取 id 参数的默认值
//
//
//
// 路径匹配规则：
//
// 可能会出现多条记录与同一请求都匹配的情况，这种情况下，
// 系统会找到一条认为最匹配的路由来处理，判断规则如下：
//  1. 普通路由优先于正则路由判断；
//  2. 正则路由优先于命名路由；
//  3. 完全匹配的路由项优先于有通配符的路由；
//  4. 带通配符的各类路由，按照 1、2 规则执行。
//
//
//
// OPTIONS
//
// 默认情况下，用户无须显示地实现它，系统会自动实现。
// 当然用户也可以使用 *.Options() 函数指定特定的数据；
// 或是直接使用 *.Add() 指定一个自定义的实现方式。
//
// 如果不需要的话，也可以在 New() 中将 disableOptions 设置为 true。
// 通过 *.Add 和 *.Remove 来显示的指定或是删除 OPTIONS，不受是否禁用的影响。
//  m := mux.New(...)
//  m.Get("/posts/{id}", nil)     // 默认情况下， OPTIONS 的报头为 GET, OPTIONS
//  m.Options("/posts/{id}", "*") // 强制改成 *
//  m.Delete("/posts/{id}", nil)  // OPTIONS 依然为 *
//
//  m.Remove("/posts/{id}", http.MethodOptions) // 在当前路由上禁用 OPTIONS
//  m.Add("/posts/{id}", h, http.MethodOptions) // 显示指定一个处理函数 h
//
//
//
// 适用范围
//
// 由于路由项采用了切片(slice) 的形式保存路由项，
// 如果在运行过程中需要大量的增删路由操作，性能上会比较差，
// 建议使用其它的库的代替。其它情况下，性能还是不错的，
// 具体的可运行 `go test -bench=.` 查看。
package mux // import "github.com/issue9/mux"
