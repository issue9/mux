// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/assert/v4/rest"

	"github.com/issue9/mux/v9/internal/tree"
)

func newGroup(a *assert.Assertion, o ...Option) *Group[http.Handler] {
	a.TB().Helper()
	g := NewGroup(call, http.NotFoundHandler(), methodNotAllowedBuilder, optionsHandlerBuilder, o...)
	a.NotNil(g)
	return g
}

func TestGroup_Use(t *testing.T) {
	a := assert.New(t, false)
	g := newGroup(a)

	h1 := g.New("h1", NewHosts(false, "h1.example.com"))
	h1.Get("/posts/5.html", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("h1")) }))

	g.Use(tree.BuildTestMiddleware(a, "m1"))
	h2 := g.New("h2", NewHosts(false, "h2.example.com"))
	h2.Get("/posts/5.html", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("h2")) }))

	rest.NewRequest(a, http.MethodGet, "https://h1.example.com/posts/5.html").
		Do(g).
		Status(http.StatusOK).
		StringBody("h1m1")
	rest.NewRequest(a, http.MethodGet, "https://h2.example.com/posts/5.html").
		Do(g).
		Status(http.StatusOK).
		StringBody("h2m1")

	// h2 中的 notFound
	rest.NewRequest(a, http.MethodGet, "https://h2.example.com/not-exists").
		Do(g).
		Status(http.StatusNotFound).
		StringBody("404 page not found\nm1")
	// group.notFound
	rest.NewRequest(a, http.MethodGet, "https://not-match.example.com/posts/5.html").
		Do(g).
		Status(http.StatusNotFound).
		StringBody("404 page not found\nm1")

	// 添加了新的中间件

	g.Use(tree.BuildTestMiddleware(a, "m2"))

	rest.NewRequest(a, http.MethodGet, "https://h1.example.com/posts/5.html").
		Do(g).
		Status(http.StatusOK).
		StringBody("h1m1m2")
	rest.NewRequest(a, http.MethodGet, "https://h2.example.com/posts/5.html").
		Do(g).
		Status(http.StatusOK).
		StringBody("h2m1m2")

	// h2 中的 notFound
	rest.NewRequest(a, http.MethodGet, "https://h2.example.com/not-exists").
		Do(g).
		Status(http.StatusNotFound).
		StringBody("404 page not found\nm1m2")
	// group.notFound
	rest.NewRequest(a, http.MethodGet, "https://not-match.example.com/posts/5.html").
		Do(g).
		Status(http.StatusNotFound).
		StringBody("404 page not found\nm1m2")
}

func TestGroup_Add(t *testing.T) {
	a := assert.New(t, false)
	g := newGroup(a)

	def := newRouter(a, "host")
	g.Add(&pathVersion{}, def)
	r := g.Router("host")
	a.Equal(r.Name(), "host")

	// 同名添加不成功
	def = newRouter(a, "host")
	a.PanicString(func() {
		g.Add(&pathVersion{}, def)
	}, "已经存在名为 host 的路由")
	a.PanicString(func() {
		g.New("host", &pathVersion{})
	}, "已经存在名为 host 的路由")

	a.Nil(g.Router("not-exists"))
}

func TestGroup_Remove(t *testing.T) {
	a := assert.New(t, false)
	g := newGroup(a)

	def := newRouter(a, "host")
	g.Add(&pathVersion{}, def)
	def = newRouter(a, "host-2")
	g.Add(&pathVersion{}, def)

	g.Remove("host")
	g.Remove("host") // 已经删除，不存在了
	a.Equal(1, len(g.Routers()))
	def = newRouter(a, "host")
	g.Add(&pathVersion{}, def)
	a.Equal(2, len(g.Routers()))

	// 删除空名，不出错。
	g.Remove("")
	a.Equal(2, len(g.Routers()))
}

func TestGroup_empty(t *testing.T) {
	a := assert.New(t, false)
	g := newGroup(a)
	rest.NewRequest(a, http.MethodGet, "/path").Do(g).Status(http.StatusNotFound)
}

func TestGroup_ServeHTTP(t *testing.T) {
	a := assert.New(t, false)
	g := newGroup(a)

	h := NewHosts(true, "{sub}.example.com")
	a.NotNil(h)
	def := g.New("host", h, WithDigitInterceptor("digit"))
	a.NotNil(def)

	def.Get("/posts/{id:digit}.html", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))

	rest.NewRequest(a, http.MethodGet, "https://abc.example.com/posts/5.html").
		Do(g).
		Status(http.StatusAccepted)
}

