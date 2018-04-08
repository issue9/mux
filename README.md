mux [![Build Status](https://travis-ci.org/issue9/mux.svg?branch=master)](https://travis-ci.org/issue9/mux)
[![Go version](https://img.shields.io/badge/Go-1.8-brightgreen.svg?style=flat)](https://golang.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/issue9/mux)](https://goreportcard.com/report/github.com/issue9/mux)
[![codecov](https://codecov.io/gh/issue9/mux/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/mux)
======

mux 是一个实现了 [http.Handler](https://godoc.org/net/http#Handler) 的中间件，为用户提供了以下功能：

1. 正则路由；
1. 路由参数；
1. 丰富的 OPTIONS 请求处理方式；
1. 根据路由生成地址。


##### 中间件

mux 本身就是一个实现了 [http.Handler](https://godoc.org/net/http#Handler) 接口的中间件，
所有实现官方接口 `http.Handler` 的都可以附加到 mux 上作为中间件使用。

[middleware](https://github.com/issue9/middleware) 提供了诸如按域名过滤等常用的中间件功能。


```go
m := mux.New(false, false, nil, nil).
    Get("/users/1", h).             // GET /user/1
    Post("/login", h).              // POST /api/login
    Get("/posts/{id:\\d+}", h).     // GET /blog/post/{id:\\d+} 正则路由
    Options("/users/1", "GET")      // OPTIONS /user/1 手动指定该路由项的 OPTIONS 请求方法返回内容

// 统一前缀路径的路由
p := m.Prefix("/api")
p.Get("/logout", h) // 相当于 m.Get("/api/logout", h)
p.Post("/login", h) // 相当于 m.Get("/api/login", h)

// 对同一资源的不同操作
res := p.Resource("/users/{id:\\d+}")
res.Get(h)   // 相当于 m.Get("/api/users/{id:\\d+}", h)
res.Post(h)  // 相当于 m.Post("/api/users/{id:\\d+}", h)
res.URL(map[string]string{"id": "5"}) // 构建一条基于此路由项的路径：/users/5

http.ListenAndServe(":8080", m)
```


### 安装

```shell
go get github.com/issue9/mux
```


### 文档

[![Go Walker](https://gowalker.org/api/v1/badge)](https://gowalker.org/github.com/issue9/mux)
[![GoDoc](https://godoc.org/github.com/issue9/mux?status.svg)](https://godoc.org/github.com/issue9/mux)


### 版权

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
