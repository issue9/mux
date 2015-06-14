// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
)

type Group struct {
	mux    *ServeMux
	prefix string
}

// Add相当于ServeMux.Add(prefix+pattern, h, "POST"...)的简易写法
func (g *Group) Add(pattern string, h http.Handler, methods ...string) *Group {
	g.mux.Add(g.prefix+pattern, h, methods...)
	return g
}

// Get相当于ServeMux.Get(prefix+pattern, h)的简易写法
func (g *Group) Get(pattern string, h http.Handler) *Group {
	g.mux.Get(g.prefix+pattern, h)
	return g
}

// Post相当于ServeMux.Post(prefix+pattern, h)的简易写法
func (g *Group) Post(pattern string, h http.Handler) *Group {
	g.mux.Post(g.prefix+pattern, h)
	return g
}

// Delete相当于ServeMux.Delete(prefix+pattern, h)的简易写法
func (g *Group) Delete(pattern string, h http.Handler) *Group {
	g.mux.Delete(g.prefix+pattern, h)
	return g
}

// Put相当于ServeMux.Put(prefix+pattern, h)的简易写法
func (g *Group) Put(pattern string, h http.Handler) *Group {
	g.mux.Put(g.prefix+pattern, h)
	return g
}

// Any相当于ServeMux.Any(prefix+pattern, h)的简易写法
func (g *Group) Any(pattern string, h http.Handler) *Group {
	g.mux.Any(g.prefix+pattern, h)
	return g
}

// AddFunc功能同ServeMux.AddFunc(prefix+pattern, fun, ...)
func (g *Group) AddFunc(pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) *Group {
	g.mux.AddFunc(g.prefix+pattern, fun, methods...)
	return g
}

// GetFunc相当于ServeMux.GetFunc(prefix+pattern, func)的简易写法
func (g *Group) GetFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Group {
	g.mux.GetFunc(g.prefix+pattern, fun)
	return g
}

// PutFunc相当于ServeMux.PutFunc(prefix+pattern, func)的简易写法
func (g *Group) PutFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Group {
	g.mux.PutFunc(g.prefix+pattern, fun)
	return g
}

// PostFunc相当于ServeMux.PostFunc(prefix+pattern, func)的简易写法
func (g *Group) PostFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Group {
	g.mux.PostFunc(g.prefix+pattern, fun)
	return g
}

// DeleteFunc相当于ServeMux.DeleteFunc(prefix+pattern, func)的简易写法
func (g *Group) DeleteFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Group {
	g.mux.DeleteFunc(g.prefix+pattern, fun)
	return g
}

// AnyFunc相当于ServeMux.AnyFunc(prefix+pattern, func)的简易写法
func (g *Group) AnyFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Group {
	g.mux.AnyFunc(g.prefix+pattern, fun)
	return g
}

// AnyFunc相当于ServeMux.Remove(prefix+pattern, methods...)的简易写法
func (g *Group) Remove(pattern string, methods ...string) {
	g.mux.Remove(g.prefix+pattern, methods...)
}
