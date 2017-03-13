mux [![Build Status](https://travis-ci.org/issue9/mux.svg?branch=master)](https://travis-ci.org/issue9/mux)
======

mux 是对 http.ServeMux 的扩展，添加正则路由等功能。

相对于 http.ServeMux 提供了以下功能：
1. 正则路由；
1. 自动生成 OPTIONS；


通过与 [handlers](https://github.com/issue9/handlers) 还可以实现诸如按域名过滤等功能。


```go
m := mux.NewServerMux(false).
    Get("/user/1", h).              // GET /user/1
    Post("/api/login", h).          // POST /api/login
    Get("/blog/post/{id:\\d+}", h). // GET /blog/post/{id:\d+} 正则路由
    Options("/user/1", "GET")       // OPTIONS /user/1 手动指定该路由项的 OPTIONS 请求方法返回内容

// 统一前缀名称的路由
p := m.Prefix("/api")
p.Get("/logout", h) // 相当于m.Get("/api/logout", h)
p.Post("/login", h) // 相当于m.Get("/api/login", h)

http.ListenAndServe("8080", m)
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
