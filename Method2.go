// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
)

// 功能同Method，但是在接口上稍微有些变化。函数不再单独返回错误信息，
// 而是通过一个专门的函数判断是否存在错误。方便以函数链的形式写代码。
//
// 所有添加路由的方法不再返回错误信息，所以必须在开始路由之前，调用
// HasError()方法判断当前实例是否存在错误。否则在路由开始时会触发panic。
type Method2 struct {
	errors []error
	method *Method
}

// 声明一个Method2实例。
func NewMethod2() *Method2 {
	return &Method2{
		errors: []error{},
		method: NewMethod(),
	}
}

// 是否存在错误信息。
func (m *Method2) HasError() bool {
	return len(m.errors) > 0
}

// 返回所有的错误信息。
func (m *Method2) Errors() []error {
	return m.errors
}

// 添加一条路由数据。
// 具体参数可参考Method.Add()方法。
func (m *Method2) Add(pattern string, h http.Handler, methods ...string) *Method2 {
	if err := m.method.Add(pattern, h, methods...); err != nil {
		m.errors = append(m.errors, err)
	}

	return m
}

// Get相当于m.Add(h, "GET")的简易写法
func (m *Method2) Get(pattern string, h http.Handler) *Method2 {
	return m.Add(pattern, h, "GET")
}

// Post相当于m.Add(h, "POST")的简易写法
func (m *Method2) Post(pattern string, h http.Handler) *Method2 {
	return m.Add(pattern, h, "POST")
}

// Delete相当于m.Add(h, "DELETE")的简易写法
func (m *Method2) Delete(pattern string, h http.Handler) *Method2 {
	return m.Add(pattern, h, "DELETE")
}

// Put相当于m.Add(h, "PUT")的简易写法
func (m *Method2) Put(pattern string, h http.Handler) *Method2 {
	return m.Add(pattern, h, "PUT")
}

// Any相当于m.Add(h, "*")的简易写法
func (m *Method2) Any(pattern string, h http.Handler) *Method2 {
	return m.Add(pattern, h, "*")
}

func (m *Method2) AddFunc(pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) *Method2 {
	return m.Add(pattern, http.HandlerFunc(fun), methods...)
}

func (m *Method2) GetFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Method2 {
	return m.AddFunc(pattern, fun, "GET")
}

func (m *Method2) PostFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Method2 {
	return m.AddFunc(pattern, fun, "POST")
}

func (m *Method2) PutFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Method2 {
	return m.AddFunc(pattern, fun, "PUT")
}

func (m *Method2) DeleteFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Method2 {
	return m.AddFunc(pattern, fun, "DELETE")
}

func (m *Method2) AnyFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Method2 {
	return m.AddFunc(pattern, fun, "*")
}

// implement http.Handler.ServerHTTP()
func (m *Method2) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if m.HasError() {
		panic("ServeHTTP:存在错误信息，具体请调用Method2.Errors()函数查看！")
	}

	m.method.ServeHTTP(w, req)
}
