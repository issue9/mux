// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"container/list"
	"context"
	"errors"
	"net/http"
	"path"
	"sync"

	"github.com/issue9/mux/internal/entry"
	"github.com/issue9/mux/internal/method"
)

// 两个默认处理函数
var (
	defaultNotFound = func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	defaultMethodNotAllowed = func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
)

// Mux 提供了强大的路由匹配功能，可以处理正则路径和按请求方法进行匹配。
//
// 用法如下：
//  m := mux.New()
//  m.Get("/abc/h1", h1).
//    Post("/abc/h2", h2).
//    Add("/api/{version:\\d+}",h3, http.MethodGet, http.MethodPost) // 只匹配 GET 和 POST
//  http.ListenAndServe(m)
type Mux struct {
	mu sync.RWMutex

	// 路由项，按资源进行分类。
	entries *list.List

	// 是否禁用自动产生 OPTIONS 请求方法。
	// 该值不能中途修改，否则会出现部分有 OPTIONS，部分没有的情况。
	disableOptions bool

	// 是否不对提交的路径作处理。
	skipCleanPath bool

	// 404 的处理方式
	notFound http.HandlerFunc

	// 405 的处理方式
	methodNotAllowed http.HandlerFunc
}

// New 声明一个新的 Mux。
//
// disableOptions 是否禁用自动生成 OPTIONS 功能。
// skipCleanPath 是否忽略对访问路径作处理，比如 "//api" ==> "/api"
// notFound 404 页面的处理方式，为 nil 时会调用 http.Error 进行处理
// methodNotAllowed 405 页面的处理方式，为 nil 时会调用 http.Error 进行处理
func New(disableOptions, skipCleanPath bool, notFound, methodNotAllowed http.HandlerFunc) *Mux {
	if notFound == nil {
		notFound = defaultNotFound
	}
	if methodNotAllowed == nil {
		methodNotAllowed = defaultMethodNotAllowed
	}

	return &Mux{
		entries:          list.New(),
		disableOptions:   disableOptions,
		skipCleanPath:    skipCleanPath,
		notFound:         notFound,
		methodNotAllowed: methodNotAllowed,
	}
}

// Clean 清除所有的路由项
func (mux *Mux) Clean() *Mux {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	mux.entries.Init()

	return mux
}

// Remove 移除指定的路由项。
//
// 当未指定 methods 时，将删除所有 method 匹配的项。
// 指定错误的 methods 值，将自动忽略该值。
func (mux *Mux) Remove(pattern string, methods ...string) {
	if len(methods) == 0 { // 删除所有 method 下匹配的项
		methods = method.Supported
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
		return // 只可能有一相完全匹配，找到之后，即可返回
	}
}

// Add 添加一条路由数据。
//
// pattern 为路由匹配模式，可以是正则匹配也可以是字符串匹配，
// methods 参数应该只能为 method.Default 中的字符串，若不指定，默认为所有，
// 当 h 或是 pattern 为空时，将触发 panic。
func (mux *Mux) Add(pattern string, h http.Handler, methods ...string) error {
	if len(pattern) == 0 {
		return errors.New("参数 pattern 不能为空")
	}
	if h == nil {
		return errors.New("参数 h 不能为空")
	}

	ety := mux.findEntry(pattern)
	if ety == nil { // 不存在相同的资源项，则声明新的。
		var err error
		if ety, err = entry.New(pattern, h); err != nil {
			return err
		}

		if mux.disableOptions { // 禁用 OPTIONS
			ety.Remove(http.MethodOptions)
		}

		mux.mu.Lock()
		defer mux.mu.Unlock()
		if ety.Type() == entry.TypeRegexp { // 正则路由，在后端插入
			mux.entries.PushBack(ety)
		} else {
			mux.entries.PushFront(ety)
		}
	}

	return ety.Add(h, methods...)
}

func (mux *Mux) findEntry(pattern string) entry.Entry {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	for item := mux.entries.Front(); item != nil; item = item.Next() {
		e := item.Value.(entry.Entry)
		if e.Pattern() == pattern {
			return e
		}
	}

	return nil
}

// Options 手动指定 OPTIONS 请求方法的报头 allow 的值。
//
// 若无特殊需求，不用调用此方法，系统会自动计算符合当前路由的请求方法列表。
// 如果想实现对处理方法的自定义，可以显示地调用 Add 方法:
//  Mux.Add("/api/1", handle, http.MethodOptions)
func (mux *Mux) Options(pattern string, allow string) *Mux {
	for item := mux.entries.Front(); item != nil; item = item.Next() {
		e := item.Value.(entry.Entry)
		if e.Pattern() != pattern {
			continue
		}

		e.SetAllow(allow)
		break
	}
	return mux
}

func (mux *Mux) add(pattern string, h http.Handler, methods ...string) *Mux {
	if err := mux.Add(pattern, h, methods...); err != nil {
		panic(err)
	}
	return mux
}

