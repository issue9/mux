// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
)

// Matcher相对于http.Handler多了一个ServeHTTP2()函数，用以弥补
// http.Handler一些功能上的不足。
type Matcher interface {
	http.Handler

	// ServeHTTP2()相较于ServeHTTP()：多了一个返回参数isMatched，用于
	// 确定当前的Handler是否被正确匹配。即当前的请求如果符合既定的条
	// 件，则执行此Handler并返回true，否则返回false。
	ServeHTTP2(w http.ResponseWriter, r *http.Request) (isMatched bool)
}

// MatcherFunc用于将一个符合Matcher.ServeHTTP2()声明格式的函数转换成
// Matcher对象。
type MatcherFunc func(w http.ResponseWriter, r *http.Request) bool

func (m MatcherFunc) ServeHTTP2(w http.ResponseWriter, r *http.Request) bool {
	return m(w, r)
}

func (m MatcherFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m(w, r)
}

// HandlerFunc用于将一个符合http.HandlerFunc声明格式的函数转换成Matcher对象。
// 该类型的对象，Matcher.ServeHTTP2()接口的返回值永远都是true。
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// 返回值永远为true
func (m HandlerFunc) ServeHTTP2(w http.ResponseWriter, r *http.Request) bool {
	m(w, r)
	return true
}

func (m HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m(w, r)
}

// 用于将一个http.Handler对象转换成Matcher对象。
type matche struct {
	h http.Handler
}

var _ Matcher = &matche{}

func (m *matche) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.h.ServeHTTP(w, r)
}

func (m *matche) ServeHTTP2(w http.ResponseWriter, r *http.Request) bool {
	m.h.ServeHTTP(w, r)
	return true
}

// 将http.Handler转换成Matcher
// 该类型的对象，Matcher.ServeHTTP2()接口的返回值永远都是true。
func Handler2Matcher(h http.Handler) *matche {
	return &matche{h: h}
}
