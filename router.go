// SPDX-License-Identifier: MIT

package mux

import (
	"context"
	"net/http"
	"sort"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v4/group"
	"github.com/issue9/mux/v4/internal/tree"
	"github.com/issue9/mux/v4/params"
)

// Router 提供了基本的路由项添加和删除等功能
type Router struct {
	mux         *Mux
	name        string
	matcher     group.Matcher
	tree        *tree.Tree
	middlewares *Middlewares
	last        bool // 在多路由中，有此标记的排在最后。
}

// Routers 返回当前路由所属的子路由组列表
func (mux *Mux) Routers() []*Router { return mux.routers }

// NewRouter 添加子路由
//
// 该路由只有符合 group.Matcher 的要求才会进入，其它与 Router 功能相同。
// 当 group.Matcher 与其它路由组的判断有重复时，第一条返回 true 的路由组获胜，
// 即使该路由组最终返回 404，也不会再在其它路由组里查找相应的路由。
// 所以在有多条子路由的情况下，第一条子路由不应该永远返回 true，
// 这样会造成其它子路由永远无法到达。
//
// name 表示该路由组的名称，需要唯一，否则返回 false；
// matcher 路由的准入条件，如果为空，则此条路由匹配时会被排在最后，
// 只有一个路由的 matcher 为空，否则会 panic。
func (mux *Mux) NewRouter(name string, matcher group.Matcher) (r *Router, ok bool) {
	if name == "" {
		panic("参数 name 不能为空")
	}

	// 重名检测
	index := sliceutil.Index(mux.routers, func(i int) bool {
		return mux.routers[i].name == name
	})
	if index > -1 {
		return nil, false
	}

	last := matcher == nil
	if last {
		matcher = group.MatcherFunc(group.Any)
		index := sliceutil.Index(mux.routers, func(i int) bool { return mux.routers[i].last })
		if index > -1 {
			panic("已经存在一个 matcher 参数为空的路由")
		}
	}

	r = &Router{
		mux:     mux,
		name:    name,
		matcher: matcher,
		tree:    tree.New(mux.disableOptions, mux.disableHead),
		last:    last,
	}
	r.middlewares = NewMiddlewares(http.HandlerFunc(r.serveHTTP))
	mux.routers = append(mux.routers, r)
	sortRouters(mux.routers)
	return r, true
}

func sortRouters(rs []*Router) {
	sort.SliceStable(rs, func(i, j int) bool {
		if rs[i].last {
			return rs[j].last
		}
		return rs[j].last
	})
}

// RemoveRouter 删除子路由
func (mux *Mux) RemoveRouter(name string) {
	size := sliceutil.Delete(mux.routers, func(i int) bool {
		return mux.routers[i].name == name
	})
	mux.routers = mux.routers[:size]
}

// Name 当前路由组的名称
func (r *Router) Name() string { return r.name }

// Clean 清除当前路由组的所有路由项
func (r *Router) Clean() *Router {
	r.tree.Clean("")
	return r
}

// Routes 返回当前路由组的路由项
//
// ignoreHead 是否忽略自动生成的 HEAD 请求；
// ignoreOptions 是否忽略自动生成的 OPTIONS 请求；
func (r *Router) Routes(ignoreHead, ignoreOptions bool) map[string][]string {
	return r.tree.All(ignoreHead, ignoreOptions)
}

// Remove 移除指定的路由项
//
// 当未指定 methods 时，将删除所有 method 匹配的项。
// 指定错误的 methods 值，将自动忽略该值。
func (r *Router) Remove(pattern string, methods ...string) *Router {
	r.tree.Remove(pattern, methods...)
	return r
}

// Handle 添加一条路由数据
//
// pattern 为路由匹配模式，可以是正则匹配也可以是字符串匹配，
// 若语法不正确，则直接 panic，可以通过 IsWell 检测语法的有效性，其它接口也相同；
// methods 该路由项对应的请求方法，如果未指定值，则表示所有支持的请求方法，
// 但不包含 OPTIONS 和 HEAD。
func (r *Router) Handle(pattern string, h http.Handler, methods ...string) error {
	return r.tree.Add(pattern, h, methods...)
}

// SetAllow 将 OPTIONS 请求方法的报头 allow 值固定为指定的值
//
// 若无特殊需求，不用调用此方法，系统会自动计算符合当前路由的请求方法列表。
// 如果想实现对处理方法的自定义，可以显示地调用 Handle 方法:
//  Router.Handle("/api/1", handle, http.MethodOptions)
//
// Options 与 SetAllow 功能上完全相同，只是对错误处理上有所有区别。
// Options 在出错时 panic，而 SetAllow 会返回错误信息。
func (r *Router) SetAllow(pattern string, allow string) error {
	return r.tree.SetAllow(pattern, allow)
}

