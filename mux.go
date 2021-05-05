// SPDX-License-Identifier: MIT

// Package mux 功能完备的路由中间件
package mux

import (
	"context"
	"net/http"
	"strings"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v4/group"
	"github.com/issue9/mux/v4/internal/handlers"
	"github.com/issue9/mux/v4/internal/syntax"
	"github.com/issue9/mux/v4/params"
)

var (
	defaultNotFound = func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	defaultMethodNotAllowed = func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
)

// Mux 提供了强大的路由匹配功能
//
// 可以对路径按正则或是请求方法进行匹配。用法如下：
//  m := mux.Default()
//  m.Get("/abc/h1", h1).
//    Post("/abc/h2", h2).
//    Handle("/api/{version:\\d+}",h3, http.MethodGet, http.MethodPost) // 只匹配 GET 和 POST
//  http.ListenAndServe(m)
type Mux struct {
	*Router
	routers []*Router
}

// Default New 的默认参数版本
func Default() *Mux { return New(false, false, false, nil, nil, "", nil) }

// New 声明一个新的 Mux
//
// disableOptions 是否禁用自动生成 OPTIONS 功能；
// disableHead 是否禁用根据 Get 请求自动生成 HEAD 请求；
// skipCleanPath 是否不对访问路径作处理，比如 "//api" ==> "/api"；
// notFound 404 页面的处理方式，为 nil 时会调用默认的方式进行处理；
// methodNotAllowed 405 页面的处理方式，为 nil 时会调用默认的方式进行处理；
// name 为当前路由组指定名称，可以为空，该名称即在 Router.Routes() 返回标记属于哪个路由组；
// m 当前路由组的匹配规则，可以为空，表示无规则；
func New(disableOptions, disableHead, skipCleanPath bool,
	notFound, methodNotAllowed http.HandlerFunc,
	name string, m group.Matcher) *Mux {
	r := newRouter(disableOptions, disableHead, skipCleanPath, notFound, methodNotAllowed, name, m)

	return &Mux{
		Router: r,
	}
}

// Routers 返回当前路由所属的子路由组列表
func (mux *Mux) Routers() []*Router { return mux.routers }

func (mux *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !mux.skipCleanPath {
		req.URL.Path = cleanPath(req.URL.Path)
	}

	hs, ps := mux.match(req)
	if hs == nil {
		mux.notFound(w, req)
		return
	}

	h := hs.Handler(req.Method)
	if h == nil {
		w.Header().Set("Allow", hs.Options())
		mux.methodNotAllowed(w, req)
		return
	}

	if len(ps) > 0 {
		req = req.WithContext(context.WithValue(req.Context(), params.ContextKeyParams, ps))
	}

	h.ServeHTTP(w, req)
}

func (mux *Mux) match(req *http.Request) (*handlers.Handlers, params.Params) {
	for _, m := range mux.routers {
		if m.matcher.Match(req) {
			return m.tree.Handler(req.URL.Path)
		}
	}

	if mux.matcher.Match(req) {
		return mux.tree.Handler(req.URL.Path)
	}
	return nil, nil
}

// NewRouter 添加子路由组
//
// 该路由只有符合 group.Matcher 的要求才会进入，其它与 Router 功能相同。
//
// name 表示该路由组的名称，需要唯一，否则返回 false；
func (mux *Mux) NewRouter(name string, matcher group.Matcher) (*Router, bool) {
	if matcher == nil {
		panic("参数 matcher 不能为空")
	}

	if mux.routers == nil {
		mux.routers = make([]*Router, 0, 5)
	}

	dup := sliceutil.Count(mux.routers, func(i int) bool {
		return mux.routers[i].name == name
	}) > 0
	if mux.Name() == name || dup {
		return nil, false
	}

	m := newRouter(mux.disableOptions, mux.disableHead, mux.skipCleanPath, mux.notFound, mux.methodNotAllowed, name, matcher)
	mux.routers = append(mux.routers, m)
	return m, true
}

// Params 获取路由的参数集合
//
// NOTE: 详细情况可参考 params.Get
func Params(r *http.Request) params.Params { return params.Get(r) }

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

	var b strings.Builder
	b.Grow(len(p) + 1)

	if p[0] != '/' {
		b.WriteByte('/')
	}

	index := strings.Index(p, "//")
	if index == -1 {
		b.WriteString(p)
		return b.String()
	}

	b.WriteString(p[:index+1])

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
		b.WriteByte(p[i])
	}

	return b.String()
}
