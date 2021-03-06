// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"

	"github.com/issue9/mux/v5/interceptor"
)

var _ http.Handler = &Router{}

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

// mux 的测试工具
type tester struct {
	router *Router
	srv    *rest.Server
}

func newTester(t testing.TB, disableHead bool) *tester {
	r, err := NewRouter(disableHead, AllowedCORS(), nil, nil)
	assert.NotError(t, err)
	assert.NotNil(t, r)

	return &tester{
		router: r,
		srv:    rest.NewServer(t, r, nil),
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
	test := newTester(t, true)

	// 测试 / 和 "" 是否访问同一地址
	test.router.Get("/", buildHandler(201))
	test.matchTrue(http.MethodGet, "", 201)
	test.matchTrue(http.MethodGet, "/", 201)
	test.matchTrue(http.MethodHead, "/", http.StatusMethodNotAllowed) // 未启用 autoHead
	test.matchTrue(http.MethodGet, "/abc", http.StatusNotFound)

	test.router.Get("/h/1", buildHandler(201))
	test.matchTrue(http.MethodGet, "/h/1", 201)
	test.router.GetFunc("/f/1", buildHandlerFunc(201))
	test.matchTrue(http.MethodGet, "/f/1", 201)

	test.router.Post("/h/1", buildHandler(202))
	test.matchTrue(http.MethodPost, "/h/1", 202)
	test.router.PostFunc("/f/1", buildHandlerFunc(202))
	test.matchTrue(http.MethodPost, "/f/1", 202)

	test.router.Put("/h/1", buildHandler(203))
	test.matchTrue(http.MethodPut, "/h/1", 203)
	test.router.PutFunc("/f/1", buildHandlerFunc(203))
	test.matchTrue(http.MethodPut, "/f/1", 203)

	test.router.Patch("/h/1", buildHandler(204))
	test.matchTrue(http.MethodPatch, "/h/1", 204)
	test.router.PatchFunc("/f/1", buildHandlerFunc(204))
	test.matchTrue(http.MethodPatch, "/f/1", 204)

	test.router.Delete("/h/1", buildHandler(205))
	test.matchTrue(http.MethodDelete, "/h/1", 205)
	test.router.DeleteFunc("/f/1", buildHandlerFunc(205))
	test.matchTrue(http.MethodDelete, "/f/1", 205)

	// Any
	test.router.Any("/h/any", buildHandler(206))
	test.matchTrue(http.MethodGet, "/h/any", 206)
	test.matchTrue(http.MethodPost, "/h/any", 206)
	test.matchTrue(http.MethodPut, "/h/any", 206)
	test.matchTrue(http.MethodPatch, "/h/any", 206)
	test.matchTrue(http.MethodDelete, "/h/any", 206)
	test.matchTrue(http.MethodTrace, "/h/any", 206)

	test.router.AnyFunc("/f/any", buildHandlerFunc(206))
	test.matchTrue(http.MethodGet, "/f/any", 206)
	test.matchTrue(http.MethodPost, "/f/any", 206)
	test.matchTrue(http.MethodPut, "/f/any", 206)
	test.matchTrue(http.MethodPatch, "/f/any", 206)
	test.matchTrue(http.MethodDelete, "/f/any", 206)
	test.matchTrue(http.MethodTrace, "/f/any", 206)
}

func TestRouter_Head(t *testing.T) {
	test := newTester(t, false)

	test.router.Get("/", buildHandler(201))
	test.matchTrue(http.MethodGet, "", 201)
	test.matchTrue(http.MethodGet, "/", 201)
	test.matchTrue(http.MethodHead, "", 201)
	test.matchTrue(http.MethodHead, "/", 201)
	test.matchContent(http.MethodHead, "/", 201, "")

	test.router.Get("/h/1", buildHandler(201))
	test.matchTrue(http.MethodGet, "/h/1", 201)
	test.matchTrue(http.MethodHead, "/h/1", 201)
	test.router.GetFunc("/f/1", buildHandlerFunc(201))
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

	test.router.AnyFunc("/f/any", buildHandlerFunc(206))
	test.matchTrue(http.MethodGet, "/f/any", 206)
	test.matchTrue(http.MethodHead, "/f/any", 206)
	test.matchTrue(http.MethodPost, "/f/any", 206)
	test.matchTrue(http.MethodPut, "/f/any", 206)
	test.matchTrue(http.MethodPatch, "/f/any", 206)
	test.matchTrue(http.MethodDelete, "/f/any", 206)
	test.matchTrue(http.MethodTrace, "/f/any", 206)

	// 不能主动添加 Head
	assert.Error(t, test.router.HandleFunc("/head", buildHandlerFunc(202), http.MethodHead))
}

func TestRouter_Handle_Remove(t *testing.T) {
	a := assert.New(t)
	test := newTester(t, true)

	// 添加 GET /api/1
	// 添加 PUT /api/1
	// 添加 GET /api/2
	a.NotError(test.router.HandleFunc("/api/1", buildHandlerFunc(201), http.MethodGet))
	a.NotError(test.router.HandleFunc("/api/1", buildHandlerFunc(201), http.MethodPut))
	a.NotError(test.router.HandleFunc("/api/2", buildHandlerFunc(202), http.MethodGet))

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
	a.NotError(test.router.Handle("/api/1", buildHandlerFunc(201), http.MethodPost))
	test.matchTrue(http.MethodPost, "/api/1", 201)

	// 删除 ANY /api/1
	test.router.Remove("/api/1")
	test.matchTrue(http.MethodPost, "/api/1", http.StatusNotFound) // 404 表示整个节点都没了
}

func TestRouter_Routes(t *testing.T) {
	a := assert.New(t)

	def := DefaultRouter()
	a.NotNil(def)
	def.Get("/m", buildHandler(1))
	def.Post("/m", buildHandler(1))
	a.Equal(def.Routes(), map[string][]string{"/m": {"GET", "HEAD", "OPTIONS", "POST"}})

	def, err := NewRouter(true, nil, nil, nil)
	a.NotError(err).NotNil(def)
	def.Get("/m", buildHandler(1))
	def.Post("/m", buildHandler(1))
	a.Equal(def.Routes(), map[string][]string{"/m": {"GET", "OPTIONS", "POST"}})
}

func TestRouter_Params(t *testing.T) {
	a := assert.New(t)
	router := DefaultRouter()
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
	router.Clean()
	a.NotError(router.Patch("/api/v2/{version:\\d*}", buildParamsHandler()))
	requestParams(http.MethodPatch, "/api/v2/2", http.StatusOK, map[string]string{"version": "2"})
	requestParams(http.MethodPatch, "/api/v2/", http.StatusOK, map[string]string{"version": ""})

	// 添加 patch /api/v2/{version:\\d*}/test
	router.Clean()
	a.NotError(router.Patch("/api/v2/{version:\\d*}/test", buildParamsHandler()))
	requestParams(http.MethodPatch, "/api/v2/2/test", http.StatusOK, map[string]string{"version": "2"})
	requestParams(http.MethodPatch, "/api/v2//test", http.StatusOK, map[string]string{"version": ""})

	// 中文作为值
	router.Clean()
	a.NotError(router.Patch("/api/v3/{版本:digit}", buildParamsHandler()))
	requestParams(http.MethodPatch, "/api/v3/2", http.StatusOK, map[string]string{"版本": "2"})
}

func TestRouter_Clean(t *testing.T) {
	a := assert.New(t)

	def := DefaultRouter()
	a.NotNil(def)
	def.Get("/m1", buildHandler(200)).
		Post("/m1", buildHandler(201))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://localhost:88/m1", nil)
	def.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 200)

	def.Clean()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/m1", nil)
	def.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)
}

