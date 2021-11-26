// SPDX-License-Identifier: MIT

package mux

import (
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
	a.Error(err)

	url, err = URL("/posts/{id:\\\\d+}/author/{page}/", map[string]string{"id": "100", "page": "200"})
	a.NotError(err).Equal(url, "/posts/100/author/200/")
}
