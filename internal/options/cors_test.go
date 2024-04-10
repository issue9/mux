// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package options

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/assert/v4/rest"

	"github.com/issue9/mux/v8/header"
	"github.com/issue9/mux/v8/internal/syntax"
	"github.com/issue9/mux/v8/internal/tree"
	"github.com/issue9/mux/v8/types"
)

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
	tr := tree.NewTestTree(a, false, false, syntax.NewInterceptors())
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
	a.Empty(w.Header().Get(header.AccessControlAllowOrigin))

	// allowed

	c = allowedCORS()
	a.NotError(c.sanitize())
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Request()
	c.Handle(node, w.Header(), r)
	a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "*")
	// 非预检，没有此报头
	a.Empty(w.Header().Get(header.AccessControlAllowMethods)).
		Empty(w.Header().Get(header.AccessControlMaxAge)).
		Empty(w.Header().Get(header.AccessControlAllowHeaders))

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header(header.Origin, "http://example.com").Request()

	c.Handle(node, w.Header(), r)
	a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "*")
	// 非预检，没有此报头
	a.Empty(w.Header().Get(header.AccessControlAllowMethods)).
		Empty(w.Header().Get(header.AccessControlMaxAge)).
		Empty(w.Header().Get(header.AccessControlAllowHeaders))

	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").Header(header.Origin, "http://example.com").Request()

	c.Handle(node, w.Header(), r)
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
	c.Handle(node, w.Header(), r)
	a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "*")
	a.Equal(w.Header().Get(header.AccessControlAllowMethods), "DELETE, GET, HEAD, OPTIONS")

	// preflight，但是方法不被允许
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").
		Header(header.Origin, "http://example.com").
		Header(header.AccessControlRequestMethod, "PATCH").
		Request()
	c.Handle(node, w.Header(), r)
	a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "")
	a.Equal(w.Header().Get(header.AccessControlAllowMethods), "")

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
		Header(header.Origin, "https://example.com/").
		Request()
	c.Handle(node, w.Header(), r)
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
	c.Handle(node, w.Header(), r)
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
	c.Handle(node, w.Header(), r)
	a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "")
	a.Equal(w.Header().Get(header.AccessControlAllowHeaders), "")
	a.Equal(w.Header().Get(header.AccessControlAllowCredentials), "")

	// preflight，origin 不匹配
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path").
		Header(header.Origin, "https://deny.com/").
		Header(header.AccessControlRequestMethod, "GET").
		Request()
	c.Handle(node, w.Header(), r)
	a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "")
	a.Equal(w.Header().Get(header.AccessControlAllowHeaders), "")
	a.Equal(w.Header().Get(header.AccessControlAllowCredentials), "")

	// deny

	c = &CORS{}
	a.NotError(c.sanitize())
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Request()
	c.Handle(node, w.Header(), r)
	a.Empty(w.Header().Get(header.AccessControlAllowOrigin))
}

func TestCORS_headerIsAllowed(t *testing.T) {
	a := assert.New(t, false)

	// Deny

	c := &CORS{}
	a.NotError(c.sanitize())

	r := rest.Get(a, "/").Request()
	a.True(c.headerIsAllowed(r))

	r = rest.Get(a, "/").Header(header.AccessControlRequestHeaders, "h1").Request()
	a.False(c.headerIsAllowed(r))

	// Allowed

	c = allowedCORS()
	a.NotNil(c).NotError(c.sanitize())

	r = rest.Get(a, "/").Request()
	a.True(c.headerIsAllowed(r))

	r = rest.Get(a, "/").Header(header.AccessControlRequestHeaders, "h1").Request()
	a.True(c.headerIsAllowed(r))

	// 自定义
	c = &CORS{AllowHeaders: []string{"h1", "h2"}}
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

func allowedCORS() *CORS {
	return &CORS{
		Origins:      []string{"*"},
		AllowHeaders: []string{"*"},
		MaxAge:       3600,
	}
}