func TestMux_ServeHTTP(t *testing.T) {
	a := assert.New(t)

	test := newTester(t, true)

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

	test := newTester(t, true)
	a.NotError(test.router.GetFunc("/posts/{id}", buildHandlerFunc(203)))        // f3
	a.NotError(test.router.GetFunc("/posts/{id:\\d+}", buildHandlerFunc(202)))   // f2
	a.NotError(test.router.GetFunc("/posts/1", buildHandlerFunc(201)))           // f1
	a.NotError(test.router.GetFunc("/posts/{id:[0-9]+}", buildHandlerFunc(199))) // f0 两个正则，后添加的永远匹配不到
	test.matchTrue(http.MethodGet, "/posts/1", 201)                              // f1 普通路由项完全匹配
	test.matchTrue(http.MethodGet, "/posts/2", 202)                              // f1 正则路由
	test.matchTrue(http.MethodGet, "/posts/abc", 203)                            // f3 命名路由
	test.matchTrue(http.MethodGet, "/posts/", 203)                               // f3

	// interceptor
	test = newTester(t, true)
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

	test = newTester(t, true)
	a.NotError(test.router.GetFunc("/p1/{p1}/p2/{p2:\\d+}", buildHandlerFunc(201))) // f1
	a.NotError(test.router.GetFunc("/p1/{p1}/p2/{p2:\\w+}", buildHandlerFunc(202))) // f2
	test.matchTrue(http.MethodGet, "/p1/1/p2/1", 201)                               // f1
	test.matchTrue(http.MethodGet, "/p1/2/p2/s", 202)                               // f2

	test = newTester(t, true)
	a.NotError(test.router.GetFunc("/posts/{id}/{page}", buildHandlerFunc(202))) // f2
	a.NotError(test.router.GetFunc("/posts/{id}/1", buildHandlerFunc(201)))      // f1
	test.matchTrue(http.MethodGet, "/posts/1/1", 201)                            // f1 普通路由项完全匹配
	test.matchTrue(http.MethodGet, "/posts/2/5", 202)                            // f2 命名完全匹配

	test = newTester(t, true)
	a.NotError(test.router.GetFunc("/tags/{id}.html", buildHandlerFunc(201))) // f1
	a.NotError(test.router.GetFunc("/tags.html", buildHandlerFunc(202)))      // f2
	a.NotError(test.router.GetFunc("/{path}", buildHandlerFunc(203)))         // f3
	test.matchTrue(http.MethodGet, "/tags", 203)                              // f3 // 正好与 f1 的第一个节点匹配
	test.matchTrue(http.MethodGet, "/tags/1.html", 201)                       // f1
	test.matchTrue(http.MethodGet, "/tags.html", 202)                         // f2
}
