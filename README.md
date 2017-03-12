mux [![Build Status](https://travis-ci.org/issue9/mux.svg?branch=master)](https://travis-ci.org/issue9/mux)
======

mux是对http.ServeMux的扩展，添加正则路由等功能。
```go
m := mux.NewServerMux(false).
    Get("/user/1", h). // 不限定域名，必须以/开头
    Post("www.example/api/login", h). // 限定了域名
    Get("/blog/post/{id:\\d+}", h). // 正则路由
    Options("/user/1", "GET") // 手动指定该路由项的OPTIONS请求方法返回内容

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
