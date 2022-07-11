// SPDX-License-Identifier: MIT

// Package group 提供了针对一组路由的操作
package group

import (
	"fmt"
	"net/http"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v7"
	"github.com/issue9/mux/v7/internal/options"
	"github.com/issue9/mux/v7/internal/tree"
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
		options []options.Option

		ms []types.MiddlewareOf[T]
	}

	routerOf[T any] struct {
		r *mux.RouterOf[T]
		m Matcher
	}
)

// NewGroupOf 声明 GroupOf 对象
//
// 初始化参数与 mux.NewRouterOf 相同，这些参数最终也会被 GroupOf.New 传递给新对象。
func NewGroupOf[T any](call mux.CallOf[T], notFound T, methodNotAllowedBuilder, optionsBuilder types.BuildNodeHandleOf[T], o ...mux.Option) *GroupOf[T] {
	return &GroupOf[T]{
		routers: make([]*routerOf[T], 0, 1),

		call:                    call,
		notFound:                notFound,
		methodNotAllowedBuilder: methodNotAllowedBuilder,
		optionsBuilder:          optionsBuilder,
		options:                 o,

		ms: make([]types.MiddlewareOf[T], 0, 10),
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
//
// 新路由会继承 NewGroupOf 中指定的参数，其中的 o 可以覆盖由 NewGroupOf 中的相关参数；
func (g *GroupOf[T]) New(name string, matcher Matcher, o ...mux.Option) *mux.RouterOf[T] {
	o = g.mergeOption(o...)
	r := mux.NewRouterOf(name, g.call, g.notFound, g.methodNotAllowedBuilder, g.optionsBuilder, o...)
	g.Add(matcher, r)
	return r
}

// 将 g.options 与 o 合并，保证 g.options 在前且不会被破坏
func (g *GroupOf[T]) mergeOption(o ...mux.Option) []mux.Option {
	l1 := len(g.options)
	if l1 == 0 {
		return o
	}

	l2 := len(o)
	if l2 == 0 {
		return g.options
	}

	ret := make([]mux.Option, l1+l2)
	size := copy(ret, g.options)
	copy(ret[size:], o)
	return ret
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

	r.Use(g.ms...)

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
	g.notFound = tree.ApplyMiddlewares(g.notFound, m...)

	g.ms = append(g.ms, m...)
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
