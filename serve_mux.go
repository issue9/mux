// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/issue9/context"
)

// http.ServeMux的升级版，可处理正则匹配和method匹配。定义了以下六组函数：
//  Add()    / AddFunc()
//  Get()    / GetFunc()
//  Post()   / PostFunc()
//  Delete() / DeleteFunc()
//  Put()    / PutFunc()
//  Any()    / AnyFunc()
//
// 简单的用法如下：
//  m := mux.NewServeMux()
//  m.Get("www.example.com/abc", h1). // 只匹配www.example.com域名下的路径
//    Post("/abc/"", h2). // 不限定域名的路径匹配
//    Add("api/1",h3, "GET", "POST")
//  http.ListenAndServe(m)
//
// 还有一个功能与之相同的ServeMux2，用法上有些稍微的差别。具体可参考ServeMux的文档。
type ServeMux struct {
	mu      sync.Mutex
	methods map[string]*entries
}

type entries struct {
	list  []*entry
	named map[string]*entry
}

// 声明一个新的ServeMux
func NewServeMux() *ServeMux {
	return &ServeMux{
		// 至少5个：get/post/delete/put/*
		methods: make(map[string]*entries, 5),
	}
}

// 添加一条路由数据。
//
// pattern为路由匹配模式，可以是正则匹配也可以是字符串匹配，
// 若第一个字符为'?'，则将第一个字符去掉之后，将其后的字符串当作一个正则表达式来处理；
// 否则整个字符串就当作一个普通的字符串来进行比较。
// pattern可以带上域名，当第一个字符为'/'当作是一个路径，否则就将'/'之前的当作域名或IP。
// methods参数应该只能为http.Request.Method中合法的字符串以及代表所有方法的"*"，
// 其它任何字符串都是无效的，但不会提示错误。当methods或是h为空时，将返回错误信息。
func (mux *ServeMux) Add(pattern string, h http.Handler, methods ...string) error {
	if len(methods) == 0 {
		return errors.New("Add:请至少指定一个methods参数")
	}

	if h == nil {
		return errors.New("Add:参数h不能为空")
	}

	if len(pattern) == 0 {
		return errors.New("Add:pattern正则内容不能为空")
	}

	e := newEntry(pattern, h)

	mux.mu.Lock()
	defer mux.mu.Unlock()

	for _, method := range methods {
		method = strings.ToUpper(method)

		methodEntries, found := mux.methods[method]
		if !found { // 为新的method分配空间
			methodEntries = &entries{
				list:  []*entry{},
				named: map[string]*entry{},
			}
			mux.methods[method] = methodEntries
		}

		if _, found = methodEntries.named[pattern]; found {
			return fmt.Errorf("Add:该表达式[%v]已经存在", pattern)
		}

		methodEntries.list = append(methodEntries.list, e)
		methodEntries.named[pattern] = e
	} // end for methods

	return nil
}

// Get相当于ServeMux.Add(pattern, h, "GET")的简易写法
func (mux *ServeMux) Get(pattern string, h http.Handler) error {
	return mux.Add(pattern, h, "GET")
}

// Post相当于ServeMux.Add(pattern, h, "POST")的简易写法
func (mux *ServeMux) Post(pattern string, h http.Handler) error {
	return mux.Add(pattern, h, "POST")
}

// Delete相当于ServeMux.Add(pattern, h, "DELETE")的简易写法
func (mux *ServeMux) Delete(pattern string, h http.Handler) error {
	return mux.Add(pattern, h, "DELETE")
}

// Put相当于ServeMux.Add(pattern, h, "PUT")的简易写法
func (mux *ServeMux) Put(pattern string, h http.Handler) error {
	return mux.Add(pattern, h, "PUT")
}

// Any相当于ServeMux.Add(pattern, h, "*")的简易写法
func (mux *ServeMux) Any(pattern string, h http.Handler) error {
	return mux.Add(pattern, h, "*")
}

// 功能同ServeMux.Add()，但是将第二个参数从http.Handler换成了func(http.ResponseWriter, *http.Request)
func (mux *ServeMux) AddFunc(pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) error {
	return mux.Add(pattern, http.HandlerFunc(fun), methods...)
}

// GetFunc相当于ServeMux.AddFunc(pattern, func, "GET")的简易写法
func (mux *ServeMux) GetFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return mux.AddFunc(pattern, fun, "GET")
}

// PutFunc相当于ServeMux.Add(pattern, func, "PUT")的简易写法
func (mux *ServeMux) PutFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return mux.AddFunc(pattern, fun, "PUT")
}

// PostFunc相当于ServeMux.Add(pattern, func, "POST")的简易写法
func (mux *ServeMux) PostFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return mux.AddFunc(pattern, fun, "POST")
}

// DeleteFunc相当于ServeMux.Add(pattern, func, "DELETE")的简易写法
func (mux *ServeMux) DeleteFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return mux.AddFunc(pattern, fun, "DELETE")
}

// AnyFunc相当于ServeMux.Add(pattern, func, "*")的简易写法
func (mux *ServeMux) AnyFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return mux.AddFunc(pattern, fun, "*")
}

// 仅供Remove()使用
func deleteFromEntries(entries *entries, pattern string) {
	delete(entries.named, pattern)

	for index, entry := range entries.list {
		if entry.pattern == pattern {
			entries.list = append(entries.list[:index], entries.list[index+1:]...)
			return
		}
	}
}

// 移除指定的路由项，通过路由表达式和method来匹配。
// 当未指定methods时，将不发生任何删除操作。
func (mux *ServeMux) Remove(pattern string, methods ...string) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	// 若存在method值为"*"的情部，则从任何一个method中删除路由项。
	for _, m := range methods {
		if m != "*" {
			continue
		}

		for _, entries := range mux.methods {
			deleteFromEntries(entries, pattern)
		}
		return
	}

	// 否则只从匹配的method中删除路由项。
	for _, method := range methods {
		if entries, found := mux.methods[method]; found {
			deleteFromEntries(entries, pattern)
		}

	}
}

// 获取符合当前Method的所有路由项。即req.Method与"*"的内容
// 仅供ServeHTTP()调用。
func (mux *ServeMux) getList(req *http.Request) []*entry {
	if methods, found := mux.methods[req.Method]; found {
		if anyMethods, found := mux.methods["*"]; found {
			return append(methods.list, anyMethods.list...)
		}
		return methods.list
	}

	if methods, found := mux.methods["*"]; found {
		return methods.list
	}

	panic("ServeHTTP:没有找到与之匹配的方法：" + req.Method)
}

// implement http.Handler.ServerHTTP()
func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	list := mux.getList(req)
	hostURL := req.Host + req.URL.Path
	size := -1
	var e *entry
	var p string

	for _, entry := range list {
		url := req.URL.Path
		if entry.pattern[0] != '/' {
			url = hostURL
		}

		s := entry.isMatch(url)
		if s == 0 { // 完全匹配，可以中止匹配过程
			size = 0
			e = entry
			p = url
			break
		}
		if s < 0 || (size > 0 && s > size) { // 不匹配
			continue
		}
		size = s
		e = entry
		p = url
	} // end for methods.list
	if size < 0 {
		panic(fmt.Sprintf("没有找到与之前匹配的路径，Host:[%v],Path:[%v]", req.Host, req.URL.Path))
	}

	ctx := context.Get(req)
	ctx.Set("params", e.getParams(p))
	e.handler.ServeHTTP(w, req)
}
