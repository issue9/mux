// SPDX-License-Identifier: MIT

package mux

import (
	"fmt"
	"net/http"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v6/params"
)

type (
	// RoutersOf 一组路由的集合
	RoutersOf[T any] struct {
		routers  []*RouterOf[T]
		call     CallOf[T]
		notFound http.Handler
	}

	// Matcher 验证一个请求是否符合要求
	//
	// Matcher 用于路由项的前置判断，用于对路由项进行归类，
	// 符合同一个 Matcher 的路由项，再各自进行路由。比如按域名进行分组路由。
	Matcher interface {
		// Match 验证请求是否符合当前对象的要求
		//
		// ps 为匹配过程中生成的参数信息，可以返回 nil；
		// ok 表示是否匹配成功；
		Match(r *http.Request) (ps params.Params, ok bool)
	}

	// MatcherFunc 验证请求是否符合要求
	//
	// ps 为匹配过程中生成的参数信息，可以返回 nil；
	// ok 表示是否匹配成功；
	MatcherFunc func(r *http.Request) (ps Params, ok bool)
)

// Match 实现 Matcher 接口
func (f MatcherFunc) Match(r *http.Request) (params.Params, bool) { return f(r) }

func anyRouter(*http.Request) (params.Params, bool) { return nil, true }

// NewRoutersOf 声明一个新的 RoutersOf
//
// notFound 表示所有路由都不匹配时的处理方式，如果为空，则调用 http.NotFoundHandler。
func NewRoutersOf[T any](b CallOf[T], notFound http.Handler) *RoutersOf[T] {
	if notFound == nil {
		notFound = http.NotFoundHandler()
	}

	return &RoutersOf[T]{
		routers:  make([]*RouterOf[T], 0, 1),
		call:     b,
		notFound: notFound,
	}
}

func (rs *RoutersOf[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, router := range rs.routers {
		if ps, ok := router.matcher.Match(r); ok {
			router.serveHTTP(w, r, ps)
			return
		}
	}
	rs.notFound.ServeHTTP(w, r)
}

// New 声明新路由
func (rs *RoutersOf[T]) New(name string, matcher Matcher, o *OptionsOf[T]) *RouterOf[T] {
	r := NewRouterOf(name, rs.call, o)
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
