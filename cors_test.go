// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v5/internal/tree"
)

func TestCORS_sanitize(t *testing.T) {
	a := assert.New(t)

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
	a.ErrorString(c.sanitize(), "MaxAge 的值只能是 >= -1")

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
	a := assert.New(t)
	tree2 := tree.New(true, true)
	a.NotError(tree2.Add("/path", nil, http.MethodGet, http.MethodDelete))
	hs, ps := tree2.Route("/path")
	a.NotNil(hs).Empty(ps)

	// deny

	c := DeniedCORS()
	a.NotError(c.sanitize())
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	c.handle(hs, w, r)
	a.Empty(w.Header().Get("Access-Control-Allow-Origin"))

	// allowed

	c = AllowedCORS()
	a.NotError(c.sanitize())

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	c.handle(hs, w, r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "*")
	// 非预检，没有此报头
	a.Empty(w.Header().Get("Access-Control-Allow-Methods")).
		Empty(w.Header().Get("Access-Control-Max-Age")).
		Empty(w.Header().Get("Access-Control-Allow-Headers"))

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Origin", "http://example.com")
	c.handle(hs, w, r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "*")
	// 非预检，没有此报头
	a.Empty(w.Header().Get("Access-Control-Allow-Methods")).
		Empty(w.Header().Get("Access-Control-Max-Age")).
		Empty(w.Header().Get("Access-Control-Allow-Headers"))

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodOptions, "/path", nil)
	r.Header.Set("Origin", "http://example.com")
	c.handle(hs, w, r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "*")
	// 非预检，没有此报头
	a.Empty(w.Header().Get("Access-Control-Allow-Methods")).
		Empty(w.Header().Get("Access-Control-Max-Age")).
		Empty(w.Header().Get("Access-Control-Allow-Headers"))

	// preflight
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodOptions, "/path", nil)
	r.Header.Set("Origin", "http://example.com")
	r.Header.Set("Access-Control-Request-Method", "GET")
	c.handle(hs, w, r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "*")
	a.Equal(w.Header().Get("Access-Control-Allow-Methods"), "DELETE, GET, OPTIONS")

	// preflight，但是方法不被允许
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodOptions, "/path", nil)
	r.Header.Set("Origin", "http://example.com")
	r.Header.Set("Access-Control-Request-Method", "PATCH")
	c.handle(hs, w, r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Methods"), "")

	// custom cors
	c = &CORS{
		ExposedHeaders:   []string{"h1"},
		Origins:          []string{"https://example.com/"},
		AllowCredentials: true,
		MaxAge:           50,
	}
	a.NotError(c.sanitize())

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Origin", "https://example.com/")
	c.handle(hs, w, r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "https://example.com/")
	// 非预检，没有此报头
	a.Empty(w.Header().Get("Access-Control-Allow-Methods")).
		Empty(w.Header().Get("Access-Control-Max-Age")).
		Empty(w.Header().Get("Access-Control-Allow-Headers"))

	// preflight
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodOptions, "/path", nil)
	r.Header.Set("Origin", "https://example.com/")
	r.Header.Set("Access-Control-Request-Headers", "h1")
	c.handle(hs, w, r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "https://example.com/")
	a.Equal(w.Header().Get("Access-Control-Allow-Headers"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Credentials"), "true")
	a.Equal(w.Header().Get("Access-Control-Expose-Headers"), "h1")
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "https://example.com/")

	// preflight，但是报头不被允许
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodOptions, "/path", nil)
	r.Header.Set("Origin", "https://example.com/")
	r.Header.Set("Access-Control-Request-Method", "GET")
	r.Header.Set("Access-Control-Request-Headers", "deny")
	c.handle(hs, w, r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Headers"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Credentials"), "")

	// preflight，origin 不匹配
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodOptions, "/path", nil)
	r.Header.Set("Origin", "https://deny.com/")
	r.Header.Set("Access-Control-Request-Method", "GET")
	c.handle(hs, w, r)
	a.Equal(w.Header().Get("Access-Control-Allow-Origin"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Headers"), "")
	a.Equal(w.Header().Get("Access-Control-Allow-Credentials"), "")
}

func TestCORS_headerIsAllowed(t *testing.T) {
	a := assert.New(t)

	// Deny

	c := DeniedCORS()
	a.NotNil(c).NotError(c.sanitize())

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	a.True(c.headerIsAllowed(r))

	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Access-Control-Request-Headers", "h1")
	a.False(c.headerIsAllowed(r))

	// Allowed

	c = AllowedCORS()
	a.NotNil(c).NotError(c.sanitize())

	r = httptest.NewRequest(http.MethodGet, "/", nil)
	a.True(c.headerIsAllowed(r))

	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Access-Control-Request-Headers", "h1")
	a.True(c.headerIsAllowed(r))

	// 自定义
	c = &CORS{
		AllowHeaders: []string{"h1", "h2"},
	}
	a.NotError(c.sanitize())

	r = httptest.NewRequest(http.MethodGet, "/", nil)
	a.True(c.headerIsAllowed(r))

	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Access-Control-Request-Headers", "h1")
	a.True(c.headerIsAllowed(r))

	// 不存在的报头
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	a.True(c.headerIsAllowed(r))

	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Access-Control-Request-Headers", "h100")
	a.False(c.headerIsAllowed(r))
}
