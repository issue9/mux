// SPDX-License-Identifier: MIT

package mux

import "net/http"

// Prefix 可以将具有统一前缀的路由项集中在一起操作
//
// example:
//  p := srv.Prefix("/api")
//  p.Get("/users")  // 相当于 srv.Get("/api/users")
//  p.Get("/user/1") // 相当于 srv.Get("/api/user/1")
type Prefix struct {
	mux    *Mux
	prefix string
}

// SetAllow 手动指定 OPTIONS 请求方法的值
func (p *Prefix) SetAllow(pattern string, allow string) error {
	return p.mux.SetAllow(p.prefix+pattern, allow)
}

// Options 手动指定 OPTIONS 请求方法的值
func (p *Prefix) Options(pattern string, allow string) *Prefix {
	if err := p.SetAllow(pattern, allow); err != nil {
		panic(err)
	}
	return p
}

// Handle 相当于 Router.Handle(prefix+pattern, h, methods...) 的简易写法
func (p *Prefix) Handle(pattern string, h http.Handler, methods ...string) error {
	return p.mux.Handle(p.prefix+pattern, h, methods...)
}

func (p *Prefix) handle(pattern string, h http.Handler, methods ...string) *Prefix {
	if err := p.Handle(pattern, h, methods...); err != nil {
		panic(err)
	}

	return p
}

// Get 相当于 Router.Get(prefix+pattern, h) 的简易写法
func (p *Prefix) Get(pattern string, h http.Handler) *Prefix {
	return p.handle(pattern, h, http.MethodGet)
}

// Post 相当于 Router.Post(prefix+pattern, h) 的简易写法
func (p *Prefix) Post(pattern string, h http.Handler) *Prefix {
	return p.handle(pattern, h, http.MethodPost)
}

// Delete 相当于 Router.Delete(prefix+pattern, h)的简易写法
func (p *Prefix) Delete(pattern string, h http.Handler) *Prefix {
	return p.handle(pattern, h, http.MethodDelete)
}

// Put 相当于 Router.Put(prefix+pattern, h) 的简易写法
func (p *Prefix) Put(pattern string, h http.Handler) *Prefix {
	return p.handle(pattern, h, http.MethodPut)
}

// Patch 相当于 Router.Patch(prefix+pattern, h) 的简易写法
func (p *Prefix) Patch(pattern string, h http.Handler) *Prefix {
	return p.handle(pattern, h, http.MethodPatch)
}

// Any 相当于 Router.Any(prefix+pattern, h) 的简易写法
func (p *Prefix) Any(pattern string, h http.Handler) *Prefix {
	return p.handle(pattern, h)
}

// HandleFunc 功能同 Router.HandleFunc(prefix+pattern, fun, ...)
func (p *Prefix) HandleFunc(pattern string, fun http.HandlerFunc, methods ...string) error {
	return p.Handle(pattern, fun, methods...)
}

func (p *Prefix) handleFunc(pattern string, fun http.HandlerFunc, methods ...string) *Prefix {
	if err := p.HandleFunc(pattern, fun, methods...); err != nil {
		panic(err)
	}
	return p
}

// GetFunc 相当于 Router.GetFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) GetFunc(pattern string, fun http.HandlerFunc) *Prefix {
	return p.handleFunc(pattern, fun, http.MethodGet)
}

// PutFunc 相当于 Router.PutFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) PutFunc(pattern string, fun http.HandlerFunc) *Prefix {
	return p.handleFunc(pattern, fun, http.MethodPut)
}

// PostFunc 相当 于Mux.PostFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) PostFunc(pattern string, fun http.HandlerFunc) *Prefix {
	return p.handleFunc(pattern, fun, http.MethodPost)
}

// DeleteFunc 相当于 Router.DeleteFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) DeleteFunc(pattern string, fun http.HandlerFunc) *Prefix {
	return p.handleFunc(pattern, fun, http.MethodDelete)
}

// PatchFunc 相当于 Router.PatchFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) PatchFunc(pattern string, fun http.HandlerFunc) *Prefix {
	return p.handleFunc(pattern, fun, http.MethodPatch)
}

// AnyFunc 相当于 Router.AnyFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) AnyFunc(pattern string, fun http.HandlerFunc) *Prefix {
	return p.handleFunc(pattern, fun)
}

// Remove 删除指定匹配模式的路由项
func (p *Prefix) Remove(pattern string, methods ...string) *Prefix {
	p.mux.Remove(p.prefix+pattern, methods...)
	return p
}

// Clean 清除所有以 Prefix.prefix 开头的路由项
//
// 当指定多个相同的 Prefix 时，调用其中的一个 Clean 也将会清除其它的：
//  p1 := mux.Prefix("prefix")
//  p2 := mux.Prefix("prefix")
//  p2.Clean() 将同时清除 p1 的内容，因为有相同的前缀。
func (p *Prefix) Clean() *Prefix {
	p.mux.tree.Clean(p.prefix)
	return p
}

// URL 根据参数生成地址
//
// name 为路由的名称，或是直接为路由项的定义内容，
// 若 name 作为路由项定义，会加上 Prefix.prefix 作为前缀；
// params 为路由项中的参数，键名为参数名，键值为参数值。
func (p *Prefix) URL(pattern string, params map[string]string) (string, error) {
	return p.mux.tree.URL(p.prefix+pattern, params)
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
		mux:    p.mux,
		prefix: p.prefix + prefix,
	}
}

// Prefix 声明一个 Prefix 实例
func (mux *Mux) Prefix(prefix string) *Prefix {
	return &Prefix{
		mux:    mux,
		prefix: prefix,
	}
}

// Mux 返回与当前关联的 *Mux 实例
func (p *Prefix) Mux() *Mux { return p.mux }
