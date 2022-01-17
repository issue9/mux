// SPDX-License-Identifier: MIT

// Package routertest 提供针对路由的测试用例
//
// NOTE: 只提供针对不同的类型参数 T 可能造成不周结果的测试。
package routertest

import (
	"net/http"
	"net/http/httptest"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"

	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/internal/syntax"
)

type Tester[T any] struct {
	c mux.CallOf[T]
}

func NewTester[T any](c mux.CallOf[T]) *Tester[T] {
	return &Tester[T]{
		c: c,
	}
}

// Params 测试参数是否正常
//
// f 返回一个路由处理函数，该函数必须要将获得的参数写入 params.
func (t *Tester[T]) Params(a *assert.Assertion, f func(params *mux.Params) T) {
	router := mux.NewRouterOf("test", t.c, nil, mux.Interceptor(mux.InterceptorDigit, "digit"))
	a.NotNil(router)

	var globalParams mux.Params = syntax.NewParams("")

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
		globalParams = syntax.NewParams("") // 清空全局的 globalParams
	}

	// 添加 patch /api/{version:\\d+}
	router.Patch("/api/{version:\\d+}", f(&globalParams))
	requestParams(http.MethodPatch, "/api/256", http.StatusOK, map[string]string{"version": "256"})
	requestParams(http.MethodPatch, "/api/2", http.StatusOK, map[string]string{"version": "2"})
	requestParams(http.MethodGet, "/api/256", http.StatusMethodNotAllowed, nil) // 不存在的请求方法

	// 添加 patch /api/v2/{version:\\d*}
	router.Clean()
	router.Patch("/api/v2/{version:\\d*}", f(&globalParams))
	requestParams(http.MethodPatch, "/api/v2/2", http.StatusOK, map[string]string{"version": "2"})
	requestParams(http.MethodPatch, "/api/v2/", http.StatusOK, map[string]string{"version": ""})

	// 忽略名称捕获
	router.Clean()
	router.Patch("/api/v3/{-version:\\d*}", f(&globalParams))
	requestParams(http.MethodPatch, "/api/v3/2", http.StatusOK, nil)
	requestParams(http.MethodPatch, "/api/v3/", http.StatusOK, nil)

	// 添加 patch /api/v2/{version:\\d*}/test
	router.Clean()
	router.Patch("/api/v2/{version:\\d*}/test", f(&globalParams))
	requestParams(http.MethodPatch, "/api/v2/2/test", http.StatusOK, map[string]string{"version": "2"})
	requestParams(http.MethodPatch, "/api/v2//test", http.StatusOK, map[string]string{"version": ""})

	// 中文作为值
	router.Clean()
	router.Patch("/api/v3/{版本:digit}", f(&globalParams))
	requestParams(http.MethodPatch, "/api/v3/2", http.StatusOK, map[string]string{"版本": "2"})
}

// Serve 测试路由是否正常
//
// h 返回路由处理函数，该函数只要输出 status 作为其状态码即可。
func (t *Tester[T]) Serve(a *assert.Assertion, h func(status int) T) {
	router := mux.NewRouterOf("test", t.c, nil, mux.Interceptor(mux.InterceptorDigit, "digit"), mux.Interceptor(mux.InterceptorAny, "any"))
	a.NotNil(router)
	srv := rest.NewServer(a, router, nil)

	router.Handle("/posts/{path}.html", h(201))
	srv.NewRequest(http.MethodGet, "/posts/2017/1.html").Do(nil).Status(201)
	srv.NewRequest(http.MethodGet, "/Posts/2017/1.html").Do(nil).Status(404) // 大小写不一样

	router.Handle("/posts/{path:.+}.html", h(202))
	srv.NewRequest(http.MethodGet, "/posts/2017/1.html").Do(nil).Status(202)

	router.Handle("/posts/{id:digit}123", h(203))
	srv.NewRequest(http.MethodGet, "/posts/123123").Do(nil).Status(203)

	router.Get("///", h(201))
	srv.NewRequest(http.MethodGet, "///").Do(nil).Status(201)
	srv.NewRequest(http.MethodGet, "//").Do(nil).Status(404)

	// 对 any 和空参数的测试

	router.Get("/posts1-{id}-{page}.html", h(201))
	srv.NewRequest(http.MethodGet, "/posts1--.html").Do(nil).Status(201)
	srv.NewRequest(http.MethodGet, "/posts1-1-0.html").Do(nil).Status(201)

	router.Get("/posts2-{id:any}-{page:any}.html", h(201))
	srv.NewRequest(http.MethodGet, "/posts2--.html").Do(nil).Status(404)
	srv.NewRequest(http.MethodGet, "/posts2-1-0.html").Do(nil).Status(201)

	router.Get("/posts3-{id}-{page:any}.html", h(201))
	srv.NewRequest(http.MethodGet, "/posts3--.html").Do(nil).Status(404)
	srv.NewRequest(http.MethodGet, "/posts3-1-0.html").Do(nil).Status(201)
	srv.NewRequest(http.MethodGet, "/posts3--0.html").Do(nil).Status(201)

	// 忽略大小写测试

	router = mux.NewRouterOf("test", t.c, nil, mux.CaseInsensitive)
	srv = rest.NewServer(a, router, nil)

	router.Handle("/posts/{path}.html", h(201))
	srv.NewRequest(http.MethodGet, "/posts/2017/1.html").Do(nil).Status(201)
	srv.NewRequest(http.MethodGet, "/Posts/2017/1.html").Do(nil).Status(201) // 忽略大小写
}
