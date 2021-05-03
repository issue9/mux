// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"

	"github.com/issue9/mux/v4/group"
)

func buildHandler(code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	})
}

func buildFunc(code int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	}
}

// mux 的测试工具
type tester struct {
	router *Router
	srv    *rest.Server
}

func newTester(t testing.TB, disableOptions, disableHead, skipClean bool) *tester {
	router := NewRouter(disableOptions, disableHead, skipClean, nil, nil, "", nil)
	return &tester{
		router: router,
		srv:    rest.NewServer(t, router, nil),
	}
}

// 确保能正常匹配到指定的 URL
func (t *tester) matchTrue(method, path string, code int) {
	t.srv.NewRequest(method, path).Do().Status(code)
}

// 确保能正常匹配到指定的 URL
func (t *tester) matchContent(method, path string, code int, content string) {
	t.srv.NewRequest(method, path).Do().Status(code).StringBody(content)
}

// 确保能正确匹配地址，且拿到正确的 options 报头
func (t *tester) optionsTrue(path string, code int, allow string) {
	t.srv.NewRequest(http.MethodOptions, path).Do().Status(code).Header("Allow", allow)
}

func TestRouter(t *testing.T) {
	test := newTester(t, false, true, false)

	// 测试 / 和 "" 是否访问同一地址
	test.router.Get("/", buildHandler(201))
	test.matchTrue(http.MethodGet, "", 201)
	test.matchTrue(http.MethodGet, "/", 201)
	test.matchTrue(http.MethodHead, "/", http.StatusMethodNotAllowed) // 未启用 autoHead
	test.matchTrue(http.MethodGet, "/abc", http.StatusNotFound)

	test.router.Get("/h/1", buildHandler(201))
	test.matchTrue(http.MethodGet, "/h/1", 201)
	test.router.GetFunc("/f/1", buildFunc(201))
	test.matchTrue(http.MethodGet, "/f/1", 201)

	test.router.Post("/h/1", buildHandler(202))
	test.matchTrue(http.MethodPost, "/h/1", 202)
	test.router.PostFunc("/f/1", buildFunc(202))
	test.matchTrue(http.MethodPost, "/f/1", 202)

	test.router.Put("/h/1", buildHandler(203))
	test.matchTrue(http.MethodPut, "/h/1", 203)
	test.router.PutFunc("/f/1", buildFunc(203))
	test.matchTrue(http.MethodPut, "/f/1", 203)

	test.router.Patch("/h/1", buildHandler(204))
	test.matchTrue(http.MethodPatch, "/h/1", 204)
	test.router.PatchFunc("/f/1", buildFunc(204))
	test.matchTrue(http.MethodPatch, "/f/1", 204)

	test.router.Delete("/h/1", buildHandler(205))
	test.matchTrue(http.MethodDelete, "/h/1", 205)
	test.router.DeleteFunc("/f/1", buildFunc(205))
	test.matchTrue(http.MethodDelete, "/f/1", 205)

	// Any
	test.router.Any("/h/any", buildHandler(206))
	test.matchTrue(http.MethodGet, "/h/any", 206)
	test.matchTrue(http.MethodPost, "/h/any", 206)
	test.matchTrue(http.MethodPut, "/h/any", 206)
	test.matchTrue(http.MethodPatch, "/h/any", 206)
	test.matchTrue(http.MethodDelete, "/h/any", 206)
	test.matchTrue(http.MethodTrace, "/h/any", 206)

	test.router.AnyFunc("/f/any", buildFunc(206))
	test.matchTrue(http.MethodGet, "/f/any", 206)
	test.matchTrue(http.MethodPost, "/f/any", 206)
	test.matchTrue(http.MethodPut, "/f/any", 206)
	test.matchTrue(http.MethodPatch, "/f/any", 206)
	test.matchTrue(http.MethodDelete, "/f/any", 206)
	test.matchTrue(http.MethodTrace, "/f/any", 206)
}

