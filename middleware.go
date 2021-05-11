// SPDX-License-Identifier: MIT

package mux

import "net/http"

// MiddlewareFunc 将一个 http.Handler 封装成另一个 http.Handler
type MiddlewareFunc func(http.Handler) http.Handler

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

// AddMiddleware 添加中间件
//
// first 是否添加到顶部，顶部的中间件在运行过程中，最早被调用，多次添加，最后一次在顶部。
// first == false 添加在尾部，末次添加的元素在最末尾。
func (mux *Mux) AddMiddleware(first bool, f MiddlewareFunc) *Mux {
	if first {
		mux.insertFirst(f)
	} else {
		mux.insertLast(f)
	}

	mux.handler = ApplyMiddlewaresFunc(mux.serveHTTP, mux.middlewares...)
	return mux
}

func (mux *Mux) insertFirst(f MiddlewareFunc) {
	// NOTE: 当允许传递多个参数时，不同用户对添加顺序理解可能会不一样：
	// - 按顺序一次性添加；
	// - 单个元素依次添加；
	ms := make([]MiddlewareFunc, 0, 1+len(mux.middlewares))
	ms = append(ms, f)
	if len(mux.middlewares) > 0 {
		ms = append(ms, mux.middlewares...)
	}
	mux.middlewares = ms
}

func (mux *Mux) insertLast(f MiddlewareFunc) {
	if mux.middlewares == nil {
		mux.middlewares = make([]MiddlewareFunc, 0, 5)
	}
	mux.middlewares = append(mux.middlewares, f)
}

// Reset 清除中间件
func (mux *Mux) Reset() {
	mux.middlewares = mux.middlewares[:0]
	mux.handler = http.HandlerFunc(mux.serveHTTP)
}

func (mux *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux.handler.ServeHTTP(w, r)
}
