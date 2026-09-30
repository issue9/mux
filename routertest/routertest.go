// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package routertest 提供针对路由的测试用例
package routertest

import (
	"net/http"
	"net/http/httptest"

	"github.com/issue9/assert/v5"
	"github.com/issue9/assert/v5/rest"

	"github.com/issue9/mux/v10"
	"github.com/issue9/mux/v10/types"
)

type TestRouter[T any] struct {
	ServeHTTP http.HandlerFunc
	Handle    func(pattern string, h T, methods ...string)
	Clean     func()
	Remove    func(pattern string, methods ...string)
	URL       func(strict bool, pattern string, params map[string]string) (string, error)
}

type Tester[T any] struct {
	loader func(o ...mux.Option) *TestRouter[T]
}

// NewTester 声明一个用于测试的对象
//
// loader 路由的加载方法，用户需要在该函数中初始化自己的路由对象，并用此对象构建一个测试路由。
func NewTester[T any](loader func(o ...mux.Option) *TestRouter[T]) *Tester[T] {
	return &Tester[T]{
		loader: loader,
	}
}

// Params 测试参数是否正常
//
// f 返回一个路由处理函数，该函数必须要将获得的参数写入 route。
func (t *Tester[T]) Params(a *assert.Assertion, f func(route *types.Route) T) {
	router := t.loader(mux.WithAnyInterceptor("any"), mux.WithDigitInterceptor("digit"))

	globalParams := types.NewRoute()

	requestParams := func(method, url string, status int, ps map[string]string) {
		a.TB().Helper()

		w := httptest.NewRecorder()
		r := rest.NewRequest(a, method, url).Request()

		router.ServeHTTP(w, r)

		a.Equal(w.Code, status)
		if len(ps) > 0 { // 由于 globalParams 是公用数据，会保存上一次获取的值，所以只在有值时才比较
			a.Equal(len(ps), globalParams.Count())
			for k, v := range ps {
				vv, found := globalParams.Get(k)
				a.True(found).Equal(vv, v)
			}
		}
		globalParams.Reset() // 清空全局的 globalParams
	}

	// 添加 patch /api/{version:\\d+}
	router.Handle("/api/{version:\\d+}", f(globalParams), http.MethodPatch)
	requestParams(http.MethodPatch, "/api/256", http.StatusOK, map[string]string{"version": "256"})
	requestParams(http.MethodPatch, "/api/2", http.StatusOK, map[string]string{"version": "2"})
	requestParams(http.MethodGet, "/api/256", http.StatusMethodNotAllowed, nil) // 不存在的请求方法

	// 添加 patch /api/v2/{version:\\d*}
	router.Clean()
	router.Handle("/api/v2/{version:\\d*}", f(globalParams), http.MethodPatch)
	requestParams(http.MethodPatch, "/api/v2/2", http.StatusOK, map[string]string{"version": "2"})
	requestParams(http.MethodPatch, "/api/v2/", http.StatusOK, map[string]string{"version": ""})

	// 忽略名称捕获
	router.Clean()
	router.Handle("/api/v3/{-version:\\d*}", f(globalParams), http.MethodPatch)
	requestParams(http.MethodPatch, "/api/v3/2", http.StatusOK, nil)
	requestParams(http.MethodPatch, "/api/v3/", http.StatusOK, nil)

	// 添加 patch /api/v2/{version:\\d*}/test
	router.Clean()
	router.Handle("/api/v2/{version:\\d*}/test", f(globalParams), http.MethodPatch)
	requestParams(http.MethodPatch, "/api/v2/2/test", http.StatusOK, map[string]string{"version": "2"})
	requestParams(http.MethodPatch, "/api/v2//test", http.StatusOK, map[string]string{"version": ""})

	// 中文作为值
	router.Clean()
	router.Handle("/api/v3/{版本:digit}", f(globalParams), http.MethodPatch)
	requestParams(http.MethodPatch, "/api/v3/2", http.StatusOK, map[string]string{"版本": "2"})
}

// Serve 测试路由是否正常
//
// h 返回路由处理函数，该函数只要输出 status 作为其状态码即可。
func (t *Tester[T]) Serve(a *assert.Assertion, h func(status int) T) {
	router := t.loader(mux.WithAnyInterceptor("any"), mux.WithDigitInterceptor("digit"))
	a.NotNil(router)
	srv := rest.NewServer(a, router.ServeHTTP, nil)

	router.Handle("/posts/{path}.html", h(201))
	srv.Get("/posts/2017/1.html").Do(nil).Status(201)
	srv.Get("/Posts/2017/1.html").Do(nil).Status(404) // 大小写不一样

	router.Handle("/posts/{path:.+}.html", h(202))
	srv.Get("/posts/2017/1.html").Do(nil).Status(202)

	router.Handle("/posts/{id:digit}123", h(203))
	srv.Get("/posts/123123").Do(nil).Status(203)

	router.Handle("///", h(201), http.MethodGet)
	srv.Get("///").Do(nil).Status(201)
	srv.Get("//").Do(nil).Status(404)

	// 对 any 拦截器和空参数的测试

	router.Handle("/posts1-{id}-{page}.html", h(201), http.MethodGet)
	srv.Get("/posts1--.html").Do(nil).Status(201)
	srv.Get("/posts1-1-0.html").Do(nil).Status(201)

	router.Handle("/posts2-{id:any}-{page:any}.html", h(201), http.MethodGet)
	srv.Get("/posts2--.html").Do(nil).Status(404)
	srv.Get("/posts2-1-0.html").Do(nil).Status(201)

	router.Handle("/posts3-{id}-{page:any}.html", h(201), http.MethodGet)
	srv.Get("/posts3--.html").Do(nil).Status(404)
	srv.Get("/posts3-1-0.html").Do(nil).Status(201)
	srv.Get("/posts3--0.html").Do(nil).Status(201)
}
