// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v4/group"
)

func buildMiddleware(a *assert.Assertion, text string) MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r) // 先输出被包含的内容
			a.NotError(w.Write([]byte(text)))
		})
	}
}

func TestMux_PrependMiddleware(t *testing.T) {
	a := assert.New(t)
	m := Default()
	a.NotNil(m)
	def, ok := m.NewRouter("def", group.MatcherFunc(group.Any), Allowed())
	a.True(ok).NotNil(def)

	def.Get("/get", buildHandler(201))
	m.PrependMiddleware(buildMiddleware(a, "1")).
		PrependMiddleware(buildMiddleware(a, "2"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/get", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Equal(w.Body.String(), "12")

	// CleanMiddlewares

	m.CleanMiddlewares()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/get", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Empty(w.Body.String())
}

func TestMux_AppendMiddleware(t *testing.T) {
	a := assert.New(t)
	m := Default()
	a.NotNil(m)
	def, ok := m.NewRouter("def", group.MatcherFunc(group.Any), Allowed())
	a.True(ok).NotNil(def)

	def.Get("/get", buildHandler(201))
	m.AppendMiddleware(buildMiddleware(a, "1")).
		AppendMiddleware(buildMiddleware(a, "2"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/get", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Equal(w.Body.String(), "21")

	// CleanMiddlewares

	m.CleanMiddlewares()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/get", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Empty(w.Body.String())
}

func TestMux_AddMiddleware(t *testing.T) {
	a := assert.New(t)
	m := Default()
	a.NotNil(m)
	def, ok := m.NewRouter("def", group.MatcherFunc(group.Any), Allowed())
	a.True(ok).NotNil(def)

	def.Get("/get", buildHandler(201))
	m.AppendMiddleware(buildMiddleware(a, "p1")).
		PrependMiddleware(buildMiddleware(a, "a1")).
		AppendMiddleware(buildMiddleware(a, "p2")).
		PrependMiddleware(buildMiddleware(a, "a2"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/get", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Equal(w.Body.String(), "p2p1a1a2") // buildHandler 导致顶部的后输出

	// CleanMiddlewares

	m.CleanMiddlewares()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/get", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Empty(w.Body.String())
}

func TestRouter_AddMiddleware(t *testing.T) {
	a := assert.New(t)
	m := Default()
	a.NotNil(m)

	m.AppendMiddleware(buildMiddleware(a, "p1")).
		PrependMiddleware(buildMiddleware(a, "a1")).
		AppendMiddleware(buildMiddleware(a, "p2")).
		PrependMiddleware(buildMiddleware(a, "a2"))

	def, ok := m.NewRouter("def", group.MatcherFunc(group.Any), Allowed())
	a.True(ok).NotNil(def)
	def.Get("/get", buildHandler(201))
	def.AppendMiddleware(buildMiddleware(a, "rp1")).
		PrependMiddleware(buildMiddleware(a, "ra1")).
		AppendMiddleware(buildMiddleware(a, "rp2")).
		PrependMiddleware(buildMiddleware(a, "ra2"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/get", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Equal(w.Body.String(), "rp2rp1ra1ra2p2p1a1a2") // buildHandler 导致顶部的后输出

	// CleanMiddlewares

	m.CleanMiddlewares()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/get", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Equal(w.Body.String(), "rp2rp1ra1ra2")

	def.CleanMiddlewares()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/get", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Empty(w.Body.String())
}
