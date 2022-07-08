// SPDX-License-Identifier: MIT

package options

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"

	"github.com/issue9/mux/v7/internal/syntax"
	"github.com/issue9/mux/v7/internal/tree"
	"github.com/issue9/mux/v7/types"
)

func TestOptions_sanitize(t *testing.T) {
	a := assert.New(t, false)

	o, err := Build()
	a.NotError(err).
		NotNil(o).
		NotNil(o.CORS)

	// URLDomain

	o, err = Build(func(o *Options) { o.URLDomain = "https://example.com" })
	a.NotError(err).NotNil(o).Equal(o.URLDomain, "https://example.com")

	o, err = Build(func(o *Options) { o.URLDomain = "https://example.com/" })
	a.NotError(err).NotNil(o).Equal(o.URLDomain, "https://example.com")

	o, err = Build(func(o *Options) { o.CORS = &CORS{AllowCredentials: true, Origins: []string{"*"}} })
	a.Error(err).Nil(o)
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

func TestCORS_Handle(t *testing.T) {
	a := assert.New(t, false)
	tr := tree.NewTestTree(a, false, syntax.NewInterceptors())
	a.NotError(tr.Add("/path", nil, http.MethodGet, http.MethodDelete))
	ctx := types.NewContext()
	ctx.Path = "/path"
	node, _, exists := tr.Handler(ctx, http.MethodGet)
	a.NotNil(node).Zero(ctx.Count()).True(exists)

	// deny

	c := &CORS{}
	a.NotError(c.sanitize())
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	c.Handle(node, w.Header(), r)
	a.Empty(w.Header().Get("Access-Control-Allow-Origin"))

	// allowed

	c = allowedCORS()
	a.NotError(c.sanitize())
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Request()
	c.Handle(node, w.Header(), r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "*")
	// 非预检，没有此报头
	a.Empty(w.Header().Get("Access-Control-Allow-Methods")).
		Empty(w.Header().Get("Access-Control-Max-Age")).
		Empty(w.Header().Get("Access-Control-Allow-Headers"))

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Origin", "http://example.com").Request()

	c.Handle(node, w.Header(), r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "*")
	// 非预检，没有此报头
	a.Empty(w.Header().Get("Access-Control-Allow-Methods")).
		Empty(w.Header().Get("Access-Control-Max-Age")).
		Empty(w.Header().Get("Access-Control-Allow-Headers"))

	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").Header("Origin", "http://example.com").Request()

	c.Handle(node, w.Header(), r)
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
	c.Handle(node, w.Header(), r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "*")
	a.Equal(w.Header().Get("Access-Control-Allow-Methods"), "DELETE, GET, HEAD, OPTIONS")

	// preflight，但是方法不被允许
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").
		Header("Origin", "http://example.com").
		Header("Access-Control-Request-Method", "PATCH").
		Request()
	c.Handle(node, w.Header(), r)
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
	c.Handle(node, w.Header(), r)
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
	c.Handle(node, w.Header(), r)
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
	c.Handle(node, w.Header(), r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Headers"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Credentials"), "")

	// preflight，origin 不匹配
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").
		Header("Origin", "https://deny.com/").
		Header("Access-Control-Request-Method", "GET").
		Request()
	c.Handle(node, w.Header(), r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Headers"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Credentials"), "")

	// deny

	c = &CORS{}
	a.NotError(c.sanitize())
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Request()
	c.Handle(node, w.Header(), r)
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

	c = allowedCORS()
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

func allowedCORS() *CORS {
	return &CORS{
		Origins:      []string{"*"},
		AllowHeaders: []string{"*"},
		MaxAge:       3600,
	}
}
