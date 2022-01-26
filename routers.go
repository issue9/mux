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
		routers []*RouterOf[T]
		options []Option
		ms      []MiddlewareFuncOf[T]
		call    CallOf[T]

		notFound http.Handler
		recovery RecoverFunc
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
// o 用于设置由 RoutersOf.New 添加的路由，有关 NotFound 与 Recovery 的设置同时会作用于 RoutersOf。
func NewRoutersOf[T any](b CallOf[T], ms []MiddlewareFuncOf[T], o ...Option) *RoutersOf[T] {
	opt, err := buildOptions(o...)
	if err != nil {
		panic(err)
	}

	g := &RoutersOf[T]{
		options: o,
		routers: make([]*RouterOf[T], 0, 1),
		ms:      ms,
		call:    b,

		notFound: opt.NotFound,
		recovery: opt.RecoverFunc,
	}
	return g
}

func (rs *RoutersOf[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if rs.recovery != nil {
		defer func() {
			if err := recover(); err != nil {
				rs.recovery(w, err)
			}
		}()
	}

	for _, router := range rs.routers {
		if ps, ok := router.matcher.Match(r); ok {
			router.serveHTTP(w, r, ps)
			return
		}
	}
	rs.notFound.ServeHTTP(w, r)
}

// New 声明新路由
//
// 初始化参数从 g.options 中获取，但是可以通过 o 作修改。
func (rs *RoutersOf[T]) New(name string, matcher Matcher, ms []MiddlewareFuncOf[T], o ...Option) *RouterOf[T] {
	mm := make([]MiddlewareFuncOf[T], 0, len(rs.ms)+len(ms))
	mm = append(mm, rs.ms...)
	mm = append(mm, ms...)
	r := NewRouterOf[T](name, rs.call, mm, append(rs.options, o...)...)
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
	index := sliceutil.Index(rs.routers, func(i int) bool {
		return rs.routers[i].Name() == r.Name()
	})
	if index > -1 {
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

// Remove 删除路由
func (rs *RoutersOf[T]) Remove(name string) {
	size := sliceutil.Delete(rs.routers, func(i int) bool {
		return rs.routers[i].Name() == name
	})
	rs.routers = rs.routers[:size]
}

func (rs *RoutersOf[T]) Routes() map[string]map[string][]string {
	routers := rs.Routers()

	routes := make(map[string]map[string][]string, len(routers))
	for _, r := range routers {
		routes[r.Name()] = r.Routes()
	}

	return routes
}
