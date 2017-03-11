// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"container/list"
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/issue9/mux/internal/entry"
)

type contextKey int

// ContextKeyParams 表示从 context 中获取的参数列表的关键字。
const ContextKeyParams contextKey = 0

// Params 用以保存请求地址中的参数内容
type Params map[string]string

var (
	// 支持的所有请求方法
	supportedMethods = []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodDelete,
		http.MethodPut,
		http.MethodPatch,
		http.MethodOptions,
		http.MethodHead,
		http.MethodTrace,
	}

	// 调用 Any 添加的列表，默认不添加 OPTIONS 请求
	defaultMethods = []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodDelete,
		http.MethodPut,
		http.MethodPatch,
		// http.MethodOptions,
		http.MethodHead,
		http.MethodTrace,
	}
)

// ServeMux 是 http.ServeMux 的升级版，可处理对 URL 的正则匹配和根据 METHOD 进行过滤。
//
// 用法如下：
//  m := mux.NewServeMux()
//  m.Get("/abc/h1", h1).
//    Post("/abc/h2", h2).
//    Add("/api/{version:\\d+}",h3, http.MethodGet, "POST") // 只匹配 GET 和 POST
//  http.ListenAndServe(m)
//
//
// 路由参数：
//
// 路由参数可通过 r.Context 获取：
//  params := r.Context().Value(mux.ContextKeyParams).(mux.Params)
//
//
// 匹配规则：
//
// 可能会出现多条记录与同一请求都匹配的情况，这种情况下，
// 系统会找到一条认为最匹配的路由来处理，判断规则如下：
//  1. 静态路由优先于正则路由判断；
//  2. 完全匹配的路由项优先于部分匹配的路由项；
//  3. 正则只能是完全匹配；
//  4. 只有以 / 结尾的静态路由才有部分匹配功能；
//  5. 同类的后插入先匹配。
//
// 正则匹配语法：
//  /post/{id}     // 匹配 /post/ 开头的任意字符串，其后的字符串保存到 id 中；
//  /post/{id:\d+} // 同上，但 id 的值只能为 \d+；
//  /post/{:\d+}   // 同上，但是没有命名；
type ServeMux struct {
	// 同时处理 entries 三个的竟争问题
	mu sync.RWMutex

	// 路由项，按请求方法进行分类，键名为请求方法名称，键值为路由项的列表。
	entries *list.List
}

// NewServeMux 声明一个新的 ServeMux
func NewServeMux() *ServeMux {
	ret := &ServeMux{
		entries: list.New(),
	}

	return ret
}

// Clean 清除所有的路由项
func (mux *ServeMux) Clean() *ServeMux {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	mux.entries.Init()

	return mux
}

// Remove 移除指定的路由项，通过路由表达式和 method 来匹配。
// 当未指定 methods 时，将删除所有 method 匹配的项。
// 指定错误的 methods 值，将自动忽略该值。
func (mux *ServeMux) Remove(pattern string, methods ...string) {
	if len(methods) == 0 { // 删除所有 method 下匹配的项
		methods = supportedMethods
	}

	mux.mu.Lock()
	defer mux.mu.Unlock()

	for item := mux.entries.Front(); item != nil; item = item.Next() {
		e := item.Value.(entry.Entry)
		if e.Pattern() != pattern {
			continue
		}

		if empty := e.Remove(methods...); empty { // 该 Entry 下已经没有路由项了
			mux.entries.Remove(item)
		}
		break
	}
}

// Add 添加一条路由数据。
//
// pattern 为路由匹配模式，可以是正则匹配也可以是字符串匹配，
// methods 参数应该只能为 supportedMethods 中的字符串，若不指定，默认为所有，
// 当 h 或是 pattern 为空时，将触发 panic。
func (mux *ServeMux) Add(pattern string, h http.Handler, methods ...string) *ServeMux {
	if len(pattern) == 0 {
		panic("参数pattern不能为空")
	}

	if h == nil {
		panic("参数h不能为空")
	}

	if len(methods) == 0 {
		methods = defaultMethods
	}

	ety := entry.New(pattern, h)
	for _, method := range methods {
		if !MethodIsSupported(method) {
			panic(fmt.Errorf("无效的 methods: %v", method))
		}
		if err := ety.Add(method, h); err != nil {
			panic(err)
		}
	}

	mux.mu.Lock()
	defer mux.mu.Unlock()

	if ety.IsRegexp() { // 正则路由，在后端插入
		mux.entries.PushBack(ety)
	}
	mux.entries.PushFront(ety)

	return mux
}

// Options 手动指定 OPTIONS 请求方法的值。
//
// 若无特殊需求，不用调用此方法，系统会自动计算符合当前路由的请求方法列表。
func (mux *ServeMux) Options(pattern string, methods string) *ServeMux {
	for item := mux.entries.Front(); item != nil; item = item.Next() {
		e := item.Value.(entry.Entry)
		if e.Pattern() != pattern {
			continue
		}

		e.SetAllow(methods)
		break
	}
	return mux
}

