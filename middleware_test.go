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
			h.ServeHTTP(w, r)
			a.NotError(w.Write([]byte(text)))
		})
	}
}

func TestMux_insertFirst(t *testing.T) {
	a := assert.New(t)
	m := Default()
	a.NotNil(m)
	def, ok := m.NewRouter("def", group.MatcherFunc(group.Any))
	a.True(ok).NotNil(def)

	def.Get("/get", buildHandler(201))
	m.AddMiddleware(true, buildMiddleware(a, "1")).
		AddMiddleware(true, buildMiddleware(a, "2"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/get", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Equal(w.Body.String(), "21")

	// reset

	m.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/get", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Empty(w.Body.String())
}

func TestMux_insertLast(t *testing.T) {
	a := assert.New(t)
	m := Default()
	a.NotNil(m)
	def, ok := m.NewRouter("def", group.MatcherFunc(group.Any))
	a.True(ok).NotNil(def)

	def.Get("/get", buildHandler(201))
	m.AddMiddleware(false, buildMiddleware(a, "1")).
		AddMiddleware(false, buildMiddleware(a, "2"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/get", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Equal(w.Body.String(), "12")

	// reset

	m.Reset()
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
	def, ok := m.NewRouter("def", group.MatcherFunc(group.Any))
	a.True(ok).NotNil(def)

	def.Get("/get", buildHandler(201))
	m.AddMiddleware(false, buildMiddleware(a, "p1")).
		AddMiddleware(true, buildMiddleware(a, "a1")).
		AddMiddleware(false, buildMiddleware(a, "p2")).
		AddMiddleware(true, buildMiddleware(a, "a2"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/get", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Equal(w.Body.String(), "a2a1p1p2")

	// reset

	m.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/get", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Empty(w.Body.String())
}
