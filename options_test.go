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

	"github.com/issue9/mux/v6/internal/syntax"
	"github.com/issue9/mux/v6/internal/tree"
)

func TestOption(t *testing.T) {
	a := assert.New(t, false)

	r := NewRouter("", nil)
	a.NotNil(r).
		False(r.options.CaseInsensitive).
		NotNil(r.options.MethodNotAllowed)

	r = NewRouter("", nil, CaseInsensitive)
	a.NotNil(r).
		True(r.options.CaseInsensitive).
		NotNil(r.options.MethodNotAllowed)

	notFound := rest.BuildHandler(a, 404, "", nil)
	methodNotAllowed := rest.BuildHandler(a, 405, "", nil)
	r = NewRouter("", nil, NotFound(notFound), MethodNotAllowed(methodNotAllowed))
	a.NotNil(r).
		False(r.options.CaseInsensitive).
		Equal(r.options.MethodNotAllowed, methodNotAllowed).
		Equal(r.options.NotFound, notFound)

	r = NewRouter("", nil, CORS([]string{"https://example.com"}, nil, nil, 3600, false))
	a.NotNil(r).
		Equal(r.options.CORS.Origins, []string{"https://example.com"}).
		Nil(r.options.CORS.AllowHeaders).
		Equal(r.options.CORS.MaxAge, 3600)

	r = NewRouter("", nil, CORS([]string{"https://example.com"}, nil, nil, 0, true))
	a.NotNil(r)

	a.Panic(func() {
		r = NewRouter("", nil, CORS([]string{"*"}, nil, nil, 0, true))
	})
}

func TestRecovery(t *testing.T) {
	a := assert.New(t, false)

	p := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic("test") })

	router := NewRouter("", nil)
	a.NotNil(router).Nil(router.options.RecoverFunc)
	router.Get("/path", p)
	a.Panic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
	})

	// WriterRecovery
	out := new(bytes.Buffer)
	router = NewRouter("", nil, WriterRecovery(404, out))
	a.NotNil(router).NotNil(router.options.RecoverFunc)
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
	router = NewRouter("", nil, LogRecovery(405, l))
	a.NotNil(router).NotNil(router.options.RecoverFunc)
	router.Get("/path", p)
	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
		a.Equal(405, w.Code).
			Contains(out.String(), "test")
	})

	// HTTPRecovery
	router = NewRouter("", nil, HTTPRecovery(406))
	a.NotNil(router).NotNil(router.options.RecoverFunc)
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

	o := &options{}
	a.NotError(o.sanitize())
	a.NotNil(o.CORS).
		NotNil(o.NotFound).
		NotNil(o.MethodNotAllowed)

	rest.Get(a, "/").Do(o.MethodNotAllowed).Status(405).StringBody(http.StatusText(http.StatusMethodNotAllowed) + "\n")

	// URLDomain

	o = &options{URLDomain: "https://example.com"}
	a.NotError(o.sanitize())
	a.Equal(o.URLDomain, "https://example.com")
	o = &options{URLDomain: "https://example.com/"}
	a.NotError(o.sanitize())
	a.Equal(o.URLDomain, "https://example.com")
}

func TestBuildOptions(t *testing.T) {
	a := assert.New(t, false)

	o, err := buildOptions()
	a.NotError(err).
		NotNil(o).
		False(o.CaseInsensitive).
		NotNil(o.CORS).
		NotNil(o.NotFound).
		NotNil(o.Interceptors)

	o, err = buildOptions(func(o *options) { o.CaseInsensitive = true })
	a.NotError(err).
		NotNil(o).
		True(o.CaseInsensitive)

	o, err = buildOptions(func(o *options) {
		o.CORS = &cors{
			Origins:          []string{"*"},
			AllowCredentials: true,
		}
	})
	a.ErrorString(err, "不能同时成立").Nil(o)
}

