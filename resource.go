// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"errors"
	"net/http"

	"github.com/issue9/mux/internal/entry"
	"github.com/issue9/mux/internal/method"
)

// ErrResourceNameExists 当为一个资源命名时，若存在相同名称的，
// 则返回此错误信息。
var ErrResourceNameExists = errors.New("存在相同名称的资源")

// Resource 以资源地址为对象的路由配置。
//  r, _ := srv.Resource("/api/users/{id}")
//  r.Get(h)  // 相当于 srv.Get("/api/users/{id}")
//  r.Post(h) // 相当于 srv.Post("/api/users/{id}")
//  url := r.URL(map[string]string{"id":5}) // 获得 /api/users/5
//
//  r.Name("user")             // 以 user 为名保存此实例，方便之后调用
//  srv.Name("user").URL(...)  // 调用名为 user 的 *Resource 实例
type Resource struct {
	mux     *Mux
	pattern string
	entry   entry.Entry
}

// Options 手动指定 OPTIONS 请求方法的值。具体说明可参考 Mux.Options 方法。
func (r *Resource) Options(allow string) *Resource {
	r.mux.Options(r.pattern, allow)
	return r
}

// Add 相当于 Mux.Add(pattern, h, methods...) 的简易写法
func (r *Resource) Add(h http.Handler, methods ...string) error {
	return r.mux.Add(r.pattern, h, methods...)
}

func (r *Resource) add(h http.Handler, methods ...string) *Resource {
	if err := r.mux.Add(r.pattern, h, methods...); err != nil {
		panic(err)
	}
	return r
}

// Get 相当于 Mux.Get(pattern, h) 的简易写法
func (r *Resource) Get(h http.Handler) *Resource {
	return r.add(h, http.MethodGet)
}

// Post 相当于 Mux.Post(pattern, h) 的简易写法
func (r *Resource) Post(h http.Handler) *Resource {
	return r.add(h, http.MethodPost)
}

// Delete 相当于 Mux.Delete(pattern, h) 的简易写法
func (r *Resource) Delete(h http.Handler) *Resource {
	return r.add(h, http.MethodDelete)
}

// Put 相当于 Mux.Put(pattern, h) 的简易写法
func (r *Resource) Put(h http.Handler) *Resource {
	return r.add(h, http.MethodPut)
}

// Patch 相当于 Mux.Patch(pattern, h) 的简易写法
func (r *Resource) Patch(h http.Handler) *Resource {
	return r.add(h, http.MethodPatch)
}

// Any 相当于 Mux.Any(pattern, h) 的简易写法
func (r *Resource) Any(h http.Handler) *Resource {
	return r.add(h, method.Default...)
}

// AddFunc 功能同 Mux.AddFunc(pattern, fun, ...)
func (r *Resource) AddFunc(fun http.HandlerFunc, methods ...string) error {
	return r.mux.AddFunc(r.pattern, fun, methods...)
}

func (r *Resource) addFunc(fun http.HandlerFunc, methods ...string) *Resource {
	if err := r.mux.AddFunc(r.pattern, fun, methods...); err != nil {
		panic(err)
	}

	return r
}

// GetFunc 相当于 Mux.GetFunc(pattern, func) 的简易写法
func (r *Resource) GetFunc(fun http.HandlerFunc) *Resource {
	return r.addFunc(fun, http.MethodGet)
}

// PutFunc 相当于 Mux.PutFunc(pattern, func) 的简易写法
func (r *Resource) PutFunc(fun http.HandlerFunc) *Resource {
	return r.addFunc(fun, http.MethodPut)
}

// PostFunc 相当于 Mux.PostFunc(pattern, func) 的简易写法
func (r *Resource) PostFunc(fun http.HandlerFunc) *Resource {
	return r.addFunc(fun, http.MethodPost)
}

// DeleteFunc 相当于 Mux.DeleteFunc(pattern, func) 的简易写法
func (r *Resource) DeleteFunc(fun http.HandlerFunc) *Resource {
	return r.addFunc(fun, http.MethodDelete)
}

// PatchFunc 相当于 Mux.PatchFunc(pattern, func) 的简易写法
func (r *Resource) PatchFunc(fun http.HandlerFunc) *Resource {
	return r.addFunc(fun, http.MethodPatch)
}

// AnyFunc 相当于 Mux.AnyFunc(pattern, func) 的简易写法
func (r *Resource) AnyFunc(fun http.HandlerFunc) *Resource {
	return r.addFunc(fun, method.Default...)
}

// Remove 删除指定匹配模式的路由项
func (r *Resource) Remove(methods ...string) *Resource {
	r.mux.Remove(r.pattern, methods...)
	return r
}

// Clean 清除当前资源的所有路由项
func (r *Resource) Clean() *Resource {
	r.mux.Remove(r.pattern, method.Supported...)
	r.entry = nil
	return r
}

// Name 给当前资源一个名称，不能与已有名称相同。
// 命名的资源会保存在 Mux 中，之后可以通过过 Mux.Name() 找到该资源。
func (r *Resource) Name(name string) error {
	r.mux.resourcesMu.RLock()
	_, exists := r.mux.resources[name]
	r.mux.resourcesMu.RUnlock()

	if exists {
		return ErrResourceNameExists
	}

	r.mux.resourcesMu.Lock()
	r.mux.resources[name] = r
	r.mux.resourcesMu.Unlock()

	return nil
}

// URL 根据参数构建一条 URL。
//
// params 匹配路由参数中的同名参数，或是不存在路由参数，比如普通的字符串路由项，
// 该参数不启作用；path 仅对通配符路由项启作用。
//  res, := m.Resource("/posts/{id}")
//  res.URL(map[string]string{"id": "1"}, "") // /posts/1
//
//  res, := m.Resource("/posts/{id}/*")
//  res.URL(map[string]string{"id": "1"}, "author/profile") // /posts/1/author/profile
func (r *Resource) URL(params map[string]string, path string) (string, error) {
	return r.entry.URL(params, path)
}

// Name 返回指定名称的 *Resource 实例，如果不存在返回 nil。
func (mux *Mux) Name(name string) *Resource {
	mux.resourcesMu.RLock()
	r := mux.resources[name]
	mux.resourcesMu.RUnlock()

	return r
}

// Resource 创建一个资源路由项。
func (mux *Mux) Resource(pattern string) (*Resource, error) {
	return newResource(mux, pattern)
}

// Resource 创建一个资源路由项。
func (p *Prefix) Resource(pattern string) (*Resource, error) {
	return newResource(p.mux, p.prefix+pattern)
}

// Mux 返回与当前资源关联的 *Mux 实例
func (r *Resource) Mux() *Mux {
	return r.mux
}

func newResource(mux *Mux, pattern string) (*Resource, error) {
	ety, err := mux.list.Entry(pattern)
	if err != nil {
		return nil, err
	}

	return &Resource{
		mux:     mux,
		pattern: pattern,
		entry:   ety,
	}, nil
}
