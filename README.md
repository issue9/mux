mux [![Build Status](https://travis-ci.org/issue9/mux.svg?branch=master)](https://travis-ci.org/issue9/mux)
======

mux是对http.Handler接口的一系列实现，提供了大部分实用的功能：
```go
m := mux.NewServerMux().
        Get("/user/logout", h). // 不限定域名，必须以/开头
        Post("www.example/api/login", h) // 限定了域名
        Get("/blog/post/{id:\\d+}", h) // 正则路由

// 统一前缀名称的路由
p := m.Prefix("/api")
p.Get("/logout", h) // 相当于mux.Get("/api/logout", h)
p.Post("/login", h) // 相当于mux.Get("/api/login", h)

// 分组路由，该分组可以在运行过程中控制是否暂停
g := m.Group("admin")
g.Get("/admin", h).
    Get("/api/admin/logout").
    Post("/api/admin/login")

http.ListenAndServe("8080", m)
```

### 安装

```shell
go get github.com/issue9/mux
```


### 文档

[![Go Walker](http://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/issue9/mux)
[![GoDoc](https://godoc.org/github.com/issue9/mux?status.svg)](https://godoc.org/github.com/issue9/mux)


### 版权

本项目采用[MIT](http://opensource.org/licenses/MIT)开源授权许可证，完整的授权说明可在[LICENSE](LICENSE)文件中找到。
