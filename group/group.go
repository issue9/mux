// SPDX-License-Identifier: MIT

// Package group 提供了按条件进行分组路由的功能
package group

import (
	"errors"
	"net/http"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v5"
	"github.com/issue9/mux/v5/internal/tree"
)

// ErrRouterExists 表示是否存在同名的路由名称
var ErrRouterExists = errors.New("该名称的路由已经存在")

type Groups struct {
	routers     []*Router
	middlewares *mux.Middlewares

	notFound,
	methodNotAllowed http.HandlerFunc
}

type Router struct {
	*mux.Router
	g       *Groups
	name    string
	matcher Matcher
}

// Default 返回默认参数的 New
func Default() *Groups { return New(nil, nil) }

// New 声明一个新的 Mux
//
// notFound 404 页面的处理方式，为 nil 时会调用默认的方式进行处理；
// methodNotAllowed 405 页面的处理方式，为 nil 时会调用默认的方式进行处理；
func New(notFound, methodNotAllowed http.HandlerFunc) *Groups {
	if notFound == nil {
		notFound = tree.DefaultNotFound
	}
	if methodNotAllowed == nil {
		methodNotAllowed = tree.DefaultMethodNotAllowed
	}

	g := &Groups{
		routers:          make([]*Router, 0, 1),
		notFound:         notFound,
		methodNotAllowed: methodNotAllowed,
	}
	g.middlewares = mux.NewMiddlewares(http.HandlerFunc(g.serveHTTP))
	return g
}

func (g *Groups) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.middlewares.ServeHTTP(w, r)
}

func (g *Groups) serveHTTP(w http.ResponseWriter, r *http.Request) {
	for _, router := range g.routers {
		if req, ok := router.matcher.Match(r); ok {
			router.ServeHTTP(w, req)
			return
		}
	}
	g.notFound(w, r)
}

// AddRouter 添加一个路由组
//
// 该路由只有符合 Matcher 的要求才会进入，其它与 Router 功能相同。
// 当 Matcher 与其它路由组的判断有重复时，第一条返回 true 的路由组获胜，
// 即使该路由组最终返回 404，也不会再在其它路由组里查找相应的路由。
// 所以在有多条子路由的情况下，第一条子路由不应该永远返回 true，
// 这样会造成其它子路由永远无法到达。
//
// name 表示该路由组的名称，需要唯一，否则返回 false；
// matcher 路由的准入条件，如果为空，则此条路由匹配时会被排在最后，
// 只能有一个路由的 matcher 为空，否则会 panic；
func (g *Groups) AddRouter(name string, matcher Matcher, r *mux.Router) error {
	if name == "" {
		panic("参数 name 不能为空")
	}
	if matcher == nil {
		panic("参数 matcher 不能为空")
	}

	// 重名检测
	index := sliceutil.Index(g.routers, func(i int) bool {
		return g.routers[i].name == name
	})
	if index > -1 {
		return ErrRouterExists
	}

	g.routers = append(g.routers, &Router{
		Router:  r,
		name:    name,
		matcher: matcher,
		g:       g,
	})

	return nil
}

// Router 返回指定名称的 *Router 实例
func (g *Groups) Router(name string) *Router {
	for _, r := range g.routers {
		if r.name == name {
			return r
		}
	}
	return nil
}

// Routers 返回当前路由所属的子路由组列表
func (g *Groups) Routers() []*Router { return g.routers }

// RemoveRouter 删除子路由
func (g *Groups) RemoveRouter(name string) {
	size := sliceutil.Delete(g.routers, func(i int) bool {
		return g.routers[i].name == name
	})
	g.routers = g.routers[:size]
}

// AppendMiddleware 添加中间件到尾部
func (g *Groups) AppendMiddleware(f mux.MiddlewareFunc) *Groups {
	g.middlewares.Append(f)
	return g
}

// PrependMiddleware 添加中间件到顶部
func (g *Groups) PrependMiddleware(f mux.MiddlewareFunc) *Groups {
	g.middlewares.Prepend(f)
	return g
}

func (g *Groups) CleanMiddlewares() { g.middlewares.Reset() }

// Name 返回名称
func (r *Router) Name() string { return r.name }
