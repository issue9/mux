// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
)

// 功能同ServeMux，但是在接口上稍微有些变化。函数不再单独返回错误信息，
// 而是通过一个专门的函数判断是否存在错误。方便以函数链的形式写代码。
//
// 所有添加路由的方法不再返回错误信息，所以必须在开始路由之前，
// 调用HasError()方法判断当前实例是否存在错误。否则在路由开始时会触发panic。
type ServeMux2 struct {
	errors []error
	mux    *ServeMux
}

// 声明一个ServeMux2实例。
func NewServeMux2() *ServeMux2 {
	return &ServeMux2{
		errors: []error{},
		mux:    NewServeMux(),
	}
}

// 是否存在错误信息。
func (m *ServeMux2) HasError() bool {
	return len(m.errors) > 0
}

// 返回所有的错误信息。
func (m *ServeMux2) Errors() []error {
	return m.errors
}

// 添加一条路由数据。
// 具体参数可参考ServeMux2.Add()方法。
func (m *ServeMux2) Add(pattern string, h http.Handler, muxs ...string) *ServeMux2 {
	if err := m.mux.Add(pattern, h, muxs...); err != nil {
		m.errors = append(m.errors, err)
	}

	return m
}

// Get相当于ServeMux2.Add(pattern, h, "GET")的简易写法
func (m *ServeMux2) Get(pattern string, h http.Handler) *ServeMux2 {
	return m.Add(pattern, h, "GET")
}

// Post相当于ServeMux2.Add(pattern, h, "POST")的简易写法
func (m *ServeMux2) Post(pattern string, h http.Handler) *ServeMux2 {
	return m.Add(pattern, h, "POST")
}

// Delete相当于ServeMux2.Add(pattern, h, "DELETE")的简易写法
func (m *ServeMux2) Delete(pattern string, h http.Handler) *ServeMux2 {
	return m.Add(pattern, h, "DELETE")
}

// Put相当于ServeMux2.Add(pattern, h, "PUT")的简易写法
func (m *ServeMux2) Put(pattern string, h http.Handler) *ServeMux2 {
	return m.Add(pattern, h, "PUT")
}

// Any相当于ServeMux2.Add(pattern, h)的简易写法
func (m *ServeMux2) Any(pattern string, h http.Handler) *ServeMux2 {
	return m.Add(pattern, h)
}

// 功能同Add()，但是将第二个参数从http.Handler换成了func(http.ResponseWriter, *http.Request)
func (m *ServeMux2) AddFunc(pattern string, fun func(http.ResponseWriter, *http.Request), muxs ...string) *ServeMux2 {
	return m.Add(pattern, http.HandlerFunc(fun), muxs...)
}

// GetFunc相当于ServeMux2.AddFunc(pattern, fun, "GET")的简易写法
func (m *ServeMux2) GetFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux2 {
	return m.AddFunc(pattern, fun, "GET")
}

// PostFunc相当于ServeMux2.AddFunc(pattern, fun, "POST")的简易写法
func (m *ServeMux2) PostFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux2 {
	return m.AddFunc(pattern, fun, "POST")
}

// PutFunc相当于ServeMux2.AddFunc(pattern, fun, "PUT")的简易写法
func (m *ServeMux2) PutFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux2 {
	return m.AddFunc(pattern, fun, "PUT")
}

// DeleteFunc相当于ServeMux2.AddFunc(pattern, fun, "DELETE")的简易写法
func (m *ServeMux2) DeleteFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux2 {
	return m.AddFunc(pattern, fun, "DELETE")
}

// AnyFunc相当于ServeMux2.AddFunc(pattern, fun)的简易写法
func (m *ServeMux2) AnyFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux2 {
	return m.AddFunc(pattern, fun)
}

// 创建一个路由组，该组中所有的操作，都会加上前缀prefix
//  g := srv.Group("/api")
//  g.Get("/users")  // 相当于 srv.Get("/api/users")
//  g.Get("/user/1") // 相当于 srv.Get("/api/user/1")
func (m *ServeMux2) Group(prefix string) *Group2 {
	return &Group2{
		mux:    m,
		prefix: prefix,
	}
}

// 移除指定匹配的路由项。
func (m *ServeMux2) Remove(pattern string, methods ...string) {
	m.mux.Remove(pattern, methods...)
}

// implement http.Handler.ServerHTTP()
// 若有错误，则会panic
func (m *ServeMux2) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if m.HasError() {
		panic("ServeHTTP:存在错误信息，具体请调用ServeMux2.Errors()函数查看！")
	}

	m.mux.ServeHTTP(w, req)
}
