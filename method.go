// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"strings"
)

// 用于匹配http.Request.Method的Handler
//  m := mux.NewMethod()
//  m.Get(h1).
//    Post(h2).
//    Add(h3, "GET", "POST")
//  http.ListenAndServe(m)
type Method struct {
	// 某一method对应的所有Handler
	entries map[string]Matches
}

func NewMethod() *Method {
	return &Method{entries: make(map[string]Matches)}
}

// 添加一条数据。
// methods参数应该只能为http.Request.Method中合法的字符串以及
// 代表所有方法的"*"，其它任何字符串都是无效的，但不会提示错误。
func (m *Method) Add(h Matcher, methods ...string) *Method {
	if len(methods) == 0 {
		panic("请至少指定一个methods参数")
	}

	for _, method := range methods {
		method = strings.ToUpper(method)
		_, found := m.entries[method]
		if !found {
			m.entries[method] = make(Matches, 0, 1)
		}

		m.entries[method] = append(m.entries[method], h)
	}

	return m
}

// Get相当于m.Add(h, "GET")的简易写法
func (m *Method) Get(h Matcher) *Method {
	return m.Add(h, "GET")
}

// Post相当于m.Add(h, "POST")的简易写法
func (m *Method) Post(h Matcher) *Method {
	return m.Add(h, "POST")
}

// Delete相当于m.Add(h, "DELETE")的简易写法
func (m *Method) Delete(h Matcher) *Method {
	return m.Add(h, "DELETE")
}

// Put相当于m.Add(h, "PUT")的简易写法
func (m *Method) Put(h Matcher) *Method {
	return m.Add(h, "PUT")
}

// Any相当于m.Add(h, "*")的简易写法
func (m *Method) Any(h Matcher) *Method {
	return m.Add(h, "*")
}

func (m *Method) ServeHTTP2(w http.ResponseWriter, r *http.Request) bool {
	if list, found := m.entries[r.Method]; found {
		if list.ServeHTTP2(w, r) {
			return true
		}
	}

	if list, found := m.entries["*"]; found {
		if list.ServeHTTP2(w, r) {
			return true
		}
	}

	return false
}

func (m *Method) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.ServeHTTP2(w, r)
}
