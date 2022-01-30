// SPDX-License-Identifier: MIT

package mux_test

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"

	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/routertest"
)

type (
	ctx struct {
		R *http.Request
		W http.ResponseWriter
		P mux.Params
	}
	ctxHandlerFunc func(ctx *ctx)
)

func contextCall(w http.ResponseWriter, r *http.Request, ps mux.Params, h ctxHandlerFunc) {
	h(&ctx{R: r, W: w, P: ps})
}

func buildMiddleware(a *assert.Assertion, text string) mux.Middleware {
	return mux.MiddlewareFunc(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r) // 先输出被包含的内容
			_, err := w.Write([]byte(text))
			a.NotError(err)
		})
	})
}

func TestRouter_Middleware(t *testing.T) {
	a := assert.New(t, false)

	def := mux.NewRouter("",
		&mux.Options{
			Middlewares: []mux.Middleware{
				buildMiddleware(a, "m1"),
				buildMiddleware(a, "m2"),
				buildMiddleware(a, "m3"),
				buildMiddleware(a, "m4"),
			},
		},
	)
	a.NotNil(def)
	def.Get("/get", rest.BuildHandler(a, 201, "", nil))

	rest.Get(a, "/get").Do(def).Status(201).StringBody("m1m2m3m4") // buildHandler 导致顶部的后输出
}

func TestDefaultRouter(t *testing.T) {
	a := assert.New(t, false)
	tt := routertest.NewTester[http.Handler](mux.DefaultCall)

	a.Run("params", func(a *assert.Assertion) {
		tt.Params(a, func(ps *mux.Params) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				p := mux.GetParams(r)
				if p != nil {
					p.Range(func(k, v string) {
						(*ps).Set(k, v)
					})
				}
			})
		})
	})

	a.Run("serve", func(a *assert.Assertion) {
		tt.Serve(a, func(status int) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
			})
		})
	})
}

func TestContextRouter_Params(t *testing.T) {
	a := assert.New(t, false)
	tt := routertest.NewTester[ctxHandlerFunc](contextCall)

	a.Run("params", func(a *assert.Assertion) {
		tt.Params(a, func(ps *mux.Params) ctxHandlerFunc {
			return func(c *ctx) {
				if c.P != nil {
					c.P.Range(func(k, v string) {
						(*ps).Set(k, v)
					})
				}
			}
		})
	})

	a.Run("serve", func(a *assert.Assertion) {
		tt.Serve(a, func(status int) ctxHandlerFunc {
			return func(c *ctx) {
				c.W.WriteHeader(status)
			}
		})
	})
}

func TestResource(t *testing.T) {
	a := assert.New(t, false)
	r := mux.NewRouter("", nil)

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
	rest.NewRequest(a, http.MethodTrace, "/h/any").Do(r).Status(206)

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
	def := mux.NewRouter("", nil)
	a.NotNil(def)

	r1 := def.Resource("/abc/1")
	a.NotNil(r1)
	a.Equal(r1.Router(), def)

	r2 := def.Resource("/abc/1")
	a.NotNil(r2)
	a.False(r1 == r2) // 不是同一个 *Resource

	r2.Delete(rest.BuildHandler(a, 201, "", nil))
	rest.Delete(a, "/abc/1").Do(def).Status(201)
}

func TestPrefix_Resource(t *testing.T) {
	a := assert.New(t, false)

	def := mux.NewRouter("", nil)
	a.NotNil(def)

	p := def.Prefix("/p1", buildMiddleware(a, "p1"), buildMiddleware(a, "p2"))
	a.NotNil(p)

	r1 := p.Resource("/abc/1", buildMiddleware(a, "r1"), buildMiddleware(a, "r2"))
	a.NotNil(r1)

	r1.Delete(rest.BuildHandler(a, 201, "-201-", nil))
	rest.Delete(a, "/p1/abc/1").Do(def).Status(201).StringBody("-201-p1p2r1r2")
}

func TestResource_URL(t *testing.T) {
	a := assert.New(t, false)
	def := mux.NewRouter("", &mux.Options{CORS: mux.AllowedCORS()})
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
	r := mux.NewRouter("", nil)
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
	rest.NewRequest(a, http.MethodTrace, "/p/h/any").Do(r).Status(206)

	// remove
	p.Remove("/h/any", http.MethodPut, http.MethodGet)
	rest.Get(a, "/p/h/any").Do(r).Status(405)    // 已经删除
	rest.Delete(a, "/p/h/any").Do(r).Status(206) // 未删除

	// clean
	p.Clean()
	rest.Delete(a, "/p/h/any").Do(r).Status(404)
	rest.NewRequest(a, http.MethodOptions, "/p/h/any").Do(r).Status(404)
}

func TestRouter_Prefix(t *testing.T) {
	a := assert.New(t, false)

	a.Run("prefix", func(a *assert.Assertion) {
		def := mux.NewRouter("", &mux.Options{CORS: mux.AllowedCORS()})
		a.NotNil(def)

		p := def.Prefix("/abc")
		a.Equal(p.Router(), def)

		p = def.Prefix("")
	}).Run("prefix with middleware", func(a *assert.Assertion) {
		def := mux.NewRouter("", &mux.Options{CORS: mux.AllowedCORS()})
		a.NotNil(def)

		p := def.Prefix("/abc")
		a.Equal(p.Router(), def)

		pp := p.Prefix("")
		pp.Delete("", rest.BuildHandler(a, 201, "", nil))
		rest.Delete(a, "/abc").Do(def).Status(201)
	}).Run("empty prefix with middleware", func(a *assert.Assertion) {
		def := mux.NewRouter("", &mux.Options{CORS: mux.AllowedCORS()})
		a.NotNil(def)

		p := def.Prefix("/abc")
		a.Equal(p.Router(), def)

		pp := p.Prefix("", buildMiddleware(a, "p1"), buildMiddleware(a, "p2"))
		pp.Delete("", rest.BuildHandler(a, 201, "-201-", nil))
		rest.Delete(a, "/abc").Do(def).Status(201).StringBody("-201-p1p2")
	})
}

func TestPrefix_Prefix(t *testing.T) {
	a := assert.New(t, false)
	def := mux.NewRouter("", &mux.Options{CORS: mux.AllowedCORS()})
	a.NotNil(def)

	p := def.Prefix("/abc", buildMiddleware(a, "p1"), buildMiddleware(a, "p2"))
	pp := p.Prefix("/def", buildMiddleware(a, "pp1"), buildMiddleware(a, "pp2"))
	a.Equal(p.Router(), def)
	pp.Delete("", rest.BuildHandler(a, 201, "-201-", nil))

	rest.Delete(a, "/abc/def").Do(def).Status(201).StringBody("-201-p1p2pp1pp2")
}

func TestPrefix_URL(t *testing.T) {
	a := assert.New(t, false)
	def := mux.NewRouter("", &mux.Options{
		CORS:      mux.AllowedCORS(),
		URLDomain: "https://example.com",
	})
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
