// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"

	"github.com/issue9/mux/internal/method"
)

// Prefix 可以将具有统一前缀的路由项集中在一起操作。
//  p := srv.Prefix("/api")
//  p.Get("/users")  // 相当于 srv.Get("/api/users")
//  p.Get("/user/1") // 相当于 srv.Get("/api/user/1")
type Prefix struct {
	mux    *Mux
	prefix string
}

// Options 手动指定 OPTIONS 请求方法的值。具体说明可参考 Mux.Options 方法。
func (p *Prefix) Options(pattern string, allow string) *Prefix {
	p.mux.Options(p.prefix+pattern, allow)
	return p
}

// Handle 相当于 Mux.Handle(prefix+pattern, h, methods...) 的简易写法
func (p *Prefix) Handle(pattern string, h http.Handler, methods ...string) error {
	return p.mux.Handle(p.prefix+pattern, h, methods...)
}

func (p *Prefix) handle(pattern string, h http.Handler, methods ...string) *Prefix {
	if err := p.mux.Handle(p.prefix+pattern, h, methods...); err != nil {
		panic(err)
	}

	return p
}

// Get 相当于 Mux.Get(prefix+pattern, h) 的简易写法
func (p *Prefix) Get(pattern string, h http.Handler) *Prefix {
	return p.handle(pattern, h, http.MethodGet)
}

// Post 相当于 Mux.Post(prefix+pattern, h) 的简易写法
func (p *Prefix) Post(pattern string, h http.Handler) *Prefix {
	return p.handle(pattern, h, http.MethodPost)
}

// Delete 相当于 Mux.Delete(prefix+pattern, h)的简易写法
func (p *Prefix) Delete(pattern string, h http.Handler) *Prefix {
	return p.handle(pattern, h, http.MethodDelete)
}

// Put 相当于 Mux.Put(prefix+pattern, h) 的简易写法
func (p *Prefix) Put(pattern string, h http.Handler) *Prefix {
	return p.handle(pattern, h, http.MethodPut)
}

// Patch 相当于 Mux.Patch(prefix+pattern, h) 的简易写法
func (p *Prefix) Patch(pattern string, h http.Handler) *Prefix {
	return p.handle(pattern, h, http.MethodPatch)
}

// Any 相当于 Mux.Any(prefix+pattern, h) 的简易写法
func (p *Prefix) Any(pattern string, h http.Handler) *Prefix {
	return p.handle(pattern, h, method.Default...)
}

// HandleFunc 功能同 Mux.HandleFunc(prefix+pattern, fun, ...)
func (p *Prefix) HandleFunc(pattern string, fun http.HandlerFunc, methods ...string) error {
	return p.mux.HandleFunc(p.prefix+pattern, fun, methods...)
}

func (p *Prefix) handleFunc(pattern string, fun http.HandlerFunc, methods ...string) *Prefix {
	if err := p.mux.HandleFunc(p.prefix+pattern, fun, methods...); err != nil {
		panic(err)
	}
	return p
}

// GetFunc 相当于 Mux.GetFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) GetFunc(pattern string, fun http.HandlerFunc) *Prefix {
	return p.handleFunc(pattern, fun, http.MethodGet)
}

// PutFunc 相当于 Mux.PutFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) PutFunc(pattern string, fun http.HandlerFunc) *Prefix {
	return p.handleFunc(pattern, fun, http.MethodPut)
}

// PostFunc 相当 于Mux.PostFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) PostFunc(pattern string, fun http.HandlerFunc) *Prefix {
	return p.handleFunc(pattern, fun, http.MethodPost)
}

// DeleteFunc 相当于 Mux.DeleteFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) DeleteFunc(pattern string, fun http.HandlerFunc) *Prefix {
	return p.handleFunc(pattern, fun, http.MethodDelete)
}

// PatchFunc 相当于 Mux.PatchFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) PatchFunc(pattern string, fun http.HandlerFunc) *Prefix {
	return p.handleFunc(pattern, fun, http.MethodPatch)
}

// AnyFunc 相当于 Mux.AnyFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) AnyFunc(pattern string, fun http.HandlerFunc) *Prefix {
	return p.handleFunc(pattern, fun, method.Default...)
}

// Remove 删除指定匹配模式的路由项
func (p *Prefix) Remove(pattern string, methods ...string) *Prefix {
	p.mux.Remove(p.prefix+pattern, methods...)
	return p
}

// Clean 清除所有以 Prefix.prefix 开头的 Entry。
//
// 当指定多个相同的 Prefix 时，调用其中的一个 Clean 也将会清除其它的：
//  p1 := mux.Prefix("prefix")
//  p2 := mux.Prefix("prefix")
//  p2.Clean() 将同时清除 p1 的内容，因为有相同的前缀。
func (p *Prefix) Clean() *Prefix {
	p.mux.tree.Clean(p.prefix)
	return p
}

// Prefix 在现在有 Prefix 的基础上声明一个新的 Prefix 实例。
//  p := mux.Prefix("/api")
//  v := p.Prefix("/v2")
//  v.Get("/users")  // 相当于 g.Get("/api/v2/users")
//  v.Get("/user/1") // 相当于 g.Get("/api/v2/user/1")
func (p *Prefix) Prefix(prefix string) *Prefix {
	return &Prefix{
		mux:    p.mux,
		prefix: p.prefix + prefix,
	}
}

// Prefix 声明一个 Prefix 实例。
func (mux *Mux) Prefix(prefix string) *Prefix {
	return &Prefix{
		mux:    mux,
		prefix: prefix,
	}
}

// Mux 返回与当前关联的 *Mux 实例
func (p *Prefix) Mux() *Mux {
	return p.mux
}
