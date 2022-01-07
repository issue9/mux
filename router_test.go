// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
)

var _ http.Handler = &Router{}

// mux 的测试工具
type tester struct {
	router *Router
	srv    *rest.Server
	a      *assert.Assertion
}

func newTester(t testing.TB, o ...Option) *tester {
	r := NewRouter("def", nil, o...)
	a := assert.New(t, false)
	a.NotNil(r)
	a.Equal("def", r.Name())

	return &tester{
		router: r,
		srv:    rest.NewServer(a, r, nil),
		a:      a,
	}
}

func (t *tester) matchCode(method, path string, code int) {
	t.a.TB().Helper()
	t.srv.NewRequest(method, path).Do(nil).Status(code)
}

func (t *tester) matchHeader(method, path string, code int, headers map[string]string) {
	t.a.TB().Helper()
	resp := t.srv.NewRequest(method, path).Do(nil).Status(code)
	for k, v := range headers {
		resp.Header(k, v)
	}
}

func (t *tester) matchContent(method, path string, code int, content string) {
	t.a.TB().Helper()
	t.srv.NewRequest(method, path).Do(nil).Status(code).StringBody(content)
}

func (t *tester) matchOptions(path string, code int, allow string) {
	t.a.TB().Helper()
	t.matchHeader(http.MethodOptions, path, code, map[string]string{"Allow": allow})
}

func (t *tester) matchOptionsAsterisk(allow string) {
	t.a.TB().Helper()
	r, err := http.NewRequest(http.MethodOptions, "*", nil)
	t.a.NotError(err).NotNil(r)

	w := httptest.NewRecorder()
	t.router.ServeHTTP(w, r) // Client.Do 无法传递 * 或是空的路径请求。改用 ServeHTTP
	t.a.Equal(w.Code, 200).
		Equal(w.Header().Get("Allow"), allow)
}

func TestRouter(t *testing.T) {
	test := newTester(t, Lock)

	test.router.Get("/", rest.BuildHandler(test.a, 201, "201", nil))
	test.router.Get("/200", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("200"))
		test.a.NotError(err)
	}))
	test.matchCode(http.MethodGet, "/", 201)
	test.matchCode(http.MethodHead, "/", 201)
	test.matchCode(http.MethodGet, "/abc", http.StatusNotFound)
	test.matchContent(http.MethodHead, "/", 201, "")
	test.matchHeader(http.MethodHead, "/", 201, nil)                                         // WriteHeader 会让 Content-Length 失效
	test.matchHeader(http.MethodHead, "/200", 200, map[string]string{"Content-Length": "3"}) // 不调用 WriteHeader
	test.matchOptionsAsterisk("GET, HEAD, OPTIONS")

	test.router.Get("/h/1", rest.BuildHandler(test.a, 201, "", nil))
	test.matchCode(http.MethodGet, "/h/1", 201)

	test.router.Post("/h/1", rest.BuildHandler(test.a, 202, "", nil))
	test.matchCode(http.MethodPost, "/h/1", 202)
	test.matchOptionsAsterisk("GET, HEAD, OPTIONS, POST")

	test.router.Put("/h/1", rest.BuildHandler(test.a, 203, "", nil))
	test.matchCode(http.MethodPut, "/h/1", 203)

	test.router.Patch("/h/1", rest.BuildHandler(test.a, 204, "", nil))
	test.matchCode(http.MethodPatch, "/h/1", 204)

	test.router.Delete("/h/1", rest.BuildHandler(test.a, 205, "", nil))
	test.matchCode(http.MethodDelete, "/h/1", 205)
	test.matchOptionsAsterisk("DELETE, GET, HEAD, OPTIONS, PATCH, POST, PUT")

	// Any
	test.router.Any("/h/any", rest.BuildHandler(test.a, 206, "", nil))
	test.matchCode(http.MethodGet, "/h/any", 206)
	test.matchCode(http.MethodPost, "/h/any", 206)
	test.matchCode(http.MethodPut, "/h/any", 206)
	test.matchCode(http.MethodPatch, "/h/any", 206)
	test.matchCode(http.MethodDelete, "/h/any", 206)
	test.matchCode(http.MethodTrace, "/h/any", 206)

	// 不能主动添加 Head
	test.a.PanicString(func() {
		test.router.Handle("/head", rest.BuildHandler(test.a, 202, "", nil), http.MethodHead)
	}, "OPTIONS/HEAD")
}

