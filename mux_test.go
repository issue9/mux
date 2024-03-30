// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/assert/v4/rest"

	"github.com/issue9/mux/v7/internal/tree"
)

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

func TestTrace(t *testing.T) {
	a := assert.New(t, false)

	r := rest.NewRequest(a, http.MethodTrace, "/path").Body([]byte("<body>")).Request()
	w := httptest.NewRecorder()
	a.NotError(Trace(w, r, false))
	body := w.Body.String()
	a.Contains(body, "/path").
		NotContains(body, "body").
		True(strings.HasPrefix(body, http.MethodTrace)).
		Equal(w.Header().Get("content-type"), traceContentType)

	w = httptest.NewRecorder()
	a.NotError(Trace(w, r, true))
	body = w.Body.String()
	a.Contains(body, "/path").
		Contains(body, "&lt;body&gt;").
		True(strings.HasPrefix(body, http.MethodTrace)).
		Equal(w.Header().Get("content-type"), traceContentType)
}
