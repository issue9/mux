// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import "net/http"

// Resource 以资源地址为对象的路由配置。
//  r := srv.Resource("/api/users/{id}")
//  r.Get(h)  // 相当于 srv.Get("/api/users/{id}")
//  r.Post(h) // 相当于 srv.Post("/api/users/{id}")
type Resource struct {
	mux     *ServeMux
	pattern string
}

// Options 手动指定 OPTIONS 请求方法的值。
//
// 若无特殊需求，不用调用些方法，系统会自动计算符合当前路由的请求方法列表。
func (r *Resource) Options(methods string) *Resource {
	r.mux.Options(r.pattern, methods)
	return r
}

// Add 相当于 ServeMux.Add(pattern, h, "POST"...) 的简易写法
func (r *Resource) Add(h http.Handler, methods ...string) *Resource {
	r.mux.Add(r.pattern, h, methods...)
	return r
}

// Get 相当于 ServeMux.Get(pattern, h) 的简易写法
func (r *Resource) Get(h http.Handler) *Resource {
	return r.Add(h, http.MethodGet)
}

// Post 相当于 ServeMux.Post(pattern, h) 的简易写法
func (r *Resource) Post(h http.Handler) *Resource {
	return r.Add(h, http.MethodPost)
}

// Delete 相当于 ServeMux.Delete(pattern, h) 的简易写法
func (r *Resource) Delete(h http.Handler) *Resource {
	return r.Add(h, http.MethodDelete)
}

// Put 相当于 ServeMux.Put(pattern, h) 的简易写法
func (r *Resource) Put(h http.Handler) *Resource {
	return r.Add(h, http.MethodPut)
}

// Patch 相当于 ServeMux.Patch(pattern, h) 的简易写法
func (r *Resource) Patch(h http.Handler) *Resource {
	return r.Add(h, http.MethodPatch)
}

// Any 相当于 ServeMux.Any(pattern, h) 的简易写法
func (r *Resource) Any(h http.Handler) *Resource {
	return r.Add(h, defaultMethods...)
}

// AddFunc 功能同 ServeMux.AddFunc(pattern, fun, ...)
func (r *Resource) AddFunc(fun func(http.ResponseWriter, *http.Request), methods ...string) *Resource {
	r.mux.AddFunc(r.pattern, fun, methods...)
	return r
}

// GetFunc 相当于 ServeMux.GetFunc(pattern, func) 的简易写法
func (r *Resource) GetFunc(fun func(http.ResponseWriter, *http.Request)) *Resource {
	return r.AddFunc(fun, http.MethodGet)
}

// PutFunc 相当于 ServeMux.PutFunc(pattern, func) 的简易写法
func (r *Resource) PutFunc(fun func(http.ResponseWriter, *http.Request)) *Resource {
	return r.AddFunc(fun, http.MethodPut)
}

// PostFunc 相当于 ServeMux.PostFunc(pattern, func) 的简易写法
func (r *Resource) PostFunc(fun func(http.ResponseWriter, *http.Request)) *Resource {
	return r.AddFunc(fun, http.MethodPost)
}

// DeleteFunc 相当于 ServeMux.DeleteFunc(pattern, func) 的简易写法
func (r *Resource) DeleteFunc(fun func(http.ResponseWriter, *http.Request)) *Resource {
	return r.AddFunc(fun, http.MethodDelete)
}

// PatchFunc 相当于 ServeMux.PatchFunc(pattern, func) 的简易写法
func (r *Resource) PatchFunc(fun func(http.ResponseWriter, *http.Request)) *Resource {
	return r.AddFunc(fun, http.MethodPatch)
}

// AnyFunc 相当于 ServeMux.AnyFunc(pattern, func) 的简易写法
func (r *Resource) AnyFunc(fun func(http.ResponseWriter, *http.Request)) *Resource {
	return r.AddFunc(fun, defaultMethods...)
}

// Remove 删除指定匹配模式的路由项
func (r *Resource) Remove(methods ...string) *Resource {
	r.mux.Remove(r.pattern, methods...)
	return r
}

// Clean 清除当前资源的所有路由项
func (r *Resource) Clean() *Resource {
	r.mux.Remove(r.pattern, supportedMethods...)
	return r
}

// Resource 创建一个路由组，该组中添加的路由项，其地址均为 pattern。
// pattern 前缀字符串，所有从 Resource 中声明的路由都将包含此前缀。
//  p := srv.Resource("/api")
//  p.Get("/users")  // 相当于 srv.Get("/api/users")
//  p.Get("/user/1") // 相当于 srv.Get("/api/user/1")
func (mux *ServeMux) Resource(pattern string) *Resource {
	return &Resource{
		mux:     mux,
		pattern: pattern,
	}
}

// Resource 创建一个路由组，该组中添加的路由项，其地址均为 pattern。
// pattern 前缀字符串，所有从 Resource 中声明的路由都将包含此前缀。
//  p := srv.Resource("/api")
//  p.Get("/users")  // 相当于 srv.Get("/api/users")
//  p.Get("/user/1") // 相当于 srv.Get("/api/user/1")
func (p *Prefix) Resource(pattern string) *Resource {
	return &Resource{
		mux:     p.mux,
		pattern: p.prefix + pattern,
	}
}
