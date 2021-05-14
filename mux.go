// SPDX-License-Identifier: MIT

// Package mux 功能完备的路由中间件
//
// 语法
//
// 路由支持以 {} 的形式包含参数，比如：/posts/{id}.html，id 在解析时会解析任意字符。
// 也可以在 {} 中约束参数的范围，比如 /posts/{id:\\d+}.html，表示 id 只能匹配数字。
// 路由地址可以是 ascii 字符，但是参数名称如果是非 ascii，在正则表达式中无法使用。
package mux

import (
	"net/http"
	"strings"

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
//  router, ok := m.NewRouter("default", group.Any, AllowedCORS())
//  router.Get("/abc/h1", h1).
//      Post("/abc/h2", h2).
//      Handle("/api/{version:\\d+}",h3, http.MethodGet, http.MethodPost) // 只匹配 GET 和 POST
//  http.ListenAndServe(m)
type Mux struct {
	routers []*Router

	notFound,
	methodNotAllowed http.HandlerFunc

	disableHead,
	skipCleanPath bool

	middlewares *Middlewares
}

// Default New 的默认参数版本
func Default() *Mux { return New(false, false, nil, nil) }

// New 声明一个新的 Mux
//
// disableHead 是否禁用根据 Get 请求自动生成 HEAD 请求；
// skipCleanPath 是否不对访问路径作处理，比如 "//api" ==> "/api"；
// notFound 404 页面的处理方式，为 nil 时会调用默认的方式进行处理；
// methodNotAllowed 405 页面的处理方式，为 nil 时会调用默认的方式进行处理；
func New(disableHead, skipCleanPath bool, notFound, methodNotAllowed http.HandlerFunc) *Mux {
	if notFound == nil {
		notFound = defaultNotFound
	}
	if methodNotAllowed == nil {
		methodNotAllowed = defaultMethodNotAllowed
	}

	mux := &Mux{
		routers: make([]*Router, 0, 1),

		notFound:         notFound,
		methodNotAllowed: methodNotAllowed,

		disableHead:   disableHead,
		skipCleanPath: skipCleanPath,
	}
	mux.middlewares = NewMiddlewares(http.HandlerFunc(mux.serveHTTP))

	return mux
}

func (mux *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux.middlewares.ServeHTTP(w, r)
}

func (mux *Mux) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if !mux.skipCleanPath {
		r.URL.Path = cleanPath(r.URL.Path)
	}

	for _, router := range mux.routers {
		if req, ok := router.matcher.Match(r); ok {
			router.middlewares.ServeHTTP(w, req)
			return
		}
	}
	mux.notFound(w, r)
}

// Params 获取路由的参数集合
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
