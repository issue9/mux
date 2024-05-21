// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package mux

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/issue9/mux/v9/internal/tree"
	"github.com/issue9/mux/v9/types"
)

type (
	// Group 一组路由的集合
	Group[T any] struct {
		routers []*Router[T]
		ms      []types.Middleware[T]

		call           CallFunc[T]
		notFound       T // 所有路由都找不差时调用的方法，该方法应用的中间件中 router 参数是为空的。
		originNotFound T // 这是应用中间件的 notFound
		methodNotAllowedBuilder,
		optionsBuilder types.BuildNodeHandler[T]
		options []Option
	}
)

// NewGroup 声明 Group 对象
//
// 初始化参数与 [NewRouter] 相同，这些参数最终也会被 [Group.New] 传递给新对象。
func NewGroup[T any](call CallFunc[T], notFound T, methodNotAllowedBuilder, optionsBuilder types.BuildNodeHandler[T], o ...Option) *Group[T] {
	return &Group[T]{
		routers: make([]*Router[T], 0, 1),
		ms:      make([]types.Middleware[T], 0, 10),

		call:                    call,
		notFound:                notFound,
		originNotFound:          notFound,
		methodNotAllowedBuilder: methodNotAllowedBuilder,
		optionsBuilder:          optionsBuilder,
		options:                 o,
	}
}

func (g *Group[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := types.NewContext()
	defer ctx.Destroy()

	// 如果已经在 [NewGroup] 中指定了 Recovery 的相关参数，那么在初始化 g.routers
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
// 新路由会继承 [NewGroup] 中指定的参数，其中的 o 可以覆盖由 [NewGroup] 中指定的相关参数；
func (g *Group[T]) New(name string, matcher Matcher, o ...Option) *Router[T] {
	o = slices.Concat(g.options, o)
	r := NewRouter(name, g.call, g.originNotFound, g.methodNotAllowedBuilder, g.optionsBuilder, o...)
	g.Add(matcher, r)
	return r
}

// Add 添加路由
//
// matcher 用于判断进入 r 的条件，如果为空，则表示不作判断。
// 如果有多个 matcher 都符合条件，第一个符合条件的 r 获得优胜；
func (g *Group[T]) Add(matcher Matcher, r *Router[T]) {
	if matcher == nil {
		matcher = MatcherFunc(anyRouter)
	}

	// 重名检测
	if slices.IndexFunc(g.routers, func(rr *Router[T]) bool { return rr.Name() == r.Name() }) >= 0 {
		panic(fmt.Sprintf("已经存在名为 %s 的路由", r.Name()))
	}

	r.Use(g.ms...)
	r.matcher = matcher
	g.routers = append(g.routers, r)
}

// Router 返回指定名称的路由
func (g *Group[T]) Router(name string) *Router[T] {
	for _, r := range g.routers {
		if r.Name() == name {
			return r
		}
	}
	return nil
}

// Use 为所有已经注册的路由添加中间件
func (g *Group[T]) Use(m ...types.Middleware[T]) {
	for _, r := range g.routers {
		r.Use(m...)
	}

	g.notFound = tree.ApplyMiddleware(g.notFound, "", "", "", m...)
	g.ms = append(g.ms, m...)
}

// Routers 返回路由列表
func (g *Group[T]) Routers() []*Router[T] { return g.routers }

func (g *Group[T]) Remove(name string) {
	g.routers = slices.DeleteFunc(g.routers, func(r *Router[T]) bool { return r.Name() == name })
}

func (g *Group[T]) Routes() map[string]map[string][]string {
	routers := g.Routers()

	routes := make(map[string]map[string][]string, len(routers))
	for _, r := range routers {
		routes[r.Name()] = r.Routes()
	}

	return routes
}
