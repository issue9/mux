// SPDX-License-Identifier: MIT

package mux

import (
	"fmt"
	"net/http"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v7/internal/params"
	"github.com/issue9/mux/v7/types"
)

type (
	// RoutersOf 一组路由的集合
	RoutersOf[T any] struct {
		routers []*RouterOf[T]
		call    CallOf[T]

		notFound T
		methodNotAllowedBuilder,
		optionsBuilder types.BuildNodeHandleOf[T]
	}

	// Matcher 验证一个请求是否符合要求
	//
	// Matcher 用于路由项的前置判断，用于对路由项进行归类，
	// 符合同一个 Matcher 的路由项，再各自进行路由。比如按域名进行分组路由。
	Matcher interface {
		// Match 验证请求是否符合当前对象的要求
		//
		// p 为匹配过程中生成的参数信息，必须为非空值；
		// ok 表示是否匹配成功；
		Match(r *http.Request, p types.Params) (ok bool)
	}

	// MatcherFunc 验证请求是否符合要求
	MatcherFunc func(*http.Request, types.Params) (ok bool)
)

// Match 实现 Matcher 接口
func (f MatcherFunc) Match(r *http.Request, p types.Params) bool { return f(r, p) }

func anyRouter(*http.Request, types.Params) bool { return true }

// NewRoutersOf 声明一个新的 RoutersOf
func NewRoutersOf[T any](b CallOf[T], notFound T, methodNotAllowedBuilder, opt types.BuildNodeHandleOf[T]) *RoutersOf[T] {
	return &RoutersOf[T]{
		routers:                 make([]*RouterOf[T], 0, 1),
		call:                    b,
		notFound:                notFound,
		methodNotAllowedBuilder: methodNotAllowedBuilder,
		optionsBuilder:          opt,
	}
}

func (rs *RoutersOf[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, router := range rs.routers {
		ps := params.New("")
		defer ps.Destroy()

		if ok := router.matcher.Match(r, ps); ok {
			router.serveHTTP(w, r, ps)
			return
		}
	}
	rs.call(w, r, nil, rs.notFound)
}

// New 声明新路由
func (rs *RoutersOf[T]) New(name string, matcher Matcher, o *Options) *RouterOf[T] {
	r := NewRouterOf(name, rs.call, rs.notFound, rs.methodNotAllowedBuilder, rs.optionsBuilder, o)
	rs.Add(matcher, r)
	return r
}

// Add 添加路由
//
// matcher 用于判断进入 r 的条件，如果为空，则表示不作判断。
// 如果有多个 matcher 都符合条件，第一个符合条件的 r 获得优胜；
func (rs *RoutersOf[T]) Add(matcher Matcher, r *RouterOf[T]) {
	if matcher == nil {
		matcher = MatcherFunc(anyRouter)
	}

	// 重名检测
	if sliceutil.Exists(rs.routers, func(rr *RouterOf[T]) bool {
		return rr.Name() == r.Name()
	}) {
		panic(fmt.Sprintf("已经存在名为 %s 的路由", r.Name()))
	}

	r.matcher = matcher
	rs.routers = append(rs.routers, r)
}

// Router 返回指定名称的路由
func (rs *RoutersOf[T]) Router(name string) *RouterOf[T] {
	for _, r := range rs.routers {
		if r.Name() == name {
			return r
		}
	}
	return nil
}

// Use 为所有已经注册的路由添加中间件
func (rs *RoutersOf[T]) Use(m ...types.MiddlewareOf[T]) {
	for _, r := range rs.Routers() {
		r.Use(m...)
	}
}

// Routers 返回路由列表
func (rs *RoutersOf[T]) Routers() []*RouterOf[T] { return rs.routers }

func (rs *RoutersOf[T]) Remove(name string) {
	rs.routers = sliceutil.Delete(rs.routers, func(r *RouterOf[T]) bool {
		return r.Name() == name
	})
}

func (rs *RoutersOf[T]) Routes() map[string]map[string][]string {
	routers := rs.Routers()

	routes := make(map[string]map[string][]string, len(routers))
	for _, r := range routers {
		routes[r.Name()] = r.Routes()
	}

	return routes
}
