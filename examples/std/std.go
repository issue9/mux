// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package std 兼容标准库的路由
package std

import (
	"context"
	"net/http"

	"github.com/issue9/mux/v9"
	"github.com/issue9/mux/v9/header"
	"github.com/issue9/mux/v9/types"
)

const contextKeyParams contextKey = 0

type (
	Routers         = mux.Group[http.Handler]
	Router          = mux.Router[http.Handler]
	Prefix          = mux.Prefix[http.Handler]
	Resource        = mux.Resource[http.Handler]
	Middleware      = types.Middleware[http.Handler]
	MiddlewareFunc  = types.MiddlewareFunc[http.Handler]
	BuildNodeHandle = types.BuildNodeHandler[http.Handler]

	contextKey int
)

func call(w http.ResponseWriter, r *http.Request, ps types.Route, h http.Handler) {
	h.ServeHTTP(w, WithValue(r, ps))
}

func methodNotAllowedBuilder(p types.Node) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(header.Allow, p.AllowHeader())
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}

func optionsHandlerBuilder(p types.Node) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(header.Allow, p.AllowHeader())
	})
}

func NewRouters(o ...mux.Option) *Routers {
	return mux.NewGroup(call, http.NotFoundHandler(), methodNotAllowedBuilder, optionsHandlerBuilder, o...)
}

// NewRouter 声明适用于官方 [http.Handler] 接口的路由
func NewRouter(name string, o ...mux.Option) *Router {
	return mux.NewRouter(name, call, http.NotFoundHandler(), methodNotAllowedBuilder, optionsHandlerBuilder, o...)
}

// GetParams 获取当前请求实例上的参数列表
func GetParams(r *http.Request) types.Route {
	if ps := r.Context().Value(contextKeyParams); ps != nil {
		return ps.(types.Route)
	}
	return nil
}

// WithValue 将参数 ps 附加在 r 上
//
// 与 context.WithValue 功能相同，但是考虑了在同一个 r 上调用多次 WithValue 的情况。
func WithValue(r *http.Request, ps types.Route) *http.Request {
	if ps == nil {
		return r
	}

	if ps2 := GetParams(r); ps2 != nil && ps2.Params().Count() > 0 {
		ps2.Params().Range(func(k, v string) {
			ps.Params().Set(k, v)
		})
	}

	return r.WithContext(context.WithValue(r.Context(), contextKeyParams, ps))
}