// Get 相当于 ServeMux.Add(pattern, h, "GET") 的简易写法
func (mux *ServeMux) Get(pattern string, h http.Handler) *ServeMux {
	return mux.Add(pattern, h, http.MethodGet)
}

// Post 相当于 ServeMux.Add(pattern, h, "POST") 的简易写法
func (mux *ServeMux) Post(pattern string, h http.Handler) *ServeMux {
	return mux.Add(pattern, h, http.MethodPost)
}

// Delete 相当于 ServeMux.Add(pattern, h, "DELETE") 的简易写法
func (mux *ServeMux) Delete(pattern string, h http.Handler) *ServeMux {
	return mux.Add(pattern, h, http.MethodDelete)
}

// Put 相当于 ServeMux.Add(pattern, h, "PUT") 的简易写法
func (mux *ServeMux) Put(pattern string, h http.Handler) *ServeMux {
	return mux.Add(pattern, h, http.MethodPut)
}

// Patch 相当于 ServeMux.Add(pattern, h, "PATCH") 的简易写法
func (mux *ServeMux) Patch(pattern string, h http.Handler) *ServeMux {
	return mux.Add(pattern, h, http.MethodPatch)
}

// Any 相当于 ServeMux.Add(pattern, h) 的简易写法
func (mux *ServeMux) Any(pattern string, h http.Handler) *ServeMux {
	return mux.Add(pattern, h, defaultMethods...)
}

// AddFunc 功能同 ServeMux.Add()，但是将第二个参数从 http.Handler 换成了 func(http.ResponseWriter, *http.Request)
func (mux *ServeMux) AddFunc(pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) *ServeMux {
	return mux.Add(pattern, http.HandlerFunc(fun), methods...)
}

// GetFunc 相当于 ServeMux.AddFunc(pattern, func, "GET") 的简易写法
func (mux *ServeMux) GetFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.AddFunc(pattern, fun, http.MethodGet)
}

// PutFunc 相当于 ServeMux.AddFunc(pattern, func, "PUT") 的简易写法
func (mux *ServeMux) PutFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.AddFunc(pattern, fun, http.MethodPut)
}

// PostFunc 相当于 ServeMux.AddFunc(pattern, func, "POST") 的简易写法
func (mux *ServeMux) PostFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.AddFunc(pattern, fun, http.MethodPost)
}

// DeleteFunc 相当于 ServeMux.AddFunc(pattern, func, "DELETE") 的简易写法
func (mux *ServeMux) DeleteFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.AddFunc(pattern, fun, http.MethodDelete)
}

// PatchFunc 相当于 ServeMux.AddFunc(pattern, func, "PATCH") 的简易写法
func (mux *ServeMux) PatchFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.AddFunc(pattern, fun, http.MethodPatch)
}

// AnyFunc 相当于 ServeMux.AddFunc(pattern, func) 的简易写法
func (mux *ServeMux) AnyFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.AddFunc(pattern, fun, defaultMethods...)
}

// 查找最匹配的路由项
func (mux *ServeMux) match(r *http.Request) (p string, e entry.Entry) {
	size := -1 // 匹配度，0 表示完全匹配，负数表示完全不匹配，其它值越小匹配度越高
	p = cleanPath(r.URL.Path)

	mux.mu.RLock()
	defer mux.mu.RUnlock()

	for item := mux.entries.Front(); item != nil; item = item.Next() {
		ety := item.Value.(entry.Entry)

		s := ety.Match(p)

		if s == 0 { // 完全匹配，可以中止匹配过程
			return p, ety
		}

		if s == -1 || (size > 0 && s >= size) { // 完全不匹配，或是匹配度没有当前的高
			continue
		}

		// 匹配度比当前的高
		size = s
		e = ety
	} // end for

	if size < 0 {
		return "", nil
	}
	return p, e
}

// 若没有找到匹配路由，返回 404
func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p, e := mux.match(r)

	if e == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	h := e.Handler(r.Method)
	if h == nil {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	params := e.Params(p)
	if params != nil {
		ctx := context.WithValue(r.Context(), ContextKeyParams, Params(params))
		r = r.WithContext(ctx)
	}
	h.ServeHTTP(w, r)
}

// 清除路径中的怪异符号
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}

	if p[0] != '/' {
		p = "/" + p
	}

	pp := path.Clean(p)
	if pp == "/" {
		return pp
	}

	// path.Clean 会去掉最后的 / 符号，所以原来有 / 的，需要加回去
	if p[len(p)-1] == '/' {
		pp += "/"
	}
	return pp
}

// MethodIsSupported 是否支持某请求方法
func MethodIsSupported(method string) bool {
	method = strings.ToUpper(method)
	for _, m := range supportedMethods {
		if m == method {
			return true
		}
	}

	return false
}
