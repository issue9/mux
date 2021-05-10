// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v4/group"
	"github.com/issue9/mux/v4/interceptor"
	"github.com/issue9/mux/v4/internal/handlers"
)

func TestMux_NewRouter(t *testing.T) {
	a := assert.New(t)

	m := Default()

	// group.Matcher 为空
	a.Panic(func() {
		m.NewRouter("v1", nil)
	})

	// name 为空
	a.Panic(func() {
		m.NewRouter("", group.NewHosts("example.com"))
	})

	r, ok := m.NewRouter("host", group.NewHosts())
	a.True(ok).NotNil(r)
	a.Equal(r.name, "host").Equal(r.Name(), "host")

	r, ok = m.NewRouter("host", group.NewHosts())
	a.False(ok).Nil(r)

	r, ok = m.NewRouter("host-2", group.NewHosts())
	a.True(ok).NotNil(r)
	a.Equal(r.name, "host-2").Equal(r.Name(), "host-2")
}

func TestMux_RemoveRouter(t *testing.T) {
	a := assert.New(t)

	m := Default()
	r, ok := m.NewRouter("host", group.NewHosts())
	a.True(ok).NotNil(r)
	a.Equal(r.name, "host").Equal(r.Name(), "host")

	r, ok = m.NewRouter("host-2", group.NewHosts())
	a.True(ok).NotNil(r)
	a.Equal(2, len(m.Routers()))

	// 同名，添加不成功
	r, ok = m.NewRouter("host", group.NewHosts())
	a.False(ok).Nil(r)
	a.Equal(2, len(m.Routers()))

	m.RemoveRouter("host")
	m.RemoveRouter("host") // 已经删除，不存在了
	a.Equal(1, len(m.Routers()))
	r, ok = m.NewRouter("host", group.NewHosts())
	a.True(ok).NotNil(r)
	a.Equal(2, len(m.Routers()))

	// 删除空名，不出错。
	m.RemoveRouter("")
	a.Equal(2, len(m.Routers()))
}

func TestRouter_routers(t *testing.T) {
	a := assert.New(t)

	m := Default()
	def, ok := m.NewRouter("host", group.NewHosts("localhost"))
	a.True(ok).NotNil(def)
	w := httptest.NewRecorder()
	def.Get("/t1", buildHandler(201))
	r := httptest.NewRequest(http.MethodGet, "/t1", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://localhost/t1", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 201)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://localhost/t1", nil)
	m.ServeHTTP(w, r) // 由 h 直接访问
	a.Equal(w.Result().StatusCode, 201)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/t1", nil)
	m.ServeHTTP(w, r) // 由 h 直接访问
	a.Equal(w.Result().StatusCode, 404)

	// resource
	m = Default()
	a.NotNil(m)
	def, ok = m.NewRouter("def", group.NewHosts("localhost"))
	a.True(ok).NotNil(def)
	res := def.Resource("/r1")
	res.Get(buildHandler(202))
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/r1", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://localhost/r1", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 202)

	// prefix
	m = Default()
	a.NotNil(m)
	def, ok = m.NewRouter("def", group.NewHosts("localhost"))
	a.True(ok).NotNil(def)
	p := def.Prefix("/prefix1")
	p.Get("/p1", buildHandler(203))
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/prefix1/p1", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://localhost:88/prefix1/p1", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 203)

	// prefix prefix
	m = New(false, false, false, nil, nil)
	a.NotNil(m)
	def, ok = m.NewRouter("def", group.NewHosts("localhost"))
	a.True(ok).NotNil(def)
	p1 := def.Prefix("/prefix1")
	p2 := p1.Prefix("/prefix2")
	p2.GetFunc("/p2", buildHandlerFunc(204))
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/prefix1/prefix2/p2", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://localhost/prefix1/prefix2/p2", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 204)

	// 第二个 Prefix 为域名
	m = Default()
	def, ok = m.NewRouter("def", group.MatcherFunc(group.Any))
	a.True(ok).NotNil(def)
	p1 = def.Prefix("/prefix1")
	p2 = p1.Prefix("example.com")
	p2.GetFunc("/p2", buildHandlerFunc(205))
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/prefix1example.com/p2", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 205)
}

func TestRouter_routers_multiple(t *testing.T) {
	a := assert.New(t)

	m := New(false, false, false, nil, nil)
	a.NotNil(m)
	def, ok := m.NewRouter("default", group.NewHosts("localhost"))
	a.True(ok).NotNil(def)
	def.Get("/t1", buildHandler(201))

	v1, ok := m.NewRouter("v1", group.NewPathVersion("v1"))
	a.True(ok).NotNil(v1)
	v1.Get("/path", buildHandler(202))
	v2, ok := m.NewRouter("v2", group.NewPathVersion("v1", "v2"))
	a.True(ok).NotNil(v2)
	v2.Get("/path", buildHandler(203))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "https://localhost/t1", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 201)

	// 永远指向 def
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://localhost/v1/path", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)

	// 永远指向 def
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://localhost/v2/path", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)

	// 不匹配 def，转向 v2
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://example.com/v2/path", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 203)
}