func TestGroup_routers(t *testing.T) {
	a := assert.New(t, false)
	h := NewHosts(false, "localhost")
	a.NotNil(h)

	g := newGroup(a)
	def := newRouter(a, "host")
	g.Add(h, def)
	def.Get("/t1", rest.BuildHandler(a, 201, "", nil))

	// g 限制了域名
	rest.NewRequest(a, http.MethodGet, "/t1").Do(g).Status(404)
	rest.NewRequest(a, http.MethodGet, "https://localhost/t1").Do(g).Status(201)

	// def 本身不限制域名
	rest.NewRequest(a, http.MethodGet, "https://localhost/t1").Do(def).Status(201)
	rest.NewRequest(a, http.MethodGet, "https://example.com/t1").Do(def).Status(201)
	rest.NewRequest(a, http.MethodGet, "/t1").Do(def).Status(201)

	// resource

	g = newGroup(a)
	def = newRouter(a, "def")
	g.Add(h, def)
	res := def.Resource("/r1")
	res.Get(rest.BuildHandler(a, 202, "", nil))
	rest.NewRequest(a, http.MethodGet, "/r1").Do(g).Status(404)
	rest.NewRequest(a, http.MethodGet, "https://localhost/r1").Do(g).Status(202)

	// prefix
	g = newGroup(a)
	def = newRouter(a, "def")
	g.Add(h, def)
	p := def.Prefix("/prefix1")
	p.Get("/p1", rest.BuildHandler(a, 203, "", nil))
	rest.NewRequest(a, http.MethodGet, "/prefix1/p1").Do(g).Status(404)
	rest.NewRequest(a, http.MethodGet, "https://localhost:88/prefix1/p1").Do(g).Status(203)

	// prefix prefix
	g = newGroup(a)
	def = newRouter(a, "def")
	g.Add(h, def)
	p1 := def.Prefix("/prefix1")
	p2 := p1.Prefix("/prefix2")
	p2.Get("/p2", rest.BuildHandler(a, 204, "", nil))

	rest.NewRequest(a, http.MethodGet, "/prefix1/prefix2/p2").Do(g).Status(404)
	rest.NewRequest(a, http.MethodGet, "https://localhost/prefix1/prefix2/p2").Do(g).Status(204)

	// 第二个 Prefix 为域名
	g = newGroup(a)
	def = newRouter(a, "def")
	g.Add(nil, def)
	p1 = def.Prefix("/prefix1")
	p2 = p1.Prefix("example.com")
	p2.Get("/p2", rest.BuildHandler(a, 205, "", nil))
	rest.NewRequest(a, http.MethodGet, "/prefix1example.com/p2").Do(g).Status(205)
}

func TestGroup_routers_multiple(t *testing.T) {
	a := assert.New(t, false)
	g := newGroup(a)

	v1 := newRouter(a, "v1")
	g.Add(NewPathVersion("", "v1"), v1)
	v1.Get("/path", rest.BuildHandler(a, 202, "", nil))

	v2 := newRouter(a, "v2")
	g.Add(NewPathVersion("", "v1", "v2"), v2)
	v2.Get("/path", rest.BuildHandler(a, 203, "", nil))

	// def 匹配任意内容，放在最后。
	def := newRouter(a, "default")
	g.Add(nil, def)
	def.Get("/t1", rest.BuildHandler(a, 201, "", nil))

	a.Equal(g.Routes(), map[string]map[string][]string{
		"v1": {
			"*":     {http.MethodOptions},
			"/path": {http.MethodGet, http.MethodHead, http.MethodOptions},
		},
		"v2": {
			"*":     {http.MethodOptions},
			"/path": {http.MethodGet, http.MethodHead, http.MethodOptions},
		},
		"default": {
			"*":   {http.MethodOptions},
			"/t1": {http.MethodGet, http.MethodHead, http.MethodOptions},
		},
	})

	// 指向 def
	rest.NewRequest(a, http.MethodGet, "https://localhost/t1").Do(g).Status(201)

	// 同时匹配 v1、v2，指向 v1
	rest.NewRequest(a, http.MethodGet, "https://localhost/v1/path").Do(g).Status(202)

	// 指向 v2
	rest.NewRequest(a, http.MethodGet, "https://localhost/v2/path").Do(g).Status(203)

	// 指向 v2
	rest.NewRequest(a, http.MethodGet, "https://example.com/v2/path").Do(g).Status(203)
}
