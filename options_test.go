// SPDX-License-Identifier: MIT

package mux

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"

	"github.com/issue9/mux/v7/internal/syntax"
	"github.com/issue9/mux/v7/internal/tree"
	"github.com/issue9/mux/v7/types"
)

func newRouter(name string, o *Options) *RouterOf[http.Handler] {
	callFunc := func(w http.ResponseWriter, r *http.Request, p types.Params, h http.Handler) {
		h.ServeHTTP(w, r)
	}
	m := tree.BuildTestNodeHandlerFunc(http.StatusMethodNotAllowed)
	opt := tree.BuildTestNodeHandlerFunc(http.StatusOK)
	return NewRouterOf(name, callFunc, http.NotFoundHandler(), m, opt, o)
}

func TestOptions(t *testing.T) {
	a := assert.New(t, false)

	r := newRouter("", nil)
	a.NotNil(r)

	r = newRouter("", &Options{CORS: &CORS{
		Origins: []string{"https://example.com"},
		MaxAge:  3600,
	}})
	a.NotNil(r).
		Equal(r.cors.Origins, []string{"https://example.com"}).
		Nil(r.cors.AllowHeaders).
		Equal(r.cors.MaxAge, 3600)

	r = newRouter("", &Options{CORS: &CORS{
		Origins:          []string{"https://example.com"},
		AllowCredentials: true,
	}})
	a.NotNil(r)

	a.Panic(func() {
		r = newRouter("", &Options{CORS: &CORS{
			Origins:          []string{"*"},
			AllowCredentials: true,
		}})
	})
}

func TestRecovery(t *testing.T) {
	a := assert.New(t, false)

	p := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic("test") })

	router := newRouter("", nil)
	a.NotNil(router).Nil(router.recoverFunc)
	router.Get("/path", p)
	a.Panic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
	})

	// WriterRecovery
	out := new(bytes.Buffer)
	router = newRouter("", &Options{RecoverFunc: WriterRecovery(404, out)})
	a.NotNil(router).NotNil(router.recoverFunc)
	router.Get("/path", p)
	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
		a.Contains(out.String(), "test").
			Equal(w.Code, 404)
	})

	// LogRecovery
	out = new(bytes.Buffer)
	l := log.New(out, "test:", 0)
	router = newRouter("", &Options{RecoverFunc: LogRecovery(405, l)})
	a.NotNil(router).NotNil(router.recoverFunc)
	router.Get("/path", p)
	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
		a.Equal(405, w.Code).
			Contains(out.String(), "test")
	})

	// HTTPRecovery
	router = newRouter("", &Options{RecoverFunc: HTTPRecovery(406)})
	a.NotNil(router).NotNil(router.recoverFunc)
	router.Get("/path", p)
	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
		a.Equal(w.Code, 406)
	})
}

func TestOptions_sanitize(t *testing.T) {
	a := assert.New(t, false)

	o := &Options{}
	a.NotError(o.sanitize())
	a.NotNil(o.CORS)

	// URLDomain

	o = &Options{URLDomain: "https://example.com"}
	a.NotError(o.sanitize())
	a.Equal(o.URLDomain, "https://example.com")
	o = &Options{URLDomain: "https://example.com/"}
	a.NotError(o.sanitize())
	a.Equal(o.URLDomain, "https://example.com")
}

func TestCORS_sanitize(t *testing.T) {
	a := assert.New(t, false)

	c := &CORS{}
	a.NotError(c.sanitize())
	a.True(c.deny).
		False(c.anyHeaders).
		Empty(c.allowHeadersString).
		False(c.anyOrigins).
		Empty(c.exposedHeadersString).
		Empty(c.maxAgeString)

	c = &CORS{
		Origins: []string{"*"},
		MaxAge:  50,
	}
	a.NotError(c.sanitize())
	a.True(c.anyOrigins).Equal(c.maxAgeString, "50")

	c = &CORS{
		Origins: []string{"*"},
		MaxAge:  -1,
	}
	a.NotError(c.sanitize())
	a.True(c.anyOrigins).Equal(c.maxAgeString, "-1")

	c = &CORS{
		MaxAge: -2,
	}
	a.ErrorString(c.sanitize(), "maxAge 的值只能是 >= -1")

	c = &CORS{
		Origins:          []string{"*"},
		AllowCredentials: true,
	}
	a.ErrorString(c.sanitize(), "不能同时成立")

	c = &CORS{
		AllowHeaders:   []string{"*"},
		ExposedHeaders: []string{"h1", "h2"},
	}
	a.NotError(c.sanitize())
	a.True(c.anyHeaders).
		Equal(c.allowHeadersString, "*").
		Equal(c.exposedHeadersString, "h1,h2")
}

