// SPDX-License-Identifier: MIT

package mux

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"

	"github.com/issue9/mux/v6/internal/tree"
)

func buildMiddleware(a *assert.Assertion, text string) MiddlewareFuncOf[http.Handler] {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r) // 先输出被包含的内容
			_, err := w.Write([]byte(text))
			a.NotError(err)
		})
	}
}

func TestRouter_Middleware(t *testing.T) {
	a := assert.New(t, false)

	def := NewRouter("",
		[]MiddlewareFuncOf[http.Handler]{
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

func TestOption(t *testing.T) {
	a := assert.New(t, false)

	r := NewRouter("", nil)
	a.NotNil(r).
		False(r.options.CaseInsensitive).
		NotNil(r.options.MethodNotAllowed)

	r = NewRouter("", nil, CaseInsensitive)
	a.NotNil(r).
		True(r.options.CaseInsensitive).
		NotNil(r.options.MethodNotAllowed)

	notFound := rest.BuildHandler(a, 404, "", nil)
	methodNotAllowed := rest.BuildHandler(a, 405, "", nil)
	r = NewRouter("", nil, NotFound(notFound), MethodNotAllowed(methodNotAllowed))
	a.NotNil(r).
		False(r.options.CaseInsensitive).
		Equal(r.options.MethodNotAllowed, methodNotAllowed).
		Equal(r.options.NotFound, notFound)

	r = NewRouter("", nil, CORS([]string{"https://example.com"}, nil, nil, 3600, false))
	a.NotNil(r).
		Equal(r.options.CORS.Origins, []string{"https://example.com"}).
		Nil(r.options.CORS.AllowHeaders).
		Equal(r.options.CORS.MaxAge, 3600)

	r = NewRouter("", nil, CORS([]string{"https://example.com"}, nil, nil, 0, true))
	a.NotNil(r)

	a.Panic(func() {
		r = NewRouter("", nil, CORS([]string{"*"}, nil, nil, 0, true))
	})
}

func TestRecovery(t *testing.T) {
	a := assert.New(t, false)

	p := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic("test") })

	router := NewRouter("", nil)
	a.NotNil(router).Nil(router.options.RecoverFunc)
	router.Get("/path", p)
	a.Panic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
	})

	// WriterRecovery
	out := new(bytes.Buffer)
	router = NewRouter("", nil, WriterRecovery(404, out))
	a.NotNil(router).NotNil(router.options.RecoverFunc)
	router.Get("/path", p)
	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
		a.Contains(out.String(), "test").
			Equal(w.Code, 404)
	})

	// LogRecovery
	out = new(bytes.Buffer)
	l := log.New(out, "test:", 0)
	router = NewRouter("", nil, LogRecovery(405, l))
	a.NotNil(router).NotNil(router.options.RecoverFunc)
	router.Get("/path", p)
	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
		a.Equal(405, w.Code).
			Contains(out.String(), "test")
	})

	// HTTPRecovery
	router = NewRouter("", nil, HTTPRecovery(406))
	a.NotNil(router).NotNil(router.options.RecoverFunc)
	router.Get("/path", p)
	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
		a.Equal(w.Code, 406)
	})
}

func TestMethods(t *testing.T) {
	a := assert.New(t, false)
	a.Equal(Methods(), tree.Methods)
}

func TestCheckSyntax(t *testing.T) {
	a := assert.New(t, false)

	a.NotError(CheckSyntax("/{path"))
	a.NotError(CheckSyntax("/path}"))
	a.Error(CheckSyntax(""))
}

func TestURL(t *testing.T) {
	a := assert.New(t, false)

	url, err := URL("/posts/{id:}", map[string]string{"id": "100"})
	a.NotError(err).Equal(url, "/posts/100")

	url, err = URL("/posts/{id:}", nil)
	a.NotError(err).Equal(url, "/posts/{id:}")

	url, err = URL("/posts/{id:}", map[string]string{"other-": "id"})
	a.Error(err).Empty(url)

	url, err = URL("/posts/{id:\\\\d+}/author/{page}/", map[string]string{"id": "100", "page": "200"})
	a.NotError(err).Equal(url, "/posts/100/author/200/")
}
