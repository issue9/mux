// SPDX-License-Identifier: MIT

package mux_test

import (
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/internal/tree"
)

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
