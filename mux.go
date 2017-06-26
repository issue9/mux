// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/issue9/mux/internal/tree"
	ps "github.com/issue9/mux/params"
)

var (
	defaultNotFound = func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	defaultMethodNotAllowed = func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
)

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

	resources   map[string]*Resource
	resourcesMu sync.RWMutex
}

// New 声明一个新的 Mux。
//
// disableOptions 是否禁用自动生成 OPTIONS 功能；
// skipCleanPath 是否不对对访问路径作处理，比如 "//api" ==> "/api"；
// notFound 404 页面的处理方式，为 nil 时会调用 defaultNotFound 进行处理；
// methodNotAllowed 405 页面的处理方式，为 nil 时会调用 defaultMethodNotAllowed 进行处理。
func New(disableOptions, skipCleanPath bool, notFound, methodNotAllowed http.HandlerFunc) *Mux {
	if notFound == nil {
		notFound = defaultNotFound
	}
	if methodNotAllowed == nil {
		methodNotAllowed = defaultMethodNotAllowed
	}

	return &Mux{
		tree:             tree.New(),
		resources:        make(map[string]*Resource, 500),
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

	node := mux.tree.Match(p)
	if node == nil {
		mux.notFound(w, r)
		return
	}

	h := node.Handler(r.Method)
	if h == nil {
		mux.methodNotAllowed(w, r)
		return
	}

	if params := node.Params(p); len(params) > 0 {
		ctx := context.WithValue(r.Context(), ps.ContextKeyParams, ps.Params(params))
		r = r.WithContext(ctx)
	}

	h.ServeHTTP(w, r)
}

// GetParams 获取路由的参数集合
func GetParams(r *http.Request) ps.Params {
	return ps.Get(r)
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

	pp := make([]byte, 0, len(p))
	pp = append(pp, p[:index+1]...)
	slash := true

	p = p[index+2:]
	for i := 0; i < len(p); i++ {
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
