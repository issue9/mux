// SPDX-License-Identifier: MIT

package mux

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/issue9/mux/v3/internal/handlers"
	"github.com/issue9/mux/v3/internal/syntax"
	"github.com/issue9/mux/v3/internal/tree"
	"github.com/issue9/mux/v3/params"
)

var (
	defaultNotFound = func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	defaultMethodNotAllowed = func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
)

// ErrNameExists 存在相同名称
//
// 当为一个路由项命名时，若存在相同名称的，则返回此错误信息。
var ErrNameExists = errors.New("存在相同名称的路由项")

// Mux 提供了强大的路由匹配功能
//
// 可以对路径按正则或是请求方法进行匹配。用法如下：
//  m := mux.New()
//  m.Get("/abc/h1", h1).
//    Post("/abc/h2", h2).
//    Handle("/api/{version:\\d+}",h3, http.MethodGet, http.MethodPost) // 只匹配 GET 和 POST
//  http.ListenAndServe(m)
type Mux struct {
	name             string     // 当前路由的名称
	routers          []*Mux     // 子路由
	matcher          Matcher    // 当前路由的先决条件
	tree             *tree.Tree // 当前路由的路由项
	notFound         http.HandlerFunc
	methodNotAllowed http.HandlerFunc

	disableOptions, disableHead, skipCleanPath bool

	// names 保存着路由项与其名称的对应关系，默认情况下，
	// 路由项不存在名称，但可以通过 Mux.Name() 为其指定一个名称，
	// 之后即可以在 Mux.URL() 使用名称来查找路由项。
	names   map[string]string
	namesMu sync.RWMutex
}

// Router 用于描述 Mux.All 返回的参数
type Router struct {
	Name   string
	Routes map[string][]string
}

// Default New 的默主人参数版本要
func Default() *Mux {
	return New(false, false, false, nil, nil)
}

// New 声明一个新的 Mux
//
// disableOptions 是否禁用自动生成 OPTIONS 功能；
// disableHead 是否禁用根据 Get 请求自动生成 HEAD 请求；
// skipCleanPath 是否不对访问路径作处理，比如 "//api" ==> "/api"；
// notFound 404 页面的处理方式，为 nil 时会调用默认的方式进行处理；
// methodNotAllowed 405 页面的处理方式，为 nil 时会调用默认的方式进行处理，
// 调用此方法前，会设置 Allow 报头，如果不需要，则要在 methodNotAllowed 中去掉。
func New(disableOptions, disableHead, skipCleanPath bool, notFound, methodNotAllowed http.HandlerFunc) *Mux {
	if notFound == nil {
		notFound = defaultNotFound
	}
	if methodNotAllowed == nil {
		methodNotAllowed = defaultMethodNotAllowed
	}

	mux := &Mux{
		tree: tree.New(disableOptions, disableHead),

		disableOptions: disableOptions,
		disableHead:    disableHead,
		skipCleanPath:  skipCleanPath,

		names:            make(map[string]string, 50),
		notFound:         notFound,
		methodNotAllowed: methodNotAllowed,
	}

	return mux
}

// Clean 清除所有的路由项
//
// 包括子匹配项的 matchers
func (mux *Mux) Clean() *Mux {
	for _, m := range mux.routers {
		m.Clean()
	}
	mux.tree.Clean("")

	return mux
}

// All 返回所有的路由项
//
// ignoreHead 是否忽略自动生成的 HEAD 请求；
// ignoreOptions 是否忽略自动生成的 OPTIONS 请求；
// 返回值中，键名为路路地址，键值为该路由项对应的可用请求方法。
func (mux *Mux) All(ignoreHead, ignoreOptions bool) []*Router {
	routers := make([]*Router, 0, len(mux.routers)+1)

	for _, router := range mux.routers {
		routers = append(routers, router.All(ignoreHead, ignoreOptions)...)
	}

	return append(routers, &Router{
		Name:   mux.name,
		Routes: mux.tree.All(ignoreHead, ignoreOptions),
	})
}

// Remove 移除指定的路由项
//
// 当未指定 methods 时，将删除所有 method 匹配的项。
// 指定错误的 methods 值，将自动忽略该值。
func (mux *Mux) Remove(pattern string, methods ...string) *Mux {
	mux.tree.Remove(pattern, methods...)
	return mux
}

// Handle 添加一条路由数据
//
// pattern 为路由匹配模式，可以是正则匹配也可以是字符串匹配，
// 若语法不正确，则直接 panic，可以通过 IsWell 检测语法的有效性，其它接口也相同；
// methods 该路由项对应的请求方法，如果未指定值，则表示所有支持的请求方法，
// 但不包含 OPTIONS 和 HEAD。
func (mux *Mux) Handle(pattern string, h http.Handler, methods ...string) error {
	return mux.tree.Add(pattern, h, methods...)
}

// SetAllow 将 OPTIONS 请求方法的报头 allow 值固定为指定的值
//
// 若无特殊需求，不用调用此方法，系统会自动计算符合当前路由的请求方法列表。
// 如果想实现对处理方法的自定义，可以显示地调用 Handle 方法:
//  Mux.Handle("/api/1", handle, http.MethodOptions)
//
// Options 与 SetAllow 功能上完全相同，只是对错误处理上有所有区别。
// Options 在出错时 panic，而 SetAllow 会返回错误信息。
func (mux *Mux) SetAllow(pattern string, allow string) error {
	return mux.tree.SetAllow(pattern, allow)
}

// Options 将 OPTIONS 请求方法的报头 allow 值固定为指定的值
//
// 若无特殊需求，不用调用此方法，系统会自动计算符合当前路由的请求方法列表。
// 如果想实现对处理方法的自定义，可以显示地调用 Handle 方法:
//  Mux.Handle("/api/1", handle, http.MethodOptions)
//
// Options 与 SetAllow 功能上完全相同，只是对错误处理上有所有区别。
// Options 在出错时 panic，而 SetAllow 会返回错误信息。
func (mux *Mux) Options(pattern string, allow string) *Mux {
	if err := mux.SetAllow(pattern, allow); err != nil {
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
	hs, ps := mux.match(r)
	if hs == nil {
		mux.notFound(w, r)
		return
	}

	h := hs.Handler(r.Method)
	if h == nil {
		w.Header().Set("Allow", hs.Options())
		mux.methodNotAllowed(w, r)
		return
	}

	if len(ps) > 0 {
		ctx := context.WithValue(r.Context(), params.ContextKeyParams, ps)
		r = r.WithContext(ctx)
	}

	h.ServeHTTP(w, r)
}

// Name 为一条路由项命名
//
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

// URL 根据参数生成地址
//
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

// Params 获取路由的参数集合
//
// NOTE: 详细情况可参考 params.Get
func Params(r *http.Request) params.Params {
	return params.Get(r)
}

// IsWell 语法格式是否正确
//
// 如果出错，则会返回具体的错误信息。
func IsWell(pattern string) error {
	_, err := syntax.Split(pattern)
	return err
}

// Methods 返回所有支持的请求方法
func Methods() []string {
	methods := make([]string, len(handlers.Methods))
	copy(methods, handlers.Methods)
	return methods
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
