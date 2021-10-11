# mux

[![Go](https://github.com/issue9/mux/workflows/Go/badge.svg)](https://github.com/issue9/mux/actions?query=workflow%3AGo)
[![Go version](https://img.shields.io/github/go-mod/go-version/issue9/mux)](https://golang.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/issue9/mux)](https://goreportcard.com/report/github.com/issue9/mux)
[![license](https://img.shields.io/github/license/issue9/mux)](LICENSE)
[![codecov](https://codecov.io/gh/issue9/mux/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/mux)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/issue9/mux/v5)](https://pkg.go.dev/github.com/issue9/mux/v5)

mux 功能完备的 Go 路由器：

1. 路由参数；
1. 支持正则表达式作为路由项匹配方式；
1. 自动生成 OPTIONS 请求处理方式；
1. 自动生成 HEAD 请求处理方式；
1. 根据路由反向生成地址；
1. 任意风格的路由，比如 discuz 这种不以 / 作为分隔符的；
1. 分组路由，比如按域名，或是版本号等；
1. CORS 跨域资源的处理；
1. 支持中间件；

```go
import "github.com/issue9/middleware/v4/compress"
import "github.com/issue9/mux/v5"

c := compress.New()

router := mux.DefaultRouter()
router.Middlewares().Append(c)
router.Get("/users/1", h).
    Post("/login", h).
    Get("/pages/{id:\\d+}.html", h). // 匹配 /pages/123.html 等格式，path = 123
    Get("/posts/{path}.html", h).    // 匹配 /posts/2020/11/11/title.html 等格式，path = 2020/11/11/title

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

## 语法

### 正则表达式

路由中支持以正则表达式的方式进行匹配，表达式以大括号包含，内部以冒号分隔，
前半部分为变量的名称，后半部分为变量可匹配类型的正则表达式。比如：

```text
/posts/{id:\\d+} // 将被转换成 /posts/(?P<id>\\d+)
/posts/{:\\d+}   // 将被转换成 /posts/\\d+
```

### 命名参数

若路由字符串中，所有的正则表达式冒号之后的内容是特定的内容，或是无内容，
则会被转换成命名参数，因为有专门的验证方法，性能会比较正则稍微好上一些。

```text
 /posts/{id}.html                  // 匹配 /posts/1.html
 /posts-{id}-{page}.html           // 匹配 /posts-1-10.html
 /posts/{id:digit}.html            // 匹配 /posts/1.html
 /posts/{path}.html                // 匹配 /posts/2020/11/11/title.html
```

目前支持以下作为命名参数的内容约束：

- digit 限定为数字字符，相当于正则的 `[0-9]`；
- word 相当于正则的 `[a-zA-Z0-9]`；
- any 表示匹配任意非空内容；
- "" 为空表示任意内容，包括非内容；

用户也可以自行添加新的约束符。具体可参考 <https://pkg.go.dev/github.com/issue9/mux/v5/interceptor>

在路由字符串中若是以命名参数结尾的，则表示可以匹配之后的任意字符。

```text
/blog/assets/{path}       // 可以匹配 /blog/assets/2020/11/11/file.ext 等格式
/blog/{tags:\\w+}/{path}
/blog/assets{path}
```

### 路径匹配规则

可能会出现多条记录与同一请求都匹配的情况，这种情况下，
系统会找到一条认为最匹配的路由来处理，判断规则如下：

 1. 普通路由优先于正则路由；
 1. 正则路由优先于命名路由；

比如：

```text
/posts/{id}.html              // 1
/posts/{id:\\d+}.html         // 2
/posts/1.html                 // 3

/posts/1.html      // 匹配 3
/posts/11.html     // 匹配 2
/posts/index.html  // 匹配 1
```

### 路由参数

通过正则表达式匹配的路由，其中带命名的参数可通过 `Params()` 获取：

```go
params := Params(r)

id, err := params.Int("id")
 // 或是
id := params.MustInt("id", 0) // 0 表示在无法获取 id 参数的默认值
```

## 高级用法

### 分组路由

可以通过匹配 `group.Matcher` 接口，定义了一组特定要求的路由项。

```go
// server.go

import "github.com/issue9/mux/v5"
import "github.com/issue9/mux/v5/group"

m := group.Default()

def := mux.DefaultRouter()
err := m.AddRouter("default", group.NewPathVersion("version-key", "v1"), def)
def.Get("/path", h1)

host := mux.DefaultRouter()
err := m.AddRouter("host", group.NewHosts("*.example.com"), host)
host.Get("/path", h2)

http.ListenAndServe(":8080", m)

// client.go

// 访问 h2 的内容
r := http.NewRequest(http.MethodGet, "https://abc.example.com/path", nil)
r.Do()

// 访问 h1 的内容
r := http.NewRequest(http.MethodGet, "https://other_domain.com/v1/path", nil)
r.Do()
```

### interceptor

正常情况下，`/posts/{id:\d+}` 或是 `/posts/{id:[0-9]+}` 会被当作正则表达式处理，
但是正则表达式的性能并不是很好，这个时候我们可以通完 `interceptor` 包进行拦截，
采用自己的特定方法进行处理：

```go
import "github.com/issue9/mux/v5/interceptor"

func digit(path string) bool {
    for _, c := range path {
        if c < '0' || c > '9' {
            return false
        }
    }
    return len(path) > 0
}

// 路由中的 \d+ 和 [0-9]+ 均采用 digit 函数进行处理，不再是正则表达式。
interceptor.Register(digit, "\\d+", "[0-9]+")
```

### CORS

CORS 不再是以中间件的形式提供，而是通过 NewRouter 直接传递有关 CORS 的配置信息，
这样可以更好地处理每个地址支持的请求方法。

OPTIONS 请求方法由系统自动生成。

```go
r, ok := mux.NewRouter(false, AllowedCORS()) // 任意跨域请求

r.Get("/posts/{id}", nil)     // 默认情况下， OPTIONS 的报头为 GET, OPTIONS

http.ListenAndServe(":8080", m)

// client.go

// 访问 h2 的内容
r := http.NewRequest(http.MethodGet, "https://localhost:8080/posts/1", nil)
r.Header.Set("Origin", "http://example.com")
r.Do() // 跨域，可以正常访问


r = http.NewRequest(http.MethodOptions, "https://localhost:8080/posts/1", nil)
r.Header.Set("Origin", "http://example.com")
r.Header.Set("Access-Control-Request-Method", "GET")
r.Do() // 预检请求，可以正常访问
```

### HEAD

默认情况下，用户无须显示地实现 `HEAD` 请求， 系统会为每一个 `GET` 请求自动实现一个对应的 `HEAD` 请求，
当然也与 `OPTIONS` 一样，你也可以自通过 `mux.Handle()` 自己实现 `HEAD` 请求。

### 中间件

mux 本身就是一个实现了 [http.Handler](https://pkg.go.dev/net/http#Handler) 接口的中间件，
所有实现官方接口 `http.Handler` 的都可以附加到 mux 上作为中间件使用。

mux 本身也提供了对中间件的管理功能，同时 [middleware](https://github.com/issue9/middleware) 提供了常用的中间件功能。

```go
import "github.com/issue9/middleware/v4/compress"

c := compress.New(log.Default(), "*")

r := mux.DefaultRouter()

// 添加中间件
r.Middlewares().Append(c.Middleware)
```

## 性能

<https://caixw.github.io/go-http-routers-testing/> 提供了与其它几个框架的对比情况。

## 版权

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
