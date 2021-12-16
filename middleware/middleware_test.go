// SPDX-License-Identifier: MIT

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
)

func buildMiddleware(a *assert.Assertion, text string) Func {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r) // 先输出被包含的内容
			_, err := w.Write([]byte(text))
			a.NotError(err)
		})
	}
}

func TestMiddlewares(t *testing.T) {
	a := assert.New(t, false)

	ms := NewMiddlewares(rest.BuildHandler(a, 201, "", nil))
	a.NotNil(ms)
	ms.Append(buildMiddleware(a, "rp1")).
		Prepend(buildMiddleware(a, "ra1")).
		Append(buildMiddleware(a, "rp2")).
		Prepend(buildMiddleware(a, "ra2"))

	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/get", nil)
	a.NotError(err).NotNil(r)
	ms.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Equal(w.Body.String(), "rp2rp1ra1ra2") // buildHandler 导致顶部的后输出

	// Middlewares.Reset()

	ms.Reset()
	w = httptest.NewRecorder()
	r, err = http.NewRequest(http.MethodGet, "/get", nil)
	a.NotError(err).NotNil(r)
	ms.ServeHTTP(w, r)
	a.Equal(w.Code, 201).
		Empty(w.Body.String())
}