// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"

	"github.com/issue9/mux/v5/internal/syntax"
	"github.com/issue9/mux/v5/internal/tree"
	"github.com/issue9/mux/v5/params"
)

// Router 提供了基本的路由项添加和删除等功能
//
// 可以对路径按正则或是请求方法进行匹配。用法如下：
//  router := DefaultRouter()
//  router.Get("/abc/h1", h1).
//      Post("/abc/h2", h2).
//      Handle("/api/{version:\\d+}",h3, http.MethodGet, http.MethodPost) // 只匹配 GET 和 POST
//  http.ListenAndServe(m)
type Router struct {
	skipCleanPath bool
	tree          *tree.Tree
	ms            *Middlewares
	cors          *CORS

	notFound,
	methodNotAllowed http.HandlerFunc
}

// DefaultRouter 返回默认参数的 NewRouter
//
// 相当于调用 NewRouter(false, false, DeniedCORS(), nil, nil)
func DefaultRouter() *Router {
	r, err := NewRouter(false, false, DeniedCORS(), nil, nil)
	if err != nil {
		panic(err)
	}
	return r
}

// NewRouter 添加子路由
//
// disableHead 是否禁用根据 Get 请求自动生成 HEAD 请求；
// skipCleanPath 是否不对访问路径作处理，比如 "//api" ==> "/api"；
// cors 跨域请求的相关设置项；
// notFound 404 页面的处理方式，为 nil 时会调用默认的方式进行处理；
// methodNotAllowed 405 页面的处理方式，为 nil 时会调用默认的方式进行处理；
func NewRouter(disableHead, skipCleanPath bool, cors *CORS, notFound, methodNotAllowed http.HandlerFunc) (*Router, error) {
	if cors == nil {
		cors = DeniedCORS()
	}
	if notFound == nil {
		notFound = tree.DefaultNotFound
	}
	if methodNotAllowed == nil {
		methodNotAllowed = tree.DefaultMethodNotAllowed
	}

	if err := cors.sanitize(); err != nil {
		return nil, err
	}

	r := &Router{
		skipCleanPath:    skipCleanPath,
		tree:             tree.New(disableHead),
		cors:             cors,
		notFound:         notFound,
		methodNotAllowed: methodNotAllowed,
	}
	r.ms = NewMiddlewares(http.HandlerFunc(r.serveHTTP))
	return r, nil
}

// Clean 清除当前路由组的所有路由项
func (r *Router) Clean() error {
	r.tree.Clean("")
	return nil
}

// Routes 返回当前路由组的路由项
func (r *Router) Routes() map[string][]string { return r.tree.Routes() }

// Remove 移除指定的路由项
//
// 当未指定 methods 时，将删除所有 method 匹配的项。
// 指定错误的 methods 值，将自动忽略该值。
func (r *Router) Remove(pattern string, methods ...string) error {
	return r.tree.Remove(pattern, methods...)
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
	return syntax.URL(pattern, params)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.ms.ServeHTTP(w, req)
}

func (r *Router) serveHTTP(w http.ResponseWriter, req *http.Request) {
	if !r.skipCleanPath {
		req.URL.Path = cleanPath(req.URL.Path)
	}

	hs, ps := r.tree.Route(req.URL.Path)
	if ps == nil {
		r.notFound(w, req)
		return
	}

	r.cors.handle(hs, w, req) // 处理跨域问题

	h := hs.Handler(req.Method)
	if h == nil {
		w.Header().Set("Allow", hs.Options())
		r.methodNotAllowed(w, req)
		return
	}

	if len(ps) > 0 {
		req = params.WithValue(req, ps)
	}

	h.ServeHTTP(w, req)
}
