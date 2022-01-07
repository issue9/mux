// SPDX-License-Identifier: MIT

package group

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"

	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/internal/syntax"
)

var _ http.Handler = &Group{}

func newRouter(a *assert.Assertion, name string) *mux.Router {
	a.TB().Helper()

	r := mux.NewRouter(name, nil)
	a.NotNil(r)
	return r
}

func TestGroup_Add(t *testing.T) {
	a := assert.New(t, false)

	g := New(nil)

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
		g.New("host", &PathVersion{}, nil)
	}, "已经存在名为 host 的路由")

	a.Nil(g.Router("not-exists"))
}

func TestGroup_Remove(t *testing.T) {
	a := assert.New(t, false)
	g := New(nil)
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

func TestGroup_empty(t *testing.T) {
	a := assert.New(t, false)
	g := New(nil)
	a.NotNil(g)

	rest.NewRequest(a, http.MethodGet, "/path").Do(g).Status(http.StatusNotFound)
}

func TestGroup(t *testing.T) {
	a := assert.New(t, false)
	g := New(nil, mux.Interceptor(mux.InterceptorDigit, "digit"))
	exit := make(chan bool, 1)

	h := NewHosts(true, "{sub}.example.com")
	a.NotNil(h)
	def := g.New("host", h, nil)
	a.NotNil(def)

	def.Get("/posts/{id:digit}.html", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ps := syntax.GetParams(r)
		a.Equal(ps.MustString("sub", "not-found"), "abc").
			Equal(ps.MustInt("id", -1), 5)
		w.WriteHeader(http.StatusAccepted)
		exit <- true
	}))

	rest.NewRequest(a, http.MethodGet, "https://abc.example.com/posts/5.html").
		Do(g).
		Status(http.StatusAccepted)
	<-exit
}

func TestGroup_recovery(t *testing.T) {
	a := assert.New(t, false)

	out := new(bytes.Buffer)
	g := New(nil, mux.WriterRecovery(405, out))
	a.NotNil(g)
	h := NewPathVersion("v", "v2")
	a.NotNil(h)
	def := g.New("version", h, nil)
	a.NotNil(def)
	def.Get("/path", http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("test") }))

	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/v2/path", nil)
	a.NotError(err).NotNil(r)
	g.ServeHTTP(w, r)
	a.Equal(w.Code, 405).
		Contains(out.String(), "test")

	out.Reset()
	w = httptest.NewRecorder()
	r, err = http.NewRequest(http.MethodGet, "/v2/path", nil)
	a.NotError(err).NotNil(r)
	g.ServeHTTP(w, r)
	a.Equal(w.Code, 405).
		Contains(out.String(), "test")

	// no recovery

	g = New(nil)
	a.NotNil(g)
	h = NewPathVersion("v", "v2")
	a.NotNil(h)
	def = g.New("version", h, nil)
	a.NotNil(def)
	def.Get("/path", http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("test") }))

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

func TestGroup_routers(t *testing.T) {
	a := assert.New(t, false)
	h := NewHosts(false, "localhost")
	a.NotNil(h)

	g := New(nil)
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

	g = New(nil)
	a.NotNil(g)
	def = newRouter(a, "def")
	g.Add(h, def)
	res := def.Resource("/r1")
	res.Get(rest.BuildHandler(a, 202, "", nil))
	rest.NewRequest(a, http.MethodGet, "/r1").Do(g).Status(404)
	rest.NewRequest(a, http.MethodGet, "https://localhost/r1").Do(g).Status(202)

	// prefix
	g = New(nil)
	a.NotNil(g)
	def = newRouter(a, "def")
	g.Add(h, def)
	p := def.Prefix("/prefix1")
	p.Get("/p1", rest.BuildHandler(a, 203, "", nil))
	rest.NewRequest(a, http.MethodGet, "/prefix1/p1").Do(g).Status(404)
	rest.NewRequest(a, http.MethodGet, "https://localhost:88/prefix1/p1").Do(g).Status(203)

	// prefix prefix
	g = New(nil)
	a.NotNil(g)
	def = newRouter(a, "def")
	g.Add(h, def)
	p1 := def.Prefix("/prefix1")
	p2 := p1.Prefix("/prefix2")
	p2.Get("/p2", rest.BuildHandler(a, 204, "", nil))

	rest.NewRequest(a, http.MethodGet, "/prefix1/prefix2/p2").Do(g).Status(404)
	rest.NewRequest(a, http.MethodGet, "https://localhost/prefix1/prefix2/p2").Do(g).Status(204)

	// 第二个 Prefix 为域名
	g = New(nil)
	def = newRouter(a, "def")
	g.Add(MatcherFunc(Any), def)
	p1 = def.Prefix("/prefix1")
	p2 = p1.Prefix("example.com")
	p2.Get("/p2", rest.BuildHandler(a, 205, "", nil))
	rest.NewRequest(a, http.MethodGet, "/prefix1example.com/p2").Do(g).Status(205)
}

func TestGroup_routers_multiple(t *testing.T) {
	a := assert.New(t, false)

	g := New(nil)
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
