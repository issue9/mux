// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"

	"github.com/issue9/mux/v4/group"
	"github.com/issue9/mux/v4/internal/tree"
)

// Router 提供了基本的路由项添加和删除等功能
type Router struct {
	name    string        // 当前路由的名称
	matcher group.Matcher // 当前路由的先决条件
	tree    *tree.Tree    // 当前路由的路由项
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
func (r *Router) HandleFunc(pattern string, fun http.HandlerFunc, methods ...string) error {
	return r.Handle(pattern, fun, methods...)
}

func (r *Router) handleFunc(pattern string, fun http.HandlerFunc, methods ...string) *Router {
	return r.handle(pattern, fun, methods...)
}

// GetFunc 相当于 Router.HandleFunc(pattern, func, http.MethodGet) 的简易写法
func (r *Router) GetFunc(pattern string, fun http.HandlerFunc) *Router {
	return r.handleFunc(pattern, fun, http.MethodGet)
}

// PutFunc 相当于 Router.HandleFunc(pattern, func, http.MethodPut) 的简易写法
func (r *Router) PutFunc(pattern string, fun http.HandlerFunc) *Router {
	return r.handleFunc(pattern, fun, http.MethodPut)
}

// PostFunc 相当于 Router.HandleFunc(pattern, func, "POST") 的简易写法
func (r *Router) PostFunc(pattern string, fun http.HandlerFunc) *Router {
	return r.handleFunc(pattern, fun, http.MethodPost)
}

// DeleteFunc 相当于 Router.HandleFunc(pattern, func, http.MethodDelete) 的简易写法
func (r *Router) DeleteFunc(pattern string, fun http.HandlerFunc) *Router {
	return r.handleFunc(pattern, fun, http.MethodDelete)
}

// PatchFunc 相当于 Router.HandleFunc(pattern, func, http.MethodPatch) 的简易写法
func (r *Router) PatchFunc(pattern string, fun http.HandlerFunc) *Router {
	return r.handleFunc(pattern, fun, http.MethodPatch)
}

// AnyFunc 相当于 Router.HandleFunc(pattern, func) 的简易写法
func (r *Router) AnyFunc(pattern string, fun http.HandlerFunc) *Router {
	return r.handleFunc(pattern, fun)
}

// URL 根据参数生成地址
//
// pattern 为路由项的定义内容；
// params 为路由项中的参数，键名为参数名，键值为参数值。
func (r *Router) URL(pattern string, params map[string]string) (string, error) {
	return r.tree.URL(pattern, params)
}
