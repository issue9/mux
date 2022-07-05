// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"strconv"

	"github.com/issue9/errwrap"

	"github.com/issue9/mux/v7/internal/params"
	"github.com/issue9/mux/v7/internal/tree"
	"github.com/issue9/mux/v7/types"
)

type (
	// RouterOf 路由
	//
	// 可以对路径按正则或是请求方法进行匹配。用法如下：
	//  router := NewRouterOf[http.Handler](...)
	//  router.Get("/abc/h1", h1).
	//      Post("/abc/h2", h2).
	//      Handle("/api/{version:\\d+}",h3, http.MethodGet, http.MethodPost) // 只匹配 GET 和 POST
	//  http.ListenAndServe(router)
	//
	// 如果需要同时对多个 RouterOf 实例进行路由，可以采用 RoutersOf 对象管理多个 RouterOf 实例。
	RouterOf[T any] struct {
		name    string
		tree    *tree.Tree[T]
		call    CallOf[T]
		matcher Matcher

		cors        *CORS
		urlDomain   string
		recoverFunc RecoverFunc

		ms []types.MiddlewareOf[T]
	}

	// CallOf 指定如何调用用户自定义的对象 T
	CallOf[T any] func(http.ResponseWriter, *http.Request, types.Params, T)

	// ResourceOf 以资源地址为对象的路由
	ResourceOf[T any] struct {
		router  *RouterOf[T]
		pattern string
		ms      []types.MiddlewareOf[T]
	}

	// PrefixOf 操纵统一前缀的路由
	PrefixOf[T any] struct {
		router *RouterOf[T]
		prefix string
		ms     []types.MiddlewareOf[T]
	}

	headResponse struct {
		size int
		http.ResponseWriter
	}
)

// NewRouterOf 声明路由
//
// name 路由名称，可以为空；
// o 修改路由的默认行为，可以为空；
// T 表示用户用于处理路由项的方法。
func NewRouterOf[T any](name string, call CallOf[T], notFound T, methodNotAllowedBuilder, optionsBuilder types.BuildNodeHandleOf[T], o *Options) *RouterOf[T] {
	o, err := buildOptions(o)
	if err != nil {
		panic(err)
	}

	r := &RouterOf[T]{
		name: name,
		tree: tree.New(o.Lock, o.interceptors, notFound, methodNotAllowedBuilder, optionsBuilder),
		call: call,

		cors:        o.CORS,
		urlDomain:   o.URLDomain,
		recoverFunc: o.RecoverFunc,
	}

	return r
}

// Clean 清除当前路由组的所有路由项
func (r *RouterOf[T]) Clean() { r.tree.Clean("") }

// Routes 返回当前路由组的路由项
//
// 键名为请求地址，键值为对应的请求方法。
func (r *RouterOf[T]) Routes() map[string][]string { return r.tree.Routes() }

// Remove 移除指定的路由项
//
// 当未指定 methods 时，将删除所有 method 匹配的项。
// 指定错误的 methods 值，将自动忽略该值。
func (r *RouterOf[T]) Remove(pattern string, methods ...string) {
	r.tree.Remove(pattern, methods...)
}

// Use 将中间件应用到所有匹配的路由项
//
// OPTIONS、404、405 等没有明确归属的请求只受此函数添加的中间件影响。
func (r *RouterOf[T]) Use(m ...types.MiddlewareOf[T]) {
	r.ms = append(r.ms, m...)
	r.tree.ApplyMiddleware(m...)
}

// Handle 添加一条路由数据
//
// pattern 为路由匹配模式，可以是正则匹配也可以是字符串匹配，
// 若语法不正确，则直接 panic，可以通过 CheckSyntax 检测语法的有效性，其它接口也相同。
// methods 该路由项对应的请求方法，如果未指定值，则表示所有支持的请求方法，
// 但不包含 OPTIONS 和 HEAD。
func (r *RouterOf[T]) Handle(pattern string, h T, methods ...string) *RouterOf[T] {
	r.handle(pattern, h, methods...)
	return r
}

func (r *RouterOf[T]) handle(pattern string, h T, methods ...string) {
	h = tree.ApplyMiddlewares(h, r.ms...)
	if err := r.tree.Add(pattern, h, methods...); err != nil {
		panic(err)
	}
}

