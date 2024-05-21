// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/issue9/errwrap"

	"github.com/issue9/mux/v9/header"
	"github.com/issue9/mux/v9/internal/trace"
	"github.com/issue9/mux/v9/internal/tree"
	"github.com/issue9/mux/v9/types"
)

type (
	// Router 可自定义处理函数类型的路由
	//
	//  router := NewRouter[http.Handler](...)
	//  router.Get("/abc/h1", h1).
	//      Post("/abc/h2", h2).
	//      Handle("/api/{version:\\d+}",h3, http.MethodGet, http.MethodPost) // 只匹配 GET 和 POST
	//  http.ListenAndServe(router)
	Router[T any] struct {
		tree *tree.Tree[T]
		call CallFunc[T]
		ms   []types.Middleware[T]

		cors        *cors
		urlDomain   string
		recoverFunc RecoverFunc
		trace       bool
		matcher     Matcher
	}

	// CallFunc 指定如何调用用户给定的类型 T
	CallFunc[T any] func(http.ResponseWriter, *http.Request, types.Route, T)

	// Resource 以资源地址为对象的路由
	Resource[T any] struct {
		router  *Router[T]
		pattern string
		ms      []types.Middleware[T]
	}

	// Prefix 操纵统一前缀的路由
	Prefix[T any] struct {
		router  *Router[T]
		pattern string
		ms      []types.Middleware[T]
	}

	headResponse struct {
		size int
		http.ResponseWriter
	}
)

// NewRouter 声明路由
//
// name 路由名称，可以为空；
// methodNotAllowedBuilder 和 optionsBuilder 可以自定义 405 和 OPTIONS 请求的处理方式；
// o 用于指定一些可选的参数；
//
// T 表示用户用于处理路由项的方法。
func NewRouter[T any](
	name string,
	call CallFunc[T],
	notFound T,
	methodNotAllowedBuilder, optionsBuilder types.BuildNodeHandler[T],
	o ...Option,
) *Router[T] {
	opt, err := buildOption(o...)
	if err != nil {
		panic(err)
	}

	r := &Router[T]{
		tree: tree.New(name, opt.lock, opt.interceptors, notFound, opt.trace, methodNotAllowedBuilder, optionsBuilder),
		call: call,

		cors:        opt.cors,
		urlDomain:   opt.urlDomain,
		recoverFunc: opt.recoverFunc,
		trace:       opt.trace,
	}

	return r
}

// Clean 清除当前路由组的所有路由项
func (r *Router[T]) Clean() { r.tree.Clean("") }

// Routes 返回当前路由组的路由项
//
// 键名为请求地址，键值为对应的请求方法。
func (r *Router[T]) Routes() map[string][]string { return r.tree.Routes() }

// Remove 移除指定的路由项
//
// 当未指定 methods 时，将删除所有 method 匹配的项。
// 指定错误的 methods 值，将自动忽略该值。
func (r *Router[T]) Remove(pattern string, methods ...string) { r.tree.Remove(pattern, methods...) }

// Use 使用中间件
//
// 对于中间件的使用，除了此方法，还可以 [Router.Prefix]、[Router.Resource]、[Prefix.Prefix]
// 以及添加路由项时指定。调用顺序如下：
//   - Get/Delete 等添加路由项的方法；
//   - [Router.Prefix]、[Router.Resource]、[Prefix.Prefix]；
//   - [Router.Use]；
//
// NOTE: 对于 404 路由项的只有通过 [Router.Use] 应用的中间件有效。
func (r *Router[T]) Use(m ...types.Middleware[T]) {
	r.ms = append(r.ms, m...)
	r.tree.ApplyMiddleware(m...)
}

// Handle 添加一条路由数据
//
// pattern 为路由匹配模式，可以是正则匹配也可以是字符串匹配，
// 若语法不正确，则直接 panic，可以通过 [CheckSyntax] 检测语法的有效性，其它接口也相同；
// m 为应用于当前路由项的中间件；
// methods 该路由项对应的请求方法，如果未指定值，则表示所有支持的请求方法，其中 OPTIONS、TRACE 和 HEAD 不受控；
func (r *Router[T]) Handle(pattern string, h T, m []types.Middleware[T], methods ...string) *Router[T] {
	if err := r.tree.Add(pattern, h, slices.Concat(m, r.ms), methods...); err != nil {
		panic(err)
	}
	return r
}

// Get 相当于 Router.Handle(pattern, h, http.MethodGet) 的简易写法
//
// h 不应该主动调用 WriteHeader，否则会导致 HEAD 请求获取不到 Content-Length 报头。
func (r *Router[T]) Get(pattern string, h T, m ...types.Middleware[T]) *Router[T] {
	return r.Handle(pattern, h, m, http.MethodGet)
}

