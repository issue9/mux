// SPDX-License-Identifier: MIT

package mux

import (
	"context"
	"net/http"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"

	"github.com/issue9/mux/v6/internal/syntax"
)

var (
	_ http.Handler = &Router{}
	_ Middleware   = MiddlewareFunc(func(http.Handler) http.Handler { return nil })
)

func TestWithValue(t *testing.T) {
	a := assert.New(t, false)

	r := rest.Get(a, "/to/path").Request()
	a.Equal(WithValue(r, &syntax.Params{}), r)

	r = rest.Get(a, "/to/path").Request()
	pp := NewParams()
	pp.Set("k1", "v1")
	r = WithValue(r, pp)

	pp = NewParams()
	pp.Set("k2", "v2")
	r = WithValue(r, pp)
	ps := GetParams(r)
	a.NotNil(ps).
		Equal(ps.MustString("k2", "def"), "v2").
		Equal(ps.MustString("k1", "def"), "v1")
}

func TestGetParams(t *testing.T) {
	a := assert.New(t, false)

	r := rest.Get(a, "/to/path").Request()
	ps := GetParams(r)
	a.Nil(ps)

	kvs := []syntax.Param{{K: "key1", V: "1"}}
	r = rest.Get(a, "/to/path").Request()
	ctx := context.WithValue(r.Context(), contextKeyParams, &syntax.Params{Params: kvs})
	r = r.WithContext(ctx)
	a.Equal(GetParams(r).MustString("key1", "def"), "1")
}

func TestRouter(t *testing.T) {
	a := assert.New(t, false)
	r := NewRouter("def", &Options{Lock: true})

	r.Get("/", rest.BuildHandler(a, 201, "201", nil))
	r.Get("/200", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("200"))
		a.NotError(err)
	}))
	rest.Get(a, "/").Do(r).Status(201).StringBody("201")
	rest.NewRequest(a, http.MethodHead, "/").Do(r).Status(201).BodyEmpty()
	rest.Get(a, "/abc").Do(r).Status(http.StatusNotFound)
	rest.NewRequest(a, http.MethodHead, "/200").Do(r).Status(200).BodyEmpty() // 不调用 WriteHeader
	rest.NewRequest(a, http.MethodOptions, "*").Do(r).Status(200).Header("Allow", "GET, OPTIONS")

	r.Get("/h/1", rest.BuildHandler(a, 201, "", nil))
	rest.Get(a, "/h/1").Do(r).Status(201)

	r.Post("/h/1", rest.BuildHandler(a, 202, "", nil))
	rest.Post(a, "/h/1", nil).Do(r).Status(202)
	rest.NewRequest(a, http.MethodOptions, "*").Do(r).Status(200).Header("Allow", "GET, OPTIONS, POST")

	r.Put("/h/1", rest.BuildHandler(a, 203, "", nil))
	rest.Put(a, "/h/1", nil).Do(r).Status(203)

	r.Patch("/h/1", rest.BuildHandler(a, 204, "", nil))
	rest.Patch(a, "/h/1", nil).Do(r).Status(204)

	r.Delete("/h/1", rest.BuildHandler(a, 205, "", nil))
	rest.Delete(a, "/h/1").Do(r).Status(205)
	rest.NewRequest(a, http.MethodOptions, "*").Do(r).Status(200).Header("Allow", "DELETE, GET, OPTIONS, PATCH, POST, PUT")

	// Any
	r.Any("/h/any", rest.BuildHandler(a, 206, "", nil))
	rest.Delete(a, "/h/any").Do(r).Status(206)
	rest.Get(a, "/h/any").Do(r).Status(206)
	rest.Patch(a, "/h/any", nil).Do(r).Status(206)
	rest.Put(a, "/h/any", nil).Do(r).Status(206)
	rest.Post(a, "/h/any", nil).Do(r).Status(206)
	rest.NewRequest(a, http.MethodTrace, "/h/any").Do(r).Status(206)

	// 不能主动添加 Head
	a.PanicString(func() {
		r.Handle("/options", rest.BuildHandler(a, 202, "", nil), http.MethodOptions)
	}, "OPTIONS")
}

