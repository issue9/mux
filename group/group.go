// SPDX-License-Identifier: MIT

// Package group 提供了按条件进行分组路由的功能
package group

import (
	"errors"
	"net/http"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v5"
)

// ErrRouterExists 表示是否存在同名的路由名称
var ErrRouterExists = errors.New("该名称的路由已经存在")

// Groups 管理一组路由
type Groups struct {
	routers []*router
	ms      *mux.Middlewares

	disableHead bool
	cors        *mux.CORS
}

type router struct {
	*mux.Router
	matcher Matcher
}

// Default 返回默认参数的 New
func Default() *Groups { return New(false, mux.DeniedCORS()) }

// New 声明一个新的 Mux
//
// disableHead 是否禁用根据 Get 请求自动生成 HEAD 请求；
// cors 跨域请求的相关设置项；
func New(disableHead bool, cors *mux.CORS) *Groups {
	g := &Groups{
		routers:     make([]*router, 0, 1),
		disableHead: disableHead,
		cors:        cors,
	}
	g.ms = mux.NewMiddlewares(http.HandlerFunc(g.serveHTTP))
	return g
}

func (g *Groups) ServeHTTP(w http.ResponseWriter, r *http.Request) { g.ms.ServeHTTP(w, r) }

func (g *Groups) serveHTTP(w http.ResponseWriter, r *http.Request) {
	for _, router := range g.routers {
		if req, ok := router.matcher.Match(r); ok {
			router.ServeHTTP(w, req)
			return
		}
	}
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

// NewRouter 声明新路由
//
// 与 AddRouter 的区别在于，NewRouter 参数从 Groups 继承，而 AddRouter 的路由，其参数可自定义。
func (g *Groups) NewRouter(name string, matcher Matcher) (*mux.Router, error) {
	r, err := mux.NewRouter(name, g.disableHead, g.cors)
	if err != nil {
		return nil, err
	}
	return g.addRouter(matcher, r)
}

// AddRouter 添加路由
//
// 该路由只有符合 Matcher 的要求才会进入，其它与 mux.Router 功能相同。
// 当 Matcher 与其它路由组的判断有重复时，第一条返回 true 的路由组获胜，
// 即使该路由组最终返回 404，也不会再在其它路由组里查找相应的路由。
// 所以在有多条子路由的情况下，第一条子路由不应该永远返回 true，
// 这样会造成其它子路由永远无法到达。
//
// matcher 路由的准入条件；
func (g *Groups) AddRouter(matcher Matcher, r *mux.Router) error {
	_, err := g.addRouter(matcher, r)
	return err
}

func (g *Groups) addRouter(matcher Matcher, r *mux.Router) (*mux.Router, error) {
	if r.Name() == "" {
		panic("参数 name 不能为空")
	}
	if matcher == nil {
		panic("参数 matcher 不能为空")
	}

	// 重名检测
	index := sliceutil.Index(g.routers, func(i int) bool {
		return g.routers[i].Name() == r.Name()
	})
	if index > -1 {
		return nil, ErrRouterExists
	}

	router := &router{
		Router:  r,
		matcher: matcher,
	}
	g.routers = append(g.routers, router)

	return router.Router, nil
}

// Router 返回指定名称的 *mux.Router 实例
func (g *Groups) Router(name string) *mux.Router {
	for _, r := range g.routers {
		if r.Name() == name {
			return r.Router
		}
	}
	return nil
}

// Routers 返回当前路由所属的子路由组列表
func (g *Groups) Routers() []*mux.Router {
	routers := make([]*mux.Router, 0, len(g.routers))
	for _, r := range g.routers {
		routers = append(routers, r.Router)
	}
	return routers
}

// RemoveRouter 删除子路由
func (g *Groups) RemoveRouter(name string) {
	size := sliceutil.Delete(g.routers, func(i int) bool {
		return g.routers[i].Name() == name
	})
	g.routers = g.routers[:size]
}

// Middlewares 返回中间件管理接口
func (g *Groups) Middlewares() *mux.Middlewares { return g.ms }
