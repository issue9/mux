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
	entries map[string]*methodEntries
}

type methodEntries struct {
	list  []*entry
	named map[string]*entry
}

// 声明一个新的ServeMux
func NewServeMux() *ServeMux {
	return &ServeMux{
		// 至少5个：get/post/delete/put/*
		entries: make(map[string]*methodEntries, 5),
	}
}

// 添加一条路由数据。
//
// pattern为路由匹配模式，可以是正则匹配也可以是字符串匹配，
// 以第一个字符来确定是否为一个正则匹配，若第一个字符为'?'，
// 则将'?'之后的所有字符当作一个正则表达式来匹配路由；
// 否则为一个普通的字符串匹配。
// methods参数应该只能为http.Request.Method中合法的字符串以及代表所有方法的"*"，
// 其它任何字符串都是无效的，但不会提示错误。
//
// 当methods或是h为空时，将返回错误信息。
func (mux *ServeMux) Add(pattern string, h http.Handler, methods ...string) error {
	if len(methods) == 0 {
		return errors.New("Add:请至少指定一个methods参数")
	}

	if h == nil {
		return errors.New("Add:参数h不能为空")
	}

	mux.mu.Lock()
	defer mux.mu.Unlock()

	for _, method := range methods {
		method = strings.ToUpper(method)

		entries, found := mux.entries[method]
		if !found {
			entries = &methodEntries{
				list:  []*entry{},
				named: map[string]*entry{},
			}
			mux.entries[method] = entries
		}

		if _, found = entries.named[pattern]; found {
			return fmt.Errorf("Add:该表达式[%v]已经存在", pattern)
		}

		entry, err := newEntry(pattern, h)
		if err != nil {
			return err
		}
		entries.list = append(entries.list, entry)
		entries.named[pattern] = entry
	} // end for methods

	return nil
}

// Get相当于m.Add(h, "GET")的简易写法
func (mux *ServeMux) Get(pattern string, h http.Handler) error {
	return mux.Add(pattern, h, "GET")
}

// Post相当于m.Add(h, "POST")的简易写法
func (mux *ServeMux) Post(pattern string, h http.Handler) error {
	return mux.Add(pattern, h, "POST")
}

// Delete相当于m.Add(h, "DELETE")的简易写法
func (mux *ServeMux) Delete(pattern string, h http.Handler) error {
	return mux.Add(pattern, h, "DELETE")
}

// Put相当于m.Add(h, "PUT")的简易写法
func (mux *ServeMux) Put(pattern string, h http.Handler) error {
	return mux.Add(pattern, h, "PUT")
}

// Any相当于m.Add(h, "*")的简易写法
func (mux *ServeMux) Any(pattern string, h http.Handler) error {
	return mux.Add(pattern, h, "*")
}

// 功能同Add()，但是将第二个参数从http.Handler换成了func(http.ResponseWriter, *http.Request)
func (mux *ServeMux) AddFunc(pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) error {
	return mux.Add(pattern, http.HandlerFunc(fun), methods...)
}

// GetFunc相当于m.Add(h, "GET")的简易写法
func (mux *ServeMux) GetFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return mux.AddFunc(pattern, fun, "GET")
}

// PutFunc相当于m.Add(h, "PUT")的简易写法
func (mux *ServeMux) PutFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return mux.AddFunc(pattern, fun, "PUT")
}

// PostFunc相当于m.Add(h, "POST")的简易写法
func (mux *ServeMux) PostFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return mux.AddFunc(pattern, fun, "POST")
}

// DeleteFunc相当于m.Add(h, "DELETE")的简易写法
func (mux *ServeMux) DeleteFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return mux.AddFunc(pattern, fun, "DELETE")
}

// AnyFunc相当于m.Add(h, "*")的简易写法
func (mux *ServeMux) AnyFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return mux.AddFunc(pattern, fun, "*")
}

// implement http.Handler.ServerHTTP()
func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	// 找到与req.ServeMux对应的map，若不存在，则尝试找*
	entries, found := mux.entries[req.Method]
	if !found {
		entries, found = mux.entries["*"]
		if !found { // 也不存在*
			panic("ServeHTTP:没有找到与之匹配的方法：" + req.Method)
		}
	}

	pattern := req.URL.Path
	for _, entry := range entries.list {
		if entry.pattern[0] != '/' {
			pattern = req.Host + req.URL.Path
		}

		if entry.expr == nil { // 普通字符串匹配
			if entry.pattern != pattern {
				continue
			}
			ctx := context.Get(req)
			ctx.Set("params", nil)
			entry.handler.ServeHTTP(w, req)
			return
		}

		if !entry.expr.MatchString(pattern) { //不能匹配正则表达式
			continue
		}

		// 正确匹配正则表达式，则获相关的正则表达式命名变量。
		mapped := make(map[string]string)
		subexps := entry.expr.SubexpNames()
		args := entry.expr.FindStringSubmatch(pattern)
		for index, name := range subexps {
			if len(name) > 0 {
				mapped[name] = args[index]
			}
		}
		ctx := context.Get(req)
		ctx.Set("params", mapped)
		entry.handler.ServeHTTP(w, req)
		return
	} // end for entries.list

	panic("没有找到与之前匹配的路径：" + pattern)
}