func (r *Router[T]) Post(pattern string, h T, m ...types.Middleware[T]) *Router[T] {
	return r.Handle(pattern, h, m, http.MethodPost)
}

func (r *Router[T]) Delete(pattern string, h T, m ...types.Middleware[T]) *Router[T] {
	return r.Handle(pattern, h, m, http.MethodDelete)
}

func (r *Router[T]) Put(pattern string, h T, m ...types.Middleware[T]) *Router[T] {
	return r.Handle(pattern, h, m, http.MethodPut)
}

func (r *Router[T]) Patch(pattern string, h T, m ...types.Middleware[T]) *Router[T] {
	return r.Handle(pattern, h, m, http.MethodPatch)
}

// Any 添加一条包含全部请求方法的路由
func (r *Router[T]) Any(pattern string, h T, m ...types.Middleware[T]) *Router[T] {
	return r.Handle(pattern, h, m)
}

// URL 根据参数生成地址
//
// strict 是否检查路由是否真实存在以及参数是否符合要求；
// pattern 为路由项的定义内容；
// params 为路由项中的参数，键名为参数名，键值为参数值。
func (r *Router[T]) URL(strict bool, pattern string, params map[string]string) (string, error) {
	buf := errwrap.StringBuilder{}
	buf.Grow(len(r.urlDomain) + len(pattern))

	if r.urlDomain != "" {
		buf.WString(r.urlDomain)
	}

	switch {
	case len(pattern) == 0: // 无需要处理
	case len(params) == 0:
		buf.WString(pattern)
	case strict:
		if err := r.tree.URL(&buf, pattern, params); err != nil {
			return "", err
		}
	default:
		if err := emptyInterceptors.URL(&buf, pattern, params); err != nil {
			return "", err
		}
	}

	return buf.String(), buf.Err
}

func (r *Router[T]) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := types.NewContext()
	r.serveContext(w, req, ctx)
	ctx.Destroy()
}

func (r *Router[T]) serveContext(w http.ResponseWriter, req *http.Request, ctx *types.Context) {
	if r.recoverFunc != nil {
		defer func() {
			if err := recover(); err != nil {
				r.recoverFunc(w, err)
			}
		}()
	}

	if r.trace && req.Method == http.MethodTrace {
		trace.Trace(w, req, false)
		return
	}

	ctx.Path = req.URL.Path
	node, h, ok := r.tree.Handler(ctx, req.Method)
	ctx.SetNode(node)
	ctx.SetRouterName(r.Name())

	if ok { // !ok 即为 405 或是 404 状态
		r.cors.Handle(node, w.Header(), req)
		if req.Method == http.MethodHead {
			w = &headResponse{ResponseWriter: w}
		}
	}
	r.call(w, req, ctx, h)
}

// Name 路由名称
func (r *Router[T]) Name() string { return r.tree.Name() }

func (p *Prefix[T]) Handle(pattern string, h T, m []types.Middleware[T], methods ...string) *Prefix[T] {
	p.router.Handle(p.Pattern()+pattern, h, slices.Concat(m, p.ms), methods...)
	return p
}

func (p *Prefix[T]) Get(pattern string, h T, m ...types.Middleware[T]) *Prefix[T] {
	return p.Handle(pattern, h, m, http.MethodGet)
}

func (p *Prefix[T]) Post(pattern string, h T, m ...types.Middleware[T]) *Prefix[T] {
	return p.Handle(pattern, h, m, http.MethodPost)
}

func (p *Prefix[T]) Delete(pattern string, h T, m ...types.Middleware[T]) *Prefix[T] {
	return p.Handle(pattern, h, m, http.MethodDelete)
}

func (p *Prefix[T]) Put(pattern string, h T, m ...types.Middleware[T]) *Prefix[T] {
	return p.Handle(pattern, h, m, http.MethodPut)
}

func (p *Prefix[T]) Patch(pattern string, h T, m ...types.Middleware[T]) *Prefix[T] {
	return p.Handle(pattern, h, m, http.MethodPatch)
}

func (p *Prefix[T]) Any(pattern string, h T, m ...types.Middleware[T]) *Prefix[T] {
	return p.Handle(pattern, h, m)
}

// Pattern 当前对象的路径
func (p *Prefix[T]) Pattern() string { return p.pattern }

// Remove 删除指定匹配模式的路由项
func (p *Prefix[T]) Remove(pattern string, methods ...string) {
	p.router.Remove(p.Pattern()+pattern, methods...)
}

