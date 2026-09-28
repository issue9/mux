// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

package std

import (
	"context"
	"net/http"
	"testing"

	"github.com/issue9/assert/v5"
	"github.com/issue9/assert/v5/rest"

	"github.com/issue9/mux/v10/routertest"
	"github.com/issue9/mux/v10/types"
)

var (
	_ http.Handler = &Router{}
	_ http.Handler = &Routers{}
	_ Middleware   = MiddlewareFunc(func(_ http.Handler, _, _, _ string) http.Handler { return nil })
)

func TestRouter(t *testing.T) {
	tt := routertest.NewTester[http.Handler](call, http.NotFoundHandler(), http.HandlerFunc(trace), methodNotAllowedBuilder, optionsHandlerBuilder)

	t.Run("params", func(t *testing.T) {
		a := assert.New(t, false)
		tt.Params(a, func(ctx *types.Route) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				p := GetParams(r)
				if p != nil {
					for k, v := range p.Params() {
						ctx.Set(k, v)
					}
				}
			})
		})
	})

	t.Run("serve", func(t *testing.T) {
		a := assert.New(t, false)
		tt.Serve(a, func(status int) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
			})
		})
	})
}

func TestWithValue(t *testing.T) {
	a := assert.New(t, false)

	r := rest.Get(a, "/to/path").Request()
	a.NotEqual(WithValue(r, &types.Route{}), r)

	r = rest.Get(a, "/to/path").Request()
	pp := types.NewRoute()
	pp.Set("k1", "v1")
	r = WithValue(r, pp)

	pp = types.NewRoute()
	pp.Set("k2", "v2")
	r = WithValue(r, pp)
	ps := GetParams(r)
	a.NotNil(ps).
		Equal(ps.MustString("k2", "def"), "v2").
		Equal(ps.MustString("k1", "def"), "v1")
}

func TestGetParams(t *testing.T) {
	a := assert.New(t, false)

	r := rest.Get(a, "/to/path").Request()
	ps := GetParams(r)
	a.Nil(ps)

	c := types.NewRoute()
	c.Set("key1", "1")
	r = rest.Get(a, "/to/path").Request()
	ctx := context.WithValue(r.Context(), contextKeyParams, c)
	r = r.WithContext(ctx)
	a.Equal(GetParams(r).MustString("key1", "def"), "1")
}