// Get 相当于 Mux.Add(pattern, h, "GET") 的简易写法
func (mux *Mux) Get(pattern string, h http.Handler) *Mux {
	return mux.add(pattern, h, http.MethodGet)
}

// Post 相当于 Mux.Add(pattern, h, "POST") 的简易写法
func (mux *Mux) Post(pattern string, h http.Handler) *Mux {
	return mux.add(pattern, h, http.MethodPost)
}

// Delete 相当于 Mux.Add(pattern, h, "DELETE") 的简易写法
func (mux *Mux) Delete(pattern string, h http.Handler) *Mux {
	return mux.add(pattern, h, http.MethodDelete)
}

// Put 相当于 Mux.Add(pattern, h, "PUT") 的简易写法
func (mux *Mux) Put(pattern string, h http.Handler) *Mux {
	return mux.add(pattern, h, http.MethodPut)
}

// Patch 相当于 Mux.Add(pattern, h, "PATCH") 的简易写法
func (mux *Mux) Patch(pattern string, h http.Handler) *Mux {
	return mux.add(pattern, h, http.MethodPatch)
}

// Any 相当于 Mux.Add(pattern, h) 的简易写法
func (mux *Mux) Any(pattern string, h http.Handler) *Mux {
	return mux.add(pattern, h, method.Default...)
}

// AddFunc 功能同 Mux.Add()，但是将第二个参数从 http.Handler 换成了 func(http.ResponseWriter, *http.Request)
func (mux *Mux) AddFunc(pattern string, fun http.HandlerFunc, methods ...string) error {
	return mux.Add(pattern, http.HandlerFunc(fun), methods...)
}

func (mux *Mux) addFunc(pattern string, fun http.HandlerFunc, methods ...string) *Mux {
	return mux.add(pattern, http.HandlerFunc(fun), methods...)
}

// GetFunc 相当于 Mux.AddFunc(pattern, func, "GET") 的简易写法
func (mux *Mux) GetFunc(pattern string, fun http.HandlerFunc) *Mux {
	return mux.addFunc(pattern, fun, http.MethodGet)
}

// PutFunc 相当于 Mux.AddFunc(pattern, func, "PUT") 的简易写法
func (mux *Mux) PutFunc(pattern string, fun http.HandlerFunc) *Mux {
	return mux.addFunc(pattern, fun, http.MethodPut)
}

// PostFunc 相当于 Mux.AddFunc(pattern, func, "POST") 的简易写法
func (mux *Mux) PostFunc(pattern string, fun http.HandlerFunc) *Mux {
	return mux.addFunc(pattern, fun, http.MethodPost)
}

// DeleteFunc 相当于 Mux.AddFunc(pattern, func, "DELETE") 的简易写法
func (mux *Mux) DeleteFunc(pattern string, fun http.HandlerFunc) *Mux {
	return mux.addFunc(pattern, fun, http.MethodDelete)
}

// PatchFunc 相当于 Mux.AddFunc(pattern, func, "PATCH") 的简易写法
func (mux *Mux) PatchFunc(pattern string, fun http.HandlerFunc) *Mux {
	return mux.addFunc(pattern, fun, http.MethodPatch)
}

// AnyFunc 相当于 Mux.AddFunc(pattern, func) 的简易写法
func (mux *Mux) AnyFunc(pattern string, fun http.HandlerFunc) *Mux {
	return mux.addFunc(pattern, fun, method.Default...)
}

func (mux *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p, e := mux.match(r)

	if e == nil {
		mux.notFound(w, r)
		return
	}

	h := e.Handler(r.Method)
	if h == nil {
		mux.methodNotAllowed(w, r)
		return
	}

	if params := e.Params(p); params != nil {
		ctx := context.WithValue(r.Context(), contextKeyParams, Params(params))
		r = r.WithContext(ctx)
	}
	h.ServeHTTP(w, r)
}

// 查找最匹配的路由项
//
// p 为整理后的当前请求路径；
// e 为当前匹配的 entry.Entry 实例。
func (mux *Mux) match(r *http.Request) (p string, e entry.Entry) {
	size := -1 // 匹配度，0 表示完全匹配，-1 表示完全不匹配，其它值越小匹配度越高
	if mux.skipCleanPath {
		p = r.URL.Path
	} else {
		p = cleanPath(r.URL.Path)
	}

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

		// 匹配度比当前的高，则保存下来
		size = s
		e = ety
	} // end for

	if size < 0 {
		return "", nil
	}
	return p, e
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

// MethodIsSuppotred 检测请求方法当前包是否支持
func MethodIsSuppotred(m string) bool {
	return method.IsSupported(m)
}

// SupportedMethods 返回所有支持的请求方法
func SupportedMethods() []string {
	return method.Supported
}
