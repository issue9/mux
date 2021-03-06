// SPDX-License-Identifier: MIT

// Package group 提供了按条件进行分组路由的功能
package group

import (
	"errors"
	"net/http"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v5"
	"github.com/issue9/mux/v5/internal/tree"
)

// ErrRouterExists 表示是否存在同名的路由名称
var ErrRouterExists = errors.New("该名称的路由已经存在")

// Groups 管理一组路由
type Groups struct {
	routers []*Router
	ms      *mux.Middlewares

	disableHead bool
	cors        *mux.CORS
	notFound,
	methodNotAllowed http.HandlerFunc
}

// Router 单个路由
type Router struct {
	*mux.Router
	name    string
	matcher Matcher
}

// Default 返回默认参数的 New
func Default() *Groups { return New(false, mux.DeniedCORS(), nil, nil) }

// New 声明一个新的 Mux
//
// disableHead 是否禁用根据 Get 请求自动生成 HEAD 请求；
// cors 跨域请求的相关设置项；
// notFound 404 页面的处理方式，为 nil 时会调用默认的方式进行处理；
// methodNotAllowed 405 页面的处理方式，为 nil 时会调用默认的方式进行处理；
func New(disableHead bool, cors *mux.CORS, notFound, methodNotAllowed http.HandlerFunc) *Groups {
	if notFound == nil {
		notFound = tree.DefaultNotFound
	}
	if methodNotAllowed == nil {
		methodNotAllowed = tree.DefaultMethodNotAllowed
	}

	g := &Groups{
		routers: make([]*Router, 0, 1),

		disableHead:      disableHead,
		cors:             cors,
		notFound:         notFound,
		methodNotAllowed: methodNotAllowed,
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
	g.notFound(w, r)
}

// NewRouter 声明新路由
//
// 与 AddRouter 的区别在于，NewRouter 参数从 Groups 继承，而 AddRouter 的路由，其参数可自定义。
func (g *Groups) NewRouter(name string, matcher Matcher) (*Router, error) {
	r, err := mux.NewRouter(g.disableHead, g.cors, g.notFound, g.methodNotAllowed)
	if err != nil {
		return nil, err
	}
	return g.addRouter(name, matcher, r)
}

// AddRouter 添加路由
//
// 该路由只有符合 Matcher 的要求才会进入，其它与 Router 功能相同。
// 当 Matcher 与其它路由组的判断有重复时，第一条返回 true 的路由组获胜，
// 即使该路由组最终返回 404，也不会再在其它路由组里查找相应的路由。
// 所以在有多条子路由的情况下，第一条子路由不应该永远返回 true，
// 这样会造成其它子路由永远无法到达。
//
// name 表示该路由组的名称，需要唯一，否则返回 false；
// matcher 路由的准入条件，如果为空，则此条路由匹配时会被排在最后，
// 只能有一个路由的 matcher 为空，否则会 panic；
func (g *Groups) AddRouter(name string, matcher Matcher, r *mux.Router) error {
	_, err := g.addRouter(name, matcher, r)
	return err
}

func (g *Groups) addRouter(name string, matcher Matcher, r *mux.Router) (*Router, error) {
	if name == "" {
		panic("参数 name 不能为空")
	}
	if matcher == nil {
		panic("参数 matcher 不能为空")
	}

	// 重名检测
	index := sliceutil.Index(g.routers, func(i int) bool {
		return g.routers[i].name == name
	})
	if index > -1 {
		return nil, ErrRouterExists
	}

	router := &Router{
		Router:  r,
		name:    name,
		matcher: matcher,
	}
	g.routers = append(g.routers, router)

	return router, nil
}

// Router 返回指定名称的 *Router 实例
func (g *Groups) Router(name string) *Router {
	for _, r := range g.routers {
		if r.name == name {
			return r
		}
	}
	return nil
}

// Routers 返回当前路由所属的子路由组列表
func (g *Groups) Routers() []*Router { return g.routers }

// RemoveRouter 删除子路由
func (g *Groups) RemoveRouter(name string) {
	size := sliceutil.Delete(g.routers, func(i int) bool {
		return g.routers[i].name == name
	})
	g.routers = g.routers[:size]
}

// Middlewares 返回中间件管理接口
func (g *Groups) Middlewares() *mux.Middlewares { return g.ms }

// Name 返回名称
func (r *Router) Name() string { return r.name }
