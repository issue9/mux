// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/assert/v4/rest"

	"github.com/issue9/mux/v9/header"
	"github.com/issue9/mux/v9/internal/tree"
	"github.com/issue9/mux/v9/types"
)

func call(w http.ResponseWriter, r *http.Request, ps types.Route, h http.Handler) {
	h.ServeHTTP(w, r)
}

var (
	methodNotAllowedBuilder = tree.BuildTestNodeHandlerFunc(http.StatusMethodNotAllowed)
	optionsHandlerBuilder   = tree.BuildTestNodeHandlerFunc(http.StatusOK)
)

func newRouter(a *assert.Assertion, name string, o ...Option) *Router[http.Handler] {
	a.TB().Helper()
	r := NewRouter(name, call, http.NotFoundHandler(), methodNotAllowedBuilder, optionsHandlerBuilder, o...)
	a.NotNil(r)
	return r
}

func TestRouter(t *testing.T) {
	a := assert.New(t, false)
	r := newRouter(a, "def", WithLock(true))

	r.Get("/", rest.BuildHandler(a, 201, "201", nil))
	r.Get("/200", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("200"))
		a.NotError(err)
	}))
	rest.Get(a, "/").Do(r).Status(201).StringBody("201")
	rest.NewRequest(a, http.MethodHead, "/").Do(r).Status(201).BodyEmpty()
	rest.Get(a, "/abc").Do(r).Status(http.StatusNotFound)
	rest.NewRequest(a, http.MethodHead, "/200").Do(r).Status(200).BodyEmpty() // 不调用 WriteHeader
	rest.NewRequest(a, http.MethodOptions, "*").Do(r).Status(200).Header(header.Allow, "GET, OPTIONS")

	r.Get("/h/1", rest.BuildHandler(a, 201, "", nil))
	rest.Get(a, "/h/1").Do(r).Status(201)

	r.Post("/h/1", rest.BuildHandler(a, 202, "", nil))
	rest.Post(a, "/h/1", nil).Do(r).Status(202)
	rest.NewRequest(a, http.MethodOptions, "*").Do(r).Status(200).Header(header.Allow, "GET, OPTIONS, POST")

	r.Put("/h/1", rest.BuildHandler(a, 203, "", nil))
	rest.Put(a, "/h/1", nil).Do(r).Status(203)

	r.Patch("/h/1", rest.BuildHandler(a, 204, "", nil))
	rest.Patch(a, "/h/1", nil).Do(r).Status(204)

	r.Delete("/h/1", rest.BuildHandler(a, 205, "", nil))
	rest.Delete(a, "/h/1").Do(r).Status(205)
	rest.NewRequest(a, http.MethodOptions, "*").Do(r).Status(200).Header(header.Allow, "DELETE, GET, OPTIONS, PATCH, POST, PUT")

	// Any
	r.Any("/h/any", rest.BuildHandler(a, 206, "", nil))
	rest.Delete(a, "/h/any").Do(r).Status(206)
	rest.Get(a, "/h/any").Do(r).Status(206)
	rest.Patch(a, "/h/any", nil).Do(r).Status(206)
	rest.Put(a, "/h/any", nil).Do(r).Status(206)
	rest.Post(a, "/h/any", nil).Do(r).Status(206)
	rest.NewRequest(a, http.MethodConnect, "/h/any").Do(r).Status(206)

	// 不能主动添加 Head
	a.PanicString(func() {
		r.Handle("/options", rest.BuildHandler(a, 202, "", nil), nil, http.MethodOptions)
	}, "OPTIONS")
}

