// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"

	"github.com/issue9/mux/v5/middleware"
)

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
	ms      []middleware.Func
}

func (r *Resource) Handle(h http.Handler, methods ...string) *Resource {
	r.router.Handle(r.pattern, middleware.Apply(h, r.ms...), methods...)
	return r
}

func (r *Resource) Get(h http.Handler) *Resource {
	return r.Handle(h, http.MethodGet)
}

func (r *Resource) Post(h http.Handler) *Resource {
	return r.Handle(h, http.MethodPost)
}

func (r *Resource) Delete(h http.Handler) *Resource {
	return r.Handle(h, http.MethodDelete)
}

func (r *Resource) Put(h http.Handler) *Resource {
	return r.Handle(h, http.MethodPut)
}

func (r *Resource) Patch(h http.Handler) *Resource {
	return r.Handle(h, http.MethodPatch)
}

func (r *Resource) Any(h http.Handler) *Resource { return r.Handle(h) }

func (r *Resource) HandleFunc(f http.HandlerFunc, methods ...string) *Resource {
	return r.Handle(f, methods...)
}

func (r *Resource) GetFunc(f http.HandlerFunc) *Resource {
	return r.HandleFunc(f, http.MethodGet)
}

func (r *Resource) PutFunc(f http.HandlerFunc) *Resource {
	return r.HandleFunc(f, http.MethodPut)
}

func (r *Resource) PostFunc(f http.HandlerFunc) *Resource {
	return r.HandleFunc(f, http.MethodPost)
}

func (r *Resource) DeleteFunc(f http.HandlerFunc) *Resource {
	return r.HandleFunc(f, http.MethodDelete)
}

func (r *Resource) PatchFunc(f http.HandlerFunc) *Resource {
	return r.HandleFunc(f, http.MethodPatch)
}

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
func (r *Resource) URL(strict bool, params map[string]string) (string, error) {
	return r.router.URL(strict, r.pattern, params)
}

// Resource 创建一个资源路由项
//
// pattern 资源地址；
// m 中间件函数，按顺序调用；
func (r *Router) Resource(pattern string, m ...middleware.Func) *Resource {
	return &Resource{router: r, pattern: pattern, ms: m}
}

// Resource 创建一个资源路由项
//
// pattern 资源地址；
// m 中间件函数，按顺序调用，会继承 p 的中间件并按在 m 之前；
func (p *Prefix) Resource(pattern string, m ...middleware.Func) *Resource {
	ms := make([]middleware.Func, 0, len(p.ms)+len(m))
	ms = append(ms, p.ms...)
	ms = append(ms, m...)
	return p.router.Resource(p.prefix+pattern, ms...)
}

// Router 返回与当前资源关联的 *Router 实例
func (r *Resource) Router() *Router { return r.router }
