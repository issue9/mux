mux [![Build Status](https://travis-ci.org/issue9/mux.svg?branch=master)](https://travis-ci.org/issue9/mux)
======

mux是对http.Handler接口的一系列实现，提供了大部分实用的功能：
```go
var h1, h2, h3, h4 http.Handler

// 声明一个带method匹配的实例
m1 := mux.NewMethod().
          MustGet("api/logout", h1).
          MustPost("api/login", h2)

srv := http.NewServeMux()
srv.Handle(h3, "/")

// 添加到各自的域名下
h := mux.NewHost()
h.Add("api.example.com", m1)
h.Add("?(\\w+).example.com", srv)

http.ListenAndServe("8080", h)
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
