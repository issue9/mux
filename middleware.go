// SPDX-License-Identifier: MIT

package mux

import "net/http"

// Middleware 将一个 http.Handler 封装成另一个 http.Handler
type Middleware func(http.Handler) http.Handler

// Append 添加中间件到尾部
func (mux *Mux) Append(m ...Middleware) {
	if mux.middlewares == nil {
		mux.middlewares = make([]Middleware, 0, len(m))
	}
	mux.middlewares = append(mux.middlewares, m...)
	mux.buildMiddlewares()
}

// Prepend 添加中间件到顶部
func (mux *Mux) Prepend(m ...Middleware) {
	ms := make([]Middleware, 0, len(m)+len(mux.middlewares))
	ms = append(ms, m...)
	if len(mux.middlewares) > 0 {
		ms = append(ms, mux.middlewares...)
	}
	mux.middlewares = ms
	mux.buildMiddlewares()
}

// Reset 清除中间件
func (mux *Mux) Reset() {
	mux.middlewares = mux.middlewares[:0]
	mux.handler = http.HandlerFunc(mux.serveHTTP)
}

func (mux *Mux) buildMiddlewares() {
	mux.handler = http.HandlerFunc(mux.serveHTTP)

	if l := len(mux.middlewares); l > 0 {
		for i := l - 1; i >= 0; i-- {
			mux.handler = mux.middlewares[i](mux.handler)
		}
	}
}

func (mux *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux.handler.ServeHTTP(w, r)
}
