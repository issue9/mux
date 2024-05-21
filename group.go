// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package mux

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/issue9/mux/v8/internal/tree"
	"github.com/issue9/mux/v8/types"
)

type (
	// GroupOf 一组路由的集合
	GroupOf[T any] struct {
		routers []*RouterOf[T]

		call           CallOf[T]
		notFound       T // 所有路由都找不差时调用的方法，该方法应用的中间件中 router 参数是为空的。
		originNotFound T // 这是应用中间件的 notFound
		methodNotAllowedBuilder,
		optionsBuilder types.BuildNodeHandleOf[T]
		options []Option

		middleware []types.MiddlewareOf[T]
	}
)

// NewGroupOf 声明 GroupOf 对象
//
// 初始化参数与 [NewRouterOf] 相同，这些参数最终也会被 [GroupOf.New] 传递给新对象。
func NewGroupOf[T any](call CallOf[T], notFound T, methodNotAllowedBuilder, optionsBuilder types.BuildNodeHandleOf[T], o ...Option) *GroupOf[T] {
	return &GroupOf[T]{
		routers: make([]*RouterOf[T], 0, 1),

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

	// 如果已经在 [NewGroupOf] 中指定了 Recovery 的相关参数，那么在初始化 g.routers
	// 时会自动为各个路由添加，无需在此处再次添加 Recovery 的处理。

	for _, router := range g.routers {
		if ok := router.matcher.Match(r, ctx); ok {
			router.serveContext(w, r, ctx)
			return
		}
		ctx.Reset()
	}

	// BUG 可能 panic，应该根据参数 recovery
	g.call(w, r, ctx, g.notFound)
}

// New 声明新路由
//
// 新路由会继承 [NewGroupOf] 中指定的参数，其中的 o 可以覆盖由 [NewOf] 中指定的相关参数；
func (g *GroupOf[T]) New(name string, matcher Matcher, o ...Option) *RouterOf[T] {
	o = slices.Concat(g.options, o)
	r := NewRouterOf(name, g.call, g.originNotFound, g.methodNotAllowedBuilder, g.optionsBuilder, o...)
	g.Add(matcher, r)
	return r
}

// Add 添加路由
//
// matcher 用于判断进入 r 的条件，如果为空，则表示不作判断。
// 如果有多个 matcher 都符合条件，第一个符合条件的 r 获得优胜；
func (g *GroupOf[T]) Add(matcher Matcher, r *RouterOf[T]) {
	if matcher == nil {
		matcher = MatcherFunc(anyRouter)
	}

	// 重名检测
	if slices.IndexFunc(g.routers, func(rr *RouterOf[T]) bool { return rr.Name() == r.Name() }) >= 0 {
		panic(fmt.Sprintf("已经存在名为 %s 的路由", r.Name()))
	}

	r.Use(g.middleware...)
	r.matcher = matcher
	g.routers = append(g.routers, r)
}

// Router 返回指定名称的路由
func (g *GroupOf[T]) Router(name string) *RouterOf[T] {
	for _, r := range g.routers {
		if r.Name() == name {
			return r
		}
	}
	return nil
}

// Use 为所有已经注册的路由添加中间件
func (g *GroupOf[T]) Use(m ...types.MiddlewareOf[T]) {
	for _, r := range g.routers {
		r.Use(m...)
	}

	g.notFound = tree.ApplyMiddleware(g.notFound, "", "", "", m...)
	g.middleware = append(g.middleware, m...)
}

// Routers 返回路由列表
func (g *GroupOf[T]) Routers() []*RouterOf[T] { return g.routers }

func (g *GroupOf[T]) Remove(name string) {
	g.routers = slices.DeleteFunc(g.routers, func(r *RouterOf[T]) bool { return r.Name() == name })
}

func (g *GroupOf[T]) Routes() map[string]map[string][]string {
	routers := g.Routers()

	routes := make(map[string]map[string][]string, len(routers))
	for _, r := range routers {
		routes[r.Name()] = r.Routes()
	}

	return routes
}
