// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func buildMiddleware(a *assert.Assertion, text string) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
			a.NotError(w.Write([]byte(text)))
		})
	}
}

func TestMux_Append(t *testing.T) {
	a := assert.New(t)
	mux := Default()
	a.NotNil(mux)

	mux.Get("/get", buildHandler(201))
	mux.Append(buildMiddleware(a, "1"), buildMiddleware(a, "2"))
	mux.Append(buildMiddleware(a, "3"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/get", nil)
	mux.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Equal(w.Body.String(), "321")

	// reset

	mux.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/get", nil)
	mux.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Empty(w.Body.String())
}

func TestMux_Prepend(t *testing.T) {
	a := assert.New(t)
	mux := Default()
	a.NotNil(mux)

	mux.Get("/get", buildHandler(201))
	mux.Prepend(buildMiddleware(a, "1"), buildMiddleware(a, "2"))
	mux.Prepend(buildMiddleware(a, "3"), buildMiddleware(a, "4"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/get", nil)
	mux.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Equal(w.Body.String(), "2143")

	// reset

	mux.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/get", nil)
	mux.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Empty(w.Body.String())
}

func TestMux_Append_Prepend(t *testing.T) {
	a := assert.New(t)
	mux := Default()
	a.NotNil(mux)

	mux.Get("/get", buildHandler(201))
	mux.Prepend(buildMiddleware(a, "p1"))
	mux.Append(buildMiddleware(a, "a1"))
	mux.Prepend(buildMiddleware(a, "p2"))
	mux.Append(buildMiddleware(a, "a2"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/get", nil)
	mux.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Equal(w.Body.String(), "a2a1p1p2")

	// reset

	mux.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/get", nil)
	mux.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Empty(w.Body.String())
}
