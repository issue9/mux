mux [![Build Status](https://travis-ci.org/issue9/mux.svg?branch=master)](https://travis-ci.org/issue9/mux)
======

mux是对http.Handler接口的一系列实现，提供了大部分实用的功能：
```go
var h1, h2, h3, h4 http.Handler

// 声明一个带method匹配的实例
m1 := mux.NewMethod().
          Get(mux.NewPath(h1, "api/logout")).
          Post(mux.NewPath(h2, "api/login"))

// 将srv和一个正则路由压入到m2中
m2 := mux.NewMethod().
          Get(mux.NewPath(h4, "/")).
          Any(mux.NewPath(h3, "/test"))

// 添加到各自的域名下
h1 := mux.NewHost(m1, "api.example.com")
h2 := mux.NewHost(m2, "(\\w+).example.com")

http.ListenAndServe("8080", NewMatches(h1, h2))
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