func TestRouter_Handle_Remove(t *testing.T) {
	a := assert.New(t, false)
	r := newRouter(a, "def")
	a.NotNil(r)

	// 添加 GET /api/1
	// 添加 PUT /api/1
	// 添加 GET /api/2
	r.Handle("/api/1", rest.BuildHandler(a, 201, "", nil), nil, http.MethodGet)
	r.Handle("/api/1", rest.BuildHandler(a, 201, "", nil), nil, http.MethodPut)
	r.Handle("/api/2", rest.BuildHandler(a, 202, "", nil), nil, http.MethodGet)

	rest.Get(a, "/api/1").Do(r).Status(201)
	rest.Put(a, "/api/1", nil).Do(r).Status(201)
	rest.Get(a, "/api/2").Do(r).Status(202)
	rest.Delete(a, "/api/1").Do(r).Status(http.StatusMethodNotAllowed) // 未实现

	// 删除 GET /api/1
	r.Remove("/api/1", http.MethodGet)
	rest.Get(a, "/api/1").Do(r).Status(http.StatusMethodNotAllowed)
	rest.Put(a, "/api/1", nil).Do(r).Status(201) // 不影响 PUT
	rest.Get(a, "/api/2").Do(r).Status(202)

	// 删除 GET /api/2，只有一个，所以相当于整个节点被删除
	r.Remove("/api/2", http.MethodGet)
	rest.Get(a, "/api/1").Do(r).Status(http.StatusMethodNotAllowed)
	rest.Put(a, "/api/1", nil).Do(r).Status(201)            // 不影响 PUT
	rest.Get(a, "/api/2").Do(r).Status(http.StatusNotFound) // 整个节点被删除

	// 添加 POST /api/1
	r.Handle("/api/1", rest.BuildHandler(a, 201, "", nil), nil, http.MethodPost)
	rest.Post(a, "/api/1", nil).Do(r).Status(201)

	// 删除 ANY /api/1
	r.Remove("/api/1")
	rest.Get(a, "/api/1").Do(r).Status(http.StatusNotFound) // 整个节点被删除
}

func TestRouter_Routes(t *testing.T) {
	a := assert.New(t, false)

	def := newRouter(a, "def")
	a.NotNil(def)
	def.Get("/m", rest.BuildHandler(a, 1, "", nil))
	def.Post("/m", rest.BuildHandler(a, 1, "", nil))
	a.Equal(def.Routes(), map[string][]string{"*": {"OPTIONS"}, "/m": {"GET", "HEAD", "OPTIONS", "POST"}})
}

func TestRouter_Clean(t *testing.T) {
	a := assert.New(t, false)

	def := newRouter(a, "def")
	a.NotNil(def)
	def.Get("/m1", rest.BuildHandler(a, 200, "", nil)).
		Post("/m1", rest.BuildHandler(a, 201, "", nil))

	rest.Get(a, "http://localhost:88/m1").Do(def).Status(200)

	def.Clean()
	rest.Get(a, "/m1").Do(def).Status(404)
}