func TestCORS_sanitize(t *testing.T) {
	a := assert.New(t, false)

	c := &cors{}
	a.NotError(c.sanitize())
	a.True(c.deny).
		False(c.anyHeaders).
		Empty(c.allowHeadersString).
		False(c.anyOrigins).
		Empty(c.exposedHeadersString).
		Empty(c.maxAgeString)

	c = &cors{
		Origins: []string{"*"},
		MaxAge:  50,
	}
	a.NotError(c.sanitize())
	a.True(c.anyOrigins).Equal(c.maxAgeString, "50")

	c = &cors{
		Origins: []string{"*"},
		MaxAge:  -1,
	}
	a.NotError(c.sanitize())
	a.True(c.anyOrigins).Equal(c.maxAgeString, "-1")

	c = &cors{
		MaxAge: -2,
	}
	a.ErrorString(c.sanitize(), "maxAge 的值只能是 >= -1")

	c = &cors{
		Origins:          []string{"*"},
		AllowCredentials: true,
	}
	a.ErrorString(c.sanitize(), "不能同时成立")

	c = &cors{
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
	tr := tree.New(false, syntax.NewInterceptors())
	a.NotError(tr.Add("/path", nil, http.MethodGet, http.MethodDelete))
	node, ps := tr.Route("/path")
	a.NotNil(node).Zero(ps.Count())

	a.Panic(func() {
		buildOptions(DenyCORS, nil)
	}, "option 不能为空值")

	// deny

	o, err := buildOptions(DenyCORS)
	a.NotError(err).NotNil(o)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	o.handleCORS(node, w, r)
	a.Empty(w.Header().Get("Access-Control-Allow-Origin"))

	// allowed

	o, err = buildOptions(AllowedCORS)
	a.NotError(err).NotNil(o)

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Request()
	o.handleCORS(node, w, r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "*")
	// 非预检，没有此报头
	a.Empty(w.Header().Get("Access-Control-Allow-Methods")).
		Empty(w.Header().Get("Access-Control-Max-Age")).
		Empty(w.Header().Get("Access-Control-Allow-Headers"))

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Origin", "http://example.com").Request()

	o.handleCORS(node, w, r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "*")
	// 非预检，没有此报头
	a.Empty(w.Header().Get("Access-Control-Allow-Methods")).
		Empty(w.Header().Get("Access-Control-Max-Age")).
		Empty(w.Header().Get("Access-Control-Allow-Headers"))

	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").Header("Origin", "http://example.com").Request()

	o.handleCORS(node, w, r)
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
	o.handleCORS(node, w, r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "*")
	a.Equal(w.Header().Get("Access-Control-Allow-Methods"), "DELETE, GET, HEAD, OPTIONS")

	// preflight，但是方法不被允许
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").
		Header("Origin", "http://example.com").
		Header("Access-Control-Request-Method", "PATCH").
		Request()
	o.handleCORS(node, w, r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Methods"), "")

	// custom cors
	o, err = buildOptions(CORS([]string{"https://example.com/"}, nil, []string{"h1"}, 50, true))
	a.NotError(err).NotNil(o)

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Origin", "https://example.com/").
		Request()
	o.handleCORS(node, w, r)
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
	o.handleCORS(node, w, r)
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
	o.handleCORS(node, w, r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Headers"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Credentials"), "")

	// preflight，origin 不匹配
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").
		Header("Origin", "https://deny.com/").
		Header("Access-Control-Request-Method", "GET").
		Request()
	o.handleCORS(node, w, r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Headers"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Credentials"), "")

	// deny

	o, err = buildOptions()
	a.NotError(err).NotNil(o)
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Request()
	o.handleCORS(node, w, r)
	a.Empty(w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_headerIsAllowed(t *testing.T) {
	a := assert.New(t, false)
	o := &options{}

	// Deny

	DenyCORS(o)
	c := o.CORS
	a.NotNil(c).NotError(c.sanitize())

	r := rest.Get(a, "/").Request()
	a.True(c.headerIsAllowed(r))

	r = rest.Get(a, "/").Header("Access-Control-Request-Headers", "h1").Request()
	a.False(c.headerIsAllowed(r))

	// Allowed

	AllowedCORS(o)
	c = o.CORS
	a.NotNil(c).NotError(c.sanitize())

	r = rest.Get(a, "/").Request()
	a.True(c.headerIsAllowed(r))

	r = rest.Get(a, "/").Header("Access-Control-Request-Headers", "h1").Request()
	a.True(c.headerIsAllowed(r))

	// 自定义
	c = &cors{AllowHeaders: []string{"h1", "h2"}}
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
