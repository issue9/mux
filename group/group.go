// SPDX-License-Identifier: MIT

// Package group 提供了按条件进行分组路由的功能
package group

import (
	"fmt"
	"net/http"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v5"
	"github.com/issue9/mux/v5/internal/options"
	"github.com/issue9/mux/v5/middleware"
)

// Group 一组路由
//
// 当路由关联的 Matcher 返回 true 时，就会进入该路由。
// 如果多条路由的 Matcher 都返回 true，则第一条路由获得权限，
// 即使该路由最终返回 404，也不会再在其它路由里查找相应的路由。
// 所以在有多条子路由的情况下，第一条子路由不应该永远返回 true，
// 否则其它子路由永远无法到达。
type Group struct {
	routers []*router
	ms      *middleware.Middlewares
	options []mux.Option

	notFound http.Handler
	recovery mux.RecoverFunc
}

type router struct {
	*mux.Router
	matcher Matcher
}

// New 声明一个新的 Groups
//
// o 用于设置由 New 添加的路由，有关 NotFound 与 Recovery 的设置同时会作用于 Group。
func New(o ...mux.Option) *Group {
	opt, err := options.Build(o...)
	if err != nil {
		panic(err)
	}

	g := &Group{
		options: o,
		routers: make([]*router, 0, 1),

		notFound: opt.NotFound,
		recovery: opt.RecoverFunc,
	}
	g.ms = middleware.NewMiddlewares(http.HandlerFunc(g.serveHTTP))
	return g
}

func (g *Group) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if g.recovery != nil {
		defer func() {
			if err := recover(); err != nil {
				g.recovery(w, err)
			}
		}()
	}

	g.ms.ServeHTTP(w, r)
}

func (g *Group) serveHTTP(w http.ResponseWriter, r *http.Request) {
	for _, router := range g.routers {
		if req, ok := router.matcher.Match(r); ok {
			router.ServeHTTP(w, req)
			return
		}
	}
	g.notFound.ServeHTTP(w, r)
}

// New 声明新路由
//
// 初始化参数从 g.options 中获取，但是可以通过 o 作修改。
func (g *Group) New(name string, matcher Matcher, o ...mux.Option) *mux.Router {
	r := mux.NewRouter(name, append(g.options, o...)...)
	g.Add(matcher, r)
	return r
}

// Add 添加路由
func (g *Group) Add(matcher Matcher, r *mux.Router) {
	if matcher == nil {
		panic("参数 matcher 不能为空")
	}

	if r.Name() == "" {
		panic("r.Name() 不能为空")
	}

	// 重名检测
	index := sliceutil.Index(g.routers, func(i int) bool {
		return g.routers[i].Name() == r.Name()
	})
	if index > -1 {
		panic(fmt.Sprintf("已经存在名为 %s 的路由", r.Name()))
	}

	g.routers = append(g.routers, &router{Router: r, matcher: matcher})
}

// Router 返回指定名称的路由
func (g *Group) Router(name string) *mux.Router {
	for _, r := range g.routers {
		if r.Name() == name {
			return r.Router
		}
	}
	return nil
}

// Routers 返回路由列表
func (g *Group) Routers() []*mux.Router {
	routers := make([]*mux.Router, 0, len(g.routers))
	for _, r := range g.routers {
		routers = append(routers, r.Router)
	}
	return routers
}

// Remove 删除路由
func (g *Group) Remove(name string) {
	size := sliceutil.Delete(g.routers, func(i int) bool {
		return g.routers[i].Name() == name
	})
	g.routers = g.routers[:size]
}

// Middlewares 返回中间件管理接口
func (g *Group) Middlewares() *middleware.Middlewares { return g.ms }

func (g *Group) Routes() map[string]map[string][]string {
	routers := g.Routers()

	routes := make(map[string]map[string][]string, len(routers))
	for _, r := range routers {
		routes[r.Name()] = r.Routes()
	}

	return routes
}
