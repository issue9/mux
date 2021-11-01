// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func buildMiddleware(a *assert.Assertion, text string) MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r) // 先输出被包含的内容
			a.NotError(w.Write([]byte(text)))
		})
	}
}

func TestRouter_Middleware(t *testing.T) {
	a := assert.New(t)

	def := NewRouter("")
	a.NotNil(def)
	def.Get("/get", buildHandler(201))
	ms := def.Middlewares()
	ms.Append(buildMiddleware(a, "rp1")).
		Prepend(buildMiddleware(a, "ra1")).
		Append(buildMiddleware(a, "rp2")).
		Prepend(buildMiddleware(a, "ra2"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/get", nil)
	def.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Equal(w.Body.String(), "rp2rp1ra1ra2") // buildHandler 导致顶部的后输出

	// CleanMiddlewares

	ms.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/get", nil)
	def.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Empty(w.Body.String())
}