func TestMux_ServeHTTP(t *testing.T) {
	a := assert.New(t)

	test := newTester(t, false, true, false)

	a.NotError(test.router.Handle("/posts/{path}.html", buildHandler(201)))
	test.matchTrue(http.MethodGet, "/posts/2017/1.html", 201)

	a.NotError(test.router.Handle("/posts/{path:.+}.html", buildHandler(202)))
	test.matchTrue(http.MethodGet, "/posts/2017/1.html", 202)

	a.NotError(test.router.Handle("/posts/{id:digit}123", buildHandler(203)))
	test.matchTrue(http.MethodGet, "/posts/123123", 203)
}

// 测试匹配顺序是否正确
func TestMux_ServeHTTP_Order(t *testing.T) {
	a := assert.New(t)

	test := newTester(t, false, true, false)
	a.NotError(test.router.GetFunc("/posts/{id}", buildHandlerFunc(203)))        // f3
	a.NotError(test.router.GetFunc("/posts/{id:\\d+}", buildHandlerFunc(202)))   // f2
	a.NotError(test.router.GetFunc("/posts/1", buildHandlerFunc(201)))           // f1
	a.NotError(test.router.GetFunc("/posts/{id:[0-9]+}", buildHandlerFunc(199))) // f0 两个正则，后添加的永远匹配不到
	test.matchTrue(http.MethodGet, "/posts/1", 201)                              // f1 普通路由项完全匹配
	test.matchTrue(http.MethodGet, "/posts/2", 202)                              // f1 正则路由
	test.matchTrue(http.MethodGet, "/posts/abc", 203)                            // f3 命名路由
	test.matchTrue(http.MethodGet, "/posts/", 203)                               // f3

	// interceptor
	test = newTester(t, false, true, false)
	a.NotError(interceptor.Register(interceptor.MatchDigit, "[0-9]+"))
	a.NotError(test.router.GetFunc("/posts/{id}", buildHandlerFunc(203)))        // f3
	a.NotError(test.router.GetFunc("/posts/{id:\\d+}", buildHandlerFunc(202)))   // f2 永远匹配不到
	a.NotError(test.router.GetFunc("/posts/1", buildHandlerFunc(201)))           // f1
	a.NotError(test.router.GetFunc("/posts/{id:[0-9]+}", buildHandlerFunc(210))) // f0 interceptor 权限比正则要高
	test.matchTrue(http.MethodGet, "/posts/1", 201)                              // f1 普通路由项完全匹配
	test.matchTrue(http.MethodGet, "/posts/2", 210)                              // f1 interceptor
	test.matchTrue(http.MethodGet, "/posts/abc", 203)                            // f3 命名路由
	test.matchTrue(http.MethodGet, "/posts/", 203)                               // f3
	interceptor.Deregister("[0-9]+")

	test = newTester(t, false, true, false)
	a.NotError(test.router.GetFunc("/p1/{p1}/p2/{p2:\\d+}", buildHandlerFunc(201))) // f1
	a.NotError(test.router.GetFunc("/p1/{p1}/p2/{p2:\\w+}", buildHandlerFunc(202))) // f2
	test.matchTrue(http.MethodGet, "/p1/1/p2/1", 201)                               // f1
	test.matchTrue(http.MethodGet, "/p1/2/p2/s", 202)                               // f2

	test = newTester(t, false, true, false)
	a.NotError(test.router.GetFunc("/posts/{id}/{page}", buildHandlerFunc(202))) // f2
	a.NotError(test.router.GetFunc("/posts/{id}/1", buildHandlerFunc(201)))      // f1
	test.matchTrue(http.MethodGet, "/posts/1/1", 201)                            // f1 普通路由项完全匹配
	test.matchTrue(http.MethodGet, "/posts/2/5", 202)                            // f2 命名完全匹配

	test = newTester(t, false, true, false)
	a.NotError(test.router.GetFunc("/tags/{id}.html", buildHandlerFunc(201))) // f1
	a.NotError(test.router.GetFunc("/tags.html", buildHandlerFunc(202)))      // f2
	a.NotError(test.router.GetFunc("/{path}", buildHandlerFunc(203)))         // f3
	test.matchTrue(http.MethodGet, "/tags", 203)                              // f3 // 正好与 f1 的第一个节点匹配
	test.matchTrue(http.MethodGet, "/tags/1.html", 201)                       // f1
	test.matchTrue(http.MethodGet, "/tags.html", 202)                         // f2
}

func TestMethods(t *testing.T) {
	a := assert.New(t)
	a.Equal(Methods(), handlers.Methods)
}

func TestIsWell(t *testing.T) {
	a := assert.New(t)

	a.Error(IsWell("/{path"))
	a.Error(IsWell("/path}"))
	a.Error(IsWell(""))
}

func TestClearPath(t *testing.T) {
	a := assert.New(t)

	a.Equal(cleanPath(""), "/")

	a.Equal(cleanPath("/api//"), "/api/")
	a.Equal(cleanPath("api/"), "/api/")
	a.Equal(cleanPath("api/////"), "/api/")
	a.Equal(cleanPath("//api/////1"), "/api/1")

	a.Equal(cleanPath("/api/"), "/api/")
	a.Equal(cleanPath("/api/./"), "/api/./")

	a.Equal(cleanPath("/api/.."), "/api/..")
	a.Equal(cleanPath("/api/../"), "/api/../")
	a.Equal(cleanPath("/api/../../"), "/api/../../")
}
