// SPDX-License-Identifier: MIT

package group

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"

	"github.com/issue9/mux/v5"
	"github.com/issue9/mux/v5/internal/syntax"
)

var _ http.Handler = &Group{}

func buildMiddleware(a *assert.Assertion, text string) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r) // 先输出被包含的内容
			_, err := w.Write([]byte(text))
			a.NotError(err)
		})
	}
}

func newRouter(a *assert.Assertion, name string) *mux.Router {
	a.TB().Helper()

	r := mux.NewRouter(name)
	a.NotNil(r)
	return r
}

func TestGroups_PrependMiddleware(t *testing.T) {
	a := assert.New(t, false)
	g := New()
	a.NotNil(g)
	def := newRouter(a, "def")
	g.Add(MatcherFunc(Any), def)

	def.Get("/get", rest.BuildHandler(a, 201, "", nil))
	g.Middlewares().Prepend(buildMiddleware(a, "1")).
		Prepend(buildMiddleware(a, "2"))

	rest.NewRequest(a, http.MethodGet, "/get").Do(g).Status(201).StringBody("12")

	// Reset

	g.Middlewares().Reset()
	rest.NewRequest(a, http.MethodGet, "/get").Do(g).Status(201).BodyEmpty()
}

func TestGroups_AppendMiddleware(t *testing.T) {
	a := assert.New(t, false)
	g := New()
	a.NotNil(g)
	def := newRouter(a, "def")
	g.Add(MatcherFunc(Any), def)

	def.Get("/get", rest.BuildHandler(a, 201, "", nil))
	g.Middlewares().Append(buildMiddleware(a, "1")).
		Append(buildMiddleware(a, "2"))

	rest.NewRequest(a, http.MethodGet, "/get").Do(g).Status(201).StringBody("21")

	// Reset

	g.Middlewares().Reset()
	rest.NewRequest(a, http.MethodGet, "/get").Do(g).Status(201).BodyEmpty()
}

func TestGroups_AddMiddleware(t *testing.T) {
	a := assert.New(t, false)
	g := New()
	a.NotNil(g)
	def := newRouter(a, "def")
	g.Add(MatcherFunc(Any), def)

	def.Get("/get", rest.BuildHandler(a, 201, "", nil))
	g.Middlewares().Append(buildMiddleware(a, "p1")).
		Prepend(buildMiddleware(a, "a1")).
		Append(buildMiddleware(a, "p2")).
		Prepend(buildMiddleware(a, "a2"))

	rest.NewRequest(a, http.MethodGet, "/get").Do(g).Status(201).StringBody("p2p1a1a2") // buildHandler 导致顶部的后输出

	// Reset

	g.Middlewares().Reset()
	rest.NewRequest(a, http.MethodGet, "/get").Do(g).Status(201).BodyEmpty()
}

func TestRouter_AddMiddleware(t *testing.T) {
	a := assert.New(t, false)
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

	rest.NewRequest(a, http.MethodGet, "/get").Do(g).Status(201).StringBody("rp2rp1ra1ra2p2p1a1a2") // buildHandler 导致顶部的后输出

	// Reset

	g.Middlewares().Reset()
	rest.NewRequest(a, http.MethodGet, "/get").Do(g).Status(201).StringBody("rp2rp1ra1ra2")

	def.Middlewares().Reset()
	rest.NewRequest(a, http.MethodGet, "/get").Do(g).Status(201).BodyEmpty()
}

func TestGroups_Add(t *testing.T) {
	a := assert.New(t, false)

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
	a := assert.New(t, false)
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
	a := assert.New(t, false)
	g := New()
	a.NotNil(g)

	rest.NewRequest(a, http.MethodGet, "/path").Do(g).Status(http.StatusNotFound)
}

func TestGroup(t *testing.T) {
	a := assert.New(t, false)
	g := New(mux.Interceptor(mux.InterceptorDigit, "digit"))
	exit := make(chan bool, 1)

	h := NewHosts(true, "{sub}.example.com")
	a.NotNil(h)
	def := g.New("host", h)
	a.NotNil(def)

	def.GetFunc("/posts/{id:digit}.html", func(w http.ResponseWriter, r *http.Request) {
		ps := syntax.GetParams(r)
		a.Equal(ps.MustString("sub", "not-found"), "abc").
			Equal(ps.MustInt("id", -1), 5)
		w.WriteHeader(http.StatusAccepted)
		exit <- true
	})

	rest.NewRequest(a, http.MethodGet, "https://abc.example.com/posts/5.html").
		Do(g).
		Status(http.StatusAccepted)
	<-exit
}

