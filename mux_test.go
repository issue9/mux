// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"

	"github.com/issue9/mux/v2/internal/handlers"
)

func buildHandler(code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	})
}

func buildFunc(code int) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	})
}

// mux 的测试工具
type tester struct {
	mux *Mux
	srv *rest.Server
}

func newTester(t testing.TB, disableOptions, disableHead, skipClean bool) *tester {
	mux := New(disableOptions, disableHead, skipClean, nil, nil)
	return &tester{
		mux: mux,
		srv: rest.NewServer(t, mux, nil),
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

func TestMux(t *testing.T) {
	test := newTester(t, false, true, false)

	// 测试 / 和 "" 是否访问同一地址
	test.mux.Get("/", buildHandler(201))
	test.matchTrue(http.MethodGet, "", 201)
	test.matchTrue(http.MethodGet, "/", 201)
	test.matchTrue(http.MethodHead, "/", http.StatusMethodNotAllowed) // 未启用 autoHead
	test.matchTrue(http.MethodGet, "/abc", http.StatusNotFound)

	test.mux.Get("/h/1", buildHandler(201))
	test.matchTrue(http.MethodGet, "/h/1", 201)
	test.mux.GetFunc("/f/1", buildFunc(201))
	test.matchTrue(http.MethodGet, "/f/1", 201)

	test.mux.Post("/h/1", buildHandler(202))
	test.matchTrue(http.MethodPost, "/h/1", 202)
	test.mux.PostFunc("/f/1", buildFunc(202))
	test.matchTrue(http.MethodPost, "/f/1", 202)

	test.mux.Put("/h/1", buildHandler(203))
	test.matchTrue(http.MethodPut, "/h/1", 203)
	test.mux.PutFunc("/f/1", buildFunc(203))
	test.matchTrue(http.MethodPut, "/f/1", 203)

	test.mux.Patch("/h/1", buildHandler(204))
	test.matchTrue(http.MethodPatch, "/h/1", 204)
	test.mux.PatchFunc("/f/1", buildFunc(204))
	test.matchTrue(http.MethodPatch, "/f/1", 204)

	test.mux.Delete("/h/1", buildHandler(205))
	test.matchTrue(http.MethodDelete, "/h/1", 205)
	test.mux.DeleteFunc("/f/1", buildFunc(205))
	test.matchTrue(http.MethodDelete, "/f/1", 205)

	// Any
	test.mux.Any("/h/any", buildHandler(206))
	test.matchTrue(http.MethodGet, "/h/any", 206)
	test.matchTrue(http.MethodPost, "/h/any", 206)
	test.matchTrue(http.MethodPut, "/h/any", 206)
	test.matchTrue(http.MethodPatch, "/h/any", 206)
	test.matchTrue(http.MethodDelete, "/h/any", 206)
	test.matchTrue(http.MethodTrace, "/h/any", 206)

	test.mux.AnyFunc("/f/any", buildFunc(206))
	test.matchTrue(http.MethodGet, "/f/any", 206)
	test.matchTrue(http.MethodPost, "/f/any", 206)
	test.matchTrue(http.MethodPut, "/f/any", 206)
	test.matchTrue(http.MethodPatch, "/f/any", 206)
	test.matchTrue(http.MethodDelete, "/f/any", 206)
	test.matchTrue(http.MethodTrace, "/f/any", 206)
}

func TestMux_All(t *testing.T) {
	a := assert.New(t)

	m := Default()
	a.NotNil(m)

	m.Get("/m", buildHandler(1))
	m.Post("/m", buildHandler(1))
	a.Equal(m.All(false, false), []*Router{
		{
			Name: "",
			Routes: map[string][]string{
				"/m": {"GET", "HEAD", "OPTIONS", "POST"},
			},
		},
	})

	r, ok := m.NewMux("host-1", NewHosts())
	a.True(ok).NotNil(r)
	r.Get("/m", buildHandler(1))
	r.Post("/m", buildHandler(1))
	a.Equal(m.All(false, false), []*Router{
		{
			Name: "host-1",
			Routes: map[string][]string{
				"/m": {"GET", "HEAD", "OPTIONS", "POST"},
			},
		},
		{
			Name: "",
			Routes: map[string][]string{
				"/m": {"GET", "HEAD", "OPTIONS", "POST"},
			},
		},
	})
}

func TestMux_Head(t *testing.T) {
	test := newTester(t, false, false, false)

	test.mux.Get("/", buildHandler(201))
	test.matchTrue(http.MethodGet, "", 201)
	test.matchTrue(http.MethodGet, "/", 201)
	test.matchTrue(http.MethodHead, "", 201)
	test.matchTrue(http.MethodHead, "/", 201)
	test.matchContent(http.MethodHead, "/", 201, "")

	test.mux.Get("/h/1", buildHandler(201))
	test.matchTrue(http.MethodGet, "/h/1", 201)
	test.matchTrue(http.MethodHead, "/h/1", 201)
	test.mux.GetFunc("/f/1", buildFunc(201))
	test.matchTrue(http.MethodGet, "/f/1", 201)
	test.matchTrue(http.MethodHead, "/f/1", 201)

	test.mux.Post("/h/post", buildHandler(202))
	test.matchTrue(http.MethodPost, "/h/post", 202)
	test.matchTrue(http.MethodHead, "/h/post", http.StatusMethodNotAllowed)

	// Any
	test.mux.Any("/h/any", buildHandler(206))
	test.matchTrue(http.MethodGet, "/h/any", 206)
	test.matchTrue(http.MethodHead, "/h/any", 206)
	test.matchTrue(http.MethodPost, "/h/any", 206)
	test.matchTrue(http.MethodPut, "/h/any", 206)
	test.matchTrue(http.MethodPatch, "/h/any", 206)
	test.matchTrue(http.MethodDelete, "/h/any", 206)
	test.matchTrue(http.MethodTrace, "/h/any", 206)

	test.mux.AnyFunc("/f/any", buildFunc(206))
	test.matchTrue(http.MethodGet, "/f/any", 206)
	test.matchTrue(http.MethodHead, "/f/any", 206)
	test.matchTrue(http.MethodPost, "/f/any", 206)
	test.matchTrue(http.MethodPut, "/f/any", 206)
	test.matchTrue(http.MethodPatch, "/f/any", 206)
	test.matchTrue(http.MethodDelete, "/f/any", 206)
	test.matchTrue(http.MethodTrace, "/f/any", 206)
}

func TestMux_Handle_Remove(t *testing.T) {
	a := assert.New(t)
	test := newTester(t, false, true, false)

	// 添加 GET /api/1
	// 添加 PUT /api/1
	// 添加 GET /api/2
	a.NotError(test.mux.HandleFunc("/api/1", buildFunc(201), http.MethodGet))
	a.NotError(test.mux.HandleFunc("/api/1", buildFunc(201), http.MethodPut))
	a.NotError(test.mux.HandleFunc("/api/2", buildFunc(202), http.MethodGet))

	test.matchTrue(http.MethodGet, "/api/1", 201)
	test.matchTrue(http.MethodPut, "/api/1", 201)
	test.matchTrue(http.MethodGet, "/api/2", 202)
	test.matchTrue(http.MethodDelete, "/api/1", http.StatusMethodNotAllowed) // 未实现

	// 删除 GET /api/1
	test.mux.Remove("/api/1", http.MethodGet)
	test.matchTrue(http.MethodGet, "/api/1", http.StatusMethodNotAllowed)
	test.matchTrue(http.MethodPut, "/api/1", 201) // 不影响 PUT
	test.matchTrue(http.MethodGet, "/api/2", 202)

	// 删除 GET /api/2，只有一个，所以相当于整个节点被删除
	test.mux.Remove("/api/2", http.MethodGet)
	test.matchTrue(http.MethodGet, "/api/1", http.StatusMethodNotAllowed)
	test.matchTrue(http.MethodPut, "/api/1", 201)                 // 不影响 PUT
	test.matchTrue(http.MethodGet, "/api/2", http.StatusNotFound) // 整个节点被删除

	// 添加 POST /api/1
	a.NotError(test.mux.Handle("/api/1", buildFunc(201), http.MethodPost))
	test.matchTrue(http.MethodPost, "/api/1", 201)

	// 删除 ANY /api/1
	test.mux.Remove("/api/1")
	test.matchTrue(http.MethodPost, "/api/1", http.StatusNotFound) // 404 表示整个节点都没了
}

func TestMux_Options(t *testing.T) {
	a := assert.New(t)
	test := newTester(t, false, true, false)

	// 添加 GET /api/1
	a.NotError(test.mux.Handle("/api/1", buildHandler(201), http.MethodGet))
	test.optionsTrue("/api/1", http.StatusOK, "GET, OPTIONS")

	// 添加 DELETE /api/1
	a.NotError(test.mux.Handle("/api/1", buildHandler(201), http.MethodDelete))
	test.optionsTrue("/api/1", http.StatusOK, "DELETE, GET, OPTIONS")

	// 删除 DELETE /api/1
	test.mux.Remove("/api/1", http.MethodDelete)
	test.optionsTrue("/api/1", http.StatusOK, "GET, OPTIONS")

	// 通过 Options 自定义 Allow 报头
	test.mux.Options("/api/1", "CUSTOM OPTIONS1")
	test.optionsTrue("/api/1", http.StatusOK, "CUSTOM OPTIONS1")
	test.mux.Options("/api/1", "CUSTOM OPTIONS2")
	test.optionsTrue("/api/1", http.StatusOK, "CUSTOM OPTIONS2")

	test.mux.HandleFunc("/api/1", buildFunc(201), http.MethodOptions)
	test.optionsTrue("/api/1", 201, "")

	// disableOptions 为 true
	test = newTester(t, true, true, false)
	test.optionsTrue("/api/1", http.StatusNotFound, "")
	test.mux.Options("/api/1", "CUSTOM OPTIONS1") // 显示指定
	test.optionsTrue("/api/1", http.StatusOK, "CUSTOM OPTIONS1")
}

func TestMux_Params(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, false, false, nil, nil)
	a.NotNil(srvmux)
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

		srvmux.ServeHTTP(w, r)

		a.Equal(w.Code, status)
		if ps != nil { // 由于 params 是公用数据，会保存上一次获取的值，所以只在有值时才比较
			a.Equal(params, ps)
		}
		params = nil // 清空全局的 params
	}

	// 添加 patch /api/{version:\\d+}
	a.NotError(srvmux.Patch("/api/{version:\\d+}", buildParamsHandler()))
	requestParams(http.MethodPatch, "/api/2", http.StatusOK, map[string]string{"version": "2"})
	requestParams(http.MethodPatch, "/api/256", http.StatusOK, map[string]string{"version": "256"})
	requestParams(http.MethodGet, "/api/256", http.StatusMethodNotAllowed, nil) // 不存在的请求方法

	// 添加 patch /api/v2/{version:\\d*}
	a.NotError(srvmux.Patch("/api/v2/{version:\\d*}", buildParamsHandler()))
	requestParams(http.MethodPatch, "/api/v2/2", http.StatusOK, map[string]string{"version": "2"})
	requestParams(http.MethodPatch, "/api/v2/", http.StatusOK, map[string]string{"version": ""})

	// 添加 patch /api/v2/{version:\\d+}/test
	a.NotError(srvmux.Patch("/api/v2/{version:\\d*}/test", buildParamsHandler()))
	requestParams(http.MethodPatch, "/api/v2/2/test", http.StatusOK, map[string]string{"version": "2"})
	requestParams(http.MethodPatch, "/api/v2//test", http.StatusNotFound, nil) // 可选参数不能在路由中间
}

