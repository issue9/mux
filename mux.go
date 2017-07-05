// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/issue9/mux/internal/tree"
	"github.com/issue9/mux/params"
)

var (
	defaultNotFound = func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	defaultMethodNotAllowed = func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
)

// ErrNameExists 当为一个路由项命名时，若存在相同名称的，则返回此错误信息。
var ErrNameExists = errors.New("存在相同名称的路由项")

// Mux 提供了强大的路由匹配功能，可以对路径按正则或是请求方法进行匹配。
//
// 用法如下：
//  m := mux.New()
//  m.Get("/abc/h1", h1).
//    Post("/abc/h2", h2).
//    Handle("/api/{version:\\d+}",h3, http.MethodGet, http.MethodPost) // 只匹配 GET 和 POST
//  http.ListenAndServe(m)
type Mux struct {
	tree             *tree.Tree
	skipCleanPath    bool
	notFound         http.HandlerFunc
	methodNotAllowed http.HandlerFunc

	names   map[string]string
	namesMu sync.RWMutex
}

// New 声明一个新的 Mux。
//
// disableOptions 是否禁用自动生成 OPTIONS 功能；
// skipCleanPath 是否不对访问路径作处理，比如 "//api" ==> "/api"；
// notFound 404 页面的处理方式，为 nil 时会调用默认的方式进行处理；
// methodNotAllowed 405 页面的处理方式，为 nil 时会调用默认的方式进行处理。
func New(disableOptions, skipCleanPath bool, notFound, methodNotAllowed http.HandlerFunc) *Mux {
	if notFound == nil {
		notFound = defaultNotFound
	}
	if methodNotAllowed == nil {
		methodNotAllowed = defaultMethodNotAllowed
	}

	return &Mux{
		tree:             tree.New(disableOptions),
		names:            make(map[string]string, 50),
		skipCleanPath:    skipCleanPath,
		notFound:         notFound,
		methodNotAllowed: methodNotAllowed,
	}
}

// Clean 清除所有的路由项
func (mux *Mux) Clean() *Mux {
	mux.tree.Clean("")
	return mux
}

// Remove 移除指定的路由项。
//
// 当未指定 methods 时，将删除所有 method 匹配的项。
// 指定错误的 methods 值，将自动忽略该值。
func (mux *Mux) Remove(pattern string, methods ...string) *Mux {
	mux.tree.Remove(pattern, methods...)
	return mux
}

// Handle 添加一条路由数据。
//
// pattern 为路由匹配模式，可以是正则匹配也可以是字符串匹配；
// methods 该路由项对应的请求方法，可通过 SupportedMethods() 获得当前支持的请求方法。
func (mux *Mux) Handle(pattern string, h http.Handler, methods ...string) error {
	return mux.tree.Add(pattern, h, methods...)
}

// Options 将 OPTIONS 请求方法的报头 allow 值固定为指定的值。
//
// 若无特殊需求，不用调用此方法，系统会自动计算符合当前路由的请求方法列表。
// 如果想实现对处理方法的自定义，可以显示地调用 Handle 方法:
//  Mux.Handle("/api/1", handle, http.MethodOptions)
func (mux *Mux) Options(pattern string, allow string) *Mux {
	err := mux.tree.SetAllow(pattern, allow)
	if err != nil {
		panic(err)
	}

	return mux
}

func (mux *Mux) handle(pattern string, h http.Handler, methods ...string) *Mux {
	if err := mux.Handle(pattern, h, methods...); err != nil {
		panic(err)
	}
	return mux
}

// Get 相当于 Mux.Handle(pattern, h, http.MethodGet) 的简易写法
func (mux *Mux) Get(pattern string, h http.Handler) *Mux {
	return mux.handle(pattern, h, http.MethodGet)
}

// Post 相当于 Mux.Handle(pattern, h, http.MethodPost) 的简易写法
func (mux *Mux) Post(pattern string, h http.Handler) *Mux {
	return mux.handle(pattern, h, http.MethodPost)
}

// Delete 相当于 Mux.Handle(pattern, h, http.MethodDelete) 的简易写法
func (mux *Mux) Delete(pattern string, h http.Handler) *Mux {
	return mux.handle(pattern, h, http.MethodDelete)
}

// Put 相当于 Mux.Handle(pattern, h, http.MethodPut) 的简易写法
func (mux *Mux) Put(pattern string, h http.Handler) *Mux {
	return mux.handle(pattern, h, http.MethodPut)
}

