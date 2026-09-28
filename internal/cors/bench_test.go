// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v5"
	"github.com/issue9/assert/v5/rest"
	"github.com/issue9/mux/v10/header"
	"github.com/issue9/mux/v10/internal/syntax"
	"github.com/issue9/mux/v10/internal/tree"
	"github.com/issue9/mux/v10/types"
)

func BenchmarkCORS_Handle(b *testing.B) {
	a := assert.New(b, false)
	tr := tree.NewTestTree(a, false, nil, syntax.NewInterceptors())
	a.NotError(tr.Add("/path", nil, nil, http.MethodGet, http.MethodDelete))
	ctx := types.NewContext()
	ctx.Path = "/path"
	node, _, exists := tr.Handler(ctx, http.MethodGet)
	a.NotNil(node).Zero(ctx.Count()).True(exists)

	b.Run("deny", func(b *testing.B) {
		c := &CORS{}
		a.NotError(c.Sanitize())
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		for b.Loop() {
			c.Handle(node, w.Header(), r)
		}
		a.Empty(w.Header().Get(header.AccessControlAllowOrigin))
	})

	b.Run("allowed-1", func(b *testing.B) {
		c := &CORS{MaxAge: 3600, Origins: []string{"*"}, AllowHeaders: []string{"*"}}
		a.NotError(c.Sanitize())
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()

		for b.Loop() {
			c.Handle(node, w.Header(), r)
		}

		a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "*")
		// 非预检，没有此报头
		a.Empty(w.Header().Get(header.AccessControlAllowMethods)).
			Empty(w.Header().Get(header.AccessControlMaxAge)).
			Empty(w.Header().Get(header.AccessControlAllowHeaders))
	})

	b.Run("allowed-2", func(b *testing.B) {
		c := &CORS{MaxAge: 3600, Origins: []string{"*"}, AllowHeaders: []string{"*"}}
		a.NotError(c.Sanitize())

		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Header(header.Origin, "http://example.com").Request()

		for b.Loop() {
			c.Handle(node, w.Header(), r)
		}

		a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "*")
		// 非预检，没有此报头
		a.Empty(w.Header().Get(header.AccessControlAllowMethods)).
			Empty(w.Header().Get(header.AccessControlMaxAge)).
			Empty(w.Header().Get(header.AccessControlAllowHeaders))
	})

	b.Run("allowed-3", func(b *testing.B) {
		c := &CORS{MaxAge: 3600, Origins: []string{"*"}, AllowHeaders: []string{"*"}}
		a.NotError(c.Sanitize())

		w := httptest.NewRecorder()
		r := rest.NewRequest(a, http.MethodOptions, "/path").Header(header.Origin, "http://example.com").Request()

		for b.Loop() {
			c.Handle(node, w.Header(), r)
		}

		a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "*")
		// 非预检，没有此报头
		a.Empty(w.Header().Get(header.AccessControlAllowMethods)).
			Empty(w.Header().Get(header.AccessControlMaxAge)).
			Empty(w.Header().Get(header.AccessControlAllowHeaders))
	})

	b.Run("preflight-1", func(b *testing.B) {
		c := &CORS{MaxAge: 3600, Origins: []string{"*"}, AllowHeaders: []string{"*"}}
		a.NotError(c.Sanitize())

		w := httptest.NewRecorder()
		r := rest.NewRequest(a, http.MethodOptions, "/path").
			Header(header.Origin, "http://example.com").
			Header(header.AccessControlRequestMethod, "GET").
			Request()

		for b.Loop() {
			c.Handle(node, w.Header(), r)
		}

		a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "*")
		a.Equal(w.Header().Get(header.AccessControlAllowMethods), "DELETE, GET, HEAD, OPTIONS")
	})

	// preflight，但是方法不被允许
	b.Run("preflight-2", func(b *testing.B) {
		c := &CORS{MaxAge: 3600, Origins: []string{"*"}, AllowHeaders: []string{"*"}}
		a.NotError(c.Sanitize())

		w := httptest.NewRecorder()
		r := rest.NewRequest(a, http.MethodOptions, "/path").
			Header(header.Origin, "http://example.com").
			Header(header.AccessControlRequestMethod, "PATCH").
			Request()

		for b.Loop() {
			c.Handle(node, w.Header(), r)
		}

		a.Equal(w.Header().Get(header.AccessControlAllowOrigin), "")
		a.Equal(w.Header().Get(header.AccessControlAllowMethods), "")
	})
}
