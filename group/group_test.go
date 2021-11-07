// SPDX-License-Identifier: MIT

package group

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"

	"github.com/issue9/mux/v5"
	"github.com/issue9/mux/v5/params"
)

var _ http.Handler = &Group{}

func buildMiddleware(a *assert.Assertion, text string) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r) // 先输出被包含的内容
			a.NotError(w.Write([]byte(text)))
		})
	}
}

func newRouter(a *assert.Assertion, name string) *mux.Router {
	r := mux.NewRouter(name)
	a.NotNil(r)
	return r
}

func TestGroups_PrependMiddleware(t *testing.T) {
	a := assert.New(t)
	g := New()
	a.NotNil(g)
	def := newRouter(a, "def")
	g.Add(MatcherFunc(Any), def)

	def.Get("/get", rest.BuildHandler(a, 201, "", nil))
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
	g := New()
	a.NotNil(g)
	def := newRouter(a, "def")
	g.Add(MatcherFunc(Any), def)

	def.Get("/get", rest.BuildHandler(a, 201, "", nil))
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
	g := New()
	a.NotNil(g)
	def := newRouter(a, "def")
	g.Add(MatcherFunc(Any), def)

	def.Get("/get", rest.BuildHandler(a, 201, "", nil))
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
	g := New()

	g.Middlewares().Append(buildMiddleware(a, "p1")).
		Prepend(buildMiddleware(a, "a1")).
		Append(buildMiddleware(a, "p2")).
		Prepend(buildMiddleware(a, "a2"))

	def := newRouter(a, "def")
	g.Add(MatcherFunc(Any), def)
	def.Get("/get", rest.BuildHandler(a, 201, "", nil))
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

func TestGroups_Add(t *testing.T) {
	a := assert.New(t)

	g := New()

	// name 为空
	a.PanicString(func() {
		def := newRouter(a, "")
		g.Add(&HeaderVersion{}, def)
	}, "r.Name() 不能为空")

	a.PanicString(func() {
		def := newRouter(a, "def")
		g.Add(nil, def)
	}, "matcher")

	def := newRouter(a, "host")
	g.Add(&PathVersion{}, def)
	r := g.Router("host")
	a.Equal(r.Name(), "host")

	// 同名添加不成功
	def = newRouter(a, "host")
	a.PanicString(func() {
		g.Add(&PathVersion{}, def)
	}, "已经存在名为 host 的路由")
	a.PanicString(func() {
		g.New("host", &PathVersion{})
	}, "已经存在名为 host 的路由")

	a.Nil(g.Router("not-exists"))
}

func TestGroups_Remove(t *testing.T) {
	a := assert.New(t)
	g := New()
	a.NotNil(g)

	def := newRouter(a, "host")
	g.Add(&PathVersion{}, def)
	def = newRouter(a, "host-2")
	g.Add(&PathVersion{}, def)

	g.Remove("host")
	g.Remove("host") // 已经删除，不存在了
	a.Equal(1, len(g.Routers()))
	def = newRouter(a, "host")
	g.Add(&PathVersion{}, def)
	a.Equal(2, len(g.Routers()))

	// 删除空名，不出错。
	g.Remove("")
	a.Equal(2, len(g.Routers()))
}

func TestGroups_empty(t *testing.T) {
	a := assert.New(t)
	g := New()
	a.NotNil(g)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusNotFound)
}

func TestGroup(t *testing.T) {
	a := assert.New(t)
	g := New()
	exit := make(chan bool, 1)

	h := NewHosts(true, "{sub}.example.com")
	a.NotNil(h)

	def := g.New("host", h)
	a.NotNil(def)

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
	h := NewHosts(false, "localhost")
	a.NotNil(h)

	g := New()
	a.NotNil(g)

	def := newRouter(a, "host")
	g.Add(h, def)
	w := httptest.NewRecorder()
	def.Get("/t1", rest.BuildHandler(a, 201, "", nil))
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
	g = New()
	a.NotNil(g)
	def = newRouter(a, "def")
	g.Add(h, def)
	res := def.Resource("/r1")
	res.Get(rest.BuildHandler(a, 202, "", nil))
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/r1", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://localhost/r1", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 202)

	// prefix
	g = New()
	a.NotNil(g)
	def = newRouter(a, "def")
	g.Add(h, def)
	p := def.Prefix("/prefix1")
	p.Get("/p1", rest.BuildHandler(a, 203, "", nil))
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/prefix1/p1", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://localhost:88/prefix1/p1", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 203)

	// prefix prefix
	g = New()
	a.NotNil(g)
	def = newRouter(a, "def")
	g.Add(h, def)
	p1 := def.Prefix("/prefix1")
	p2 := p1.Prefix("/prefix2")
	p2.GetFunc("/p2", rest.BuildHandlerFunc(a, 204, "", nil))
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/prefix1/prefix2/p2", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://localhost/prefix1/prefix2/p2", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 204)

	// 第二个 Prefix 为域名
	g = New()
	def = newRouter(a, "def")
	g.Add(MatcherFunc(Any), def)
	p1 = def.Prefix("/prefix1")
	p2 = p1.Prefix("example.com")
	p2.GetFunc("/p2", rest.BuildHandlerFunc(a, 205, "", nil))
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/prefix1example.com/p2", nil)
	g.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 205)
}

func TestGroups_routers_multiple(t *testing.T) {
	a := assert.New(t)

	g := New()
	a.NotNil(g)

	v1 := newRouter(a, "v1")
	g.Add(NewPathVersion("", "v1"), v1)
	v1.Get("/path", rest.BuildHandler(a, 202, "", nil))

	v2 := newRouter(a, "v2")
	g.Add(NewPathVersion("", "v1", "v2"), v2)
	v2.Get("/path", rest.BuildHandler(a, 203, "", nil))

	// def 匹配任意内容，放在最后。
	def := newRouter(a, "default")
	g.Add(MatcherFunc(Any), def)
	def.Get("/t1", rest.BuildHandler(a, 201, "", nil))

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
