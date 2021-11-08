// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"strings"

	"github.com/issue9/mux/v5/internal/options"
	"github.com/issue9/mux/v5/internal/syntax"
	"github.com/issue9/mux/v5/internal/tree"
)

// Router 路由
//
// 可以对路径按正则或是请求方法进行匹配。用法如下：
//  router := NewRouter("")
//  router.Get("/abc/h1", h1).
//      Post("/abc/h2", h2).
//      Handle("/api/{version:\\d+}",h3, http.MethodGet, http.MethodPost) // 只匹配 GET 和 POST
//  http.ListenAndServe(router)
//
// 如果需要同时对多个 Router 实例进行路由，可以采用  group.Group 对象管理多个 Router 实例。
type Router struct {
	tree    *tree.Tree
	ms      *Middlewares
	name    string
	options *options.Options
}

// NewRouter 声明路由
//
// name string 路由名称，可以为空；
//
// o 修改路由的默认形为。比如 CaseInsensitive 会让路由忽略大小写，
// 相同类型的函数会相互覆盖，比如 CORS 和 AllowedCORS，后传递会覆盖前传递的值。
func NewRouter(name string, o ...Option) *Router {
	opt, err := options.Build(o...)
	if err != nil {
		panic(err)
	}

	r := &Router{
		tree:    tree.New(opt.Lock, opt.Interceptors),
		name:    name,
		options: opt,
	}
	r.ms = NewMiddlewares(http.HandlerFunc(r.serveHTTP))

	return r
}

// Clean 清除当前路由组的所有路由项
func (r *Router) Clean() { r.tree.Clean("") }

// Routes 返回当前路由组的路由项
//
// 键名为请求地址，键值为对应的请求方法。
func (r *Router) Routes() map[string][]string { return r.tree.Routes() }

// Remove 移除指定的路由项
//
// 当未指定 methods 时，将删除所有 method 匹配的项。
// 指定错误的 methods 值，将自动忽略该值。
func (r *Router) Remove(pattern string, methods ...string) {
	r.tree.Remove(pattern, methods...)
}

// Handle 添加一条路由数据
//
// pattern 为路由匹配模式，可以是正则匹配也可以是字符串匹配，
// 若语法不正确，则直接 panic，可以通过 CheckSyntax 检测语法的有效性，其它接口也相同。
// methods 该路由项对应的请求方法，如果未指定值，则表示所有支持的请求方法，
// 但不包含 OPTIONS 和 HEAD。
func (r *Router) Handle(pattern string, h http.Handler, methods ...string) *Router {
	if err := r.tree.Add(pattern, h, methods...); err != nil {
		panic(err)
	}
	return r
}

// Get 相当于 Router.Handle(pattern, h, http.MethodGet) 的简易写法
//
// h 不应该主动调用 WriteHeader，否则会导致 HEAD 请求获取不到 Content-Length 报头。
func (r *Router) Get(pattern string, h http.Handler) *Router {
	return r.Handle(pattern, h, http.MethodGet)
}

// Post 相当于 Router.Handle(pattern, h, http.MethodPost) 的简易写法
func (r *Router) Post(pattern string, h http.Handler) *Router {
	return r.Handle(pattern, h, http.MethodPost)
}

// Delete 相当于 Router.Handle(pattern, h, http.MethodDelete) 的简易写法
func (r *Router) Delete(pattern string, h http.Handler) *Router {
	return r.Handle(pattern, h, http.MethodDelete)
}

// Put 相当于 Router.Handle(pattern, h, http.MethodPut) 的简易写法
func (r *Router) Put(pattern string, h http.Handler) *Router {
	return r.Handle(pattern, h, http.MethodPut)
}

// Patch 相当于 Router.Handle(pattern, h, http.MethodPatch) 的简易写法
func (r *Router) Patch(pattern string, h http.Handler) *Router {
	return r.Handle(pattern, h, http.MethodPatch)
}

// Any 相当于 Router.Handle(pattern, h) 的简易写法
func (r *Router) Any(pattern string, h http.Handler) *Router {
	return r.Handle(pattern, h)
}

// HandleFunc 功能同 Router.Handle()，但是将第二个参数从 http.Handler 换成了 http.HandlerFunc
func (r *Router) HandleFunc(pattern string, f http.HandlerFunc, methods ...string) *Router {
	return r.Handle(pattern, f, methods...)
}

// GetFunc 相当于 Router.HandleFunc(pattern, func, http.MethodGet) 的简易写法
func (r *Router) GetFunc(pattern string, f http.HandlerFunc) *Router {
	return r.HandleFunc(pattern, f, http.MethodGet)
}

// PutFunc 相当于 Router.HandleFunc(pattern, func, http.MethodPut) 的简易写法
func (r *Router) PutFunc(pattern string, f http.HandlerFunc) *Router {
	return r.HandleFunc(pattern, f, http.MethodPut)
}

// PostFunc 相当于 Router.HandleFunc(pattern, func, "POST") 的简易写法
func (r *Router) PostFunc(pattern string, f http.HandlerFunc) *Router {
	return r.HandleFunc(pattern, f, http.MethodPost)
}

// DeleteFunc 相当于 Router.HandleFunc(pattern, func, http.MethodDelete) 的简易写法
func (r *Router) DeleteFunc(pattern string, f http.HandlerFunc) *Router {
	return r.HandleFunc(pattern, f, http.MethodDelete)
}

// PatchFunc 相当于 Router.HandleFunc(pattern, func, http.MethodPatch) 的简易写法
func (r *Router) PatchFunc(pattern string, f http.HandlerFunc) *Router {
	return r.HandleFunc(pattern, f, http.MethodPatch)
}

// AnyFunc 相当于 Router.HandleFunc(pattern, func) 的简易写法
func (r *Router) AnyFunc(pattern string, f http.HandlerFunc) *Router {
	return r.HandleFunc(pattern, f)
}

// URL 根据参数生成地址
//
// pattern 为路由项的定义内容；
// params 为路由项中的参数，键名为参数名，键值为参数值。
func (r *Router) URL(pattern string, params map[string]string) (string, error) {
	return r.tree.URL(pattern, params)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.ms.ServeHTTP(w, req)
}

func (r *Router) serveHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if r.options.CaseInsensitive {
		path = strings.ToLower(req.URL.Path)
	}

	node, ps := r.tree.Route(path)
	if node == nil {
		r.options.NotFound.ServeHTTP(w, req)
		return
	}

	if h := node.Handler(req.Method); h != nil {
		r.options.HandleCORS(node, w, req) // 处理跨域问题
		h.ServeHTTP(w, syntax.WithValue(req, ps))
		ps.Destroy()
		return
	}

	// 存在节点，但是不允许当前请求方法。
	w.Header().Set("Allow", node.Options())
	r.options.MethodNotAllowed.ServeHTTP(w, req)
}

// Name 路由名称
func (r *Router) Name() string { return r.name }