// Patch 相当于 Mux.Handle(pattern, h, http.MethodPatch) 的简易写法
func (mux *Mux) Patch(pattern string, h http.Handler) *Mux {
	return mux.handle(pattern, h, http.MethodPatch)
}

// Any 相当于 Mux.Handle(pattern, h) 的简易写法
func (mux *Mux) Any(pattern string, h http.Handler) *Mux {
	return mux.handle(pattern, h)
}

// HandleFunc 功能同 Mux.Handle()，但是将第二个参数从 http.Handler 换成了 http.HandlerFunc
func (mux *Mux) HandleFunc(pattern string, fun http.HandlerFunc, methods ...string) error {
	return mux.Handle(pattern, fun, methods...)
}

func (mux *Mux) handleFunc(pattern string, fun http.HandlerFunc, methods ...string) *Mux {
	return mux.handle(pattern, fun, methods...)
}

// GetFunc 相当于 Mux.HandleFunc(pattern, func, http.MethodGet) 的简易写法
func (mux *Mux) GetFunc(pattern string, fun http.HandlerFunc) *Mux {
	return mux.handleFunc(pattern, fun, http.MethodGet)
}

// PutFunc 相当于 Mux.HandleFunc(pattern, func, http.MethodPut) 的简易写法
func (mux *Mux) PutFunc(pattern string, fun http.HandlerFunc) *Mux {
	return mux.handleFunc(pattern, fun, http.MethodPut)
}

// PostFunc 相当于 Mux.HandleFunc(pattern, func, "POST") 的简易写法
func (mux *Mux) PostFunc(pattern string, fun http.HandlerFunc) *Mux {
	return mux.handleFunc(pattern, fun, http.MethodPost)
}

// DeleteFunc 相当于 Mux.HandleFunc(pattern, func, http.MethodDelete) 的简易写法
func (mux *Mux) DeleteFunc(pattern string, fun http.HandlerFunc) *Mux {
	return mux.handleFunc(pattern, fun, http.MethodDelete)
}

// PatchFunc 相当于 Mux.HandleFunc(pattern, func, http.MethodPatch) 的简易写法
func (mux *Mux) PatchFunc(pattern string, fun http.HandlerFunc) *Mux {
	return mux.handleFunc(pattern, fun, http.MethodPatch)
}

// AnyFunc 相当于 Mux.HandleFunc(pattern, func) 的简易写法
func (mux *Mux) AnyFunc(pattern string, fun http.HandlerFunc) *Mux {
	return mux.handleFunc(pattern, fun)
}

func (mux *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	if !mux.skipCleanPath {
		p = cleanPath(p)
	}

	h, ps := mux.tree.Handler(p, r.Method)
	if ps == nil {
		mux.notFound(w, r)
		return
	}

	if h == nil {
		mux.methodNotAllowed(w, r)
		return
	}

	if len(ps) > 0 {
		ctx := context.WithValue(r.Context(), params.ContextKeyParams, ps)
		r = r.WithContext(ctx)
	}

	h.ServeHTTP(w, r)
}

// Name 为一条路由项命名。
// URL 可以通过此属性来生成地址。
func (mux *Mux) Name(name, pattern string) error {
	mux.namesMu.Lock()
	defer mux.namesMu.Unlock()

	if _, found := mux.names[name]; found {
		return ErrNameExists
	}

	mux.names[name] = pattern
	return nil
}

// URL 根据参数生成地址。
// name 为路由的名称，或是直接为路由项的定义内容；
// params 为路由项中的参数，键名为参数名，键值为参数值。
func (mux *Mux) URL(name string, params map[string]string) (string, error) {
	mux.namesMu.RLock()
	pattern, found := mux.names[name]
	mux.namesMu.RUnlock()

	if !found {
		pattern = name
	}

	return mux.tree.URL(pattern, params)
}

// GetParams 获取路由的参数集合。详细情况可参考 params.Get
func GetParams(r *http.Request) params.Params {
	return params.Get(r)
}

// 清除路径中的重复的 / 字符
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}

	if p[0] != '/' {
		p = "/" + p
	}

	index := strings.Index(p, "//")
	if index == -1 {
		return p
	}

	pp := make([]byte, index+1, len(p))
	copy(pp, p[:index+1])

	slash := true
	for i := index + 2; i < len(p); i++ {
		if p[i] == '/' {
			if slash {
				continue
			}
			slash = true
		} else {
			slash = false
		}
		pp = append(pp, p[i])
	}

	return string(pp)
}