func TestRouter_ServeHTTP(t *testing.T) {
	test := newTester(t, Interceptor(InterceptorDigit, "digit"), Interceptor(InterceptorAny, "any"))

	test.router.Handle("/posts/{path}.html", rest.BuildHandler(test.a, 201, "", nil))
	test.matchCode(http.MethodGet, "/posts/2017/1.html", 201)
	test.matchCode(http.MethodGet, "/Posts/2017/1.html", 404) // 大小写不一样

	test.router.Handle("/posts/{path:.+}.html", rest.BuildHandler(test.a, 202, "", nil))
	test.matchCode(http.MethodGet, "/posts/2017/1.html", 202)

	test.router.Handle("/posts/{id:digit}123", rest.BuildHandler(test.a, 203, "", nil))
	test.matchCode(http.MethodGet, "/posts/123123", 203)

	test.router.Get("///", rest.BuildHandler(test.a, 201, "", nil))
	test.matchCode(http.MethodGet, "///", 201)
	test.matchCode(http.MethodGet, "//", 404)

	// 对 any 和空参数的测试

	test.router.Get("/posts1-{id}-{page}.html", rest.BuildHandler(test.a, 201, "", nil))
	test.matchCode(http.MethodGet, "/posts1--.html", 201)
	test.matchCode(http.MethodGet, "/posts1-1-0.html", 201)

	test.router.Get("/posts2-{id:any}-{page:any}.html", rest.BuildHandler(test.a, 201, "", nil))
	test.matchCode(http.MethodGet, "/posts2--.html", 404)
	test.matchCode(http.MethodGet, "/posts2-1-0.html", 201)

	test.router.Get("/posts3-{id}-{page:any}.html", rest.BuildHandler(test.a, 201, "", nil))
	test.matchCode(http.MethodGet, "/posts3--.html", 404)
	test.matchCode(http.MethodGet, "/posts3-1-0.html", 201)
	test.matchCode(http.MethodGet, "/posts3--0.html", 201)

	// 忽略大小写测试

	test = newTester(t, CaseInsensitive)
	test.router.Handle("/posts/{path}.html", rest.BuildHandler(test.a, 201, "", nil))
	test.matchCode(http.MethodGet, "/posts/2017/1.html", 201)
	test.matchCode(http.MethodGet, "/Posts/2017/1.html", 201) // 忽略大小写
}

func TestRouter_Handle_Remove(t *testing.T) {
	test := newTester(t)

	// 添加 GET /api/1
	// 添加 PUT /api/1
	// 添加 GET /api/2
	test.router.Handle("/api/1", rest.BuildHandler(test.a, 201, "", nil), http.MethodGet)
	test.router.Handle("/api/1", rest.BuildHandler(test.a, 201, "", nil), http.MethodPut)
	test.router.Handle("/api/2", rest.BuildHandler(test.a, 202, "", nil), http.MethodGet)

	test.matchCode(http.MethodGet, "/api/1", 201)
	test.matchCode(http.MethodPut, "/api/1", 201)
	test.matchCode(http.MethodGet, "/api/2", 202)
	test.matchCode(http.MethodDelete, "/api/1", http.StatusMethodNotAllowed) // 未实现

	// 删除 GET /api/1
	test.router.Remove("/api/1", http.MethodGet)
	test.matchCode(http.MethodGet, "/api/1", http.StatusMethodNotAllowed)
	test.matchCode(http.MethodPut, "/api/1", 201) // 不影响 PUT
	test.matchCode(http.MethodGet, "/api/2", 202)

	// 删除 GET /api/2，只有一个，所以相当于整个节点被删除
	test.router.Remove("/api/2", http.MethodGet)
	test.matchCode(http.MethodGet, "/api/1", http.StatusMethodNotAllowed)
	test.matchCode(http.MethodPut, "/api/1", 201)                 // 不影响 PUT
	test.matchCode(http.MethodGet, "/api/2", http.StatusNotFound) // 整个节点被删除

	// 添加 POST /api/1
	test.router.Handle("/api/1", rest.BuildHandler(test.a, 201, "", nil), http.MethodPost)
	test.matchCode(http.MethodPost, "/api/1", 201)

	// 删除 ANY /api/1
	test.router.Remove("/api/1")
	test.matchCode(http.MethodPost, "/api/1", http.StatusNotFound) // 404 表示整个节点都没了
}

