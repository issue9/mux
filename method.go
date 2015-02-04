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

// 用于处理特定method的handler
// 定义了六个函数：
//  Add(...) //
//  Get()    //
//  Post()
//  Delete()
//  Put()
//  Any()
// 以及根据这些函数延伸出来的一系列函数：
// *Func...
// Must*...
// Must*Func...
//
//  m := mux.NewMethod(nil)
//  m.MustGet(h1).
//    MustPost(h2).
//    MustAdd(h3, "GET", "POST")
//  http.ListenAndServe(m)
type Method struct {
	mu         sync.Mutex
	errHandler ErrorHandler
	entries    map[string]*methodEntries
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

		if _, found = entries.named[pattern]; found {
			return fmt.Errorf("该表达式[%v]已经存在", pattern)
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

func (m *Method) AddFunc(pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) error {
	return m.Add(pattern, http.HandlerFunc(fun), methods...)
}

func (m *Method) GetFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return m.AddFunc(pattern, fun, "GET")
}

func (m *Method) PutFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return m.AddFunc(pattern, fun, "PUT")
}

func (m *Method) PostFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return m.AddFunc(pattern, fun, "POST")
}

func (m *Method) DeleteFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return m.AddFunc(pattern, fun, "DELETE")
}

func (m *Method) AnyFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) error {
	return m.AddFunc(pattern, fun, "*")
}

// 相当于Add，但是在发生错误时不返回错误信息，直接panic
func (m *Method) MustAdd(pattern string, h http.Handler, methods ...string) *Method {
	if err := m.Add(pattern, h, methods...); err != nil {
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

func (m *Method) MustAddFunc(pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) *Method {
	return m.MustAdd(pattern, http.HandlerFunc(fun), methods...)
}

func (m *Method) MustGetFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Method {
	return m.MustAddFunc(pattern, fun, "GET")
}

func (m *Method) MustPostFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Method {
	return m.MustAddFunc(pattern, fun, "POST")
}

func (m *Method) MustPutFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Method {
	return m.MustAddFunc(pattern, fun, "PUT")
}

func (m *Method) MustDeleteFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Method {
	return m.MustAddFunc(pattern, fun, "DELETE")
}

func (m *Method) MustAnyFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Method {
	return m.MustAddFunc(pattern, fun, "*")
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