func TestRouter_Handle_Remove(t *testing.T) {
	a := assert.New(t, false)
	r := NewRouter("", nil)
	a.NotNil(r)

	// 添加 GET /api/1
	// 添加 PUT /api/1
	// 添加 GET /api/2
	r.Handle("/api/1", rest.BuildHandler(a, 201, "", nil), http.MethodGet)
	r.Handle("/api/1", rest.BuildHandler(a, 201, "", nil), http.MethodPut)
	r.Handle("/api/2", rest.BuildHandler(a, 202, "", nil), http.MethodGet)

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
	r.Handle("/api/1", rest.BuildHandler(a, 201, "", nil), http.MethodPost)
	rest.Post(a, "/api/1", nil).Do(r).Status(201)

	// 删除 ANY /api/1
	r.Remove("/api/1")
	rest.Get(a, "/api/1").Do(r).Status(http.StatusNotFound) // 整个节点被删除
}

func TestRouter_Routes(t *testing.T) {
	a := assert.New(t, false)

	def := NewRouter("", nil)
	a.NotNil(def)
	def.Get("/m", rest.BuildHandler(a, 1, "", nil))
	def.Post("/m", rest.BuildHandler(a, 1, "", nil))
	a.Equal(def.Routes(), map[string][]string{"*": {"OPTIONS"}, "/m": {"GET", "HEAD", "OPTIONS", "POST"}})
}

func TestRouter_Clean(t *testing.T) {
	a := assert.New(t, false)

	def := NewRouter("", nil)
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
	r := NewRouter("def", &Options{Interceptors: map[string]InterceptorFunc{"any": InterceptorAny}})
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
	r = NewRouter("def", &Options{Interceptors: map[string]InterceptorFunc{"[0-9]+": InterceptorDigit}})
	a.NotNil(r)
	r.Get("/posts/{id}", rest.BuildHandler(a, 203, "", nil))        // f3
	r.Get("/posts/{id:\\d+}", rest.BuildHandler(a, 202, "", nil))   // f2 永远匹配不到
	r.Get("/posts/1", rest.BuildHandler(a, 201, "", nil))           // f1
	r.Get("/posts/{id:[0-9]+}", rest.BuildHandler(a, 210, "", nil)) // f0 interceptor 权限比正则要高
	rest.Get(a, "/posts/1").Do(r).Status(201)                       // f1 普通路由项完全匹配
	rest.Get(a, "/posts/2").Do(r).Status(210)                       // f0 interceptor
	rest.Get(a, "/posts/abc").Do(r).Status(203)                     // f3 命名路由
	rest.Get(a, "/posts/").Do(r).Status(203)                        // f3

	r = NewRouter("def", nil)
	a.NotNil(r)
	r.Get("/p1/{p1}/p2/{p2:\\d+}", rest.BuildHandler(a, 201, "", nil)) // f1
	r.Get("/p1/{p1}/p2/{p2:\\w+}", rest.BuildHandler(a, 202, "", nil)) // f2
	rest.Get(a, "/p1/1/p2/1").Do(r).Status(201)                        // f1
	rest.Get(a, "/p1/2/p2/s").Do(r).Status(202)                        // f2

	r = NewRouter("def", nil)
	a.NotNil(r)
	r.Get("/posts/{id}/{page}", rest.BuildHandler(a, 202, "", nil)) // f2
	r.Get("/posts/{id}/1", rest.BuildHandler(a, 201, "", nil))      // f1
	rest.Get(a, "/posts/1/1").Do(r).Status(201)                     // f1 普通路由项完全匹配
	rest.Get(a, "/posts/2/5").Do(r).Status(202)                     // f2 命名完全匹配

	r = NewRouter("def", nil)
	a.NotNil(r)
	r.Get("/tags/{id}.html", rest.BuildHandler(a, 201, "", nil)) // f1
	r.Get("/tags.html", rest.BuildHandler(a, 202, "", nil))      // f2
	r.Get("/{path}", rest.BuildHandler(a, 203, "", nil))         // f3
	rest.Get(a, "/tags").Do(r).Status(203)                       // f3 // 正好与 f1 的第一个节点匹配
	rest.Get(a, "/tags/1.html").Do(r).Status(201)                // f1
	rest.Get(a, "/tags.html").Do(r).Status(202)                  // f2
}
