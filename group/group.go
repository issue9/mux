// SPDX-License-Identifier: MIT

// Package group 提供了按条件进行分组路由的功能
package group

import (
	"fmt"
	"net/http"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v5"
)

// Groups 管理一组路由
//
// 当路由关联的 Matcher 返回 true 时，就会进入该路由。
// 如果多条路由的 Matcher 都返回 true，则第一条路由获得权限，
// 即使该路由最终返回 404，也不会再在其它路由里查找相应的路由。
// 所以在有多条子路由的情况下，第一条子路由不应该永远返回 true，
// 否则其它子路由永远无法到达。
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
// 与 AddRouter 的区别在于：NewRouter 参数从 Groups 继承，而 AddRouter 的路由，其参数可自定义。
func (g *Groups) NewRouter(name string, matcher Matcher) (*mux.Router, error) {
	r, err := mux.NewRouter(name, g.disableHead, g.cors)
	if err == nil {
		g.AddRouter(matcher, r)
	}
	return r, err
}

// AddRouter 添加路由
func (g *Groups) AddRouter(matcher Matcher, r *mux.Router) {
	if matcher == nil {
		panic("参数 matcher 不能为空")
	}

	if r.Name() == "" {
		panic("r.Name() 不能为空")
	}

	// 重名检测
	index := sliceutil.Index(g.routers, func(i int) bool {
		return g.routers[i].Name() == r.Name()
	})
	if index > -1 {
		panic(fmt.Sprintf("已经存在名为 %s 的路由", r.Name()))
	}

	g.routers = append(g.routers, &router{Router: r, matcher: matcher})
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
