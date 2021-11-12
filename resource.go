// SPDX-License-Identifier: MIT

package mux

import "net/http"

// Resource 以资源地址为对象的路由
//
//  srv := NewRouter("")
//  r, _ := srv.Resource("/api/users/{id}")
//  r.Get(h)  // 相当于 srv.Get("/api/users/{id}")
//  r.Post(h) // 相当于 srv.Post("/api/users/{id}")
//  url := r.URL(map[string]string{"id":5}) // 获得 /api/users/5
type Resource struct {
	router  *Router
	pattern string
}

// Handle 相当于 Router.Handle(pattern, h, methods...) 的简易写法
func (r *Resource) Handle(h http.Handler, methods ...string) *Resource {
	r.router.Handle(r.pattern, h, methods...)
	return r
}

// Get 相当于 Router.Get(pattern, h) 的简易写法
func (r *Resource) Get(h http.Handler) *Resource {
	return r.Handle(h, http.MethodGet)
}

// Post 相当于 Router.Post(pattern, h) 的简易写法
func (r *Resource) Post(h http.Handler) *Resource {
	return r.Handle(h, http.MethodPost)
}

// Delete 相当于 Router.Delete(pattern, h) 的简易写法
func (r *Resource) Delete(h http.Handler) *Resource {
	return r.Handle(h, http.MethodDelete)
}

// Put 相当于 Router.Put(pattern, h) 的简易写法
func (r *Resource) Put(h http.Handler) *Resource {
	return r.Handle(h, http.MethodPut)
}

// Patch 相当于 Router.Patch(pattern, h) 的简易写法
func (r *Resource) Patch(h http.Handler) *Resource {
	return r.Handle(h, http.MethodPatch)
}

// Any 相当于 Router.Any(pattern, h) 的简易写法
func (r *Resource) Any(h http.Handler) *Resource { return r.Handle(h) }

// HandleFunc 功能同 Router.HandleFunc(pattern, fun, ...)
func (r *Resource) HandleFunc(f http.HandlerFunc, methods ...string) *Resource {
	return r.Handle(f, methods...)
}

// GetFunc 相当于 Router.GetFunc(pattern, func) 的简易写法
func (r *Resource) GetFunc(f http.HandlerFunc) *Resource {
	return r.HandleFunc(f, http.MethodGet)
}

// PutFunc 相当于 Router.PutFunc(pattern, func) 的简易写法
func (r *Resource) PutFunc(f http.HandlerFunc) *Resource {
	return r.HandleFunc(f, http.MethodPut)
}

// PostFunc 相当于 Router.PostFunc(pattern, func) 的简易写法
func (r *Resource) PostFunc(f http.HandlerFunc) *Resource {
	return r.HandleFunc(f, http.MethodPost)
}

// DeleteFunc 相当于 Router.DeleteFunc(pattern, func) 的简易写法
func (r *Resource) DeleteFunc(f http.HandlerFunc) *Resource {
	return r.HandleFunc(f, http.MethodDelete)
}

// PatchFunc 相当于 Router.PatchFunc(pattern, func) 的简易写法
func (r *Resource) PatchFunc(f http.HandlerFunc) *Resource {
	return r.HandleFunc(f, http.MethodPatch)
}

// AnyFunc 相当于 Router.AnyFunc(pattern, func) 的简易写法
func (r *Resource) AnyFunc(f http.HandlerFunc) *Resource { return r.HandleFunc(f) }

// Remove 删除指定匹配模式的路由项
func (r *Resource) Remove(methods ...string) { r.router.Remove(r.pattern, methods...) }

// Clean 清除当前资源的所有路由项
func (r *Resource) Clean() { r.router.Remove(r.pattern) }

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
	return r.router.URL(r.pattern, params)
}

// Resource 创建一个资源路由项
func (r *Router) Resource(pattern string) *Resource {
	return &Resource{router: r, pattern: pattern}
}

// Resource 创建一个资源路由项
func (p *Prefix) Resource(pattern string) *Resource {
	return p.router.Resource(p.prefix + pattern)
}

// Router 返回与当前资源关联的 *Router 实例
func (r *Resource) Router() *Router { return r.router }