func TestCORS_handle(t *testing.T) {
	a := assert.New(t, false)
	tr := tree.NewTestTree(a, false, syntax.NewInterceptors())
	a.NotError(tr.Add("/path", nil, http.MethodGet, http.MethodDelete))
	ps := syntax.NewParams("/path")
	node, _, exists := tr.Handler(ps, http.MethodGet)
	a.NotNil(node).Zero(ps.Count()).True(exists)

	// deny

	c := &CORS{}
	a.NotError(c.sanitize())
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	c.handle(node, w.Header(), r)
	a.Empty(w.Header().Get("Access-Control-Allow-Origin"))

	// allowed

	c = AllowedCORS()
	a.NotError(c.sanitize())
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Request()
	c.handle(node, w.Header(), r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "*")
	// 非预检，没有此报头
	a.Empty(w.Header().Get("Access-Control-Allow-Methods")).
		Empty(w.Header().Get("Access-Control-Max-Age")).
		Empty(w.Header().Get("Access-Control-Allow-Headers"))

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Origin", "http://example.com").Request()

	c.handle(node, w.Header(), r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "*")
	// 非预检，没有此报头
	a.Empty(w.Header().Get("Access-Control-Allow-Methods")).
		Empty(w.Header().Get("Access-Control-Max-Age")).
		Empty(w.Header().Get("Access-Control-Allow-Headers"))

	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").Header("Origin", "http://example.com").Request()

	c.handle(node, w.Header(), r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "*")
	// 非预检，没有此报头
	a.Empty(w.Header().Get("Access-Control-Allow-Methods")).
		Empty(w.Header().Get("Access-Control-Max-Age")).
		Empty(w.Header().Get("Access-Control-Allow-Headers"))

	// preflight
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").
		Header("Origin", "http://example.com").
		Header("Access-Control-Request-Method", "GET").
		Request()
	c.handle(node, w.Header(), r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "*")
	a.Equal(w.Header().Get("Access-Control-Allow-Methods"), "DELETE, GET, HEAD, OPTIONS")

	// preflight，但是方法不被允许
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").
		Header("Origin", "http://example.com").
		Header("Access-Control-Request-Method", "PATCH").
		Request()
	c.handle(node, w.Header(), r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Methods"), "")

	// custom cors
	c = &CORS{
		Origins:          []string{"https://example.com/"},
		ExposedHeaders:   []string{"h1"},
		MaxAge:           50,
		AllowCredentials: true,
	}
	a.NotError(c.sanitize())

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Origin", "https://example.com/").
		Request()
	c.handle(node, w.Header(), r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "https://example.com/")
	// 非预检，没有此报头
	a.Empty(w.Header().Get("Access-Control-Allow-Methods")).
		Empty(w.Header().Get("Access-Control-Max-Age")).
		Empty(w.Header().Get("Access-Control-Allow-Headers"))

	// preflight
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").
		Header("Origin", "https://example.com/").
		Header("Access-Control-Request-Headers", "h1").
		Request()
	c.handle(node, w.Header(), r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "https://example.com/")
	a.Equal(w.Header().Get("Access-Control-Allow-Headers"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Credentials"), "true")
	a.Equal(w.Header().Get("Access-Control-Expose-Headers"), "h1")
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "https://example.com/")

	// preflight，但是报头不被允许
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").
		Header("Origin", "https://example.com/").
		Header("Access-Control-Request-Method", "GET").
		Header("Access-Control-Request-Headers", "deny").
		Request()
	c.handle(node, w.Header(), r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Headers"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Credentials"), "")

	// preflight，origin 不匹配
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").
		Header("Origin", "https://deny.com/").
		Header("Access-Control-Request-Method", "GET").
		Request()
	c.handle(node, w.Header(), r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Headers"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Credentials"), "")

	// deny

	c = &CORS{}
	a.NotError(c.sanitize())
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Request()
	c.handle(node, w.Header(), r)
	a.Empty(w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_headerIsAllowed(t *testing.T) {
	a := assert.New(t, false)

	// Deny

	c := &CORS{}
	a.NotError(c.sanitize())

	r := rest.Get(a, "/").Request()
	a.True(c.headerIsAllowed(r))

	r = rest.Get(a, "/").Header("Access-Control-Request-Headers", "h1").Request()
	a.False(c.headerIsAllowed(r))

	// Allowed

	c = AllowedCORS()
	a.NotNil(c).NotError(c.sanitize())

	r = rest.Get(a, "/").Request()
	a.True(c.headerIsAllowed(r))

	r = rest.Get(a, "/").Header("Access-Control-Request-Headers", "h1").Request()
	a.True(c.headerIsAllowed(r))

	// 自定义
	c = &CORS{AllowHeaders: []string{"h1", "h2"}}
	a.NotError(c.sanitize())

	r = rest.Get(a, "/").Request()
	a.True(c.headerIsAllowed(r))

	r = rest.Get(a, "/").Header("Access-Control-Request-Headers", "h1").Request()
	a.True(c.headerIsAllowed(r))

	// 不存在的报头
	r = rest.Get(a, "/").Request()
	a.True(c.headerIsAllowed(r))

	r = rest.Get(a, "/").Header("Access-Control-Request-Headers", "h100").Request()
	a.False(c.headerIsAllowed(r))
}