func TestRouter_Routes(t *testing.T) {
	a := assert.New(t)

	m := Default()
	a.NotNil(m).Equal(m.Name(), "")

	m.Get("/m", buildHandler(1))
	m.Post("/m", buildHandler(1))
	a.Equal(m.Routes(false, false), map[string][]string{"/m": {"GET", "HEAD", "OPTIONS", "POST"}})

	r, ok := m.NewRouter("host-1", group.NewHosts())
	a.True(ok).NotNil(r)
	r.Get("/m", buildHandler(1))
	a.Equal(r.Routes(false, false), map[string][]string{"/m": {"GET", "HEAD", "OPTIONS"}})
}

func TestRouter_Head(t *testing.T) {
	test := newTester(t, false, false, false)

	test.router.Get("/", buildHandler(201))
	test.matchTrue(http.MethodGet, "", 201)
	test.matchTrue(http.MethodGet, "/", 201)
	test.matchTrue(http.MethodHead, "", 201)
	test.matchTrue(http.MethodHead, "/", 201)
	test.matchContent(http.MethodHead, "/", 201, "")

	test.router.Get("/h/1", buildHandler(201))
	test.matchTrue(http.MethodGet, "/h/1", 201)
	test.matchTrue(http.MethodHead, "/h/1", 201)
	test.router.GetFunc("/f/1", buildFunc(201))
	test.matchTrue(http.MethodGet, "/f/1", 201)
	test.matchTrue(http.MethodHead, "/f/1", 201)

	test.router.Post("/h/post", buildHandler(202))
	test.matchTrue(http.MethodPost, "/h/post", 202)
	test.matchTrue(http.MethodHead, "/h/post", http.StatusMethodNotAllowed)

	// Any
	test.router.Any("/h/any", buildHandler(206))
	test.matchTrue(http.MethodGet, "/h/any", 206)
	test.matchTrue(http.MethodHead, "/h/any", 206)
	test.matchTrue(http.MethodPost, "/h/any", 206)
	test.matchTrue(http.MethodPut, "/h/any", 206)
	test.matchTrue(http.MethodPatch, "/h/any", 206)
	test.matchTrue(http.MethodDelete, "/h/any", 206)
	test.matchTrue(http.MethodTrace, "/h/any", 206)

	test.router.AnyFunc("/f/any", buildFunc(206))
	test.matchTrue(http.MethodGet, "/f/any", 206)
	test.matchTrue(http.MethodHead, "/f/any", 206)
	test.matchTrue(http.MethodPost, "/f/any", 206)
	test.matchTrue(http.MethodPut, "/f/any", 206)
	test.matchTrue(http.MethodPatch, "/f/any", 206)
	test.matchTrue(http.MethodDelete, "/f/any", 206)
	test.matchTrue(http.MethodTrace, "/f/any", 206)
}

func TestRouter_Handle_Remove(t *testing.T) {
	a := assert.New(t)
	test := newTester(t, false, true, false)

	// 添加 GET /api/1
	// 添加 PUT /api/1
	// 添加 GET /api/2
	a.NotError(test.router.HandleFunc("/api/1", buildFunc(201), http.MethodGet))
	a.NotError(test.router.HandleFunc("/api/1", buildFunc(201), http.MethodPut))
	a.NotError(test.router.HandleFunc("/api/2", buildFunc(202), http.MethodGet))

	test.matchTrue(http.MethodGet, "/api/1", 201)
	test.matchTrue(http.MethodPut, "/api/1", 201)
	test.matchTrue(http.MethodGet, "/api/2", 202)
	test.matchTrue(http.MethodDelete, "/api/1", http.StatusMethodNotAllowed) // 未实现

	// 删除 GET /api/1
	test.router.Remove("/api/1", http.MethodGet)
	test.matchTrue(http.MethodGet, "/api/1", http.StatusMethodNotAllowed)
	test.matchTrue(http.MethodPut, "/api/1", 201) // 不影响 PUT
	test.matchTrue(http.MethodGet, "/api/2", 202)

	// 删除 GET /api/2，只有一个，所以相当于整个节点被删除
	test.router.Remove("/api/2", http.MethodGet)
	test.matchTrue(http.MethodGet, "/api/1", http.StatusMethodNotAllowed)
	test.matchTrue(http.MethodPut, "/api/1", 201)                 // 不影响 PUT
	test.matchTrue(http.MethodGet, "/api/2", http.StatusNotFound) // 整个节点被删除

	// 添加 POST /api/1
	a.NotError(test.router.Handle("/api/1", buildFunc(201), http.MethodPost))
	test.matchTrue(http.MethodPost, "/api/1", 201)

	// 删除 ANY /api/1
	test.router.Remove("/api/1")
	test.matchTrue(http.MethodPost, "/api/1", http.StatusNotFound) // 404 表示整个节点都没了
}

