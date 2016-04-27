// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"container/list"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/issue9/context"
)

var (
	ErrUnsupportedMethod = errors.New("不支持该请求方法")
	ErrPatternExists     = errors.New("该匹配模式已经存在")
)

// 支持的所有请求方法
var supportedMethods = []string{
	"GET",
	"POST",
	"DELETE",
	"PUT",
	"PATCH",
	"OPTIONS",
	"HEAD",
	"TRACE",
}

// http.ServeMux的升级版，可处理对URL的正则匹配和根据METHOD进行过滤。
//
// 用法如下：
//  m := mux.NewServeMux()
//  m.Get("www.example.com/abc", h1).              // 只匹配www.example.com域名下的路径
//    Post("/abc/"", h2).                          // 不限定域名的路径匹配
//    Add("/api/{version:\\d+}",h3, "GET", "POST") // 只匹配GET和POST
//  http.ListenAndServe(m)
//
//
// 路由参数：
//
// 路由参数可通过context包获取：
//  ctx := context.Get(req)
//  params := ctx.Get("params") // 若不存在路由参数，则返回一个空值
// NOTE:记得在退出整个请求之前清除context中的内容：
//  context.Free(req)
//
//
// 匹配规则：
//
// 可能会出现多条记录与同一请求都匹配的情况，这种情况下，
// 系统会找到一条认为最匹配的路由来处理，判断规则如下：
//  1. 静态路由优先于正则路由判断；
//  2. 带域名的路由项优先于不带域名的路由项；
//  3. 完全匹配的路由项优先于部分匹配的路由项；
//  4. 正则只能是完全匹配；
//  5. 只有以/结尾的静态路由才有部分匹配功能。
//
// 正则匹配语法：
//  /post/{id}     // 匹配/post/开头的任意字符串，其后的字符串保存到id中；
//  /post/{id:\d+} // 同上，但id的值只能为\d+；
//  /post/{:\d+}   // 同上，但是没有命名；
type ServeMux struct {
	// 同时处理hosts,paths,options三个的竟争问题
	mu sync.RWMutex

	// 包含域名的路由列表，键名表示method。hosts中静态路由在前，正则路由在后。
	hosts map[string]*list.List

	// 路由列表，键名表示method。paths中静态路由在前，正则路由在后。
	paths map[string]*list.List
}

// 声明一个新的ServeMux
func NewServeMux() *ServeMux {
	d := make(map[string]*list.List, len(supportedMethods))
	l := make(map[string]*list.List, len(supportedMethods))
	for _, method := range supportedMethods {
		d[method] = list.New()
		l[method] = list.New()
	}

	return &ServeMux{
		hosts: d,
		paths: l,
	}
}

// 检测匹配模式是否存在，若不存在则返回error
func (mux *ServeMux) checkExists(pattern, method string) error {
	l, found := mux.hosts[method]
	if !found {
		return ErrUnsupportedMethod
	}

	for item := l.Front(); item != nil; item = item.Next() {
		if e := item.Value.(entryer); e.getPattern() == pattern {
			return ErrPatternExists
		}
	}

	l, found = mux.paths[method]
	if !found {
		return ErrUnsupportedMethod
	}

	for item := l.Front(); item != nil; item = item.Next() {
		if e := item.Value.(entryer); e.getPattern() == pattern {
			return ErrPatternExists
		}
	}

	return nil
}

// 添加一个路由项。
func (mux *ServeMux) add(g *Group, pattern string, h http.Handler, methods ...string) *ServeMux {
	if len(pattern) == 0 {
		panic("参数pattern不能为空")
	}

	if h == nil {
		panic("参数h不能为空")
	}

	if len(methods) == 0 {
		methods = supportedMethods
	}

	e := newEntry(pattern, h, g)

	mux.mu.Lock()
	defer mux.mu.Unlock()

	for _, method := range methods {
		method = strings.ToUpper(method)

		if err := mux.checkExists(pattern, method); err != nil {
			panic(err)
		}

		switch {
		case pattern[0] == '/' && !e.isRegexp(): // 包含域名匹配的静态路由，在前端插入
			mux.paths[method].PushFront(e)
		case pattern[0] == '/' && e.isRegexp(): // 包含域名匹配的正则路由，在后端插入
			mux.paths[method].PushBack(e)
		case pattern[0] != '/' && !e.isRegexp(): // 不包含域名匹配的静态路由，在前端插入
			mux.hosts[method].PushFront(e)
		case pattern[0] != '/' && e.isRegexp(): // 不包含域名匹配的正则路由，在后端插入
			mux.hosts[method].PushFront(e)
		}
	}

	return mux
}

// Clean 清除所有的路由项
func (mux *ServeMux) Clean() *ServeMux {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	for _, method := range supportedMethods {
		l, found := mux.hosts[method]
		if !found {
			continue
		}
		l.Init()
	}

	for _, method := range supportedMethods {
		l, found := mux.paths[method]
		if !found {
			continue
		}
		l.Init()
	}

	return mux
}

// 移除指定的路由项，通过路由表达式和method来匹配。
// 当未指定methods时，将删除所有method匹配的项。
// 指定错误的method值，将自动忽略该值。
func (mux *ServeMux) Remove(pattern string, methods ...string) {
	if len(methods) == 0 { // 删除所有method下匹配的项
		methods = supportedMethods
	}

	if pattern[0] == '/' {
		mux.removePaths(pattern, methods)
		return
	}
	mux.removeHosts(pattern, methods)
}

