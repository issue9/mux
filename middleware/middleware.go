// SPDX-License-Identifier: MIT

// Package middleware 中间件管理
package middleware

import "net/http"

// Func 将一个 http.Handler 封装成另一个 http.Handler
type Func func(http.Handler) http.Handler

// Middlewares 中间件管理
type Middlewares struct {
	http.Handler
	middlewares []Func
	next        http.Handler
}

func apply(h http.Handler, f ...Func) http.Handler {
	for _, ff := range f {
		h = ff(h)
	}
	return h
}

// NewMiddlewares 声明新的 Middlewares 实例
func NewMiddlewares(next http.Handler) *Middlewares {
	return &Middlewares{
		Handler:     next,
		middlewares: make([]Func, 0, 10),
		next:        next,
	}
}

// Prepend 添加中间件到顶部
//
// 顶部的中间件在运行过程中将最早被调用，多次添加，则最后一次的在顶部。
func (ms *Middlewares) Prepend(m Func) *Middlewares {
	ms.middlewares = append(ms.middlewares, m)
	ms.Handler = apply(ms.next, ms.middlewares...)
	return ms
}

// Append 添加中间件到尾部
//
// 尾部的中间件将最后被调用，多次添加，则最后一次的在最末尾。
func (ms *Middlewares) Append(f Func) *Middlewares {
	fs := make([]Func, 0, 1+len(ms.middlewares))
	fs = append(fs, f)
	if len(ms.middlewares) > 0 {
		fs = append(fs, ms.middlewares...)
	}
	ms.middlewares = fs
	ms.Handler = apply(ms.next, ms.middlewares...)
	return ms
}

// Reset 清除中间件
func (ms *Middlewares) Reset() *Middlewares {
	ms.middlewares = ms.middlewares[:0]
	ms.Handler = ms.next
	return ms
}