func TestRouter_Options(t *testing.T) {
	a := assert.New(t)
	test := newTester(t, false, true, false)

	// 添加 GET /api/1
	a.NotError(test.router.Handle("/api/1", buildHandler(201), http.MethodGet))
	test.optionsTrue("/api/1", http.StatusOK, "GET, OPTIONS")

	// 添加 DELETE /api/1
	a.NotError(test.router.Handle("/api/1", buildHandler(201), http.MethodDelete))
	test.optionsTrue("/api/1", http.StatusOK, "DELETE, GET, OPTIONS")

	// 删除 DELETE /api/1
	test.router.Remove("/api/1", http.MethodDelete)
	test.optionsTrue("/api/1", http.StatusOK, "GET, OPTIONS")

	// 通过 Options 自定义 Allow 报头
	test.router.Options("/api/1", "CUSTOM OPTIONS1")
	test.optionsTrue("/api/1", http.StatusOK, "CUSTOM OPTIONS1")
	test.router.Options("/api/1", "CUSTOM OPTIONS2")
	test.optionsTrue("/api/1", http.StatusOK, "CUSTOM OPTIONS2")

	a.NotError(test.router.HandleFunc("/api/1", buildFunc(201), http.MethodOptions))
	test.optionsTrue("/api/1", 201, "")

	// disableOptions 为 true
	test = newTester(t, true, true, false)
	test.optionsTrue("/api/1", http.StatusNotFound, "")
	test.router.Options("/api/1", "CUSTOM OPTIONS1") // 显示指定
	test.optionsTrue("/api/1", http.StatusOK, "CUSTOM OPTIONS1")
}

func TestRouter_Params(t *testing.T) {
	a := assert.New(t)
	router := Default()
	a.NotNil(router)
	params := map[string]string{}

	buildParamsHandler := func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ps := Params(r)
			a.NotNil(ps)
			params = ps
		})
	}

	requestParams := func(method, url string, status int, ps map[string]string) {
		w := httptest.NewRecorder()
		a.NotNil(w)

		r, err := http.NewRequest(method, url, nil)
		a.NotError(err).NotNil(r)

		router.ServeHTTP(w, r)

		a.Equal(w.Code, status)
		if ps != nil { // 由于 params 是公用数据，会保存上一次获取的值，所以只在有值时才比较
			a.Equal(params, ps)
		}
		params = nil // 清空全局的 params
	}

	// 添加 patch /api/{version:\\d+}
	a.NotError(router.Patch("/api/{version:\\d+}", buildParamsHandler()))
	requestParams(http.MethodPatch, "/api/2", http.StatusOK, map[string]string{"version": "2"})
	requestParams(http.MethodPatch, "/api/256", http.StatusOK, map[string]string{"version": "256"})
	requestParams(http.MethodGet, "/api/256", http.StatusMethodNotAllowed, nil) // 不存在的请求方法

	// 添加 patch /api/v2/{version:\\d*}
	a.NotError(router.Patch("/api/v2/{version:\\d*}", buildParamsHandler()))
	requestParams(http.MethodPatch, "/api/v2/2", http.StatusOK, map[string]string{"version": "2"})
	requestParams(http.MethodPatch, "/api/v2/", http.StatusOK, map[string]string{"version": ""})

	// 添加 patch /api/v2/{version:\\d+}/test
	a.NotError(router.Patch("/api/v2/{version:\\d*}/test", buildParamsHandler()))
	requestParams(http.MethodPatch, "/api/v2/2/test", http.StatusOK, map[string]string{"version": "2"})
	requestParams(http.MethodPatch, "/api/v2//test", http.StatusNotFound, nil) // 可选参数不能在路由中间

	// 中文作为值
	a.NotError(router.Patch("/api/v3/{版本:digit}", buildParamsHandler()))
	requestParams(http.MethodPatch, "/api/v3/2", http.StatusOK, map[string]string{"版本": "2"})
}

