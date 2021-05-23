// SPDX-License-Identifier: MIT

package mux

import "net/http"

// MiddlewareFunc 将一个 http.Handler 封装成另一个 http.Handler
type MiddlewareFunc func(http.Handler) http.Handler

// Middlewares 中间件管理
type Middlewares struct {
	http.Handler
	middlewares []MiddlewareFunc
	next        http.Handler
}

// ApplyMiddlewares 按顺序将所有的中间件应用于 h
func ApplyMiddlewares(h http.Handler, f ...MiddlewareFunc) http.Handler {
	for _, ff := range f {
		h = ff(h)
	}
	return h
}

// ApplyMiddlewaresFunc 按顺序将所有的中间件应用于 h
func ApplyMiddlewaresFunc(h func(w http.ResponseWriter, r *http.Request), f ...MiddlewareFunc) http.Handler {
	return ApplyMiddlewares(http.HandlerFunc(h), f...)
}

// NewMiddlewares 声明新的 Middlewares 实例
func NewMiddlewares(next http.Handler) *Middlewares {
	return &Middlewares{
		Handler:     next,
		middlewares: make([]MiddlewareFunc, 0, 10),
		next:        next,
	}
}

// Prepend 添加中间件到顶部
//
// 顶部的中间件在运行过程中将最早被调用，多次添加，则最后一次的在顶部。
func (ms *Middlewares) Prepend(m MiddlewareFunc) *Middlewares {
	ms.middlewares = append(ms.middlewares, m)
	ms.Handler = ApplyMiddlewares(ms.next, ms.middlewares...)
	return ms
}

// Append 添加中间件到尾部
//
// 尾部的中间件将最后被调用，多次添加，则最后一次的在最末尾。
func (ms *Middlewares) Append(f MiddlewareFunc) *Middlewares {
	fs := make([]MiddlewareFunc, 0, 1+len(ms.middlewares))
	fs = append(fs, f)
	if len(ms.middlewares) > 0 {
		fs = append(fs, ms.middlewares...)
	}
	ms.middlewares = fs
	ms.Handler = ApplyMiddlewares(ms.next, ms.middlewares...)
	return ms
}

// Reset 清除中间件
func (ms *Middlewares) Reset() *Middlewares {
	ms.middlewares = ms.middlewares[:0]
	ms.Handler = ms.next
	return ms
}

// AppendMiddleware 添加中间件到尾部
func (r *Router) AppendMiddleware(f MiddlewareFunc) *Router {
	r.ms.Append(f)
	return r
}

// PrependMiddleware 添加中间件到顶部
func (r *Router) PrependMiddleware(f MiddlewareFunc) *Router {
	r.ms.Prepend(f)
	return r
}

func (r *Router) CleanMiddlewares() { r.ms.Reset() }
