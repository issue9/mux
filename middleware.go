// SPDX-License-Identifier: MIT

package mux

import "net/http"

// MiddlewareFunc 将一个 http.Handler 封装成另一个 http.Handler
type MiddlewareFunc func(http.Handler) http.Handler

// ApplyMiddlewares 按顺序将所有的中间件应用于 h
func ApplyMiddlewares(h http.Handler, f ...MiddlewareFunc) http.Handler {
	if l := len(f); l > 0 {
		for i := l - 1; i >= 0; i-- {
			h = f[i](h)
		}
	}
	return h
}

// ApplyMiddlewaresFunc 按顺序将所有的中间件应用于 h
func ApplyMiddlewaresFunc(h func(w http.ResponseWriter, r *http.Request), f ...MiddlewareFunc) http.Handler {
	return ApplyMiddlewares(http.HandlerFunc(h), f...)
}

// Append 添加中间件到尾部
func (mux *Mux) Append(f MiddlewareFunc) *Mux {
	if mux.middlewares == nil {
		mux.middlewares = make([]MiddlewareFunc, 0, 5)
	}
	mux.middlewares = append(mux.middlewares, f)
	return mux.buildMiddlewares()
}

// Prepend 添加中间件到顶部
func (mux *Mux) Prepend(f MiddlewareFunc) *Mux {
	// NOTE: 当允许传递多个参数时，不同用户对添加顺序理解可能会不一样：
	// - 按顺序一次性添加到顶部；
	// - 单个元素依次添加到顶部；
	ms := make([]MiddlewareFunc, 0, 1+len(mux.middlewares))
	ms = append(ms, f)
	if len(mux.middlewares) > 0 {
		ms = append(ms, mux.middlewares...)
	}
	mux.middlewares = ms
	return mux.buildMiddlewares()
}

// Reset 清除中间件
func (mux *Mux) Reset() {
	mux.middlewares = mux.middlewares[:0]
	mux.handler = http.HandlerFunc(mux.serveHTTP)
}

func (mux *Mux) buildMiddlewares() *Mux {
	mux.handler = ApplyMiddlewaresFunc(mux.serveHTTP, mux.middlewares...)
	return mux
}

func (mux *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux.handler.ServeHTTP(w, r)
}