func TestRouter_Routes(t *testing.T) {
	a := assert.New(t, false)

	def := NewRouter("", nil)
	a.NotNil(def)
	def.Get("/m", rest.BuildHandler(a, 1, "", nil))
	def.Post("/m", rest.BuildHandler(a, 1, "", nil))
	a.Equal(def.Routes(), map[string][]string{"*": {"OPTIONS"}, "/m": {"GET", "HEAD", "OPTIONS", "POST"}})
}

func TestRouter_Params(t *testing.T) {
	a := assert.New(t, false)
	router := NewRouter("", nil, Interceptor(InterceptorDigit, "digit"))
	a.NotNil(router)

	var globalParams Params

	buildParamsHandler := func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ps := GetParams(r)
			if ps != nil {
				globalParams = ps
			}
		})
	}

	requestParams := func(method, url string, status int, ps map[string]string) {
		w := httptest.NewRecorder()
		a.NotNil(w)

		r, err := http.NewRequest(method, url, nil)
		a.NotError(err).NotNil(r)

		router.ServeHTTP(w, r)

		a.Equal(w.Code, status)
		if len(ps) > 0 { // 由于 globalParams 是公用数据，会保存上一次获取的值，所以只在有值时才比较
			a.Equal(len(ps), globalParams.Count())
			for k, v := range ps {
				vv, found := globalParams.Get(k)
				a.True(found).Equal(vv, v)
			}
		}
		globalParams = nil // 清空全局的 globalParams
	}

	// 添加 patch /api/{version:\\d+}
	router.Patch("/api/{version:\\d+}", buildParamsHandler())
	requestParams(http.MethodPatch, "/api/2", http.StatusOK, map[string]string{"version": "2"})
	requestParams(http.MethodPatch, "/api/256", http.StatusOK, map[string]string{"version": "256"})
	requestParams(http.MethodGet, "/api/256", http.StatusMethodNotAllowed, nil) // 不存在的请求方法

	// 添加 patch /api/v2/{version:\\d*}
	router.Clean()
	router.Patch("/api/v2/{version:\\d*}", buildParamsHandler())
	requestParams(http.MethodPatch, "/api/v2/2", http.StatusOK, map[string]string{"version": "2"})
	requestParams(http.MethodPatch, "/api/v2/", http.StatusOK, map[string]string{"version": ""})

	// 忽略名称捕获
	router.Clean()
	router.Patch("/api/v3/{-version:\\d*}", buildParamsHandler())
	requestParams(http.MethodPatch, "/api/v3/2", http.StatusOK, nil)
	requestParams(http.MethodPatch, "/api/v3/", http.StatusOK, nil)

	// 添加 patch /api/v2/{version:\\d*}/test
	router.Clean()
	router.Patch("/api/v2/{version:\\d*}/test", buildParamsHandler())
	requestParams(http.MethodPatch, "/api/v2/2/test", http.StatusOK, map[string]string{"version": "2"})
	requestParams(http.MethodPatch, "/api/v2//test", http.StatusOK, map[string]string{"version": ""})

	// 中文作为值
	router.Clean()
	router.Patch("/api/v3/{版本:digit}", buildParamsHandler())
	requestParams(http.MethodPatch, "/api/v3/2", http.StatusOK, map[string]string{"版本": "2"})
}

func TestRouter_Clean(t *testing.T) {
	a := assert.New(t, false)

	def := NewRouter("", nil)
	a.NotNil(def)
	def.Get("/m1", rest.BuildHandler(a, 200, "", nil)).
		Post("/m1", rest.BuildHandler(a, 201, "", nil))

	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "http://localhost:88/m1", nil)
	a.NotError(err).NotNil(r)
	def.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 200)

	def.Clean()
	w = httptest.NewRecorder()
	r, err = http.NewRequest(http.MethodGet, "/m1", nil)
	a.NotError(err).NotNil(r)
	def.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 404)
}

