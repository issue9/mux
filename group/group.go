// SPDX-License-Identifier: MIT

// Package group 提供了针对一组路由的操作
package group

import (
	"fmt"
	"net/http"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v7"
	"github.com/issue9/mux/v7/types"
)

type (
	// GroupOf 一组路由的集合
	GroupOf[T any] struct {
		routers []*routerOf[T]

		call     mux.CallOf[T]
		notFound T
		methodNotAllowedBuilder,
		optionsBuilder types.BuildNodeHandleOf[T]
	}

	routerOf[T any] struct {
		r *mux.RouterOf[T]
		m Matcher
	}
)

func NewGroupOf[T any](call mux.CallOf[T], notFound T, methodNotAllowedBuilder, opt types.BuildNodeHandleOf[T]) *GroupOf[T] {
	return &GroupOf[T]{
		routers: make([]*routerOf[T], 0, 1),

		call:                    call,
		notFound:                notFound,
		methodNotAllowedBuilder: methodNotAllowedBuilder,
		optionsBuilder:          opt,
	}
}

func (g *GroupOf[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, router := range g.routers {
		ps := types.NewContext()
		defer ps.Destroy()

		if ok := router.m.Match(r, ps); ok {
			router.r.ServeContext(w, r, ps)
			return
		}
	}
	g.call(w, r, nil, g.notFound)
}

// New 声明新路由
func (g *GroupOf[T]) New(name string, matcher Matcher, o ...mux.Option) *mux.RouterOf[T] {
	r := mux.NewRouterOf(name, g.call, g.notFound, g.methodNotAllowedBuilder, g.optionsBuilder, o...)
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
	if sliceutil.Exists(g.routers, func(rr *routerOf[T]) bool {
		return rr.r.Name() == r.Name()
	}) {
		panic(fmt.Sprintf("已经存在名为 %s 的路由", r.Name()))
	}

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
	g.routers = sliceutil.Delete(g.routers, func(r *routerOf[T]) bool {
		return r.r.Name() == name
	})
}

func (g *GroupOf[T]) Routes() map[string]map[string][]string {
	routers := g.Routers()

	routes := make(map[string]map[string][]string, len(routers))
	for _, r := range routers {
		routes[r.Name()] = r.Routes()
	}

	return routes
}
