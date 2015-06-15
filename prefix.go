// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
)

// 封装ServeMux，使所有添加的路由项的匹配模式都带上指定的前缀。
//  p := srv.Prefix("/api")
//  p.Get("/users")  // 相当于 srv.Get("/api/users")
//  p.Get("/user/1") // 相当于 srv.Get("/api/user/1")
//
// 可以直接使用ServeMux.Prefix()方法获取一个Prefix实例。
type Prefix struct {
	mux    *ServeMux
	group  *Group
	prefix string
}

// Add相当于ServeMux.Add(prefix+pattern, h, "POST"...)的简易写法
func (p *Prefix) Add(pattern string, h http.Handler, methods ...string) *Prefix {
	p.mux.add(p.group, p.prefix+pattern, h, methods...)
	return p
}

// Get相当于ServeMux.Get(prefix+pattern, h)的简易写法
func (p *Prefix) Get(pattern string, h http.Handler) *Prefix {
	p.mux.add(p.group, p.prefix+pattern, h, "GET")
	return p
}

// Post相当于ServeMux.Post(prefix+pattern, h)的简易写法
func (p *Prefix) Post(pattern string, h http.Handler) *Prefix {
	p.mux.add(p.group, p.prefix+pattern, h, "POST")
	return p
}

// Delete相当于ServeMux.Delete(prefix+pattern, h)的简易写法
func (p *Prefix) Delete(pattern string, h http.Handler) *Prefix {
	p.mux.add(p.group, p.prefix+pattern, h, "DELETE")
	return p
}

// Put相当于ServeMux.Put(prefix+pattern, h)的简易写法
func (p *Prefix) Put(pattern string, h http.Handler) *Prefix {
	p.mux.add(p.group, p.prefix+pattern, h, "PUT")
	return p
}

// Any相当于ServeMux.Any(prefix+pattern, h)的简易写法
func (p *Prefix) Any(pattern string, h http.Handler) *Prefix {
	p.mux.add(p.group, p.prefix+pattern, h)
	return p
}

// AddFunc功能同ServeMux.AddFunc(prefix+pattern, fun, ...)
func (p *Prefix) AddFunc(pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) *Prefix {
	p.mux.addFunc(p.group, p.prefix+pattern, fun, methods...)
	return p
}

// GetFunc相当于ServeMux.GetFunc(prefix+pattern, func)的简易写法
func (p *Prefix) GetFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Prefix {
	p.mux.addFunc(p.group, p.prefix+pattern, fun, "GET")
	return p
}

// PutFunc相当于ServeMux.PutFunc(prefix+pattern, func)的简易写法
func (p *Prefix) PutFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Prefix {
	p.mux.addFunc(p.group, p.prefix+pattern, fun, "PUT")
	return p
}

// PostFunc相当于ServeMux.PostFunc(prefix+pattern, func)的简易写法
func (p *Prefix) PostFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Prefix {
	p.mux.addFunc(p.group, p.prefix+pattern, fun, "POST")
	return p
}

// DeleteFunc相当于ServeMux.DeleteFunc(prefix+pattern, func)的简易写法
func (p *Prefix) DeleteFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Prefix {
	p.mux.addFunc(p.group, p.prefix+pattern, fun, "DELETE")
	return p
}

// AnyFunc相当于ServeMux.AnyFunc(prefix+pattern, func)的简易写法
func (p *Prefix) AnyFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Prefix {
	p.mux.addFunc(p.group, p.prefix+pattern, fun)
	return p
}

// AnyFunc相当于ServeMux.Remove(prefix+pattern, methods...)的简易写法
func (p *Prefix) Remove(pattern string, methods ...string) {
	p.mux.Remove(p.prefix+pattern, methods...)
}
