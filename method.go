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

// 用于处理特定method的handler，定义了六组函数：
//  Add()    / AddFunc()
//  Get()    / GetFunc()
//  Post()   / PostFunc()
//  Delete() / DeleteFunc()
//  Put()    / PutFunc()
//  Any()    / AnyFunc()
//
//
//  m := mux.NewMethod()
//  m.MustGet(h1).
//    MustPost(h2).
//    MustAdd(h3, "GET", "POST")
//  http.ListenAndServe(m)
type Method struct {
	mu      sync.Mutex
	entries map[string]*methodEntries
}

type methodEntries struct {
	list  []*entry
	named map[string]*entry
}

// 声明一个新的Method
func NewMethod() *Method {
	return &Method{
		// 至少5个：get/post/delete/put/*
		entries: make(map[string]*methodEntries, 5),
	}
}

// 添加一条路由数据。
//
// pattern为路由匹配模式，可以是正则匹配也可以是字符串匹配，
// 以第一个字符来确定是否为一个正则匹配，若第一个字符为'?'，
// 则将'?'之后的所有字符当作一个正则表达式来匹配路由；
// 否则为一个普通的字符串匹配；若pattern以'\?'开头，则'\'仅当转换字符。
// methods参数应该只能为http.Request.Method中合法的字符串以及代表所有方法的"*"，
// 其它任何字符串都是无效的，但不会提示错误。当methods或是h为空时，将返回错误信息。
func (m *Method) Add(pattern string, h http.Handler, methods ...string) error {
	if len(methods) == 0 {
		return errors.New("Add:请至少指定一个methods参数")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, method := range methods {
		method = strings.ToUpper(method)

		entries, found := m.entries[method]
		if !found {
			entries = &methodEntries{
				list:  []*entry{},
				named: map[string]*entry{},
			}
			m.entries[method] = entries
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
func (m *Method) Get(pattern string, h http.Handler) error {
	return m.Add(pattern, h, "GET")
}

// Post相当于m.Add(h, "POST")的简易写法
func (m *Method) Post(pattern string, h http.Handler) error {
	return m.Add(pattern, h, "POST")
}

// Delete相当于m.Add(h, "DELETE")的简易写法
func (m *Method) Delete(pattern string, h http.Handler) error {
	return m.Add(pattern, h, "DELETE")
}

// Put相当于m.Add(h, "PUT")的简易写法
func (m *Method) Put(pattern string, h http.Handler) error {
	return m.Add(pattern, h, "PUT")
}

// Any相当于m.Add(h, "*")的简易写法
func (m *Method) Any(pattern string, h http.Handler) error {
	return m.Add(pattern, h, "*")
}

// 功能同Add()，但是将第二个参数从http.Handler换成了func(http.ResponseWriter, *http.Request)
func (m *Method) AddFunc(pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) error {
	return m.Add(pattern, http.HandlerFunc(fun), methods...)
}

// GetFunc相当于m.Add(h, "GET")的简易写法
func (m *Method) GetFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return m.AddFunc(pattern, fun, "GET")
}

// PutFunc相当于m.Add(h, "PUT")的简易写法
func (m *Method) PutFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return m.AddFunc(pattern, fun, "PUT")
}

// PostFunc相当于m.Add(h, "POST")的简易写法
func (m *Method) PostFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return m.AddFunc(pattern, fun, "POST")
}

// DeleteFunc相当于m.Add(h, "DELETE")的简易写法
func (m *Method) DeleteFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return m.AddFunc(pattern, fun, "DELETE")
}

// AnyFunc相当于m.Add(h, "*")的简易写法
func (m *Method) AnyFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return m.AddFunc(pattern, fun, "*")
}

// implement http.Handler.ServerHTTP()
func (m *Method) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 找到与req.Method对应的map，若不存在，则尝试找*
	entries, found := m.entries[req.Method]
	if !found {
		entries, found = m.entries["*"]
		if !found { // 也不存在*
			panic("没有找到与之匹配的方法：" + req.Method)
		}
	}

	for _, entry := range entries.list {
		if ok, mapped := entry.match(req.URL.Path); ok {
			ctx := context.Get(req)
			ctx.Set("params", mapped)
			entry.handler.ServeHTTP(w, req)
			return
		}
	}
	panic("没有找到与之前匹配的路径：" + req.URL.Path)
}
