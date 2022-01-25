// SPDX-License-Identifier: MIT

package mux

import "net/http"

// PrefixOf 操纵统一前缀的路由
type PrefixOf[T any] struct {
	router *RouterOf[T]
	prefix string
	ms     []MiddlewareFuncOf[T]
}

func (p *PrefixOf[T]) Handle(pattern string, h T, methods ...string) *PrefixOf[T] {
	p.router.handle(p.prefix+pattern, applyMiddlewares(h, p.ms...), methods...)
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
//
// name 为路由的名称，或是直接为路由项的定义内容，
// 若 name 作为路由项定义，会加上 PrefixOf.prefix 作为前缀；
func (p *PrefixOf[T]) URL(strict bool, pattern string, params map[string]string) (string, error) {
	return p.router.URL(strict, p.prefix+pattern, params)
}

// Prefix 在现有 PrefixOf 的基础上声明一个新的 PrefixOf 实例
//
// m 中间件函数，按顺序调用，会继承 p 的中间件并按在 m 之前；
//
// example:
//  r := mux.NewRouterOf(...)
//  p := r.Prefix("/api")
//  v := p.Prefix("/v2")
//  v.Get("/users")   // 相当于 g.Get("/api/v2/users")
//  v.Get("/users/1") // 相当于 g.Get("/api/v2/users/1")
//  v.Get("example.com/users/1") // 相当于 g.Get("/api/v2/example.com/users/1")
func (p *PrefixOf[T]) Prefix(prefix string, m ...MiddlewareFuncOf[T]) *PrefixOf[T] {
	ms := make([]MiddlewareFuncOf[T], 0, len(p.ms)+len(m))
	ms = append(ms, p.ms...)
	ms = append(ms, m...)
	return p.router.Prefix(p.prefix+prefix, ms...)
}

// Prefix 声明一个 Prefix 实例
//
// prefix 路由前缀字符串，可以为空；
// m 中间件函数，按顺序调用；
func (r *RouterOf[T]) Prefix(prefix string, m ...MiddlewareFuncOf[T]) *PrefixOf[T] {
	ms := make([]MiddlewareFuncOf[T], 0, len(r.ms)+len(m))
	ms = append(ms, r.ms...)
	ms = append(ms, m...)
	return &PrefixOf[T]{router: r, prefix: prefix, ms: ms}
}

// Router 返回与当前关联的 *Router 实例
func (p *PrefixOf[T]) Router() *RouterOf[T] { return p.router }
