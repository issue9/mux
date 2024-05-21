// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package mux

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"github.com/issue9/assert/v4/rest"

	"github.com/issue9/mux/v9/header"
	"github.com/issue9/mux/v9/internal/syntax"
	"github.com/issue9/mux/v9/internal/tree"
	"github.com/issue9/mux/v9/types"
)

func TestOption(t *testing.T) {
	a := assert.New(t, false)

	r := newRouter(a, "")
	a.NotNil(r)

	r = newRouter(a, "", WithCORS([]string{"https://example.com"}, nil, nil, 3600, false))
	a.NotNil(r).
		Equal(r.cors.Origins, []string{"https://example.com"}).
		Nil(r.cors.AllowHeaders).
		Equal(r.cors.MaxAge, 3600)

	r = newRouter(a, "", WithCORS([]string{"https://example.com"}, nil, nil, 0, true))
	a.NotNil(r)

	a.Panic(func() {
		r = newRouter(a, "", WithCORS([]string{"*"}, nil, nil, 0, true))
	})
}

func TestRecovery(t *testing.T) {
	a := assert.New(t, false)

	p := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic("panic test") })

	// 未指定 Recovery

	router := newRouter(a, "")
	a.NotNil(router).Nil(router.recoverFunc)
	router.Get("/path", p)
	a.Panic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
	})

	// WriterRecovery

	out := new(bytes.Buffer)
	router = newRouter(a, "", WithWriteRecovery(404, out))
	a.NotNil(router).NotNil(router.recoverFunc)
	router.Get("/path", p)

	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
		a.Wait(time.Microsecond*500).
			Contains(out.String(), "panic test", out.String()).
			Contains(out.String(), "options_test.go:48", out.String()).
			Equal(w.Code, 404)
	})

	// LogRecovery

	out = new(bytes.Buffer)
	l := log.New(out, "log:", 0)
	router = newRouter(a, "", WithLogRecovery(405, l))
	a.NotNil(router).NotNil(router.recoverFunc)
	router.Get("/path", p)
	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
		a.Equal(405, w.Code)
		lines := strings.Split(out.String(), "\n")
		a.Contains(lines[0], "panic test")                                  // 保证第一行是 panic 输出的信息
		a.Contains(lines[1], "TestRecovery.func1")                          // 保证第二行是 panic 函数名
		a.True(strings.HasSuffix(lines[2], "options_test.go:48"), lines[2]) // 保证第三行是 panic 的行号
	})

	// StatusRecovery

	router = newRouter(a, "", WithStatusRecovery(406))
	a.NotNil(router).NotNil(router.recoverFunc)
	router.Get("/path", p)
	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
		a.Equal(w.Code, 406)
	})
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