// Options 将 OPTIONS 请求方法的报头 allow 值固定为指定的值
//
// 若无特殊需求，不用调用此方法，系统会自动计算符合当前路由的请求方法列表。
// 如果想实现对处理方法的自定义，可以显示地调用 Handle 方法:
//  Router.Handle("/api/1", handle, http.MethodOptions)
//
// Options 与 SetAllow 功能上完全相同，只是对错误处理上有所有区别。
// Options 在出错时 panic，而 SetAllow 会返回错误信息。
func (r *Router) Options(pattern string, allow string) *Router {
	if err := r.SetAllow(pattern, allow); err != nil {
		panic(err)
	}
	return r
}

func (r *Router) handle(pattern string, h http.Handler, methods ...string) *Router {
	if err := r.Handle(pattern, h, methods...); err != nil {
		panic(err)
	}
	return r
}

// Get 相当于 Router.Handle(pattern, h, http.MethodGet) 的简易写法
func (r *Router) Get(pattern string, h http.Handler) *Router {
	return r.handle(pattern, h, http.MethodGet)
}

// Post 相当于 Router.Handle(pattern, h, http.MethodPost) 的简易写法
func (r *Router) Post(pattern string, h http.Handler) *Router {
	return r.handle(pattern, h, http.MethodPost)
}

// Delete 相当于 Router.Handle(pattern, h, http.MethodDelete) 的简易写法
func (r *Router) Delete(pattern string, h http.Handler) *Router {
	return r.handle(pattern, h, http.MethodDelete)
}

// Put 相当于 Router.Handle(pattern, h, http.MethodPut) 的简易写法
func (r *Router) Put(pattern string, h http.Handler) *Router {
	return r.handle(pattern, h, http.MethodPut)
}

// Patch 相当于 Router.Handle(pattern, h, http.MethodPatch) 的简易写法
func (r *Router) Patch(pattern string, h http.Handler) *Router {
	return r.handle(pattern, h, http.MethodPatch)
}

// Any 相当于 Router.Handle(pattern, h) 的简易写法
func (r *Router) Any(pattern string, h http.Handler) *Router {
	return r.handle(pattern, h)
}

// HandleFunc 功能同 Router.Handle()，但是将第二个参数从 http.Handler 换成了 http.HandlerFunc
func (r *Router) HandleFunc(pattern string, f http.HandlerFunc, methods ...string) error {
	return r.Handle(pattern, f, methods...)
}

func (r *Router) handleFunc(pattern string, f http.HandlerFunc, methods ...string) *Router {
	return r.handle(pattern, f, methods...)
}

// GetFunc 相当于 Router.HandleFunc(pattern, func, http.MethodGet) 的简易写法
func (r *Router) GetFunc(pattern string, f http.HandlerFunc) *Router {
	return r.handleFunc(pattern, f, http.MethodGet)
}

// PutFunc 相当于 Router.HandleFunc(pattern, func, http.MethodPut) 的简易写法
func (r *Router) PutFunc(pattern string, f http.HandlerFunc) *Router {
	return r.handleFunc(pattern, f, http.MethodPut)
}

// PostFunc 相当于 Router.HandleFunc(pattern, func, "POST") 的简易写法
func (r *Router) PostFunc(pattern string, f http.HandlerFunc) *Router {
	return r.handleFunc(pattern, f, http.MethodPost)
}

// DeleteFunc 相当于 Router.HandleFunc(pattern, func, http.MethodDelete) 的简易写法
func (r *Router) DeleteFunc(pattern string, f http.HandlerFunc) *Router {
	return r.handleFunc(pattern, f, http.MethodDelete)
}

// PatchFunc 相当于 Router.HandleFunc(pattern, func, http.MethodPatch) 的简易写法
func (r *Router) PatchFunc(pattern string, f http.HandlerFunc) *Router {
	return r.handleFunc(pattern, f, http.MethodPatch)
}

// AnyFunc 相当于 Router.HandleFunc(pattern, func) 的简易写法
func (r *Router) AnyFunc(pattern string, f http.HandlerFunc) *Router {
	return r.handleFunc(pattern, f)
}

// URL 根据参数生成地址
//
// pattern 为路由项的定义内容；
// params 为路由项中的参数，键名为参数名，键值为参数值。
func (r *Router) URL(pattern string, params map[string]string) (string, error) {
	return r.tree.URL(pattern, params)
}

func (r *Router) serveHTTP(w http.ResponseWriter, req *http.Request) {
	hs, ps := r.tree.Handler(req.URL.Path)
	if ps == nil {
		r.mux.notFound(w, req)
		return
	}

	h := hs.Handler(req.Method)
	if h == nil {
		w.Header().Set("Allow", hs.Options())
		r.mux.methodNotAllowed(w, req)
		return
	}

	if len(ps) > 0 {
		req = req.WithContext(context.WithValue(req.Context(), params.ContextKeyParams, ps))
	}

	h.ServeHTTP(w, req)
}