func (mux *ServeMux) removeHosts(pattern string, methods []string) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	for _, method := range methods {
		l, found := mux.hosts[method]
		if !found {
			continue
		}

		for item := l.Front(); item != nil; item = item.Next() {
			if e := item.Value.(entryer); e.getPattern() == pattern {
				l.Remove(item)
				break // 最多只有一个匹配
			}
		}
	} // end for methods
}

func (mux *ServeMux) removePaths(pattern string, methods []string) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	for _, method := range methods {
		l, found := mux.paths[method]
		if !found {
			continue
		}

		for item := l.Front(); item != nil; item = item.Next() {
			if e := item.Value.(entryer); e.getPattern() == pattern {
				l.Remove(item)
				break // 最多只有一个匹配
			}
		}
	} // end for methods
}

// Add 添加一条路由数据。
//
// pattern为路由匹配模式，可以是正则匹配也可以是字符串匹配，
// 可以带上域名，当第一个字符为'/'当作是一个路径，否则就将'/'之前的当作域名或IP。
// methods参数应该只能为supportedMethods中的字符串，若不指定，默认为所有，
// 当h或是pattern为空时，将触发panic。
func (mux *ServeMux) Add(pattern string, h http.Handler, methods ...string) *ServeMux {
	return mux.add(nil, pattern, h, methods...)
}

// Get 相当于ServeMux.Add(pattern, h, "GET")的简易写法
func (mux *ServeMux) Get(pattern string, h http.Handler) *ServeMux {
	return mux.Add(pattern, h, "GET")
}

// Post 相当于ServeMux.Add(pattern, h, "POST")的简易写法
func (mux *ServeMux) Post(pattern string, h http.Handler) *ServeMux {
	return mux.Add(pattern, h, "POST")
}

// Delete 相当于ServeMux.Add(pattern, h, "DELETE")的简易写法
func (mux *ServeMux) Delete(pattern string, h http.Handler) *ServeMux {
	return mux.Add(pattern, h, "DELETE")
}

// Put 相当于ServeMux.Add(pattern, h, "PUT")的简易写法
func (mux *ServeMux) Put(pattern string, h http.Handler) *ServeMux {
	return mux.Add(pattern, h, "PUT")
}

// Patch 相当于ServeMux.Add(pattern, h, "PATCH")的简易写法
func (mux *ServeMux) Patch(pattern string, h http.Handler) *ServeMux {
	return mux.Add(pattern, h, "PATCH")
}

// Any 相当于ServeMux.Add(pattern, h)的简易写法
func (mux *ServeMux) Any(pattern string, h http.Handler) *ServeMux {
	return mux.Add(pattern, h, supportedMethods...)
}

func (mux *ServeMux) addFunc(g *Group, pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) *ServeMux {
	return mux.add(g, pattern, http.HandlerFunc(fun), methods...)
}

// AddFunc 功能同ServeMux.Add()，但是将第二个参数从http.Handler换成了func(http.ResponseWriter, *http.Request)
func (mux *ServeMux) AddFunc(pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) *ServeMux {
	return mux.Add(pattern, http.HandlerFunc(fun), methods...)
}

// GetFunc 相当于ServeMux.AddFunc(pattern, func, "GET")的简易写法
func (mux *ServeMux) GetFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.AddFunc(pattern, fun, "GET")
}

// PutFunc 相当于ServeMux.AddFunc(pattern, func, "PUT")的简易写法
func (mux *ServeMux) PutFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.AddFunc(pattern, fun, "PUT")
}

// PostFunc 相当于ServeMux.AddFunc(pattern, func, "POST")的简易写法
func (mux *ServeMux) PostFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.AddFunc(pattern, fun, "POST")
}

// DeleteFunc 相当于ServeMux.AddFunc(pattern, func, "DELETE")的简易写法
func (mux *ServeMux) DeleteFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.AddFunc(pattern, fun, "DELETE")
}

// PatchFunc 相当于ServeMux.AddFunc(pattern, func, "PATCH")的简易写法
func (mux *ServeMux) PatchFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.AddFunc(pattern, fun, "PATCH")
}

// AnyFunc 相当于ServeMux.AddFunc(pattern, func)的简易写法
func (mux *ServeMux) AnyFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.AddFunc(pattern, fun, supportedMethods...)
}

// 查找最匹配的路由项
func (mux *ServeMux) match(r *http.Request) (p string, e entryer) {
	hostURL := r.Host + r.URL.Path
	size := -1 // 匹配度，0表示完全匹配，负数表示完全不匹配，其它值越小匹配度越高

	mux.mu.RLock()
	defer mux.mu.RUnlock()

	for item := mux.hosts[r.Method].Front(); item != nil; item = item.Next() {
		entry := item.Value.(entryer)

		s := entry.match(hostURL)
		if s == -1 || (size > 0 && s >= size) { // 完全不匹配，或是匹配度没有当前的高
			continue
		}

		size = s
		e = entry
		p = hostURL

		if s == 0 { // 完全匹配，可以中止匹配过程
			return p, e
		}
	} // end for

	for item := mux.paths[r.Method].Front(); item != nil; item = item.Next() {
		entry := item.Value.(entryer)

		s := entry.match(r.URL.Path)
		if s == -1 || (size > 0 && s >= size) { // 完全不匹配，或是匹配度没有当前的高
			continue
		}

		size = s
		e = entry
		p = r.URL.Path

		if s == 0 { // 完全匹配，可以中止匹配过程
			return p, e
		}
	} // end for

	if size < 0 {
		return "", nil
	}
	return p, e
}

// implement http.Handler.ServerHTTP()
// 若没有找到匹配路由，返回404
func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p, e := mux.match(r)

	if e == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	params := e.getParams(p)
	if params != nil {
		ctx := context.Get(r)
		ctx.Set("params", params)
	}
	e.serveHTTP(w, r)
}
