// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/issue9/mux/internal/entry"
)

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

// MethodIsSupported 是否支持该请求方法
func MethodIsSupported(method string) bool {
	method = strings.ToUpper(method)
	for _, m := range supportedMethods {
		if m == method {
			return true
		}
	}

	return false
}

// ServeMux 是 http.ServeMux 的升级版，可处理对 URL 的正则匹配和根据 METHOD 进行过滤。
//
// 用法如下：
//  m := mux.NewServeMux()
//  m.Get("/abc/h1", h1).
//    Post("/abc/h2", h2).
//    Add("/api/{version:\\d+}",h3, http.MethodGet, "POST") // 只匹配 GET 和 POST
//  http.ListenAndServe(m)
type ServeMux struct {
	mu sync.RWMutex

	// 路由项，按资源进行分类。
	entries *list.List

	// 是否禁用自动产生 OPTIONS 请求方法。
	// 该值不能中途修改，否则会出现部分有 OPTIONS，部分没有的情况。
	disableOptions bool

	// 404 的处理方式
	notFound http.HandlerFunc

	// 405 的处理方式
	methodNotAllowed http.HandlerFunc
}

// NewServeMux 声明一个新的 ServeMux。
//
// disableOptions 是否禁用自动生成 OPTIONS 功能。
// notFound 404 页面的处理方式，为 nil 时会调用 http.Error 进行处理
// methodNotAllowed 405 页面的处理方式，为 nil 时会调用 http.Error 进行处理
func NewServeMux(disableOptions bool, notFound, methodNotAllowed http.HandlerFunc) *ServeMux {
	if notFound == nil {
		notFound = func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
	}

	if methodNotAllowed == nil {
		methodNotAllowed = func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}

	return &ServeMux{
		entries:          list.New(),
		disableOptions:   disableOptions,
		notFound:         notFound,
		methodNotAllowed: methodNotAllowed,
	}
}

// Clean 清除所有的路由项
func (mux *ServeMux) Clean() *ServeMux {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	mux.entries.Init()

	return mux
}

// Remove 移除指定的路由项。
//
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
// methods 参数应该只能为 defaultMethods 中的字符串，若不指定，默认为所有，
// 当 h 或是 pattern 为空时，将触发 panic。
func (mux *ServeMux) Add(pattern string, h http.Handler, methods ...string) error {
	if len(pattern) == 0 {
		return errors.New("参数 pattern 不能为空")
	}

	if h == nil {
		return errors.New("参数 h 不能为空")
	}

	if len(methods) == 0 {
		methods = defaultMethods
	}

	mux.mu.Lock()
	defer mux.mu.Unlock()

	var ety entry.Entry

	// 查找是否存在相同的资源项。
	for item := mux.entries.Front(); item != nil; item = item.Next() {
		e := item.Value.(entry.Entry)
		if e.Pattern() == pattern {
			ety = e
			break
		}
	}

	// 不存在相同的资源项，则声明新的。
	if ety == nil {
		var err error
		ety, err = entry.New(pattern, h)
		if err != nil {
			return err
		}

		if mux.disableOptions { // 禁用 OPTIONS
			ety.Remove(http.MethodOptions)
		}

		if ety.IsRegexp() { // 正则路由，在后端插入
			mux.entries.PushBack(ety)
		} else {
			mux.entries.PushFront(ety)
		}
	}

	// 添加指定请求方法的处理函数
	for _, method := range methods {
		if !MethodIsSupported(method) {
			return fmt.Errorf("无效的 methods: %v", method)
		}
		if err := ety.Add(method, h); err != nil {
			return err
		}
	}

	return nil
}

