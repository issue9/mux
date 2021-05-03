// SPDX-License-Identifier: MIT

package mux

import (
	"context"
	"net/http"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v4/group"
	"github.com/issue9/mux/v4/internal/handlers"
	"github.com/issue9/mux/v4/internal/tree"
	"github.com/issue9/mux/v4/params"
)

// Router 提供了强大的路由匹配功能
//
// 可以对路径按正则或是请求方法进行匹配。用法如下：
//  m := mux.Default()
//  m.Get("/abc/h1", h1).
//    Post("/abc/h2", h2).
//    Handle("/api/{version:\\d+}",h3, http.MethodGet, http.MethodPost) // 只匹配 GET 和 POST
//  http.ListenAndServe(m)
type Router struct {
	name             string        // 当前路由的名称
	routers          []*Router     // 子路由
	matcher          group.Matcher // 当前路由的先决条件
	tree             *tree.Tree    // 当前路由的路由项
	notFound         http.HandlerFunc
	methodNotAllowed http.HandlerFunc

	disableOptions, disableHead, skipCleanPath bool
}

// Default NewRouter 的默认参数版本
func Default() *Router { return NewRouter(false, false, false, nil, nil, "", nil) }

// NewRouter 声明一个新的 Router
//
// disableOptions 是否禁用自动生成 OPTIONS 功能；
// disableHead 是否禁用根据 Get 请求自动生成 HEAD 请求；
// skipCleanPath 是否不对访问路径作处理，比如 "//api" ==> "/api"；
// notFound 404 页面的处理方式，为 nil 时会调用默认的方式进行处理；
// methodNotAllowed 405 页面的处理方式，为 nil 时会调用默认的方式进行处理；
// name 为当前路由组指定名称，可以为空，该名称即在 Router.Routes() 返回标记属于哪个路由组；
// m 当前路由组的匹配规则，可以为空，表示无规则；
func NewRouter(disableOptions, disableHead, skipCleanPath bool,
	notFound, methodNotAllowed http.HandlerFunc,
	name string, m group.Matcher) *Router {
	if notFound == nil {
		notFound = defaultNotFound
	}
	if methodNotAllowed == nil {
		methodNotAllowed = defaultMethodNotAllowed
	}

	if m == nil {
		m = group.MatcherFunc(group.Any)
	}

	mux := &Router{
		name:    name,
		matcher: m,
		tree:    tree.New(disableOptions, disableHead),

		notFound:         notFound,
		methodNotAllowed: methodNotAllowed,

		disableOptions: disableOptions,
		disableHead:    disableHead,
		skipCleanPath:  skipCleanPath,
	}

	return mux
}

// Name 当前路由组的名称
func (mux *Router) Name() string { return mux.name }

// Clean 清除当前路由组的所有路由项
func (mux *Router) Clean() *Router {
	mux.tree.Clean("")
	return mux
}

// Routes 返回当前路由组的路由项
//
// ignoreHead 是否忽略自动生成的 HEAD 请求；
// ignoreOptions 是否忽略自动生成的 OPTIONS 请求；
func (mux *Router) Routes(ignoreHead, ignoreOptions bool) map[string][]string {
	return mux.tree.All(ignoreHead, ignoreOptions)
}

// Routers 返回当前路由所属的子路由组列表
func (mux *Router) Routers() []*Router { return mux.routers }

// Remove 移除指定的路由项
//
// 当未指定 methods 时，将删除所有 method 匹配的项。
// 指定错误的 methods 值，将自动忽略该值。
func (mux *Router) Remove(pattern string, methods ...string) *Router {
	mux.tree.Remove(pattern, methods...)
	return mux
}

// Handle 添加一条路由数据
//
// pattern 为路由匹配模式，可以是正则匹配也可以是字符串匹配，
// 若语法不正确，则直接 panic，可以通过 IsWell 检测语法的有效性，其它接口也相同；
// methods 该路由项对应的请求方法，如果未指定值，则表示所有支持的请求方法，
// 但不包含 OPTIONS 和 HEAD。
func (mux *Router) Handle(pattern string, h http.Handler, methods ...string) error {
	return mux.tree.Add(pattern, h, methods...)
}

// SetAllow 将 OPTIONS 请求方法的报头 allow 值固定为指定的值
//
// 若无特殊需求，不用调用此方法，系统会自动计算符合当前路由的请求方法列表。
// 如果想实现对处理方法的自定义，可以显示地调用 Handle 方法:
//  Router.Handle("/api/1", handle, http.MethodOptions)
//
// Options 与 SetAllow 功能上完全相同，只是对错误处理上有所有区别。
// Options 在出错时 panic，而 SetAllow 会返回错误信息。
func (mux *Router) SetAllow(pattern string, allow string) error {
	return mux.tree.SetAllow(pattern, allow)
}

// Options 将 OPTIONS 请求方法的报头 allow 值固定为指定的值
//
// 若无特殊需求，不用调用此方法，系统会自动计算符合当前路由的请求方法列表。
// 如果想实现对处理方法的自定义，可以显示地调用 Handle 方法:
//  Router.Handle("/api/1", handle, http.MethodOptions)
//
// Options 与 SetAllow 功能上完全相同，只是对错误处理上有所有区别。
// Options 在出错时 panic，而 SetAllow 会返回错误信息。
func (mux *Router) Options(pattern string, allow string) *Router {
	if err := mux.SetAllow(pattern, allow); err != nil {
		panic(err)
	}
	return mux
}