// Clean 清除所有以 [Prefix.Pattern] 开头的路由项
//
// 当指定多个相同的 Prefix 时，调用其中的一个 [Prefix.Clean] 也将会清除其它的：
//
//	r := NewRouter(...)
//	p1 := r.Prefix("prefix")
//	p2 := r.Prefix("prefix")
//	p2.Clean() 将同时清除 p1 的内容，因为有相同的前缀。
func (p *Prefix[T]) Clean() { p.router.tree.Clean(p.Pattern()) }

// URL 根据参数生成地址
func (p *Prefix[T]) URL(strict bool, pattern string, params map[string]string) (string, error) {
	return p.router.URL(strict, p.Pattern()+pattern, params)
}

// Prefix 在现有 Prefix 的基础上声明一个新的 [Prefix] 实例
//
// m 中间件函数，按顺序调用可参考 [Router.Use] 的说明；
func (p *Prefix[T]) Prefix(prefix string, m ...types.Middleware[T]) *Prefix[T] {
	return p.router.Prefix(p.Pattern()+prefix, slices.Concat(m, p.ms)...)
}

// Prefix 声明一个 [Prefix] 实例
//
// prefix 路由前缀字符串，可以为空；
// m 中间件函数，按顺序调用可参考 [Router.Use] 的说明；
func (r *Router[T]) Prefix(prefix string, m ...types.Middleware[T]) *Prefix[T] {
	return &Prefix[T]{router: r, pattern: prefix, ms: slices.Clone(m)}
}

// Router 返回与当前关联的 [Router] 实例
func (p *Prefix[T]) Router() *Router[T] { return p.router }

func (r *Resource[T]) Handle(h T, m []types.Middleware[T], methods ...string) *Resource[T] {
	r.router.Handle(r.pattern, h, slices.Concat(m, r.ms), methods...)
	return r
}

func (r *Resource[T]) Get(h T, m ...types.Middleware[T]) *Resource[T] {
	return r.Handle(h, m, http.MethodGet)
}

func (r *Resource[T]) Post(h T, m ...types.Middleware[T]) *Resource[T] {
	return r.Handle(h, m, http.MethodPost)
}

func (r *Resource[T]) Delete(h T, m ...types.Middleware[T]) *Resource[T] {
	return r.Handle(h, m, http.MethodDelete)
}

func (r *Resource[T]) Put(h T, m ...types.Middleware[T]) *Resource[T] {
	return r.Handle(h, m, http.MethodPut)
}

func (r *Resource[T]) Patch(h T, m ...types.Middleware[T]) *Resource[T] {
	return r.Handle(h, m, http.MethodPatch)
}

func (r *Resource[T]) Any(h T, m ...types.Middleware[T]) *Resource[T] { return r.Handle(h, m) }

// Remove 删除指定匹配模式的路由项
func (r *Resource[T]) Remove(methods ...string) { r.router.Remove(r.pattern, methods...) }

// Clean 清除当前资源的所有路由项
func (r *Resource[T]) Clean() { r.router.Remove(r.pattern) }

// URL 根据参数构建一条 URL
//
// params 匹配路由参数中的同名参数，或是不存在路由参数，比如普通的字符串路由项，
// 该参数不启作用；
//
//	res, := m.Resource("/posts/{id}")
//	res.URL(map[string]string{"id": "1"}, "") // /posts/1
//
//	res, := m.Resource("/posts/{id}/{path}")
//	res.URL(map[string]string{"id": "1","path":"author/profile"}) // /posts/1/author/profile
func (r *Resource[T]) URL(strict bool, params map[string]string) (string, error) {
	return r.router.URL(strict, r.Pattern(), params)
}

// Pattern 当前对象的路径
func (r *Resource[T]) Pattern() string { return r.pattern }

// Resource 创建一个资源路由项
//
// pattern 资源地址；
// m 中间件函数，按顺序调用，会继承 r 的中间件并按在 m 之前；
func (r *Router[T]) Resource(pattern string, m ...types.Middleware[T]) *Resource[T] {
	return &Resource[T]{router: r, pattern: pattern, ms: slices.Clone(m)}
}

// Resource 创建一个资源路由项
//
// pattern 资源地址；
// m 中间件函数，按顺序调用可参考 [Router.Use] 的说明；
func (p *Prefix[T]) Resource(pattern string, m ...types.Middleware[T]) *Resource[T] {
	return p.router.Resource(p.Pattern()+pattern, slices.Concat(m, p.ms)...)
}

// Router 返回与当前资源关联的 [Router] 实例
func (r *Resource[T]) Router() *Router[T] { return r.router }

func (resp *headResponse) Write(bs []byte) (int, error) {
	l := len(bs)
	resp.size += l

	resp.Header().Set(header.ContentLength, strconv.Itoa(resp.size))
	return l, nil
}
