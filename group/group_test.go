// SPDX-License-Identifier: MIT

package group

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"

	"github.com/issue9/mux/v7"
	"github.com/issue9/mux/v7/types"
)

func call(w http.ResponseWriter, r *http.Request, ps types.Route, h http.Handler) {
	h.ServeHTTP(w, r)
}

func methodNotAllowedBuilder(n types.Node) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}

func optionsHandlerBuilder(n types.Node) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", n.AllowHeader())
	})
}

func newGroup(a *assert.Assertion, o ...mux.Option) *GroupOf[http.Handler] {
	a.TB().Helper()
	g := NewGroupOf(call, http.NotFoundHandler(), methodNotAllowedBuilder, optionsHandlerBuilder, o...)
	a.NotNil(g)
	return g
}

func newRouter(a *assert.Assertion, name string, o ...mux.Option) *mux.RouterOf[http.Handler] {
	a.TB().Helper()
	r := mux.NewRouterOf(name, call, http.NotFoundHandler(), methodNotAllowedBuilder, optionsHandlerBuilder, o...)
	a.NotNil(r)
	return r
}

func TestGroupOf_mergeOption(t *testing.T) {
	a := assert.New(t, false)
	g := newGroup(a)
	a.Equal(0, g.optionsLen).Equal(0, len(g.options))
	o := g.mergeOption(mux.Lock(true))
	a.Equal(1, len(o))

	g = newGroup(a, mux.Lock(true))
	a.Equal(1, g.optionsLen).Equal(1, len(g.options))
	o = g.mergeOption()
	a.Equal(1, len(o))
	o = g.mergeOption(mux.Lock(true), mux.Lock(false))
	a.Equal(3, len(o))
}

func TestGroupOf_Add(t *testing.T) {
	a := assert.New(t, false)
	g := newGroup(a)

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

func TestGroupOf_Remove(t *testing.T) {
	a := assert.New(t, false)
	g := newGroup(a)

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

func TestGroupOf_empty(t *testing.T) {
	a := assert.New(t, false)
	g := newGroup(a)
	rest.NewRequest(a, http.MethodGet, "/path").Do(g).Status(http.StatusNotFound)
}

func TestGroupOf_ServeHTTP(t *testing.T) {
	a := assert.New(t, false)
	rs := newGroup(a)
	exit := make(chan bool, 1)

	h := NewHosts(true, "{sub}.example.com")
	a.NotNil(h)
	def := rs.New("host", h, mux.DigitInterceptor("digit"))
	a.NotNil(def)

	def.Get("/posts/{id:digit}.html", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		exit <- true
	}))

	rest.NewRequest(a, http.MethodGet, "https://abc.example.com/posts/5.html").
		Do(rs).
		Status(http.StatusAccepted)
	<-exit
}

func TestGroupOf_routers(t *testing.T) {
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

func TestGroupOf_routers_multiple(t *testing.T) {
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