func TestMux_Clean(t *testing.T) {
	a := assert.New(t)

	m := New(false, false, false, nil, nil)
	m.Get("/m1", buildHandler(200)).
		Post("/m1", buildHandler(201))
	router, ok := m.NewMux("host", NewHosts("example.com"))
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

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://example.com/m1", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)
}

func TestMux_ServeHTTP(t *testing.T) {
	test := newTester(t, false, true, false)

	test.mux.Handle("/posts/{path}.html", buildHandler(201))
	test.matchTrue(http.MethodGet, "/posts/2017/1.html", 201)

	test.mux.Handle("/posts/{path:.+}.html", buildHandler(202))
	test.matchTrue(http.MethodGet, "/posts/2017/1.html", 202)
}

// 测试匹配顺序是否正确
func TestMux_ServeHTTP_Order(t *testing.T) {
	a := assert.New(t)
	test := newTester(t, false, true, false)

	a.NotError(test.mux.GetFunc("/posts/{id}", buildFunc(203)))      // f3
	a.NotError(test.mux.GetFunc("/posts/{id:\\d+}", buildFunc(202))) // f2
	a.NotError(test.mux.GetFunc("/posts/1", buildFunc(201)))         // f1
	test.matchTrue(http.MethodGet, "/posts/1", 201)                  // f1 普通路由项完全匹配
	test.matchTrue(http.MethodGet, "/posts/2", 202)                  // f1 正则路由
	test.matchTrue(http.MethodGet, "/posts/abc", 203)                // f3 命名路由

	test = newTester(t, false, true, false)
	a.NotError(test.mux.GetFunc("/p1/{p1}/p2/{p2:\\d+}", buildFunc(201))) // f1
	a.NotError(test.mux.GetFunc("/p1/{p1}/p2/{p2:\\w+}", buildFunc(202))) // f2
	test.matchTrue(http.MethodGet, "/p1/1/p2/1", 201)                     // f1
	test.matchTrue(http.MethodGet, "/p1/2/p2/s", 202)                     // f2

	test = newTester(t, false, true, false)
	a.NotError(test.mux.GetFunc("/posts/{id}/{page}", buildFunc(202))) // f2
	a.NotError(test.mux.GetFunc("/posts/{id}/1", buildFunc(201)))      // f1
	test.matchTrue(http.MethodGet, "/posts/1/1", 201)                  // f1 普通路由项完全匹配
	test.matchTrue(http.MethodGet, "/posts/2/5", 202)                  // f2 命名完全匹配

	test = newTester(t, false, true, false)
	a.NotError(test.mux.GetFunc("/tags/{id}.html", buildFunc(201))) // f1
	a.NotError(test.mux.GetFunc("/tags.html", buildFunc(202)))      // f2
	a.NotError(test.mux.GetFunc("/{path}", buildFunc(203)))         // f3
	test.matchTrue(http.MethodGet, "/tags", 203)                    // f3 // 正好与 f1 的第一个节点匹配
	test.matchTrue(http.MethodGet, "/tags/1.html", 201)             // f1
	test.matchTrue(http.MethodGet, "/tags.html", 202)               // f2
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
