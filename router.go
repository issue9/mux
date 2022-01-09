// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"strings"

	"github.com/issue9/errwrap"

	"github.com/issue9/mux/v6/internal/options"
	"github.com/issue9/mux/v6/internal/syntax"
	"github.com/issue9/mux/v6/internal/tree"
)

// RouterOf 路由
//
// 可以对路径按正则或是请求方法进行匹配。用法如下：
//  router := NewRouterOf("")
//  router.Get("/abc/h1", h1).
//      Post("/abc/h2", h2).
//      Handle("/api/{version:\\d+}",h3, http.MethodGet, http.MethodPost) // 只匹配 GET 和 POST
//  http.ListenAndServe(router)
//
// 如果需要同时对多个 RouterOf 实例进行路由，可以采用  group.GroupOf 对象管理多个 RouterOf 实例。
type RouterOf[T any] struct {
	name    string
	tree    *tree.Tree
	options *options.Options
	ms      []MiddlewareFuncOf[T]
	call    CallOf[T]
}

// CallOf 指定如何调用用户自定义的对象 T
type CallOf[T any] func(http.ResponseWriter, *http.Request, Params, T)

// NewRouterOf 声明路由
//
// name string 路由名称，可以为空；
//
// ms 表示中间件列表，可以为空；
//
// o 修改路由的默认形为。比如 CaseInsensitive 会让路由忽略大小写，
// 相同类型的函数会相互覆盖，比如 CORS 和 AllowedCORS，后传递会覆盖前传递的值。
//
// T 表示用户用于处理路由项的方法，该类型最终通过 NewRouterOf 中的 call 参数与
// http.ResponseWriter 和 *http.Request 相关联。
func NewRouterOf[T any](name string, call CallOf[T], ms []MiddlewareFuncOf[T], o ...Option) *RouterOf[T] {
	opt, err := options.Build(o...)
	if err != nil {
		panic(err)
	}

	r := &RouterOf[T]{
		name:    name,
		tree:    tree.New(opt.Lock, opt.Interceptors),
		options: opt,
		ms:      ms,
		call:    call,
	}

	return r
}

// Clean 清除当前路由组的所有路由项
func (r *RouterOf[T]) Clean() { r.tree.Clean("") }

// Routes 返回当前路由组的路由项
//
// 键名为请求地址，键值为对应的请求方法。
func (r *RouterOf[T]) Routes() map[string][]string { return r.tree.Routes() }

// Remove 移除指定的路由项
//
// 当未指定 methods 时，将删除所有 method 匹配的项。
// 指定错误的 methods 值，将自动忽略该值。
func (r *RouterOf[T]) Remove(pattern string, methods ...string) {
	r.tree.Remove(pattern, methods...)
}

// Handle 添加一条路由数据
//
// pattern 为路由匹配模式，可以是正则匹配也可以是字符串匹配，
// 若语法不正确，则直接 panic，可以通过 CheckSyntax 检测语法的有效性，其它接口也相同。
// methods 该路由项对应的请求方法，如果未指定值，则表示所有支持的请求方法，
// 但不包含 OPTIONS 和 HEAD。
func (r *RouterOf[T]) Handle(pattern string, h T, methods ...string) *RouterOf[T] {
	r.handle(pattern, applyMiddlewares(h, r.ms...), methods...)
	return r
}

func (r *RouterOf[T]) handle(pattern string, h T, methods ...string) {
	f := func(w http.ResponseWriter, req *http.Request, ps Params) {
		r.call(w, req, ps, h)
	}
	if err := r.tree.Add(pattern, f, methods...); err != nil {
		panic(err)
	}
}

// Get 相当于 RouterOf.Handle(pattern, h, http.MethodGet) 的简易写法
//
// h 不应该主动调用 WriteHeader，否则会导致 HEAD 请求获取不到 Content-Length 报头。
func (r *RouterOf[T]) Get(pattern string, h T) *RouterOf[T] {
	return r.Handle(pattern, h, http.MethodGet)
}

func (r *RouterOf[T]) Post(pattern string, h T) *RouterOf[T] {
	return r.Handle(pattern, h, http.MethodPost)
}

func (r *RouterOf[T]) Delete(pattern string, h T) *RouterOf[T] {
	return r.Handle(pattern, h, http.MethodDelete)
}

func (r *RouterOf[T]) Put(pattern string, h T) *RouterOf[T] {
	return r.Handle(pattern, h, http.MethodPut)
}

func (r *RouterOf[T]) Patch(pattern string, h T) *RouterOf[T] {
	return r.Handle(pattern, h, http.MethodPatch)
}

func (r *RouterOf[T]) Any(pattern string, h T) *RouterOf[T] {
	return r.Handle(pattern, h)
}

// URL 根据参数生成地址
//
// strict 是否检查路由是否真实存在以及参数是否符合要求；
// pattern 为路由项的定义内容；
// params 为路由项中的参数，键名为参数名，键值为参数值。
func (r *RouterOf[T]) URL(strict bool, pattern string, params map[string]string) (string, error) {
	buf := errwrap.StringBuilder{}
	buf.Grow(len(r.options.URLDomain) + len(pattern))

	if r.options.URLDomain != "" {
		buf.WString(r.options.URLDomain)
	}

	switch {
	case len(pattern) == 0: // 无需要处理
	case len(params) == 0:
		buf.WString(pattern)
	case strict:
		if err := r.tree.URL(&buf, pattern, params); err != nil {
			return "", err
		}
	default:
		if err := emptyInterceptors.URL(&buf, pattern, params); err != nil {
			return "", err
		}
	}

	return buf.String(), buf.Err
}

func (r *RouterOf[T]) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Serve(w, req, nil)
}

func (r *RouterOf[T]) Serve(w http.ResponseWriter, req *http.Request, ps Params) {
	if r.options.RecoverFunc != nil {
		defer func() {
			if err := recover(); err != nil {
				r.options.RecoverFunc(w, err)
			}
		}()
	}

	r.serveHTTP(w, req, ps)
}

func (r *RouterOf[T]) serveHTTP(w http.ResponseWriter, req *http.Request, p Params) {
	path := req.URL.Path
	if r.options.CaseInsensitive {
		path = strings.ToLower(req.URL.Path)
	}

	node, ps := r.tree.Route(path)
	if node == nil {
		r.options.NotFound.ServeHTTP(w, req)
		return
	}

	if p != nil && p.Count() > 0 {
		if ps.Count() == 0 {
			ps = p
		} else {
			p.Range(func(k, v string) {
				ps.Set(k, v)
			})
		}
	}

	if ps != nil {
		defer ps.(*syntax.Params).Destroy()
	}

	if h := node.Handler(req.Method); h != nil {
		r.options.HandleCORS(node, w, req) // 处理跨域问题
		h(w, req, ps)
		return
	}

	// 存在节点，但是不允许当前请求方法。
	w.Header().Set("Allow", node.Options())
	r.options.MethodNotAllowed.ServeHTTP(w, req)
}

// Name 路由名称
func (r *RouterOf[T]) Name() string { return r.name }
