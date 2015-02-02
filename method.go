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
)

// 用于指定特定method的handler
//
//  m := mux.NewMethod(nil)
//  m.Get(h1).
//    Post(h2).
//    Add(h3, "GET", "POST")
//  http.ListenAndServe(m)
type Method struct {
	mu         sync.Mutex
	errHandler ErrorHandler
	entries    map[string]*methodEntries // 某一method对应的所有Handler
}

type methodEntries struct {
	list  []*entry
	named map[string]*entry
}

// 声明一个新的Method
func NewMethod(err ErrorHandler) *Method {
	if err == nil {
		err = defaultErrorHandler
	}

	return &Method{
		// 至少5个：get/post/delete/put/*
		entries:    make(map[string]*methodEntries, 5),
		errHandler: err,
	}
}

// 添加一条数据。
// methods参数应该只能为http.Request.Method中合法的字符串以及
// 代表所有方法的"*"，其它任何字符串都是无效的，但不会提示错误。
// 当methods或是h为空时，将返回错误信息。
func (m *Method) Add(pattern string, h http.Handler, methods ...string) error {
	if h == nil {
		return errors.New("h参数不能为nil")
	}

	if len(methods) == 0 {
		return errors.New("请至少指定一个methods参数")
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

		_, found = entries.named[pattern]
		if found {
			return fmt.Errorf("该表达式[%v]已经存在", pattern)
		}

		entry := newEntry(pattern, h)
		entries.list = append(entries.list, entry)
		entries.named[pattern] = entry
	}

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

// 相当于Add，但是在发生错误时不返回错误信息，直接panic
func (m *Method) MustAdd(pattern string, h http.Handler, method string) *Method {
	if err := m.Add(pattern, h, method); err != nil {
		panic(err)
	}

	return m
}

// Get相当于m.MustAdd(h, "GET")的简易写法
func (m *Method) MustGet(pattern string, h http.Handler) *Method {
	return m.MustAdd(pattern, h, "GET")
}

// Post相当于m.MustAdd(h, "POST")的简易写法
func (m *Method) MustPost(pattern string, h http.Handler) *Method {
	return m.MustAdd(pattern, h, "POST")
}

// Delete相当于m.MustAdd(h, "DELETE")的简易写法
func (m *Method) MustDelete(pattern string, h http.Handler) *Method {
	return m.MustAdd(pattern, h, "DELETE")
}

// Put相当于m.MustAdd(h, "PUT")的简易写法
func (m *Method) MustPut(pattern string, h http.Handler) *Method {
	return m.MustAdd(pattern, h, "PUT")
}

// Any相当于m.MustAdd(h, "*")的简易写法
func (m *Method) MustAny(pattern string, h http.Handler) *Method {
	return m.MustAdd(pattern, h, "*")
}

// implement http.Handler.ServerHTTP()
func (m *Method) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entries, found := m.entries[req.Method]
	if !found {
		entries, found = m.entries["*"]
		if !found { // 也不存在于*中，则表示未找到与之匹配的method
			m.errHandler(w, "没有找到与之匹配的方法："+req.Method, 404)
			return
		}
	}

	for _, entry := range entries.list {
		if !entry.match(req.URL.Path) {
			continue
		}

		ctx := GetContext(req)
		ctx.Set("params", entry.getNamedCapture(req.URL.Path))
		entry.handler.ServeHTTP(w, req)
		return
	}

	m.errHandler(w, "没有找到与之前匹配的路径："+req.URL.Path, 404)
}
