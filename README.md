# mux

[![Go](https://github.com/issue9/mux/workflows/Go/badge.svg)](https://github.com/issue9/mux/actions?query=workflow%3AGo)
[![Go version](https://img.shields.io/github/go-mod/go-version/issue9/mux)](https://golang.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/issue9/mux)](https://goreportcard.com/report/github.com/issue9/mux)
[![license](https://img.shields.io/github/license/issue9/mux)](LICENSE)
[![codecov](https://codecov.io/gh/issue9/mux/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/mux)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/issue9/mux/v9)](https://pkg.go.dev/github.com/issue9/mux/v9)

**这是一个用于定制路由的包，适用于第三方框架实现自己的路由功能。想直接使用，需要少量的代码实例化泛型对象。**

所有实现的路由都支持以下功能：

- 路由参数；
- 支持正则表达式作为路由项匹配方式；
- 拦截正则表达式的行为；
- 自定义的 OPTIONS 请求处理方式；
- 自动生成 HEAD 请求处理方式；
- 根据路由反向生成地址；
- 任意风格的路由，比如 discuz 这种不以 / 作为分隔符的；
- 分组路由，比如按域名，或是版本号等；
- CORS 跨域资源的处理；
- 支持中间件；
- 支持 OPTIONS * 请求；
- TRACE 请求方法的支持；
- panic 处理；

```go
import "github.com/issue9/mux/v9"

router := mux.NewRouter[http.Handler]("", ...) // 采用泛型实现自定义对象
router.Get("/users/1", h).
    Post("/login", h).
    Get("/pages/{id:\\d+}.html", h). // 匹配 /pages/123.html 等格式，path = 123
    Get("/posts/{path}.html", h).    // 匹配 /posts/2020/11/11/title.html 等格式，path = 2020/11/11/title

// 统一前缀路径的路由
p := router.Prefix("/api")
p.Get("/logout", h) // 相当于 m.Get("/api/logout", h)
p.Post("/login", h) // 相当于 m.Get("/api/login", h)

// 对同一资源的不同操作
res := p.Resource("/users/{id:\\d+}")
res.Get(h)   // 相当于 m.Get("/api/users/{id:\\d+}", h)
res.Post(h)  // 相当于 m.Post("/api/users/{id:\\d+}", h)
res.URL(map[string]string{"id": "5"}) // 构建一条基于此路由项的路径：/users/5

http.ListenAndServe(":8080", router)
```

## 语法

路由参数采用大括号包含，内部包含名称和规则两部分：`{name:rule}`，
其中的 name 表示参数的名称，rule 表示对参数的约束规则。

name 可以包含 `-` 前缀，表示在实际执行过程中，不捕获该名称的对应的值，
可以在一定程序上提升性能。

rule 表示对参数的约束，一般为正则或是空，为空表示匹配任意值，
拦截器一栏中有关 rule 的高级用法。以下是一些常见的示例。

```text
/posts/{id}.html                  // 匹配 /posts/1.html
/posts-{id}-{page}.html           // 匹配 /posts-1-10.html
/posts/{path:\\w+}.html           // 匹配 /posts/2020/11/11/title.html
/tags/{tag:\\w+}/{path}           // 匹配 /tags/abc/title.html
```

### 路径匹配规则

可能会出现多条记录与同一请求都匹配的情况，这种情况下，
系统会找到一条认为最匹配的路由来处理，判断规则如下：

1. 普通路由优先于正则路由；
1. 拦截器优先于正则路由；
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

路径的匹配是以从左到右的顺序进行的，父节点不会因为没有子节点匹配而扩大自己的匹配范围，
比如 `/posts/{id}-{page:digit}.html` 可以匹配 `/posts/1-1.html`，但无法匹配
`/posts/1-1-1.html`，虽然理论上 `1-1-` 也能匹配 `{id}`，但是 `1-` 已经优先匹配了，
在子元素找不到的情况下，并不会将父元素的匹配范围扩大到 `1-1-`。

### 路由参数

通过正则表达式匹配的路由，其中带命名的参数可通过 `GetParams()` 获取：

```go
import "github.com/issue9/mux/v9"

params := mux.GetParams(r)

id, err := params.Int("id")
 // 或是
id := params.MustInt("id", 0) // 在无法获取 id 参数时采用 0 作为默认值返回
```

## 高级用法

### 分组路由

可以通过匹配 `Matcher` 接口，定义了一组特定要求的路由项。

```go
// server.go

import "github.com/issue9/mux/v9"
import "github.com/issue9/mux/v9/muxutil"

m := mux.NewRouters(...)

def := mux.NewRouter("default")
m.AddRouter(muxutil.NewPathVersion("version-key", "v1"), def)
def.Get("/path", h1)

host := mux.NewRouter("host")
m.AddRouter(muxutil.NewHosts("*.example.com"), host)
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

### 拦截器

正常情况下，`/posts/{id:\d+}` 或是 `/posts/{id:[0-9]+}` 会被当作正则表达式处理，
但是正则表达式的性能并不是很好，这个时候我们可以通过在 `NewRouter` 传递 `Interceptor` 进行拦截：

```go
import "github.com/issue9/mux/v9"

func digit(path string) bool {
    for _, c := range path {
        if c < '0' || c > '9' {
            return false
        }
    }
    return len(path) > 0
}

// 路由中的 \d+ 和 [0-9]+ 均采用 digit 函数进行处理，不再是正则表达式。
opt := mux.Options{Interceptors: map[string]mux.InterceptorFunc{"\\d+": digit, "[0-9]+": digit}
r := mux.NewRouter("", opt)
```

这样在所有路由项中的 `[0-9]+` 和 `\\d+` 将由 `digit` 函数处理，
不再会被编译为正则表达式，在性能上会有很大的提升。
通过该功能，也可以自定义一些非正常的正则表达这式，然后进行拦截，比如：

```text
/posts/{id:digit}/.html
```

如果不拦截，最终传递给正则表达式，可能会出现编译错误，通过拦截器可以将 digit 合法化。
目前提供了以下几个拦截器：

- InterceptorDigit 限定为数字字符，相当于正则的 `[0-9]`；
- InterceptorWord 相当于正则的 `[a-zA-Z0-9]`；
- InterceptorAny 表示匹配任意非空内容；

用户也可以自行实现 `InterceptorFunc` 作为拦截器。具体可参考 `Interceptor` 的介绍。

### CORS

CORS 不再是以中间件的形式提供，而是通过 `NewRouter` 直接传递有关 CORS 的配置信息，
这样可以更好地处理每个地址支持的请求方法。

OPTIONS 请求方法由系统自动生成。

```go
import "github.com/issue9/mux/v9"

r := mux.NewRouter("name" ,&mux.Options{CORS: AllowedCORS}) // 任意跨域请求

r.Get("/posts/{id}", nil)     // 默认情况下， OPTIONS 的报头为 GET, OPTIONS

http.ListenAndServe(":8080", m)

// client.go

// 访问 h2 的内容
r := http.NewRequest(http.MethodGet, "https://localhost:8080/posts/1", nil)
r.Header.Set("Origin", "https://example.com")
r.Do() // 跨域，可以正常访问


r = http.NewRequest(http.MethodOptions, "https://localhost:8080/posts/1", nil)
r.Header.Set("Origin", "https://example.com")
r.Header.Set("Access-Control-Request-Method", "GET")
r.Do() // 预检请求，可以正常访问
```

### 静态文件

可以使用 `ServeFile` 与命名参数相结合的方式实现静态文件的访问：

```go
r := NewRouter("")
r.Get("/assets/{path}", func(w http.ResponseWriter, r *http.Request){
    err := muxutil.ServeFile(os.DirFS("/static/"), "path", "index.html", w, r)
    if err!= nil {
        http.Error(err.Error(), http.StatusInternalServerError)
    }
})
```

### 自定义路由

官方提供的 `http.Handler` 未必是符合每个人的要求，通过 `Router` 用户可以很方便地实现自定义格式的 `http.Handler`，
只需要以下几个步骤：

1. 定义一个专有的路由处理类型，可以是类也可以是函数；
2. 根据此类型，生成对应的 Router、Prefix、Resource、MiddlewareFunc 等类型；
3. 定义 Call 函数；
4. 将 Call 传递给 NewRouter；

```go
type Context struct {
    *http.Request
    W http.ResponseWriter
    P Params
}

type HandlerFunc func(ctx *Context)

type Router = Router[HandlerFunc]
type Prefix = Prefix[HandlerFunc]
type Resource = Resource[HandlerFunc]
type MiddlewareFunc = MiddlewareFunc[HandlerFunc]
type Middleware = Middleware[HandlerFunc]

func New(name string)* Router {
    call := func(w http.ResponseWriter, r *http.Request, ps Params, h HandlerFunc) {
        ctx := &Context {
            R: r,
            W: w,
            P: ps,
        }
        h(ctx)
    }
    opt := func(n types.Node) Handler {
        return HandlerFunc(func(ctx* Context){
            ctx.W.Header().Set("Allow", n.AllowHeader())
        })
    }

    m := func(n types.Node) Handler {
        return HandlerFunc(func(ctx* Context){
            ctx.W.Header().Set("Allow", n.AllowHeader())
            ctx.W.WriteHeader(405)
        })
    }
    notFound func(ctx* Context) {
        ctx.W.WriteHeader(404)
    }
    return NewRouter[HandlerFunc](name, f, notFound, m, opt)
}
```

以上就是自定义路由的全部功能，之后就可以直接使用：

```go
r := New("router", nil)

r.Get("/path", func(ctx *Context){
    // TODO
    ctx.W.WriteHeader(200)
})

r.Prefix("/admin").Get("/login", func(ctx *Context){
    // TODO
    ctx.W.WriteHeader(501)
})
```

更多自定义路由的介绍可参考 <https://caixw.io/posts/2022/build-go-router-with-generics.html> 或是 [examples](examples) 下的示例。

## 性能

<https://caixw.github.io/go-http-routers-testing/> 提供了与其它几个框架的对比情况。

## 版权

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
