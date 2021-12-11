// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"

	"github.com/issue9/mux/v5/middleware"
)

type MiddlewareFunc = middleware.Func

type Middlewares = middleware.Middlewares

func ApplyMiddlewares(h http.Handler, f ...MiddlewareFunc) http.Handler {
	return middleware.Apply(h, f...)
}

func ApplyMiddlewaresFunc(h func(w http.ResponseWriter, r *http.Request), f ...MiddlewareFunc) http.Handler {
	return ApplyMiddlewares(http.HandlerFunc(h), f...)
}

// NewMiddlewares 声明新的 Middlewares 实例
func NewMiddlewares(next http.Handler) *Middlewares {
	return middleware.NewMiddlewares(next)
}

// Middlewares 返回中间件管理接口
func (r *Router) Middlewares() *Middlewares { return r.ms }
