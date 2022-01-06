// SPDX-License-Identifier: MIT

package mux

import "net/http"

// Prefix 操纵统一前缀的路由
//
// example:
// r := NewRouter("")
//  p := r.Prefix("/api")
//  p.Get("/users")  // 相当于 r.Get("/api/users")
//  p.Get("/user/1") // 相当于 r.Get("/api/user/1")
type Prefix struct {
	router *Router
	prefix string
	ms     []MiddlewareFunc
}

func (p *Prefix) Handle(pattern string, h http.Handler, methods ...string) *Prefix {
	p.router.handle(p.prefix+pattern, applyMiddlewares(h, p.ms...), methods...)
	return p
}

func (p *Prefix) Get(pattern string, h http.Handler) *Prefix {
	return p.Handle(pattern, h, http.MethodGet)
}

func (p *Prefix) Post(pattern string, h http.Handler) *Prefix {
	return p.Handle(pattern, h, http.MethodPost)
}

func (p *Prefix) Delete(pattern string, h http.Handler) *Prefix {
	return p.Handle(pattern, h, http.MethodDelete)
}

func (p *Prefix) Put(pattern string, h http.Handler) *Prefix {
	return p.Handle(pattern, h, http.MethodPut)
}

func (p *Prefix) Patch(pattern string, h http.Handler) *Prefix {
	return p.Handle(pattern, h, http.MethodPatch)
}

func (p *Prefix) Any(pattern string, h http.Handler) *Prefix {
	return p.Handle(pattern, h)
}

// Remove 删除指定匹配模式的路由项
func (p *Prefix) Remove(pattern string, methods ...string) {
	p.router.Remove(p.prefix+pattern, methods...)
}

// Clean 清除所有以 Prefix.prefix 开头的路由项
//
// 当指定多个相同的 Prefix 时，调用其中的一个 Clean 也将会清除其它的：
//  p1 := mux.Prefix("prefix")
//  p2 := mux.Prefix("prefix")
//  p2.Clean() 将同时清除 p1 的内容，因为有相同的前缀。
func (p *Prefix) Clean() { p.router.tree.Clean(p.prefix) }

// URL 根据参数生成地址
//
// name 为路由的名称，或是直接为路由项的定义内容，
// 若 name 作为路由项定义，会加上 Prefix.prefix 作为前缀；
func (p *Prefix) URL(strict bool, pattern string, params map[string]string) (string, error) {
	return p.router.URL(strict, p.prefix+pattern, params)
}

// Prefix 在现有 Prefix 的基础上声明一个新的 Prefix 实例
//
// m 中间件函数，按顺序调用，会继承 p 的中间件并按在 m 之前；
//
// example:
//  p := mux.Prefix("/api")
//  v := p.Prefix("/v2")
//  v.Get("/users")   // 相当于 g.Get("/api/v2/users")
//  v.Get("/users/1") // 相当于 g.Get("/api/v2/users/1")
//  v.Get("example.com/users/1") // 相当于 g.Get("/api/v2/example.com/users/1")
func (p *Prefix) Prefix(prefix string, m ...MiddlewareFunc) *Prefix {
	ms := make([]MiddlewareFunc, 0, len(p.ms)+len(m))
	ms = append(ms, p.ms...)
	ms = append(ms, m...)
	return p.router.Prefix(p.prefix+prefix, ms...)
}

// Prefix 声明一个 Prefix 实例
//
// prefix 路由前缀字符串，可以为空；
// m 中间件函数，按顺序调用；
func (r *Router) Prefix(prefix string, m ...MiddlewareFunc) *Prefix {
	ms := make([]MiddlewareFunc, 0, len(r.options.Middlewares)+len(m))
	ms = append(ms, r.options.Middlewares...)
	ms = append(ms, m...)
	return &Prefix{router: r, prefix: prefix, ms: ms}
}

// Router 返回与当前关联的 *Router 实例
func (p *Prefix) Router() *Router { return p.router }
