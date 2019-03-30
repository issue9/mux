mux
[![Build Status](https://travis-ci.org/issue9/mux.svg?branch=master)](https://travis-ci.org/issue9/mux)
[![Go version](https://img.shields.io/badge/Go-1.11-brightgreen.svg?style=flat)](https://golang.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/issue9/mux)](https://goreportcard.com/report/github.com/issue9/mux)
[![codecov](https://codecov.io/gh/issue9/mux/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/mux)
[![GoDoc](https://godoc.org/github.com/issue9/mux?status.svg)](https://godoc.org/github.com/issue9/mux)
======

mux 是一个实现了 [http.Handler](https://godoc.org/net/http#Handler) 的中间件，为用户提供了以下功能：

1. 正则路由；
1. 路由参数；
1. 丰富的 OPTIONS 请求处理方式；
1. 自动生成 HEAD 请求内容；
1. 限定域名；
1. 根据路由生成地址。


```go
m := mux.New(false, false, false, nil, nil).
    Get("/users/1", h).             // GET /user/1
    Post("/login", h).              // POST /api/login
    Get("/posts/{id:\\d+}", h).     // GET /blog/post/{id:\\d+} 正则路由
    Options("/users/1", "GET").     // OPTIONS /user/1 手动指定该路由项的 OPTIONS 请求方法返回内容
    Get("example.com/users/1", h)   // 限定在 example.com 域名下的地址

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


#### 正则表达式

路由中支持以正则表达式的方式进行匹配，表达式以大括号包含，内部以冒号分隔，
前半部分为变量的名称，后半部分为变量可匹配类型的正则表达式。比如：
```go
/posts/{id:\\d+} // 将被转换成 /posts/(?P<id>\\d+)
/posts/{:\\d+}   // 将被转换成 /posts/\\d+
```


#### 命名参数

若路由字符串中，所有的正则表达式都只有名称部分（没有冒号及之后的内容），
则会被转换成命名参数，因为不需要作正则验证，性能会比较正则稍微好上一些。
命名参数匹配所有字符。
```go
/posts/{id}.html                  // 匹配 /posts/1.html
/posts-{id}-{page}.html           // 匹配 /posts-1-10.html
```


#### 通配符

在路由字符串中若是以命名参数结尾的，则表示可以匹配之后的任意字符。
```go
/blog/assets/{path}
/blog/{tags:\\w+}/{path}
/blog/assets{path}
```


#### 路径匹配规则

可能会出现多条记录与同一请求都匹配的情况，这种情况下，
系统会找到一条认为最匹配的路由来处理，判断规则如下：
 1. 普通路由优先于正则路由；
 1. 正则路由优先于命名路由；

比如：
```go
/posts/{id}.html              // 1
/posts/{id:\\d+}.html         // 2
/posts/1.html                 // 3

/posts/1.html      // 匹配 3
/posts/11.html     // 匹配 2
/posts/index.html  // 匹配 1
```


#### 域名限定

如果路由以非 / 开头，则自动将第一个 / 之前的判定为域名，并对其进行域名限定。
比如以下格式：
```go
example.com/html/{id}.html   // 匹配 example.com/html/1.html
*.example.com/html/{id}.html // 支持泛域名，匹配 s1.example.com/html/1.html
```


#### 路由参数

通过正则表达式匹配的路由，其中带命名的参数可通过 Params() 获取：
```go
params := Params(r)

id, err := params.Int("id")
 // 或是
id := params.MustInt("id", 0) // 0 表示在无法获取 id 参数的默认值
```


#### OPTIONS

默认情况下，用户无须显示地实现它，系统会自动实现。
当然用户也可以使用 *.Options() 函数指定特定的数据；
或是直接使用 *.Handle() 指定一个自定义的实现方式。

如果不需要的话，也可以在 New() 中将 disableOptions 设置为 true。
显示设定 OPTIONS，不受 disableOptions 的影响。
```go
m := mux.New(...)
m.Get("/posts/{id}", nil)     // 默认情况下， OPTIONS 的报头为 GET, OPTIONS
m.Options("/posts/{id}", "*") // 强制改成 *
m.Delete("/posts/{id}", nil)  // OPTIONS 依然为 *

m.Remove("/posts/{id}", http.MethodOptions)    // 在当前路由上禁用 OPTIONS
m.Handle("/posts/{id}", h, http.MethodOptions) // 显示指定一个处理函数 h
```


#### HEAD

 默认情况下，用户无须显示地实现 HEAD 请求，
 系统会为每一个 GET 请求自动实现一个对应的 HEAD 请求，
 当然也与 OPTIONS 一样，你也可以自通过 mux.Handle() 自己实现 HEAD 请求。



### 中间件

mux 本身就是一个实现了 [http.Handler](https://godoc.org/net/http#Handler) 接口的中间件，
所有实现官方接口 `http.Handler` 的都可以附加到 mux 上作为中间件使用。

[middleware](https://github.com/issue9/middleware) 提供了诸如按域名过滤等常用的中间件功能。



### 版权

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