// 测试匹配顺序是否正确
func TestRouter_ServeHTTP_Order(t *testing.T) {
	a := assert.New(t, false)
	r := newRouter(a, "def", WithAnyInterceptor("any"))
	a.NotNil(r)

	r.Get("/posts/{id}", rest.BuildHandler(a, 203, "", nil))
	r.Get("/posts/{id:\\d+}", rest.BuildHandler(a, 202, "", nil))
	r.Get("/posts/1", rest.BuildHandler(a, 201, "", nil))
	r.Get("/posts/{id:[0-9]+}", rest.BuildHandler(a, 199, "", nil)) //  两个正则，后添加的永远匹配不到
	r.Get("/posts-{id:any}", rest.BuildHandler(a, 204, "", nil))
	r.Get("/posts-", rest.BuildHandler(a, 205, "", nil))
	rest.Get(a, "/posts/1").Do(r).Status(201)   // 普通路由项完全匹配
	rest.Get(a, "/posts/2").Do(r).Status(202)   // 正则路由
	rest.Get(a, "/posts/abc").Do(r).Status(203) // 命名路由
	rest.Get(a, "/posts/").Do(r).Status(203)    // 命名路由
	rest.Get(a, "/posts-5").Do(r).Status(204)   // 命名路由
	rest.Get(a, "/posts-").Do(r).Status(205)    // 204 只匹配非空

	// interceptor
	r = newRouter(a, "def", WithDigitInterceptor("[0-9]+"))
	a.NotNil(r)
	r.Get("/posts/{id}", rest.BuildHandler(a, 203, "", nil))        // f3
	r.Get("/posts/{id:\\d+}", rest.BuildHandler(a, 202, "", nil))   // f2 永远匹配不到
	r.Get("/posts/1", rest.BuildHandler(a, 201, "", nil))           // f1
	r.Get("/posts/{id:[0-9]+}", rest.BuildHandler(a, 210, "", nil)) // f0 interceptor 权限比正则要高
	rest.Get(a, "/posts/1").Do(r).Status(201)                       // f1 普通路由项完全匹配
	rest.Get(a, "/posts/2").Do(r).Status(210)                       // f0 interceptor
	rest.Get(a, "/posts/abc").Do(r).Status(203)                     // f3 命名路由
	rest.Get(a, "/posts/").Do(r).Status(203)                        // f3

	r = newRouter(a, "def")
	a.NotNil(r)
	r.Get("/p1/{p1}/p2/{p2:\\d+}", rest.BuildHandler(a, 201, "", nil)) // f1
	r.Get("/p1/{p1}/p2/{p2:\\w+}", rest.BuildHandler(a, 202, "", nil)) // f2
	rest.Get(a, "/p1/1/p2/1").Do(r).Status(201)                        // f1
	rest.Get(a, "/p1/2/p2/s").Do(r).Status(202)                        // f2

	r = newRouter(a, "def")
	a.NotNil(r)
	r.Get("/posts/{id}/{page}", rest.BuildHandler(a, 202, "", nil)) // f2
	r.Get("/posts/{id}/1", rest.BuildHandler(a, 201, "", nil))      // f1
	rest.Get(a, "/posts/1/1").Do(r).Status(201)                     // f1 普通路由项完全匹配
	rest.Get(a, "/posts/2/5").Do(r).Status(202)                     // f2 命名完全匹配

	r = newRouter(a, "def")
	a.NotNil(r)
	r.Get("/tags/{id}.html", rest.BuildHandler(a, 201, "", nil)) // f1
	r.Get("/tags.html", rest.BuildHandler(a, 202, "", nil))      // f2
	r.Get("/{path}", rest.BuildHandler(a, 203, "", nil))         // f3
	rest.Get(a, "/tags").Do(r).Status(203)                       // f3 // 正好与 f1 的第一个节点匹配
	rest.Get(a, "/tags/1.html").Do(r).Status(201)                // f1
	rest.Get(a, "/tags.html").Do(r).Status(202)                  // f2
}

func TestRouter_Middleware(t *testing.T) {
	a := assert.New(t, false)

	def := newRouter(a, "def")
	a.NotNil(def)
	def.Use(tree.BuildTestMiddleware(a, "m1"), tree.BuildTestMiddleware(a, "m2"), tree.BuildTestMiddleware(a, "m3"), tree.BuildTestMiddleware(a, "m4"))
	def.Get("/get", rest.BuildHandler(a, 201, "", nil), tree.BuildTestMiddleware(a, "m0"))
	def.Post("/get", rest.BuildHandler(a, 201, "", nil))

	rest.Get(a, "/get").Do(def).Status(201).StringBody("m0m1m2m3m4")
	rest.Post(a, "/get", nil).Do(def).Status(201).StringBody("m1m2m3m4")

	def.Use(tree.BuildTestMiddleware(a, "m5"), tree.BuildTestMiddleware(a, "m6"))
	rest.Get(a, "/get").Do(def).Status(201).StringBody("m0m1m2m3m4m5m6")
}

