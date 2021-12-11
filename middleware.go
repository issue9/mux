// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"

	"github.com/issue9/mux/v5/middleware"
)

type MiddlewareFunc = middleware.Func

type Middlewares = middleware.Middlewares

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
	return middleware.NewMiddlewares(next)
}

// Middlewares 返回中间件管理接口
func (r *Router) Middlewares() *Middlewares { return r.ms }
