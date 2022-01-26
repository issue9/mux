// SPDX-License-Identifier: MIT

package mux

import (
	"context"
	"net/http"
)

const contextKeyParams contextKey = 0

// 提供了标准库的默认支持

type (
	Routers        = RoutersOf[http.Handler]
	Router         = RouterOf[http.Handler]
	Prefix         = PrefixOf[http.Handler]
	Resource       = ResourceOf[http.Handler]
	MiddlewareFunc = MiddlewareFuncOf[http.Handler]

	contextKey int
)

// DefaultCall 针对 http.Handler 的 CallOf 实现
func DefaultCall(w http.ResponseWriter, r *http.Request, ps Params, h http.Handler) {
	h.ServeHTTP(w, WithValue(r, ps))
}

func NewRouters(ms []MiddlewareFunc, o ...Option) *Routers {
	return NewRoutersOf[http.Handler](DefaultCall, ms, o...)
}

// NewRouter 声明适用于官方 http.Handler 接口的路由
//
// 这是对 NewRouterOf 的特化，相当于 NewRouterOf[http.Handler]。
func NewRouter(name string, ms []MiddlewareFunc, o ...Option) *Router {
	return NewRouterOf[http.Handler](name, DefaultCall, ms, o...)
}

// GetParams 获取当前请求实例上的参数列表
//
// NOTE: 仅适用于 Router 而不是所有 RouterOf。
func GetParams(r *http.Request) Params {
	if ps := r.Context().Value(contextKeyParams); ps != nil {
		return ps.(Params)
	}
	return nil
}

// WithValue 将参数 ps 附加在 r 上
//
// 与 context.WithValue 功能相同，但是考虑了在同一个 r 上调用多次 WithValue 的情况。
//
// NOTE: 仅适用于 Router 而不是所有 RouterOf。
func WithValue(r *http.Request, ps Params) *http.Request {
	if ps == nil || ps.Count() == 0 {
		return r
	}

	if ps2 := GetParams(r); ps2 != nil && ps2.Count() > 0 {
		ps2.Range(func(k, v string) {
			ps.Set(k, v)
		})
	}

	return r.WithContext(context.WithValue(r.Context(), contextKeyParams, ps))
}