func TestResource(t *testing.T) {
	a := assert.New(t, false)
	r := newRouter(a, "def")

	h := r.Resource("/h/1")
	a.NotNil(h)

	h.Get(rest.BuildHandler(a, 201, "", nil))
	rest.Get(a, "/h/1").Do(r).Status(201)

	h.Post(rest.BuildHandler(a, 202, "", nil))
	rest.Post(a, "/h/1", nil).Do(r).Status(202)

	h.Put(rest.BuildHandler(a, 203, "", nil))
	rest.Put(a, "/h/1", nil).Do(r).Status(203)

	h.Patch(rest.BuildHandler(a, 204, "", nil))
	rest.Patch(a, "/h/1", nil).Do(r).Status(204)

	h.Delete(rest.BuildHandler(a, 205, "", nil))
	rest.Delete(a, "/h/1").Do(r).Status(205)

	// Any
	h = r.Resource("/h/any")
	h.Any(rest.BuildHandler(a, 206, "", nil))
	rest.Delete(a, "/h/any").Do(r).Status(206)
	rest.Get(a, "/h/any").Do(r).Status(206)
	rest.Post(a, "/h/any", nil).Do(r).Status(206)
	rest.Put(a, "/h/any", nil).Do(r).Status(206)
	rest.Patch(a, "/h/any", nil).Do(r).Status(206)
	rest.NewRequest(a, http.MethodConnect, "/h/any").Do(r).Status(206)

	// remove
	h.Remove(http.MethodGet, http.MethodPut)
	rest.Get(a, "/h/any").Do(r).Status(405)
	rest.Delete(a, "/h/any").Do(r).Status(206)

	r.Clean()
	rest.Get(a, "/f/any").Do(r).Status(404)
	rest.Delete(a, "/f/any").Do(r).Status(404)
}

func TestRouter_Resource(t *testing.T) {
	a := assert.New(t, false)
	def := newRouter(a, "def")
	a.NotNil(def)
	def.Use(tree.BuildTestMiddleware(a, "d1"))

	r1 := def.Resource("/abc/1", tree.BuildTestMiddleware(a, "r1"))
	a.NotNil(r1)
	a.Equal(r1.Router(), def)

	r2 := def.Resource("/abc/1", tree.BuildTestMiddleware(a, "r1"))
	a.NotNil(r2)
	a.False(r1 == r2) // 不是同一个 *Resource

	r2.Delete(rest.BuildHandler(a, 201, "-201-", nil), tree.BuildTestMiddleware(a, "d1"))
	rest.Delete(a, "/abc/1").Do(def).Status(201).StringBody("-201-d1r1d1")
}

func TestPrefix_Resource(t *testing.T) {
	a := assert.New(t, false)

	def := newRouter(a, "def", WithTrace(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { Trace(w, r, true) })))
	a.NotNil(def)

	def.Use(tree.BuildTestMiddleware(a, "r"))

	p := def.Prefix("/p1", tree.BuildTestMiddleware(a, "p1"), tree.BuildTestMiddleware(a, "p2"))
	a.NotNil(p)

	r1 := p.Resource("/abc/1", tree.BuildTestMiddleware(a, "r1"), tree.BuildTestMiddleware(a, "r2"))
	a.NotNil(r1)

	r1.Delete(rest.BuildHandler(a, 201, "-201-", nil), tree.BuildTestMiddleware(a, "m1"))
	rest.Delete(a, "/p1/abc/1").Do(def).Status(201).StringBody("-201-m1r1r2p1p2r")
	rest.Delete(a, "/p1/abc/1").Do(def).Status(201).StringBody("-201-m1r1r2p1p2r")
	rest.Post(a, "/p1/abc/1", nil).Do(def).Status(405).StringBody("m1r1r2p1p2r")             // 405 中间件正常使用
	rest.Get(a, "/p1/abc/not-exist").Do(def).Status(404).StringBody("404 page not found\nr") // 404 只有通过 [Router.Use] 添加的中间件有效
	rest.NewRequest(a, http.MethodTrace, "/p1/abc/not-exists").
		Do(def).
		Status(http.StatusOK).BodyFunc(func(a *assert.Assertion, body []byte) {
		a.Contains(string(body), "TRACE /p1/abc/not-exists").
			Contains(string(body), "\nr") // TRACE 只有 [Router.Use] 的中间件有效
	})
}

