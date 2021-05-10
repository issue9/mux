// SPDX-License-Identifier: MIT

// Package mux 功能完备的路由中间件
package mux

import (
	"net/http"
	"sort"
	"strings"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v4/group"
	"github.com/issue9/mux/v4/internal/handlers"
	"github.com/issue9/mux/v4/internal/syntax"
	"github.com/issue9/mux/v4/internal/tree"
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
//  router, ok := m.NewRouter("default", group.Any)
//  router.Get("/abc/h1", h1).
//      Post("/abc/h2", h2).
//      Handle("/api/{version:\\d+}",h3, http.MethodGet, http.MethodPost) // 只匹配 GET 和 POST
//  http.ListenAndServe(m)
type Mux struct {
	routers []*Router

	notFound,
	methodNotAllowed http.HandlerFunc

	disableOptions,
	disableHead,
	skipCleanPath bool

	middlewares []MiddlewareFunc
	handler     http.Handler
}

// Default New 的默认参数版本
func Default() *Mux { return New(false, false, false, nil, nil) }

// New 声明一个新的 Mux
//
// disableOptions 是否禁用自动生成 OPTIONS 功能；
// disableHead 是否禁用根据 Get 请求自动生成 HEAD 请求；
// skipCleanPath 是否不对访问路径作处理，比如 "//api" ==> "/api"；
// notFound 404 页面的处理方式，为 nil 时会调用默认的方式进行处理；
// methodNotAllowed 405 页面的处理方式，为 nil 时会调用默认的方式进行处理；
func New(disableOptions, disableHead, skipCleanPath bool, notFound, methodNotAllowed http.HandlerFunc) *Mux {
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

		disableOptions: disableOptions,
		disableHead:    disableHead,
		skipCleanPath:  skipCleanPath,
	}
	mux.handler = http.HandlerFunc(mux.serveHTTP)

	return mux
}

func (mux *Mux) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if !mux.skipCleanPath {
		r.URL.Path = cleanPath(r.URL.Path)
	}

	for _, router := range mux.routers {
		if router.matcher.Match(r) {
			router.serveHTTP(w, r)
			return
		}
	}
	mux.notFound(w, r)
}

// Routers 返回当前路由所属的子路由组列表
func (mux *Mux) Routers() []*Router { return mux.routers }

// NewRouter 添加子路由
//
// 该路由只有符合 group.Matcher 的要求才会进入，其它与 Router 功能相同。
// 当 group.Matcher 与其它路由组的判断有重复时，第一条返回 true 的路由组获胜，
// 即使该路由组最终返回 404，也不会再在其它路由组里查找相应的路由。
// 所以在有多条子路由的情况下，第一条子路由不应该永远返回 true，
// 这样会造成其它子路由永远无法到达。
//
// name 表示该路由组的名称，需要唯一，否则返回 false；
// matcher 路由的准入条件，如果为空，则此条路由匹配时会被排在最后，
// 只有一个路由的 matcher 为空，否则会 panic。
func (mux *Mux) NewRouter(name string, matcher group.Matcher) (r *Router, ok bool) {
	if name == "" {
		panic("参数 name 不能为空")
	}

	// 重名检测
	index := sliceutil.Index(mux.routers, func(i int) bool {
		return mux.routers[i].name == name
	})
	if index > -1 {
		return nil, false
	}

	last := matcher == nil
	if last {
		matcher = group.MatcherFunc(group.Any)
		index := sliceutil.Index(mux.routers, func(i int) bool { return mux.routers[i].last })
		if index > -1 {
			panic("已经存在一个 matcher 参数为空的路由")
		}
	}

	r = &Router{
		mux:     mux,
		name:    name,
		matcher: matcher,
		tree:    tree.New(mux.disableOptions, mux.disableHead),
		last:    last,
	}
	mux.routers = append(mux.routers, r)
	sortRouters(mux.routers)
	return r, true
}

func sortRouters(rs []*Router) {
	sort.SliceStable(rs, func(i, j int) bool {
		if rs[i].last {
			return rs[j].last
		}
		return rs[j].last
	})
}

// RemoveRouter 删除子路由
func (mux *Mux) RemoveRouter(name string) {
	size := sliceutil.Delete(mux.routers, func(i int) bool {
		return mux.routers[i].name == name
	})
	mux.routers = mux.routers[:size]
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
