// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"

	"github.com/issue9/mux/v6/internal/syntax"
)

// 提供了标准库的默认支持

type (
	Router         = RouterOf[http.Handler]
	Prefix         = PrefixOf[http.Handler]
	Resource       = ResourceOf[http.Handler]
	MiddlewareFunc = MiddlewareFuncOf[http.Handler]
)

// DefaultBuildHandlerFunc 针对 http.Handler 的实现
func DefaultBuildHandlerFunc(w http.ResponseWriter, r *http.Request, ps Params, h http.Handler) {
	h.ServeHTTP(w, syntax.WithValue(r, ps))
}

// NewRouter 声明适用于官方 http.Handler 接口的路由
//
// 这是对 NewRouterOf 的特化，相当于 NewRouterOf[http.Handler]。
func NewRouter(name string, ms []MiddlewareFunc, o ...Option) *Router {
	return NewRouterOf[http.Handler](name, DefaultBuildHandlerFunc, ms, o...)
}

// GetParams 获取路由中的参数集合
func GetParams(r *http.Request) Params { return syntax.GetParams(r) }