func TestResource_URL(t *testing.T) {
	a := assert.New(t, false)
	def := newRouter(a, "def", WithAllowedCORS(3600))
	a.NotNil(def)

	// 非正则
	res := def.Resource("/api/v1")
	a.NotNil(res)
	url, err := res.URL(false, map[string]string{"id": "1"})
	a.NotError(err).Equal(url, "/api/v1")

	// 没有参数
	url, err = res.URL(false, nil)
	a.NotError(err).Equal(url, "/api/v1")

	res = def.Resource("/api//v1")
	a.NotNil(res)
	url, err = res.URL(false, map[string]string{"id": "1"})
	a.NotError(err).Equal(url, "/api//v1")

	// 正常的单个参数
	res = def.Resource("/api/{id:\\d+}/{path}")
	a.NotNil(res)
	url, err = res.URL(false, map[string]string{"id": "1", "path": "p1"})
	a.NotError(err).Equal(url, "/api/1/p1")

	// 类型不正确
	url, err = res.URL(false, map[string]string{"id": "xxx", "path": "p1"})
	a.NotError(err).Equal(url, "/api/xxx/p1")
	url, err = res.URL(true, map[string]string{"id": "xxx", "path": "p1"})
	a.Error(err).Empty(url)

	res = def.Resource("/api/{id:\\d+}//{path}")
	a.NotNil(res)
	url, err = res.URL(false, map[string]string{"id": "1", "path": "p1"})
	a.NotError(err).Equal(url, "/api/1//p1")

	// 多个参数
	res = def.Resource("/api/{action}/{id:\\d+}")
	a.NotNil(res)
	url, err = res.URL(false, map[string]string{"id": "1", "action": "blog"})
	a.NotError(err).Equal(url, "/api/blog/1")

	// 缺少参数
	url, err = res.URL(false, map[string]string{"id": "1"})
	a.Error(err).Equal(url, "")

	url, err = res.URL(false, map[string]string{"id": "1", "action": "blog"})
	a.NotError(err).Equal(url, "/api/blog/1")
}

func TestPrefix(t *testing.T) {
	a := assert.New(t, false)
	r := newRouter(a, "def")
	a.NotNil(r)
	p := r.Prefix("/p")

	p.Get("/h/1", rest.BuildHandler(a, 201, "", nil))
	rest.Get(a, "/p/h/1").Do(r).Status(201)

	p.Post("/h/1", rest.BuildHandler(a, 202, "", nil))
	rest.Post(a, "/p/h/1", nil).Do(r).Status(202)

	p.Put("/h/1", rest.BuildHandler(a, 203, "", nil))
	rest.Put(a, "/p/h/1", nil).Do(r).Status(203)

	p.Patch("/h/1", rest.BuildHandler(a, 204, "", nil))
	rest.Patch(a, "/p/h/1", nil).Do(r).Status(204)

	p.Delete("/h/1", rest.BuildHandler(a, 205, "", nil))
	rest.Delete(a, "/p/h/1").Do(r).Status(205)

	// Any
	p.Any("/h/any", rest.BuildHandler(a, 206, "", nil))
	rest.Delete(a, "/p/h/any").Do(r).Status(206)
	rest.Get(a, "/p/h/any").Do(r).Status(206)
	rest.Patch(a, "/p/h/any", nil).Do(r).Status(206)
	rest.Put(a, "/p/h/any", nil).Do(r).Status(206)
	rest.Post(a, "/p/h/any", nil).Do(r).Status(206)
	rest.NewRequest(a, http.MethodConnect, "/p/h/any").Do(r).Status(206)

	// remove
	p.Remove("/h/any", http.MethodPut, http.MethodGet)
	methods := slices.DeleteFunc(Methods(), func(s string) bool {
		return s == http.MethodGet || s == http.MethodTrace || // 并未启用 TRACE
			s == http.MethodPut || s == http.MethodHead // 删除了 GET，HEAD 也会删除。
	})
	slices.Sort(methods)
	rest.Get(a, "/p/h/any").Do(r).Status(405).Header(header.Allow, strings.Join(methods, ", ")) // 已经删除
	rest.Delete(a, "/p/h/any").Do(r).Status(206)                                                // 未删除

	// clean
	p.Clean()
	rest.Delete(a, "/p/h/any").Do(r).Status(404)
	rest.NewRequest(a, http.MethodOptions, "/p/h/any").Do(r).Status(404)
}

