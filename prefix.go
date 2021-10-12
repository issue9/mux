// SPDX-License-Identifier: MIT

package mux

import "net/http"

// Prefix 操纵统一前缀的路由
//
// example:
// r := DefaultRouter()
//  p := r.Prefix("/api")
//  p.Get("/users")  // 相当于 r.Get("/api/users")
//  p.Get("/user/1") // 相当于 r.Get("/api/user/1")
type Prefix struct {
	router *Router
	prefix string
}

// Handle 相当于 Router.Handle(prefix+pattern, h, methods...) 的简易写法
func (p *Prefix) Handle(pattern string, h http.Handler, methods ...string) *Prefix {
	p.router.Handle(p.prefix+pattern, h, methods...)
	return p
}

// Get 相当于 Router.Get(prefix+pattern, h) 的简易写法
func (p *Prefix) Get(pattern string, h http.Handler) *Prefix {
	return p.Handle(pattern, h, http.MethodGet)
}

// Post 相当于 Router.Post(prefix+pattern, h) 的简易写法
func (p *Prefix) Post(pattern string, h http.Handler) *Prefix {
	return p.Handle(pattern, h, http.MethodPost)
}

// Delete 相当于 Router.Delete(prefix+pattern, h)的简易写法
func (p *Prefix) Delete(pattern string, h http.Handler) *Prefix {
	return p.Handle(pattern, h, http.MethodDelete)
}

// Put 相当于 Router.Put(prefix+pattern, h) 的简易写法
func (p *Prefix) Put(pattern string, h http.Handler) *Prefix {
	return p.Handle(pattern, h, http.MethodPut)
}

// Patch 相当于 Router.Patch(prefix+pattern, h) 的简易写法
func (p *Prefix) Patch(pattern string, h http.Handler) *Prefix {
	return p.Handle(pattern, h, http.MethodPatch)
}

// Any 相当于 Router.Any(prefix+pattern, h) 的简易写法
func (p *Prefix) Any(pattern string, h http.Handler) *Prefix {
	return p.Handle(pattern, h)
}

// HandleFunc 功能同 Router.HandleFunc(prefix+pattern, fun, ...)
func (p *Prefix) HandleFunc(pattern string, f http.HandlerFunc, methods ...string) *Prefix {
	return p.Handle(pattern, f, methods...)
}

// GetFunc 相当于 Router.GetFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) GetFunc(pattern string, f http.HandlerFunc) *Prefix {
	return p.HandleFunc(pattern, f, http.MethodGet)
}

// PutFunc 相当于 Router.PutFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) PutFunc(pattern string, f http.HandlerFunc) *Prefix {
	return p.HandleFunc(pattern, f, http.MethodPut)
}

// PostFunc 相当 于Mux.PostFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) PostFunc(pattern string, f http.HandlerFunc) *Prefix {
	return p.HandleFunc(pattern, f, http.MethodPost)
}

// DeleteFunc 相当于 Router.DeleteFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) DeleteFunc(pattern string, f http.HandlerFunc) *Prefix {
	return p.HandleFunc(pattern, f, http.MethodDelete)
}

// PatchFunc 相当于 Router.PatchFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) PatchFunc(pattern string, f http.HandlerFunc) *Prefix {
	return p.HandleFunc(pattern, f, http.MethodPatch)
}

// AnyFunc 相当于 Router.AnyFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) AnyFunc(pattern string, f http.HandlerFunc) *Prefix {
	return p.HandleFunc(pattern, f)
}

// Remove 删除指定匹配模式的路由项
func (p *Prefix) Remove(pattern string, methods ...string) {
	p.router.Remove(p.prefix+pattern, methods...)
}

// Clean 清除所有以 Prefix.prefix 开头的路由项
//
// 当指定多个相同的 Prefix 时，调用其中的一个 Clean 也将会清除其它的：
//  p1 := mux.Prefix("prefix")
//  p2 := mux.Prefix("prefix")
//  p2.Clean() 将同时清除 p1 的内容，因为有相同的前缀。
func (p *Prefix) Clean() { p.router.tree.Clean(p.prefix) }

// URL 根据参数生成地址
//
// name 为路由的名称，或是直接为路由项的定义内容，
// 若 name 作为路由项定义，会加上 Prefix.prefix 作为前缀；
// params 为路由项中的参数，键名为参数名，键值为参数值。
func (p *Prefix) URL(pattern string, params map[string]string) (string, error) {
	return p.router.URL(p.prefix+pattern, params)
}

// Prefix 在现有 Prefix 的基础上声明一个新的 Prefix 实例
//
// example:
//  p := mux.Prefix("/api")
//  v := p.Prefix("/v2")
//  v.Get("/users")   // 相当于 g.Get("/api/v2/users")
//  v.Get("/users/1") // 相当于 g.Get("/api/v2/users/1")
//  v.Get("example.com/users/1") // 相当于 g.Get("/api/v2/example.com/users/1")
func (p *Prefix) Prefix(prefix string) *Prefix {
	return &Prefix{
		router: p.router,
		prefix: p.prefix + prefix,
	}
}

// Prefix 声明一个 Prefix 实例
func (r *Router) Prefix(prefix string) *Prefix {
	return &Prefix{
		router: r,
		prefix: prefix,
	}
}

// Router 返回与当前关联的 *Router 实例
func (p *Prefix) Router() *Router { return p.router }