// Get 相当于 RouterOf.Handle(pattern, h, http.MethodGet) 的简易写法
//
// h 不应该主动调用 WriteHeader，否则会导致 HEAD 请求获取不到 Content-Length 报头。
func (r *RouterOf[T]) Get(pattern string, h T) *RouterOf[T] {
	return r.Handle(pattern, h, http.MethodGet)
}

func (r *RouterOf[T]) Post(pattern string, h T) *RouterOf[T] {
	return r.Handle(pattern, h, http.MethodPost)
}

func (r *RouterOf[T]) Delete(pattern string, h T) *RouterOf[T] {
	return r.Handle(pattern, h, http.MethodDelete)
}

func (r *RouterOf[T]) Put(pattern string, h T) *RouterOf[T] {
	return r.Handle(pattern, h, http.MethodPut)
}

func (r *RouterOf[T]) Patch(pattern string, h T) *RouterOf[T] {
	return r.Handle(pattern, h, http.MethodPatch)
}

func (r *RouterOf[T]) Any(pattern string, h T) *RouterOf[T] {
	return r.Handle(pattern, h)
}

// URL 根据参数生成地址
//
// strict 是否检查路由是否真实存在以及参数是否符合要求；
// pattern 为路由项的定义内容；
// params 为路由项中的参数，键名为参数名，键值为参数值。
func (r *RouterOf[T]) URL(strict bool, pattern string, params map[string]string) (string, error) {
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

func (r *RouterOf[T]) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p := params.New("")
	defer p.Destroy()
	r.serveHTTP(w, req, p)
}

func (r *RouterOf[T]) serveHTTP(w http.ResponseWriter, req *http.Request, p *params.Params) {
	if r.recoverFunc != nil {
		defer func() {
			if err := recover(); err != nil {
				r.recoverFunc(w, err)
			}
		}()
	}

	p.Path = req.URL.Path
	node, h, exists := r.tree.Handler(p, req.Method)
	p.SetNode(node)
	p.SetRouterName(r.Name())

	if exists {
		r.cors.handle(node, w.Header(), req)
		if req.Method == http.MethodHead {
			w = &headResponse{ResponseWriter: w}
		}
	}
	r.call(w, req, p, h)
}

// Name 路由名称
func (r *RouterOf[T]) Name() string { return r.name }

func (p *PrefixOf[T]) Handle(pattern string, h T, methods ...string) *PrefixOf[T] {
	p.router.handle(p.prefix+pattern, tree.ApplyMiddlewares(h, p.ms...), methods...)
	return p
}

func (p *PrefixOf[T]) Get(pattern string, h T) *PrefixOf[T] {
	return p.Handle(pattern, h, http.MethodGet)
}

func (p *PrefixOf[T]) Post(pattern string, h T) *PrefixOf[T] {
	return p.Handle(pattern, h, http.MethodPost)
}

func (p *PrefixOf[T]) Delete(pattern string, h T) *PrefixOf[T] {
	return p.Handle(pattern, h, http.MethodDelete)
}

func (p *PrefixOf[T]) Put(pattern string, h T) *PrefixOf[T] {
	return p.Handle(pattern, h, http.MethodPut)
}

func (p *PrefixOf[T]) Patch(pattern string, h T) *PrefixOf[T] {
	return p.Handle(pattern, h, http.MethodPatch)
}

func (p *PrefixOf[T]) Any(pattern string, h T) *PrefixOf[T] {
	return p.Handle(pattern, h)
}

// Remove 删除指定匹配模式的路由项
func (p *PrefixOf[T]) Remove(pattern string, methods ...string) {
	p.router.Remove(p.prefix+pattern, methods...)
}

// Clean 清除所有以 PrefixOf.prefix 开头的路由项
//
// 当指定多个相同的 PrefixOf 时，调用其中的一个 Clean 也将会清除其它的：
//  r := NewRouterOf(...)
//  p1 := r.Prefix("prefix")
//  p2 := r.Prefix("prefix")
//  p2.Clean() 将同时清除 p1 的内容，因为有相同的前缀。
func (p *PrefixOf[T]) Clean() { p.router.tree.Clean(p.prefix) }

// URL 根据参数生成地址
func (p *PrefixOf[T]) URL(strict bool, pattern string, params map[string]string) (string, error) {
	return p.router.URL(strict, p.prefix+pattern, params)
}

