// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"

	"github.com/issue9/mux/internal/method"
)

// Resource 以资源地址为对象的路由配置。
//  r := srv.Resource("/api/users/{id}")
//  r.Get(h)  // 相当于 srv.Get("/api/users/{id}")
//  r.Post(h) // 相当于 srv.Post("/api/users/{id}")
type Resource struct {
	mux     *Mux
	pattern string
}

// Options 手动指定 OPTIONS 请求方法的值。
//
// 若无特殊需求，不用调用些方法，系统会自动计算符合当前路由的请求方法列表。
func (r *Resource) Options(allow string) *Resource {
	r.mux.Options(r.pattern, allow)
	return r
}

// Add 相当于 Mux.Add(pattern, h, "POST"...) 的简易写法
func (r *Resource) Add(h http.Handler, methods ...string) error {
	return r.mux.Add(r.pattern, h, methods...)
}

func (r *Resource) add(h http.Handler, methods ...string) *Resource {
	if err := r.mux.Add(r.pattern, h, methods...); err != nil {
		panic(err)
	}
	return r
}

// Get 相当于 Mux.Get(pattern, h) 的简易写法
func (r *Resource) Get(h http.Handler) *Resource {
	return r.add(h, http.MethodGet)
}

// Post 相当于 Mux.Post(pattern, h) 的简易写法
func (r *Resource) Post(h http.Handler) *Resource {
	return r.add(h, http.MethodPost)
}

// Delete 相当于 Mux.Delete(pattern, h) 的简易写法
func (r *Resource) Delete(h http.Handler) *Resource {
	return r.add(h, http.MethodDelete)
}

// Put 相当于 Mux.Put(pattern, h) 的简易写法
func (r *Resource) Put(h http.Handler) *Resource {
	return r.add(h, http.MethodPut)
}

// Patch 相当于 Mux.Patch(pattern, h) 的简易写法
func (r *Resource) Patch(h http.Handler) *Resource {
	return r.add(h, http.MethodPatch)
}

// Any 相当于 Mux.Any(pattern, h) 的简易写法
func (r *Resource) Any(h http.Handler) *Resource {
	return r.add(h, method.Default...)
}

// AddFunc 功能同 Mux.AddFunc(pattern, fun, ...)
func (r *Resource) AddFunc(fun func(http.ResponseWriter, *http.Request), methods ...string) error {
	return r.mux.AddFunc(r.pattern, fun, methods...)
}

func (r *Resource) addFunc(fun func(http.ResponseWriter, *http.Request), methods ...string) *Resource {
	if err := r.mux.AddFunc(r.pattern, fun, methods...); err != nil {
		panic(err)
	}

	return r
}

// GetFunc 相当于 Mux.GetFunc(pattern, func) 的简易写法
func (r *Resource) GetFunc(fun func(http.ResponseWriter, *http.Request)) *Resource {
	return r.addFunc(fun, http.MethodGet)
}

// PutFunc 相当于 Mux.PutFunc(pattern, func) 的简易写法
func (r *Resource) PutFunc(fun func(http.ResponseWriter, *http.Request)) *Resource {
	return r.addFunc(fun, http.MethodPut)
}

// PostFunc 相当于 Mux.PostFunc(pattern, func) 的简易写法
func (r *Resource) PostFunc(fun func(http.ResponseWriter, *http.Request)) *Resource {
	return r.addFunc(fun, http.MethodPost)
}

// DeleteFunc 相当于 Mux.DeleteFunc(pattern, func) 的简易写法
func (r *Resource) DeleteFunc(fun func(http.ResponseWriter, *http.Request)) *Resource {
	return r.addFunc(fun, http.MethodDelete)
}

// PatchFunc 相当于 Mux.PatchFunc(pattern, func) 的简易写法
func (r *Resource) PatchFunc(fun func(http.ResponseWriter, *http.Request)) *Resource {
	return r.addFunc(fun, http.MethodPatch)
}

// AnyFunc 相当于 Mux.AnyFunc(pattern, func) 的简易写法
func (r *Resource) AnyFunc(fun func(http.ResponseWriter, *http.Request)) *Resource {
	return r.addFunc(fun, method.Default...)
}

// Remove 删除指定匹配模式的路由项
func (r *Resource) Remove(methods ...string) *Resource {
	r.mux.Remove(r.pattern, methods...)
	return r
}

// Clean 清除当前资源的所有路由项
func (r *Resource) Clean() *Resource {
	r.mux.Remove(r.pattern, method.Supported...)
	return r
}

// Resource 创建一个资源路由项，之后可以为该资源指定各种请求方法。
//  p := srv.Resource("/api")
//  p.Get("/users")  // 相当于 srv.Get("/api/users")
//  p.Get("/user/1") // 相当于 srv.Get("/api/user/1")
func (mux *Mux) Resource(pattern string) *Resource {
	return &Resource{
		mux:     mux,
		pattern: pattern,
	}
}

// Resource 创建一个资源路由项，之后可以为该资源指定各种请求方法。
//  p := srv.Resource("/api")
//  p.Get("/users")  // 相当于 srv.Get("/api/users")
//  p.Get("/user/1") // 相当于 srv.Get("/api/user/1")
func (p *Prefix) Resource(pattern string) *Resource {
	return &Resource{
		mux:     p.mux,
		pattern: p.prefix + pattern,
	}
}

// Mux 返回与当前资源关联的 *Mux 实例
func (r *Resource) Mux() *Mux {
	return r.mux
}