func TestCORS_Handle(t *testing.T) {
	a := assert.New(t, false)
	tr := tree.NewTestTree(a, false, false, syntax.NewInterceptors())
	a.NotError(tr.Add("/path", nil, nil, http.MethodGet, http.MethodDelete))
	ctx := types.NewContext()
	ctx.Path = "/path"
	node, _, exists := tr.Handler(ctx, http.MethodGet)
	a.NotNil(node).Zero(ctx.Count()).True(exists)

	// deny

	c := &cors{}
	a.NotError(c.sanitize())
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	c.handle(node, w.Header(), r)
	a.Empty(w.Header().Get(header.AccessControlAllowOrigin))

	// allowed

	c = &cors{MaxAge: 3600, Origins: []string{"*"}, AllowHeaders: []string{"*"}}
	a.NotError(c.sanitize())
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Request()
	c.handle(node, w.Header(), r)
	a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "*")
	// 非预检，没有此报头
	a.Empty(w.Header().Get(header.AccessControlAllowMethods)).
		Empty(w.Header().Get(header.AccessControlMaxAge)).
		Empty(w.Header().Get(header.AccessControlAllowHeaders))

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header(header.Origin, "http://example.com").Request()

	c.handle(node, w.Header(), r)
	a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "*")
	// 非预检，没有此报头
	a.Empty(w.Header().Get(header.AccessControlAllowMethods)).
		Empty(w.Header().Get(header.AccessControlMaxAge)).
		Empty(w.Header().Get(header.AccessControlAllowHeaders))

	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").Header(header.Origin, "http://example.com").Request()

	c.handle(node, w.Header(), r)
	a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "*")
	// 非预检，没有此报头
	a.Empty(w.Header().Get(header.AccessControlAllowMethods)).
		Empty(w.Header().Get(header.AccessControlMaxAge)).
		Empty(w.Header().Get(header.AccessControlAllowHeaders))

	// preflight
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").
		Header(header.Origin, "http://example.com").
		Header(header.AccessControlRequestMethod, "GET").
		Request()
	c.handle(node, w.Header(), r)
	a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "*")
	a.Equal(w.Header().Get(header.AccessControlAllowMethods), "DELETE, GET, HEAD, OPTIONS")

	// preflight，但是方法不被允许
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").
		Header(header.Origin, "http://example.com").
		Header(header.AccessControlRequestMethod, "PATCH").
		Request()
	c.handle(node, w.Header(), r)
	a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "")
	a.Equal(w.Header().Get(header.AccessControlAllowMethods), "")

	// custom cors
	c = &cors{
		Origins:          []string{"https://example.com/"},
		ExposedHeaders:   []string{"h1"},
		MaxAge:           50,
		AllowCredentials: true,
	}
	a.NotError(c.sanitize())

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header(header.Origin, "https://example.com/").
		Request()
	c.handle(node, w.Header(), r)
	a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "https://example.com/")
	// 非预检，没有此报头
	a.Empty(w.Header().Get(header.AccessControlAllowMethods)).
		Empty(w.Header().Get(header.AccessControlMaxAge)).
		Empty(w.Header().Get(header.AccessControlAllowHeaders))

	// preflight
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").
		Header(header.Origin, "https://example.com/").
		Header(header.AccessControlRequestHeaders, "h1").
		Request()
	c.handle(node, w.Header(), r)
	a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "https://example.com/")
	a.Equal(w.Header().Get(header.AccessControlAllowHeaders), "")
	a.Equal(w.Header().Get(header.AccessControlAllowCredentials), "true")
	a.Equal(w.Header().Get(header.AccessControlExposeHeaders), "h1")
	a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "https://example.com/")

	// preflight，但是报头不被允许
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").
		Header(header.Origin, "https://example.com/").
		Header(header.AccessControlRequestMethod, "GET").
		Header(header.AccessControlRequestHeaders, "deny").
		Request()
	c.handle(node, w.Header(), r)
	a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "")
	a.Equal(w.Header().Get(header.AccessControlAllowHeaders), "")
	a.Equal(w.Header().Get(header.AccessControlAllowCredentials), "")

	// preflight，origin 不匹配
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").
		Header(header.Origin, "https://deny.com/").
		Header(header.AccessControlRequestMethod, "GET").
		Request()
	c.handle(node, w.Header(), r)
	a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "")
	a.Equal(w.Header().Get(header.AccessControlAllowHeaders), "")
	a.Equal(w.Header().Get(header.AccessControlAllowCredentials), "")

	// deny

	c = &cors{}
	a.NotError(c.sanitize())
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Request()
	c.handle(node, w.Header(), r)
	a.Empty(w.Header().Get(header.AccessControlAllowOrigin))
}

func TestCORS_headerIsAllowed(t *testing.T) {
	a := assert.New(t, false)

	// Deny

	c := &cors{}
	a.NotError(c.sanitize())

	r := rest.Get(a, "/").Request()
	a.True(c.headerIsAllowed(r))

	r = rest.Get(a, "/").Header(header.AccessControlRequestHeaders, "h1").Request()
	a.False(c.headerIsAllowed(r))

	// Allowed

	c = &cors{MaxAge: 3600, Origins: []string{"*"}, AllowHeaders: []string{"*"}}
	a.NotNil(c).NotError(c.sanitize())

	r = rest.Get(a, "/").Request()
	a.True(c.headerIsAllowed(r))

	r = rest.Get(a, "/").Header(header.AccessControlRequestHeaders, "h1").Request()
	a.True(c.headerIsAllowed(r))

	// 自定义
	c = &cors{AllowHeaders: []string{"h1", "h2"}}
	a.NotError(c.sanitize())

	r = rest.Get(a, "/").Request()
	a.True(c.headerIsAllowed(r))

	r = rest.Get(a, "/").Header(header.AccessControlRequestHeaders, "h1").Request()
	a.True(c.headerIsAllowed(r))

	// 不存在的报头
	r = rest.Get(a, "/").Request()
	a.True(c.headerIsAllowed(r))

	r = rest.Get(a, "/").Header(header.AccessControlRequestHeaders, "h100").Request()
	a.False(c.headerIsAllowed(r))
}

func TestOptions_sanitize(t *testing.T) {
	a := assert.New(t, false)

	o, err := buildOption()
	a.NotError(err).
		NotNil(o).
		NotNil(o.cors)

	// URLDomain

	o, err = buildOption(func(o *options) { o.urlDomain = "https://example.com" })
	a.NotError(err).NotNil(o).Equal(o.urlDomain, "https://example.com")

	o, err = buildOption(func(o *options) { o.urlDomain = "https://example.com/" })
	a.NotError(err).NotNil(o).Equal(o.urlDomain, "https://example.com")

	o, err = buildOption(func(o *options) { o.cors = &cors{AllowCredentials: true, Origins: []string{"*"}} })
	a.Error(err).Nil(o)
}
