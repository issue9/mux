// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package ctx 自定义路由
package ctx

import (
	"net/http"

	"github.com/issue9/mux/v8"
	"github.com/issue9/mux/v8/group"
	"github.com/issue9/mux/v8/header"
	"github.com/issue9/mux/v8/types"
)

type (
	CTX struct {
		R *http.Request
		W http.ResponseWriter
		P types.Route
	}

	Router = mux.RouterOf[Handler]

	Routers = group.GroupOf[Handler]

	Handler interface {
		Handle(*CTX)
	}

	HandlerFunc func(*CTX)
)

func (f HandlerFunc) Handle(c *CTX) { f(c) }

func call(w http.ResponseWriter, r *http.Request, ps types.Route, h Handler) {
	h.Handle(&CTX{R: r, W: w, P: ps})
}

func optionsHandlerBuilder(p types.Node) Handler {
	return HandlerFunc(func(ctx *CTX) {
		ctx.W.Header().Set(header.Allow, p.AllowHeader())
	})
}

func methodNotAllowedBuilder(p types.Node) Handler {
	return HandlerFunc(func(ctx *CTX) {
		ctx.W.Header().Set(header.Allow, p.AllowHeader())
		ctx.W.WriteHeader(http.StatusMethodNotAllowed)
	})
}

func notFound(ctx *CTX) {
	ctx.W.WriteHeader(http.StatusNotFound)
}

func NewRouters(o ...mux.Option) *Routers {
	return group.NewOf[Handler](call, HandlerFunc(notFound), methodNotAllowedBuilder, optionsHandlerBuilder, o...)
}

// NewRouter 声明适用于官方 http.Handler 接口的路由
func NewRouter(name string, o ...mux.Option) *Router {
	return mux.NewRouterOf[Handler](name, call, HandlerFunc(notFound), methodNotAllowedBuilder, optionsHandlerBuilder, o...)
}