// Options 手动指定 OPTIONS 请求方法的报头 allow 的值。
//
// 若无特殊需求，不用调用此方法，系统会自动计算符合当前路由的请求方法列表。
// 如果想实现对处理方法的自定义，可以显示地调用 Add 方法:
//  ServeMux.Add("/api/1", handle, http.MethodOptions)
func (mux *ServeMux) Options(pattern string, allow string) *ServeMux {
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

func (mux *ServeMux) add(pattern string, h http.Handler, methods ...string) *ServeMux {
	if err := mux.Add(pattern, h, methods...); err != nil {
		panic(err)
	}
	return mux
}

// Get 相当于 ServeMux.Add(pattern, h, "GET") 的简易写法
func (mux *ServeMux) Get(pattern string, h http.Handler) *ServeMux {
	return mux.add(pattern, h, http.MethodGet)
}

// Post 相当于 ServeMux.Add(pattern, h, "POST") 的简易写法
func (mux *ServeMux) Post(pattern string, h http.Handler) *ServeMux {
	return mux.add(pattern, h, http.MethodPost)
}

// Delete 相当于 ServeMux.Add(pattern, h, "DELETE") 的简易写法
func (mux *ServeMux) Delete(pattern string, h http.Handler) *ServeMux {
	return mux.add(pattern, h, http.MethodDelete)
}

// Put 相当于 ServeMux.Add(pattern, h, "PUT") 的简易写法
func (mux *ServeMux) Put(pattern string, h http.Handler) *ServeMux {
	return mux.add(pattern, h, http.MethodPut)
}

// Patch 相当于 ServeMux.Add(pattern, h, "PATCH") 的简易写法
func (mux *ServeMux) Patch(pattern string, h http.Handler) *ServeMux {
	return mux.add(pattern, h, http.MethodPatch)
}

// Any 相当于 ServeMux.Add(pattern, h) 的简易写法
func (mux *ServeMux) Any(pattern string, h http.Handler) *ServeMux {
	return mux.add(pattern, h, defaultMethods...)
}

// AddFunc 功能同 ServeMux.Add()，但是将第二个参数从 http.Handler 换成了 func(http.ResponseWriter, *http.Request)
func (mux *ServeMux) AddFunc(pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) error {
	return mux.Add(pattern, http.HandlerFunc(fun), methods...)
}

func (mux *ServeMux) addFunc(pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) *ServeMux {
	return mux.add(pattern, http.HandlerFunc(fun), methods...)
}

// GetFunc 相当于 ServeMux.AddFunc(pattern, func, "GET") 的简易写法
func (mux *ServeMux) GetFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.addFunc(pattern, fun, http.MethodGet)
}

// PutFunc 相当于 ServeMux.AddFunc(pattern, func, "PUT") 的简易写法
func (mux *ServeMux) PutFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.addFunc(pattern, fun, http.MethodPut)
}

// PostFunc 相当于 ServeMux.AddFunc(pattern, func, "POST") 的简易写法
func (mux *ServeMux) PostFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.addFunc(pattern, fun, http.MethodPost)
}

// DeleteFunc 相当于 ServeMux.AddFunc(pattern, func, "DELETE") 的简易写法
func (mux *ServeMux) DeleteFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.addFunc(pattern, fun, http.MethodDelete)
}

// PatchFunc 相当于 ServeMux.AddFunc(pattern, func, "PATCH") 的简易写法
func (mux *ServeMux) PatchFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.addFunc(pattern, fun, http.MethodPatch)
}

// AnyFunc 相当于 ServeMux.AddFunc(pattern, func) 的简易写法
func (mux *ServeMux) AnyFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.addFunc(pattern, fun, defaultMethods...)
}

func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	params := e.Params(p)
	if params != nil {
		ctx := context.WithValue(r.Context(), contextKeyParams, Params(params))
		r = r.WithContext(ctx)
	}
	h.ServeHTTP(w, r)
}

// 查找最匹配的路由项
//
// p 为整理后的当前请求路径；
// e 为当前匹配的 entry.Entry 实例。
func (mux *ServeMux) match(r *http.Request) (p string, e entry.Entry) {
	size := -1 // 匹配度，0 表示完全匹配，-1 表示完全不匹配，其它值越小匹配度越高
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
