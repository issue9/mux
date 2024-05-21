// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package group 提供了针对一组路由的操作
package group

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/issue9/mux/v8"
	"github.com/issue9/mux/v8/internal/options"
	"github.com/issue9/mux/v8/internal/tree"
	"github.com/issue9/mux/v8/types"
)

type (
	// GroupOf 一组路由的集合
	GroupOf[T any] struct {
		routers []*routerOf[T]

		call           mux.CallOf[T]
		notFound       T // 所有路由都找不差时调用的方法，该方法应用的中间件中 router 参数是为空的。
		originNotFound T // 这是应用中间件的 notFound
		methodNotAllowedBuilder,
		optionsBuilder types.BuildNodeHandleOf[T]
		options []options.Option

		middleware []types.MiddlewareOf[T]
	}

	routerOf[T any] struct {
		r *mux.RouterOf[T]
		m Matcher
	}
)

// NewOf 声明 GroupOf 对象
//
// 初始化参数与 [mux.NewRouterOf] 相同，这些参数最终也会被 [GroupOf.New] 传递给新对象。
func NewOf[T any](call mux.CallOf[T], notFound T, methodNotAllowedBuilder, optionsBuilder types.BuildNodeHandleOf[T], o ...mux.Option) *GroupOf[T] {
	return &GroupOf[T]{
		routers: make([]*routerOf[T], 0, 1),

		call:                    call,
		notFound:                notFound,
		originNotFound:          notFound,
		methodNotAllowedBuilder: methodNotAllowedBuilder,
		optionsBuilder:          optionsBuilder,
		options:                 o,

		middleware: make([]types.MiddlewareOf[T], 0, 10),
	}
}

func (g *GroupOf[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := types.NewContext()
	defer ctx.Destroy()

	for _, router := range g.routers {
		if ok := router.m.Match(r, ctx); ok {
			router.r.ServeContext(w, r, ctx)
			return
		}
		ctx.Reset()
	}

	g.call(w, r, ctx, g.notFound)
}

// New 声明新路由
//
// 新路由会继承 [NewOf] 中指定的参数，其中的 o 可以覆盖由 [NewOf] 中指定的相关参数；
func (g *GroupOf[T]) New(name string, matcher Matcher, o ...mux.Option) *mux.RouterOf[T] {
	o = slices.Concat(g.options, o)
	r := mux.NewRouterOf(name, g.call, g.originNotFound, g.methodNotAllowedBuilder, g.optionsBuilder, o...)
	g.Add(matcher, r)
	return r
}

// Add 添加路由
//
// matcher 用于判断进入 r 的条件，如果为空，则表示不作判断。
// 如果有多个 matcher 都符合条件，第一个符合条件的 r 获得优胜；
func (g *GroupOf[T]) Add(matcher Matcher, r *mux.RouterOf[T]) {
	if matcher == nil {
		matcher = MatcherFunc(anyRouter)
	}

	// 重名检测
	if slices.IndexFunc(g.routers, func(rr *routerOf[T]) bool { return rr.r.Name() == r.Name() }) >= 0 {
		panic(fmt.Sprintf("已经存在名为 %s 的路由", r.Name()))
	}

	r.Use(g.middleware...)

	g.routers = append(g.routers, &routerOf[T]{r: r, m: matcher})
}

// Router 返回指定名称的路由
func (g *GroupOf[T]) Router(name string) *mux.RouterOf[T] {
	for _, r := range g.routers {
		if r.r.Name() == name {
			return r.r
		}
	}
	return nil
}

// Use 为所有已经注册的路由添加中间件
func (g *GroupOf[T]) Use(m ...types.MiddlewareOf[T]) {
	for _, r := range g.routers {
		r.r.Use(m...)
	}

	g.notFound = tree.ApplyMiddleware(g.notFound, "", "", "", m...)
	g.middleware = append(g.middleware, m...)
}

// Routers 返回路由列表
func (g *GroupOf[T]) Routers() []*mux.RouterOf[T] {
	rs := make([]*mux.RouterOf[T], 0, len(g.routers))
	for _, r := range g.routers {
		rs = append(rs, r.r)
	}
	return rs
}

func (g *GroupOf[T]) Remove(name string) {
	g.routers = slices.DeleteFunc(g.routers, func(r *routerOf[T]) bool { return r.r.Name() == name })
}

func (g *GroupOf[T]) Routes() map[string]map[string][]string {
	routers := g.Routers()

	routes := make(map[string]map[string][]string, len(routers))
	for _, r := range routers {
		routes[r.Name()] = r.Routes()
	}

	return routes
}
