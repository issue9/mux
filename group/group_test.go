// SPDX-License-Identifier: MIT

package group

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v5"
	"github.com/issue9/mux/v5/params"
)

var _ http.Handler = &Groups{}

func buildHandler(code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	})
}

func buildHandlerFunc(code int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	}
}

func buildMiddleware(a *assert.Assertion, text string) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r) // 先输出被包含的内容
			a.NotError(w.Write([]byte(text)))
		})
	}
}

func TestGroups_PrependMiddleware(t *testing.T) {
	a := assert.New(t)
	g := Default()
	a.NotNil(g)
	def := mux.DefaultRouter()
	a.NotError(g.AddRouter("def", MatcherFunc(Any), def))

	def.Get("/get", buildHandler(201))
	g.Middlewares().Prepend(buildMiddleware(a, "1")).
		Prepend(buildMiddleware(a, "2"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/get", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Equal(w.Body.String(), "12")

	// Reset

	g.Middlewares().Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/get", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Empty(w.Body.String())
}

func TestGroups_AppendMiddleware(t *testing.T) {
	a := assert.New(t)
	g := Default()
	a.NotNil(g)
	def := mux.DefaultRouter()
	a.NotError(g.AddRouter("def", MatcherFunc(Any), def))

	def.Get("/get", buildHandler(201))
	g.Middlewares().Append(buildMiddleware(a, "1")).
		Append(buildMiddleware(a, "2"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/get", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Equal(w.Body.String(), "21")

	// Reset

	g.Middlewares().Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/get", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Empty(w.Body.String())
}

func TestGroups_AddMiddleware(t *testing.T) {
	a := assert.New(t)
	g := Default()
	a.NotNil(g)
	def := mux.DefaultRouter()
	a.NotError(g.AddRouter("def", MatcherFunc(Any), def))

	def.Get("/get", buildHandler(201))
	g.Middlewares().Append(buildMiddleware(a, "p1")).
		Prepend(buildMiddleware(a, "a1")).
		Append(buildMiddleware(a, "p2")).
		Prepend(buildMiddleware(a, "a2"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/get", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Equal(w.Body.String(), "p2p1a1a2") // buildHandler 导致顶部的后输出

	// Reset

	g.Middlewares().Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/get", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Empty(w.Body.String())
}

func TestRouter_AddMiddleware(t *testing.T) {
	a := assert.New(t)
	g := Default()

	g.Middlewares().Append(buildMiddleware(a, "p1")).
		Prepend(buildMiddleware(a, "a1")).
		Append(buildMiddleware(a, "p2")).
		Prepend(buildMiddleware(a, "a2"))

	def := mux.DefaultRouter()
	a.NotError(g.AddRouter("def", MatcherFunc(Any), def))
	def.Get("/get", buildHandler(201))
	def.Middlewares().Append(buildMiddleware(a, "rp1")).
		Prepend(buildMiddleware(a, "ra1")).
		Append(buildMiddleware(a, "rp2")).
		Prepend(buildMiddleware(a, "ra2"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/get", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Equal(w.Body.String(), "rp2rp1ra1ra2p2p1a1a2") // buildHandler 导致顶部的后输出

	// Reset

	g.Middlewares().Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/get", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Equal(w.Body.String(), "rp2rp1ra1ra2")

	def.Middlewares().Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/get", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Empty(w.Body.String())
}

func TestGroups_AddRouter(t *testing.T) {
	a := assert.New(t)

	g := Default()

	// name 为空
	a.PanicString(func() {
		g.AddRouter("", &HeaderVersion{}, mux.DefaultRouter())
	}, "name")

	a.PanicString(func() {
		g.AddRouter("name", nil, mux.DefaultRouter())
	}, "matcher")

	a.NotError(g.AddRouter("host", &PathVersion{}, mux.DefaultRouter()))
	r := g.Router("host")
	a.Equal(r.Name(), "host")

	// 同名添加不成功
	a.ErrorIs(g.AddRouter("host", &PathVersion{}, mux.DefaultRouter()), ErrRouterExists)
	rr, err := g.NewRouter("host", &PathVersion{})
	a.ErrorIs(err, ErrRouterExists).Nil(rr)

	a.Nil(g.Router("not-exists"))
}

func TestGroups_RemoveRouter(t *testing.T) {
	a := assert.New(t)
	g := Default()
	a.NotNil(g)

	a.NotError(g.AddRouter("host", &PathVersion{}, mux.DefaultRouter()))
	a.NotError(g.AddRouter("host-2", &PathVersion{}, mux.DefaultRouter()))

	g.RemoveRouter("host")
	g.RemoveRouter("host") // 已经删除，不存在了
	a.Equal(1, len(g.Routers()))
	a.NotError(g.AddRouter("host", &PathVersion{}, mux.DefaultRouter()))
	a.Equal(2, len(g.Routers()))

	// 删除空名，不出错。
	g.RemoveRouter("")
	a.Equal(2, len(g.Routers()))
}

func TestGroups_empty(t *testing.T) {
	a := assert.New(t)
	g := Default()
	a.NotNil(g)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusNotFound)
}

func TestGroup(t *testing.T) {
	a := assert.New(t)
	g := Default()
	exit := make(chan bool, 1)

	h, err := NewHosts("{sub}.example.com")
	a.NotError(err).NotNil(h)

	def, err := g.NewRouter("host", h)
	a.NotError(err).NotNil(def)

	def.GetFunc("/posts/{id:digit}.html", func(w http.ResponseWriter, r *http.Request) {
		ps := params.Get(r)
		a.Equal(ps.MustString("sub", "not-found"), "abc").
			Equal(ps.MustInt("id", -1), 5)
		w.WriteHeader(http.StatusAccepted)
		exit <- true
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "https://abc.example.com/posts/5.html", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusAccepted)
	<-exit
}

func TestGroups_routers(t *testing.T) {
	a := assert.New(t)
	h, err := NewHosts("localhost")
	a.NotError(err).NotNil(h)

	g := Default()
	a.NotNil(g)

	def := mux.DefaultRouter()
	a.NotError(g.AddRouter("host", h, def))
	w := httptest.NewRecorder()
	def.Get("/t1", buildHandler(201))
	r := httptest.NewRequest(http.MethodGet, "/t1", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://localhost/t1", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 201)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://localhost/t1", nil)
	g.ServeHTTP(w, r) // 由 h 直接访问
	a.Equal(w.Result().StatusCode, 201)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/t1", nil)
	g.ServeHTTP(w, r) // 由 h 直接访问
	a.Equal(w.Result().StatusCode, 404)

	// resource
	g = Default()
	a.NotNil(g)
	def = mux.DefaultRouter()
	a.NotError(g.AddRouter("def", h, def))
	res := def.Resource("/r1")
	res.Get(buildHandler(202))
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/r1", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://localhost/r1", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 202)

	// prefix
	g = Default()
	a.NotNil(g)
	def = mux.DefaultRouter()
	a.NotError(g.AddRouter("def", h, def))
	p := def.Prefix("/prefix1")
	p.Get("/p1", buildHandler(203))
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/prefix1/p1", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://localhost:88/prefix1/p1", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 203)

	// prefix prefix
	g = Default()
	a.NotNil(g)
	def = mux.DefaultRouter()
	a.NotError(g.AddRouter("def", h, def))
	p1 := def.Prefix("/prefix1")
	p2 := p1.Prefix("/prefix2")
	p2.GetFunc("/p2", buildHandlerFunc(204))
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/prefix1/prefix2/p2", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://localhost/prefix1/prefix2/p2", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 204)

	// 第二个 Prefix 为域名
	g = Default()
	def = mux.DefaultRouter()
	a.NotError(g.AddRouter("def", MatcherFunc(Any), def))
	p1 = def.Prefix("/prefix1")
	p2 = p1.Prefix("example.com")
	p2.GetFunc("/p2", buildHandlerFunc(205))
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/prefix1example.com/p2", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 205)
}

func TestGroups_routers_multiple(t *testing.T) {
	a := assert.New(t)

	g := Default()
	a.NotNil(g)

	v1 := mux.DefaultRouter()
	a.NotError(g.AddRouter("v1", NewPathVersion("", "v1"), v1))
	v1.Get("/path", buildHandler(202))

	v2 := mux.DefaultRouter()
	a.NotError(g.AddRouter("v2", NewPathVersion("", "v1", "v2"), v2))
	v2.Get("/path", buildHandler(203))

	// def 匹配任意内容，放在最后。
	def := mux.DefaultRouter()
	a.NotError(g.AddRouter("default", MatcherFunc(Any), def))
	def.Get("/t1", buildHandler(201))

	// 指向 def
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "https://localhost/t1", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 201)

	// 同时匹配 v1、v2，指向 v1
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://localhost/v1/path", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 202)

	// 指向 v2
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://localhost/v2/path", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 203)

	// 指向 v2
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://example.com/v2/path", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 203)
}
