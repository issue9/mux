// SPDX-License-Identifier: MIT

package mux

import "net/http"

// Resource 以资源地址为对象的路由配置
//
//  r, _ := srv.Resource("/api/users/{id}")
//  r.Get(h)  // 相当于 srv.Get("/api/users/{id}")
//  r.Post(h) // 相当于 srv.Post("/api/users/{id}")
//  url := r.URL(map[string]string{"id":5}) // 获得 /api/users/5
type Resource struct {
	mux     *Mux
	pattern string
}

// Options 手动指定 OPTIONS 请求方法的值
//
// 具体说明可参考 Mux.Options 方法。
func (r *Resource) Options(allow string) *Resource {
	r.mux.Options(r.pattern, allow)
	return r
}

// Handle 相当于 Mux.Handle(pattern, h, methods...) 的简易写法
func (r *Resource) Handle(h http.Handler, methods ...string) error {
	return r.mux.Handle(r.pattern, h, methods...)
}

func (r *Resource) handle(h http.Handler, methods ...string) *Resource {
	if err := r.Handle(h, methods...); err != nil {
		panic(err)
	}
	return r
}

// Get 相当于 Mux.Get(pattern, h) 的简易写法
func (r *Resource) Get(h http.Handler) *Resource {
	return r.handle(h, http.MethodGet)
}

// Post 相当于 Mux.Post(pattern, h) 的简易写法
func (r *Resource) Post(h http.Handler) *Resource {
	return r.handle(h, http.MethodPost)
}

// Delete 相当于 Mux.Delete(pattern, h) 的简易写法
func (r *Resource) Delete(h http.Handler) *Resource {
	return r.handle(h, http.MethodDelete)
}

// Put 相当于 Mux.Put(pattern, h) 的简易写法
func (r *Resource) Put(h http.Handler) *Resource {
	return r.handle(h, http.MethodPut)
}

// Patch 相当于 Mux.Patch(pattern, h) 的简易写法
func (r *Resource) Patch(h http.Handler) *Resource {
	return r.handle(h, http.MethodPatch)
}

// Any 相当于 Mux.Any(pattern, h) 的简易写法
func (r *Resource) Any(h http.Handler) *Resource {
	return r.handle(h)
}

// HandleFunc 功能同 Mux.HandleFunc(pattern, fun, ...)
func (r *Resource) HandleFunc(fun http.HandlerFunc, methods ...string) error {
	return r.mux.HandleFunc(r.pattern, fun, methods...)
}

func (r *Resource) handleFunc(fun http.HandlerFunc, methods ...string) *Resource {
	if err := r.HandleFunc(fun, methods...); err != nil {
		panic(err)
	}

	return r
}

// GetFunc 相当于 Mux.GetFunc(pattern, func) 的简易写法
func (r *Resource) GetFunc(fun http.HandlerFunc) *Resource {
	return r.handleFunc(fun, http.MethodGet)
}

// PutFunc 相当于 Mux.PutFunc(pattern, func) 的简易写法
func (r *Resource) PutFunc(fun http.HandlerFunc) *Resource {
	return r.handleFunc(fun, http.MethodPut)
}

// PostFunc 相当于 Mux.PostFunc(pattern, func) 的简易写法
func (r *Resource) PostFunc(fun http.HandlerFunc) *Resource {
	return r.handleFunc(fun, http.MethodPost)
}

// DeleteFunc 相当于 Mux.DeleteFunc(pattern, func) 的简易写法
func (r *Resource) DeleteFunc(fun http.HandlerFunc) *Resource {
	return r.handleFunc(fun, http.MethodDelete)
}

// PatchFunc 相当于 Mux.PatchFunc(pattern, func) 的简易写法
func (r *Resource) PatchFunc(fun http.HandlerFunc) *Resource {
	return r.handleFunc(fun, http.MethodPatch)
}

// AnyFunc 相当于 Mux.AnyFunc(pattern, func) 的简易写法
func (r *Resource) AnyFunc(fun http.HandlerFunc) *Resource {
	return r.handleFunc(fun)
}

// Remove 删除指定匹配模式的路由项
func (r *Resource) Remove(methods ...string) *Resource {
	r.mux.Remove(r.pattern, methods...)
	return r
}

// Clean 清除当前资源的所有路由项
func (r *Resource) Clean() *Resource {
	r.mux.Remove(r.pattern)
	return r
}

// Name 为一条路由项命名
//
// URL 可以通过此属性来生成地址。
func (r *Resource) Name(name string) error {
	return r.mux.Name(name, r.pattern)
}

// URL 根据参数构建一条 URL
//
// params 匹配路由参数中的同名参数，或是不存在路由参数，比如普通的字符串路由项，
// 该参数不启作用；
//  res, := m.Resource("/posts/{id}")
//  res.URL(map[string]string{"id": "1"}, "") // /posts/1
//
//  res, := m.Resource("/posts/{id}/{path}")
//  res.URL(map[string]string{"id": "1","path":"author/profile"}) // /posts/1/author/profile
func (r *Resource) URL(params map[string]string) (string, error) {
	return r.mux.URL(r.pattern, params)
}

// Resource 创建一个资源路由项
func (mux *Mux) Resource(pattern string) *Resource {
	return &Resource{
		mux:     mux,
		pattern: pattern,
	}
}

// Resource 创建一个资源路由项
func (p *Prefix) Resource(pattern string) *Resource {
	return &Resource{
		mux:     p.mux,
		pattern: p.prefix + pattern,
	}
}

// Mux 返回与当前资源关联的 *Mux 实例
func (r *Resource) Mux() *Mux {
	return r.mux
}