func TestGroup_recovery(t *testing.T) {
	a := assert.New(t, false)

	out := new(bytes.Buffer)
	g := New(mux.WriterRecovery(405, out))
	a.NotNil(g)
	h := NewPathVersion("v", "v2")
	a.NotNil(h)
	def := g.New("version", h)
	a.NotNil(def)
	def.GetFunc("/path", func(http.ResponseWriter, *http.Request) { panic("test") })

	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/v2/path", nil)
	a.NotError(err).NotNil(r)
	g.ServeHTTP(w, r)
	a.Equal(w.Code, 405).
		Equal(out.String(), "test")

	out.Reset()
	w = httptest.NewRecorder()
	r, err = http.NewRequest(http.MethodGet, "/v2/path", nil)
	a.NotError(err).NotNil(r)
	g.ServeHTTP(w, r)
	a.Equal(w.Code, 405).
		Equal(out.String(), "test")

	// no recovery

	g = New()
	a.NotNil(g)
	h = NewPathVersion("v", "v2")
	a.NotNil(h)
	def = g.New("version", h)
	a.NotNil(def)
	def.GetFunc("/path", func(http.ResponseWriter, *http.Request) { panic("test") })

	a.PanicString(func() {
		w = httptest.NewRecorder()
		r, err = http.NewRequest(http.MethodGet, "/v2/path", nil)
		a.NotError(err).NotNil(r)
		g.ServeHTTP(w, r)
	}, "test")

	a.PanicString(func() {
		w = httptest.NewRecorder()
		r, err = http.NewRequest(http.MethodGet, "/v2/path", nil)
		a.NotError(err).NotNil(r)
		g.ServeHTTP(w, r)
	}, "test")

}

func TestGroups_routers(t *testing.T) {
	a := assert.New(t, false)
	h := NewHosts(false, "localhost")
	a.NotNil(h)

	g := New()
	a.NotNil(g)
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

	g = New()
	a.NotNil(g)
	def = newRouter(a, "def")
	g.Add(h, def)
	res := def.Resource("/r1")
	res.Get(rest.BuildHandler(a, 202, "", nil))
	rest.NewRequest(a, http.MethodGet, "/r1").Do(g).Status(404)
	rest.NewRequest(a, http.MethodGet, "https://localhost/r1").Do(g).Status(202)

	// prefix
	g = New()
	a.NotNil(g)
	def = newRouter(a, "def")
	g.Add(h, def)
	p := def.Prefix("/prefix1")
	p.Get("/p1", rest.BuildHandler(a, 203, "", nil))
	rest.NewRequest(a, http.MethodGet, "/prefix1/p1").Do(g).Status(404)
	rest.NewRequest(a, http.MethodGet, "https://localhost:88/prefix1/p1").Do(g).Status(203)

	// prefix prefix
	g = New()
	a.NotNil(g)
	def = newRouter(a, "def")
	g.Add(h, def)
	p1 := def.Prefix("/prefix1")
	p2 := p1.Prefix("/prefix2")
	p2.GetFunc("/p2", rest.BuildHandlerFunc(a, 204, "", nil))

	rest.NewRequest(a, http.MethodGet, "/prefix1/prefix2/p2").Do(g).Status(404)
	rest.NewRequest(a, http.MethodGet, "https://localhost/prefix1/prefix2/p2").Do(g).Status(204)

	// 第二个 Prefix 为域名
	g = New()
	def = newRouter(a, "def")
	g.Add(MatcherFunc(Any), def)
	p1 = def.Prefix("/prefix1")
	p2 = p1.Prefix("example.com")
	p2.GetFunc("/p2", rest.BuildHandlerFunc(a, 205, "", nil))
	rest.NewRequest(a, http.MethodGet, "/prefix1example.com/p2").Do(g).Status(205)
}

func TestGroups_routers_multiple(t *testing.T) {
	a := assert.New(t, false)

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
	rest.NewRequest(a, http.MethodGet, "https://localhost/t1").Do(g).Status(201)

	// 同时匹配 v1、v2，指向 v1
	rest.NewRequest(a, http.MethodGet, "https://localhost/v1/path").Do(g).Status(202)

	// 指向 v2
	rest.NewRequest(a, http.MethodGet, "https://localhost/v2/path").Do(g).Status(203)

	// 指向 v2
	rest.NewRequest(a, http.MethodGet, "https://example.com/v2/path").Do(g).Status(203)
}