// Prefix 在现有 PrefixOf 的基础上声明一个新的 PrefixOf 实例
//
// m 中间件函数，按顺序调用，会继承 p 的中间件并按在 m 之前；
func (p *PrefixOf[T]) Prefix(prefix string, m ...types.MiddlewareOf[T]) *PrefixOf[T] {
	ms := make([]types.MiddlewareOf[T], 0, len(p.ms)+len(m))
	ms = append(ms, p.ms...)
	ms = append(ms, m...)
	return p.router.Prefix(p.prefix+prefix, ms...)
}

// Prefix 声明一个 Prefix 实例
//
// prefix 路由前缀字符串，可以为空；
// m 中间件函数，按顺序调用，会继承 r 的中间件并按在 m 之前；
func (r *RouterOf[T]) Prefix(prefix string, m ...types.MiddlewareOf[T]) *PrefixOf[T] {
	ms := make([]types.MiddlewareOf[T], 0, len(m))
	ms = append(ms, m...)
	return &PrefixOf[T]{router: r, prefix: prefix, ms: ms}
}

// Router 返回与当前关联的 *Router 实例
func (p *PrefixOf[T]) Router() *RouterOf[T] { return p.router }

func (r *ResourceOf[T]) Handle(h T, methods ...string) *ResourceOf[T] {
	r.router.handle(r.pattern, tree.ApplyMiddlewares(h, r.ms...), methods...)
	return r
}

func (r *ResourceOf[T]) Get(h T) *ResourceOf[T] { return r.Handle(h, http.MethodGet) }

func (r *ResourceOf[T]) Post(h T) *ResourceOf[T] { return r.Handle(h, http.MethodPost) }

func (r *ResourceOf[T]) Delete(h T) *ResourceOf[T] { return r.Handle(h, http.MethodDelete) }

func (r *ResourceOf[T]) Put(h T) *ResourceOf[T] { return r.Handle(h, http.MethodPut) }

func (r *ResourceOf[T]) Patch(h T) *ResourceOf[T] { return r.Handle(h, http.MethodPatch) }

func (r *ResourceOf[T]) Any(h T) *ResourceOf[T] { return r.Handle(h) }

// Remove 删除指定匹配模式的路由项
func (r *ResourceOf[T]) Remove(methods ...string) { r.router.Remove(r.pattern, methods...) }

// Clean 清除当前资源的所有路由项
func (r *ResourceOf[T]) Clean() { r.router.Remove(r.pattern) }

// URL 根据参数构建一条 URL
//
// params 匹配路由参数中的同名参数，或是不存在路由参数，比如普通的字符串路由项，
// 该参数不启作用；
//  res, := m.Resource("/posts/{id}")
//  res.URL(map[string]string{"id": "1"}, "") // /posts/1
//
//  res, := m.Resource("/posts/{id}/{path}")
//  res.URL(map[string]string{"id": "1","path":"author/profile"}) // /posts/1/author/profile
func (r *ResourceOf[T]) URL(strict bool, params map[string]string) (string, error) {
	return r.router.URL(strict, r.pattern, params)
}

// Resource 创建一个资源路由项
//
// pattern 资源地址；
// m 中间件函数，按顺序调用，会继承 r 的中间件并按在 m 之前；
func (r *RouterOf[T]) Resource(pattern string, m ...types.MiddlewareOf[T]) *ResourceOf[T] {
	ms := make([]types.MiddlewareOf[T], 0, len(m))
	ms = append(ms, m...)
	return &ResourceOf[T]{router: r, pattern: pattern, ms: ms}
}

// Resource 创建一个资源路由项
//
// pattern 资源地址；
// m 中间件函数，按顺序调用，会继承 p 的中间件并按在 m 之前；
func (p *PrefixOf[T]) Resource(pattern string, m ...types.MiddlewareOf[T]) *ResourceOf[T] {
	ms := make([]types.MiddlewareOf[T], 0, len(p.ms)+len(m))
	ms = append(ms, p.ms...)
	ms = append(ms, m...)
	return p.router.Resource(p.prefix+pattern, ms...)
}

// Router 返回与当前资源关联的 *Router 实例
func (r *ResourceOf[T]) Router() *RouterOf[T] { return r.router }

func (resp *headResponse) Write(bs []byte) (int, error) {
	l := len(bs)
	resp.size += l

	resp.Header().Set("Content-Length", strconv.Itoa(resp.size))
	return l, nil
}
