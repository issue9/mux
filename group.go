// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
)

// 一个分组信息，可用于控制一组路由项是否启用。
//  g := srv.Group()
//  g.Get("/admin", h)
//  g.Get("/admin/login", h)
//  g.Stop() // 所有通过g绑定的路由都将停止解析。
type Group struct {
	isRunning bool
	mux       *ServeMux
}

// 当前分组的路由是否处于运行状态
func (g *Group) IsRunning() bool {
	return g.isRunning
}

// 将当前分组改为运行状态
func (g *Group) Start() {
	g.isRunning = true
}

// 将当前分组改为暂停状态。
func (g *Group) Stop() {
	g.isRunning = false
}

// Add 相当于ServeMux.Add(pattern, h, "POST"...)
func (g *Group) Add(pattern string, h http.Handler, methods ...string) *Group {
	g.mux.add(g, pattern, h, methods...)
	return g
}

// Get 相当于ServeMux.Get(pattern, h)
func (g *Group) Get(pattern string, h http.Handler) *Group {
	return g.Add(pattern, h, "GET")
}

// Post 相当于ServeMux.Post(pattern, h)
func (g *Group) Post(pattern string, h http.Handler) *Group {
	return g.Add(pattern, h, "POST")
}

// Delete 相当于ServeMux.Delete(pattern, h)
func (g *Group) Delete(pattern string, h http.Handler) *Group {
	return g.Add(pattern, h, "DELETE")
}

// Put 相当于ServeMux.Put(pattern, h)
func (g *Group) Put(pattern string, h http.Handler) *Group {
	return g.Add(pattern, h, "PUT")
}

// Patch 相当于ServeMux.Patch(pattern, h)
func (g *Group) Patch(pattern string, h http.Handler) *Group {
	return g.Add(pattern, h, "PATCH")
}

// Any 相当于ServeMux.Any(pattern, h)
func (g *Group) Any(pattern string, h http.Handler) *Group {
	return g.Add(pattern, h, supportMethods...)
}

// AddFunc 功能同ServeMux.AddFunc(pattern, fun, ...)
func (g *Group) AddFunc(pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) *Group {
	g.mux.addFunc(g, pattern, fun, methods...)
	return g
}

// GetFunc 相当于ServeMux.GetFunc(pattern, func)
func (g *Group) GetFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Group {
	return g.AddFunc(pattern, fun, "GET")
}

// PutFunc 相当于ServeMux.PutFunc(pattern, func)
func (g *Group) PutFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Group {
	return g.AddFunc(pattern, fun, "PUT")
}

// PostFunc 相当于ServeMux.PostFunc(pattern, func)
func (g *Group) PostFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Group {
	return g.AddFunc(pattern, fun, "POST")
}

// DeleteFunc 相当于ServeMux.DeleteFunc(pattern, func)
func (g *Group) DeleteFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Group {
	return g.AddFunc(pattern, fun, "DELETE")
}

// PatchFunc 相当于ServeMux.PatchFunc(pattern, func)
func (g *Group) PatchFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Group {
	return g.AddFunc(pattern, fun, "PATCH")
}

// AnyFunc 相当于ServeMux.AnyFunc(pattern, func)
func (g *Group) AnyFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Group {
	return g.AddFunc(pattern, fun, supportMethods...)
}

// Clean 清除所有与Group相关联的路由项
func (g *Group) Clean() *Group {
	g.mux.mu.Lock()
	defer g.mux.mu.Unlock()

	for _, method := range supportMethods {
		l, found := g.mux.hosts[method]
		if !found {
			continue
		}

		for item := l.Front(); item != nil; {
			curr := item
			item = item.Next()

			entry := curr.Value.(entryer)
			if entry.getGroup() == g {
				l.Remove(curr)
			}
		}
	} // end for

	for _, method := range supportMethods {
		l, found := g.mux.paths[method]
		if !found {
			continue
		}

		for item := l.Front(); item != nil; {
			curr := item
			item = item.Next()

			entry := curr.Value.(entryer)
			if entry.getGroup() == g {
				l.Remove(curr)
			}
		}
	} // end for

	return g
}

// 声明或是获取一组路由，可以控制该组的路由是否启用。
//  g := srv.Group()
//  g.Get("/admin", h)
//  g.Get("/admin/login", h)
//  g.Stop() // 所有通过g绑定的路由都将停止解析。
func (mux *ServeMux) Group() *Group {
	return &Group{
		mux:       mux,
		isRunning: true,
	}
}