// 测试匹配顺序是否正确
func TestRouter_ServeHTTP_Order(t *testing.T) {
	test := newTester(t, Interceptor(InterceptorAny, "any"))

	test.router.Get("/posts/{id}", rest.BuildHandler(test.a, 203, "", nil))
	test.router.Get("/posts/{id:\\d+}", rest.BuildHandler(test.a, 202, "", nil))
	test.router.Get("/posts/1", rest.BuildHandler(test.a, 201, "", nil))
	test.router.Get("/posts/{id:[0-9]+}", rest.BuildHandler(test.a, 199, "", nil)) //  两个正则，后添加的永远匹配不到
	test.router.Get("/posts-{id:any}", rest.BuildHandler(test.a, 204, "", nil))
	test.router.Get("/posts-", rest.BuildHandler(test.a, 205, "", nil))
	test.matchCode(http.MethodGet, "/posts/1", 201)   // 普通路由项完全匹配
	test.matchCode(http.MethodGet, "/posts/2", 202)   // 正则路由
	test.matchCode(http.MethodGet, "/posts/abc", 203) // 命名路由
	test.matchCode(http.MethodGet, "/posts/", 203)    // 命名路由
	test.matchCode(http.MethodGet, "/posts-5", 204)   // 命名路由
	test.matchCode(http.MethodGet, "/posts-", 205)    // 204 只匹配非空

	// interceptor
	test = newTester(t, Interceptor(InterceptorDigit, "[0-9]+"))
	test.router.Get("/posts/{id}", rest.BuildHandler(test.a, 203, "", nil))        // f3
	test.router.Get("/posts/{id:\\d+}", rest.BuildHandler(test.a, 202, "", nil))   // f2 永远匹配不到
	test.router.Get("/posts/1", rest.BuildHandler(test.a, 201, "", nil))           // f1
	test.router.Get("/posts/{id:[0-9]+}", rest.BuildHandler(test.a, 210, "", nil)) // f0 interceptor 权限比正则要高
	test.matchCode(http.MethodGet, "/posts/1", 201)                                // f1 普通路由项完全匹配
	test.matchCode(http.MethodGet, "/posts/2", 210)                                // f0 interceptor
	test.matchCode(http.MethodGet, "/posts/abc", 203)                              // f3 命名路由
	test.matchCode(http.MethodGet, "/posts/", 203)                                 // f3

	test = newTester(t)
	test.router.Get("/p1/{p1}/p2/{p2:\\d+}", rest.BuildHandler(test.a, 201, "", nil)) // f1
	test.router.Get("/p1/{p1}/p2/{p2:\\w+}", rest.BuildHandler(test.a, 202, "", nil)) // f2
	test.matchCode(http.MethodGet, "/p1/1/p2/1", 201)                                 // f1
	test.matchCode(http.MethodGet, "/p1/2/p2/s", 202)                                 // f2

	test = newTester(t)
	test.router.Get("/posts/{id}/{page}", rest.BuildHandler(test.a, 202, "", nil)) // f2
	test.router.Get("/posts/{id}/1", rest.BuildHandler(test.a, 201, "", nil))      // f1
	test.matchCode(http.MethodGet, "/posts/1/1", 201)                              // f1 普通路由项完全匹配
	test.matchCode(http.MethodGet, "/posts/2/5", 202)                              // f2 命名完全匹配

	test = newTester(t)
	test.router.Get("/tags/{id}.html", rest.BuildHandler(test.a, 201, "", nil)) // f1
	test.router.Get("/tags.html", rest.BuildHandler(test.a, 202, "", nil))      // f2
	test.router.Get("/{path}", rest.BuildHandler(test.a, 203, "", nil))         // f3
	test.matchCode(http.MethodGet, "/tags", 203)                                // f3 // 正好与 f1 的第一个节点匹配
	test.matchCode(http.MethodGet, "/tags/1.html", 201)                         // f1
	test.matchCode(http.MethodGet, "/tags.html", 202)                           // f2
}
