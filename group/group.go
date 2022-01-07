// SPDX-License-Identifier: MIT

// Package group 提供了按条件进行分组路由的功能
package group

import (
	"fmt"
	"net/http"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/internal/options"
)

// GroupOf 一组路由
//
// 当路由关联的 Matcher 返回 true 时，就会进入该路由。
// 如果多条路由的 Matcher 都返回 true，则第一条路由获得权限，
// 即使该路由最终返回 404，也不会再在其它路由里查找相应的路由。
// 所以在有多条子路由的情况下，第一条子路由不应该永远返回 true，
// 否则其它子路由永远无法到达。
type GroupOf[T any] struct {
	routers []*routerOf[T]
	options []mux.Option
	ms      []mux.MiddlewareFuncOf[T]
	b       mux.BuildHandlerFuncOf[T]

	notFound http.Handler
	recovery mux.RecoverFunc
}

type routerOf[T any] struct {
	*mux.RouterOf[T]
	matcher Matcher
}

// NewOf 声明一个新的 GroupOf
//
// o 用于设置由 GroupOf.New 添加的路由，有关 NotFound 与 Recovery 的设置同时会作用于 GroupOf。
func NewOf[T any](b mux.BuildHandlerFuncOf[T], ms []mux.MiddlewareFuncOf[T], o ...mux.Option) *GroupOf[T] {
	opt, err := options.Build(o...)
	if err != nil {
		panic(err)
	}

	g := &GroupOf[T]{
		options: o,
		routers: make([]*routerOf[T], 0, 1),
		ms:      ms,
		b:       b,

		notFound: opt.NotFound,
		recovery: opt.RecoverFunc,
	}
	return g
}

func (g *GroupOf[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if g.recovery != nil {
		defer func() {
			if err := recover(); err != nil {
				g.recovery(w, err)
			}
		}()
	}

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
func (g *GroupOf[T]) New(name string, matcher Matcher, ms []mux.MiddlewareFuncOf[T], o ...mux.Option) *mux.RouterOf[T] {
	mm := make([]mux.MiddlewareFuncOf[T], 0, len(g.ms)+len(ms))
	mm = append(mm, g.ms...)
	mm = append(mm, ms...)
	r := mux.NewRouterOf[T](name, g.b, mm, append(g.options, o...)...)
	g.Add(matcher, r)
	return r
}

// Add 添加路由
func (g *GroupOf[T]) Add(matcher Matcher, r *mux.RouterOf[T]) {
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

	g.routers = append(g.routers, &routerOf[T]{RouterOf: r, matcher: matcher})
}

// Router 返回指定名称的路由
func (g *GroupOf[T]) Router(name string) *mux.RouterOf[T] {
	for _, r := range g.routers {
		if r.Name() == name {
			return r.RouterOf
		}
	}
	return nil
}

// Routers 返回路由列表
func (g *GroupOf[T]) Routers() []*mux.RouterOf[T] {
	routers := make([]*mux.RouterOf[T], 0, len(g.routers))
	for _, r := range g.routers {
		routers = append(routers, r.RouterOf)
	}
	return routers
}

// Remove 删除路由
func (g *GroupOf[T]) Remove(name string) {
	size := sliceutil.Delete(g.routers, func(i int) bool {
		return g.routers[i].Name() == name
	})
	g.routers = g.routers[:size]
}

func (g *GroupOf[T]) Routes() map[string]map[string][]string {
	routers := g.Routers()

	routes := make(map[string]map[string][]string, len(routers))
	for _, r := range routers {
		routes[r.Name()] = r.Routes()
	}

	return routes
}