func (mux *Router) handle(pattern string, h http.Handler, methods ...string) *Router {
	if err := mux.Handle(pattern, h, methods...); err != nil {
		panic(err)
	}
	return mux
}

// Get 相当于 Router.Handle(pattern, h, http.MethodGet) 的简易写法
func (mux *Router) Get(pattern string, h http.Handler) *Router {
	return mux.handle(pattern, h, http.MethodGet)
}

// Post 相当于 Router.Handle(pattern, h, http.MethodPost) 的简易写法
func (mux *Router) Post(pattern string, h http.Handler) *Router {
	return mux.handle(pattern, h, http.MethodPost)
}

// Delete 相当于 Router.Handle(pattern, h, http.MethodDelete) 的简易写法
func (mux *Router) Delete(pattern string, h http.Handler) *Router {
	return mux.handle(pattern, h, http.MethodDelete)
}

// Put 相当于 Router.Handle(pattern, h, http.MethodPut) 的简易写法
func (mux *Router) Put(pattern string, h http.Handler) *Router {
	return mux.handle(pattern, h, http.MethodPut)
}

// Patch 相当于 Router.Handle(pattern, h, http.MethodPatch) 的简易写法
func (mux *Router) Patch(pattern string, h http.Handler) *Router {
	return mux.handle(pattern, h, http.MethodPatch)
}

// Any 相当于 Router.Handle(pattern, h) 的简易写法
func (mux *Router) Any(pattern string, h http.Handler) *Router {
	return mux.handle(pattern, h)
}

// HandleFunc 功能同 Router.Handle()，但是将第二个参数从 http.Handler 换成了 http.HandlerFunc
func (mux *Router) HandleFunc(pattern string, fun http.HandlerFunc, methods ...string) error {
	return mux.Handle(pattern, fun, methods...)
}

func (mux *Router) handleFunc(pattern string, fun http.HandlerFunc, methods ...string) *Router {
	return mux.handle(pattern, fun, methods...)
}

// GetFunc 相当于 Router.HandleFunc(pattern, func, http.MethodGet) 的简易写法
func (mux *Router) GetFunc(pattern string, fun http.HandlerFunc) *Router {
	return mux.handleFunc(pattern, fun, http.MethodGet)
}

// PutFunc 相当于 Router.HandleFunc(pattern, func, http.MethodPut) 的简易写法
func (mux *Router) PutFunc(pattern string, fun http.HandlerFunc) *Router {
	return mux.handleFunc(pattern, fun, http.MethodPut)
}

// PostFunc 相当于 Router.HandleFunc(pattern, func, "POST") 的简易写法
func (mux *Router) PostFunc(pattern string, fun http.HandlerFunc) *Router {
	return mux.handleFunc(pattern, fun, http.MethodPost)
}

// DeleteFunc 相当于 Router.HandleFunc(pattern, func, http.MethodDelete) 的简易写法
func (mux *Router) DeleteFunc(pattern string, fun http.HandlerFunc) *Router {
	return mux.handleFunc(pattern, fun, http.MethodDelete)
}

// PatchFunc 相当于 Router.HandleFunc(pattern, func, http.MethodPatch) 的简易写法
func (mux *Router) PatchFunc(pattern string, fun http.HandlerFunc) *Router {
	return mux.handleFunc(pattern, fun, http.MethodPatch)
}

// AnyFunc 相当于 Router.HandleFunc(pattern, func) 的简易写法
func (mux *Router) AnyFunc(pattern string, fun http.HandlerFunc) *Router {
	return mux.handleFunc(pattern, fun)
}

func (mux *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !mux.skipCleanPath {
		r.URL.Path = cleanPath(r.URL.Path)
	}

	hs, ps := mux.match(r)
	if hs == nil {
		mux.notFound(w, r)
		return
	}

	h := hs.Handler(r.Method)
	if h == nil {
		w.Header().Set("Allow", hs.Options())
		mux.methodNotAllowed(w, r)
		return
	}

	if len(ps) > 0 {
		r = r.WithContext(context.WithValue(r.Context(), params.ContextKeyParams, ps))
	}

	h.ServeHTTP(w, r)
}

// NewRouter 添加子路由组
//
// 该路由只有符合 group.Matcher 的要求才会进入，其它与 Router 功能相同。
//
// name 表示该路由组的名称，需要唯一，否则返回 false；
func (mux *Router) NewRouter(name string, matcher group.Matcher) (*Router, bool) {
	if mux.routers == nil {
		mux.routers = make([]*Router, 0, 5)
	}

	dup := sliceutil.Count(mux.routers, func(i int) bool {
		return mux.routers[i].name == name
	}) > 0
	if mux.Name() == name || dup {
		return nil, false
	}

	m := NewRouter(mux.disableOptions, mux.disableHead, mux.skipCleanPath, mux.notFound, mux.methodNotAllowed, name, matcher)
	mux.routers = append(mux.routers, m)
	return m, true
}

func (mux *Router) match(r *http.Request) (*handlers.Handlers, params.Params) {
	for _, m := range mux.routers {
		if hs, ps := m.match(r); hs != nil {
			return hs, ps
		}
	}

	if mux.matcher.Match(r) {
		return mux.tree.Handler(r.URL.Path)
	}
	return nil, nil
}

// URL 根据参数生成地址
//
// pattern 为路由项的定义内容；
// params 为路由项中的参数，键名为参数名，键值为参数值。
func (mux *Router) URL(pattern string, params map[string]string) (string, error) {
	return mux.tree.URL(pattern, params)
}
