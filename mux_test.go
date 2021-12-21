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

	"github.com/issue9/mux/v5/internal/tree"
)

func TestOption(t *testing.T) {
	a := assert.New(t, false)

	r := NewRouter("")
	a.NotNil(r).
		False(r.options.CaseInsensitive).
		NotNil(r.options.MethodNotAllowed)

	r = NewRouter("", CaseInsensitive)
	a.NotNil(r).
		True(r.options.CaseInsensitive).
		NotNil(r.options.MethodNotAllowed)

	notFound := rest.BuildHandler(a, 404, "", nil)
	methodNotAllowed := rest.BuildHandler(a, 405, "", nil)
	r = NewRouter("", NotFound(notFound), MethodNotAllowed(methodNotAllowed))
	a.NotNil(r).
		False(r.options.CaseInsensitive).
		Equal(r.options.MethodNotAllowed, methodNotAllowed).
		Equal(r.options.NotFound, notFound)

	r = NewRouter("", CORS([]string{"https://example.com"}, nil, nil, 3600, false))
	a.NotNil(r).
		Equal(r.options.CORS.Origins, []string{"https://example.com"}).
		Nil(r.options.CORS.AllowHeaders).
		Equal(r.options.CORS.MaxAge, 3600)

	r = NewRouter("", CORS([]string{"https://example.com"}, nil, nil, 0, true))
	a.NotNil(r)

	a.Panic(func() {
		r = NewRouter("", CORS([]string{"*"}, nil, nil, 0, true))
	})
}

func TestRecovery(t *testing.T) {
	a := assert.New(t, false)

	p := func(w http.ResponseWriter, r *http.Request) { panic("test") }

	router := NewRouter("")
	a.NotNil(router).Nil(router.options.RecoverFunc)
	router.GetFunc("/path", p)
	a.Panic(func() {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, "/path", nil)
		a.NotError(err).NotNil(r)
		router.ServeHTTP(w, r)
	})

	// WriterRecovery
	out := new(bytes.Buffer)
	router = NewRouter("", WriterRecovery(404, out))
	a.NotNil(router).NotNil(router.options.RecoverFunc)
	router.GetFunc("/path", p)
	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, "/path", nil)
		a.NotError(err).NotNil(r)
		router.ServeHTTP(w, r)
		a.Contains(out.String(), "test").
			Equal(w.Code, 404)
	})

	// LogRecovery
	out = new(bytes.Buffer)
	l := log.New(out, "test:", 0)
	router = NewRouter("", LogRecovery(405, l))
	a.NotNil(router).NotNil(router.options.RecoverFunc)
	router.GetFunc("/path", p)
	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, "/path", nil)
		a.NotError(err).NotNil(r)
		router.ServeHTTP(w, r)
		a.Equal(405, w.Code).
			Contains(out.String(), "test")
	})

	// HTTPRecovery
	router = NewRouter("", HTTPRecovery(406))
	a.NotNil(router).NotNil(router.options.RecoverFunc)
	router.GetFunc("/path", p)
	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, "/path", nil)
		a.NotError(err).NotNil(r)
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