func TestRouter_Clean(t *testing.T) {
	a := assert.New(t)

	m := Default()
	m.Get("/m1", buildHandler(200)).
		Post("/m1", buildHandler(201))
	router, ok := m.NewRouter("host", group.NewHosts("example.com"))
	a.True(ok).NotNil(router)
	router.Get("/m1", buildHandler(202)).
		Post("/m1", buildHandler(203))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/m1", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 200)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://example.com/m1", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 202)

	m.Clean()

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/m1", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)

	// 不清除子路由的项

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://example.com/m1", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 202)
}

func TestRouter_ServeHTTP(t *testing.T) {
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
func TestRouter_ServeHTTP_Order(t *testing.T) {
	a := assert.New(t)
	test := newTester(t, false, true, false)

	a.NotError(test.router.GetFunc("/posts/{id}", buildFunc(203)))      // f3
	a.NotError(test.router.GetFunc("/posts/{id:\\d+}", buildFunc(202))) // f2
	a.NotError(test.router.GetFunc("/posts/1", buildFunc(201)))         // f1
	test.matchTrue(http.MethodGet, "/posts/1", 201)                     // f1 普通路由项完全匹配
	test.matchTrue(http.MethodGet, "/posts/2", 202)                     // f1 正则路由
	test.matchTrue(http.MethodGet, "/posts/abc", 203)                   // f3 命名路由
	test.matchTrue(http.MethodGet, "/posts/", 203)                      // f3

	test = newTester(t, false, true, false)
	a.NotError(test.router.GetFunc("/p1/{p1}/p2/{p2:\\d+}", buildFunc(201))) // f1
	a.NotError(test.router.GetFunc("/p1/{p1}/p2/{p2:\\w+}", buildFunc(202))) // f2
	test.matchTrue(http.MethodGet, "/p1/1/p2/1", 201)                        // f1
	test.matchTrue(http.MethodGet, "/p1/2/p2/s", 202)                        // f2

	test = newTester(t, false, true, false)
	a.NotError(test.router.GetFunc("/posts/{id}/{page}", buildFunc(202))) // f2
	a.NotError(test.router.GetFunc("/posts/{id}/1", buildFunc(201)))      // f1
	test.matchTrue(http.MethodGet, "/posts/1/1", 201)                     // f1 普通路由项完全匹配
	test.matchTrue(http.MethodGet, "/posts/2/5", 202)                     // f2 命名完全匹配

	test = newTester(t, false, true, false)
	a.NotError(test.router.GetFunc("/tags/{id}.html", buildFunc(201))) // f1
	a.NotError(test.router.GetFunc("/tags.html", buildFunc(202)))      // f2
	a.NotError(test.router.GetFunc("/{path}", buildFunc(203)))         // f3
	test.matchTrue(http.MethodGet, "/tags", 203)                       // f3 // 正好与 f1 的第一个节点匹配
	test.matchTrue(http.MethodGet, "/tags/1.html", 201)                // f1
	test.matchTrue(http.MethodGet, "/tags.html", 202)                  // f2
}

func TestRouter_NewRouter(t *testing.T) {
	a := assert.New(t)

	m := Default()
	a.Equal(m.Name(), "") // 默认为空
	r, ok := m.NewRouter("host", group.NewHosts())
	a.True(ok).NotNil(r)
	a.Equal(r.name, "host").Equal(r.disableHead, m.disableHead).Equal(r.Name(), "host")

	r, ok = m.NewRouter("host", group.NewHosts())
	a.False(ok).Nil(r)

	// 空值，与 Default 相同
	r, ok = m.NewRouter("", group.NewHosts())
	a.False(ok).Nil(r)

	r, ok = m.NewRouter("host-2", group.NewHosts())
	a.True(ok).NotNil(r)
	a.Equal(r.name, "host-2").Equal(r.disableHead, m.disableHead).Equal(r.Name(), "host-2")

	// 主 name 不为空

	m = NewRouter(false, false, false, nil, nil, "v1", group.NewVersion(false, "v1"))
	a.NotNil(m).Equal(m.Name(), "v1")

	// 相同名称
	r, ok = m.NewRouter("v1", group.NewHosts())
	a.False(ok).Nil(r)

	// 空值可以
	r, ok = m.NewRouter("", group.NewHosts())
	a.True(ok).NotNil(r)
}

func TestRouter_routers(t *testing.T) {
	a := assert.New(t)

	m := Default()
	router, ok := m.NewRouter("host", group.NewHosts("localhost"))
	a.True(ok).NotNil(router)
	w := httptest.NewRecorder()
	router.Get("/t1", buildHandler(201))
	r := httptest.NewRequest(http.MethodGet, "/t1", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://localhost/t1", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 201)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://localhost/t1", nil)
	router.ServeHTTP(w, r) // 由 h 直接访问
	a.Equal(w.Result().StatusCode, 201)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/t1", nil)
	router.ServeHTTP(w, r) // 由 h 直接访问
	a.Equal(w.Result().StatusCode, 404)

	// resource
	m = Default()
	router, ok = m.NewRouter("host", group.NewHosts("localhost"))
	a.True(ok).NotNil(router)
	res := router.Resource("/r1")
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
	router, ok = m.NewRouter("host", group.NewHosts("localhost"))
	a.True(ok).NotNil(router)
	p := router.Prefix("/prefix1")
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
	m = Default()
	router, ok = m.NewRouter("host", group.NewHosts("localhost"))
	a.True(ok).NotNil(router)
	p1 := router.Prefix("/prefix1")
	p2 := p1.Prefix("/prefix2")
	p2.GetFunc("/p2", buildFunc(204))
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
	p1 = m.Prefix("/prefix1")
	p2 = p1.Prefix("example.com")
	p2.GetFunc("/p2", buildFunc(205))
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/prefix1example.com/p2", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 205)
}

func TestRouter_routers_nest(t *testing.T) {
	a := assert.New(t)

	m := Default()
	router, ok := m.NewRouter("host", group.NewHosts("localhost"))
	a.True(ok).NotNil(router)
	router.Get("/t1", buildHandler(201))
	v1, ok := router.NewRouter("v1", group.NewVersion(false, "v1"))
	a.True(ok).NotNil(v1)
	v1.Get("/path", buildHandler(202))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "https://localhost/t1", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 201)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://localhost/v1/path", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 202)

	// 不存在的路径
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://localhost/v111/path", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)
}

func TestRouter_routers_multiple(t *testing.T) {
	a := assert.New(t)

	m := Default()
	router, ok := m.NewRouter("host", group.NewHosts("localhost"))
	a.True(ok).NotNil(router)
	router.Get("/t1", buildHandler(201))
	v1, ok := router.NewRouter("v1", group.NewVersion(false, "v1"))
	a.True(ok).NotNil(v1)
	v1.Get("/path", buildHandler(202))
	v2, ok := router.NewRouter("v2", group.NewVersion(false, "v1", "v2"))
	a.True(ok).NotNil(v2)
	v2.Get("/path", buildHandler(203))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "https://localhost/t1", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 201)

	// 指向 v1
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://localhost/v1/path", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 202)

	// 指向 v2
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://localhost/v2/path", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 203)
}