func TestPrefix_Prefix(t *testing.T) {
	t.Run("prefix", func(t *testing.T) {
		a := assert.New(t, false)
		def := newRouter(a, "def", WithAllowedCORS(3600))
		a.NotNil(def)

		p := def.Prefix("/abc")
		a.Equal(p.Router(), def)
	})

	t.Run("empty prefix", func(t *testing.T) {
		a := assert.New(t, false)
		def := newRouter(a, "def2", WithAllowedCORS(3600))
		a.NotNil(def)

		p := def.Prefix("/abc")
		a.Equal(p.Router(), def)

		pp := p.Prefix("")
		pp.Delete("", rest.BuildHandler(a, 201, "", nil))
		rest.Delete(a, "/abc").Do(def).Status(201)
	})

	//
	t.Run("prefix 的中间调用顺序", func(t *testing.T) {
		a := assert.New(t, false)
		def := newRouter(a, "def3", WithAllowedCORS(3600))
		a.NotNil(def)
		def.Use(tree.BuildTestMiddleware(a, "r0"))

		p := def.Prefix("/abc", tree.BuildTestMiddleware(a, "p1"))
		a.Equal(p.Router(), def)

		pp := p.Prefix("", tree.BuildTestMiddleware(a, "p2"), tree.BuildTestMiddleware(a, "p3"))
		pp.Delete("", rest.BuildHandler(a, 201, "-201-", nil))

		rest.Delete(a, "/abc").Do(def).Status(201).StringBody("-201-p2p3p1r0")

		// Trace
		rest.NewRequest(a, http.MethodTrace, "/abc").Do(def).Status(http.StatusMethodNotAllowed)
		rest.NewRequest(a, http.MethodTrace, "/abc/not-exists").Do(def).Status(http.StatusNotFound)
	})
}

func TestPrefix_URL(t *testing.T) {
	a := assert.New(t, false)
	def := newRouter(a, "def", WithAllowedCORS(3600), WithURLDomain("https://example.com"))
	a.NotNil(def)

	// 非正则
	p := def.Prefix("/api")
	a.NotNil(p)
	url, err := p.URL(false, "/v1", map[string]string{"id": "1"})
	a.NotError(err).Equal(url, "https://example.com/api/v1")

	p = def.Prefix("//api")
	a.NotNil(p)
	url, err = p.URL(false, "/v1", map[string]string{"id": "1"})
	a.NotError(err).Equal(url, "https://example.com//api/v1")

	// 正常的单个参数
	p = def.Prefix("/api")
	a.NotNil(p)
	url, err = p.URL(false, "/{id:\\d+}/{path}", map[string]string{"id": "1", "path": "p1"})
	a.NotError(err).Equal(url, "https://example.com/api/1/p1")

	url, err = p.URL(false, "/{id:\\d+}///{path}", map[string]string{"id": "1", "path": "p1"})
	a.NotError(err).Equal(url, "https://example.com/api/1///p1")

	// 多个参数
	p = def.Prefix("/api")
	a.NotNil(p)
	url, err = p.URL(false, "/{action}/{id:\\d+}", map[string]string{"id": "1", "action": "blog"})
	a.NotError(err).Equal(url, "https://example.com/api/blog/1")

	// 缺少参数
	url, err = p.URL(false, "/{action}/{id:\\d+}", map[string]string{"id": "1"})
	a.Error(err).Equal(url, "")

	url, err = p.URL(false, "/{action}/{id:\\d+}", map[string]string{"id": "1", "action": "blog"})
	a.NotError(err).Equal(url, "https://example.com/api/blog/1")
}
