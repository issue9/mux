// SPDX-License-Identifier: MIT

package mux_test

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"

	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/internal/tree"
)

func TestRouter_Middleware(t *testing.T) {
	a := assert.New(t, false)

	def := mux.NewRouter("",
		[]mux.MiddlewareFuncOf[http.Handler]{
			buildMiddleware(a, "m1"),
			buildMiddleware(a, "m2"),
			buildMiddleware(a, "m3"),
			buildMiddleware(a, "m4"),
		},
	)
	a.NotNil(def)
	def.Get("/get", rest.BuildHandler(a, 201, "", nil))

	rest.Get(a, "/get").Do(def).Status(201).StringBody("m1m2m3m4") // buildHandler 导致顶部的后输出
}

func TestMethods(t *testing.T) {
	a := assert.New(t, false)
	a.Equal(mux.Methods(), tree.Methods)
}

func TestCheckSyntax(t *testing.T) {
	a := assert.New(t, false)

	a.NotError(mux.CheckSyntax("/{path"))
	a.NotError(mux.CheckSyntax("/path}"))
	a.Error(mux.CheckSyntax(""))
}

func TestURL(t *testing.T) {
	a := assert.New(t, false)

	url, err := mux.URL("/posts/{id:}", map[string]string{"id": "100"})
	a.NotError(err).Equal(url, "/posts/100")

	url, err = mux.URL("/posts/{id:}", nil)
	a.NotError(err).Equal(url, "/posts/{id:}")

	url, err = mux.URL("/posts/{id:}", map[string]string{"other-": "id"})
	a.Error(err).Empty(url)

	url, err = mux.URL("/posts/{id:\\\\d+}/author/{page}/", map[string]string{"id": "100", "page": "200"})
	a.NotError(err).Equal(url, "/posts/100/author/200/")
}
