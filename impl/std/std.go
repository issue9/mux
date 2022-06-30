// SPDX-License-Identifier: MIT

package std

import (
	"context"
	"net/http"

	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/types"
)

const contextKeyParams contextKey = 0

type (
	Routers        = mux.RoutersOf[http.Handler]
	Router         = mux.RouterOf[http.Handler]
	Prefix         = mux.PrefixOf[http.Handler]
	Resource       = mux.ResourceOf[http.Handler]
	Middleware     = types.MiddlewareOf[http.Handler]
	MiddlewareFunc = types.MiddlewareFuncOf[http.Handler]

	contextKey int
)

func call(w http.ResponseWriter, r *http.Request, ps types.Params, h http.Handler) {
	h.ServeHTTP(w, WithValue(r, ps))
}

func optionsHandlerBuilder(p types.Node) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", p.AllowHeader())
	})
}

func NewRouters(notFound http.Handler) *Routers {
	return mux.NewRoutersOf(call, optionsHandlerBuilder, notFound)
}

// NewRouter 声明适用于官方 http.Handler 接口的路由
func NewRouter(name string, o *mux.Options) *Router {
	return mux.NewRouterOf(name, call, optionsHandlerBuilder, o)
}

// GetParams 获取当前请求实例上的参数列表
func GetParams(r *http.Request) types.Params {
	if ps := r.Context().Value(contextKeyParams); ps != nil {
		return ps.(types.Params)
	}
	return nil
}

// WithValue 将参数 ps 附加在 r 上
//
// 与 context.WithValue 功能相同，但是考虑了在同一个 r 上调用多次 WithValue 的情况。
func WithValue(r *http.Request, ps types.Params) *http.Request {
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
