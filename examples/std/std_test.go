// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package std

import (
	"context"
	"net/http"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/assert/v4/rest"

	"github.com/issue9/mux/v7/routertest"
	"github.com/issue9/mux/v7/types"
)

var (
	_ http.Handler = &Router{}
	_ http.Handler = &Routers{}
	_ Middleware   = MiddlewareFunc(func(http.Handler) http.Handler { return nil })
)

func TestRouter(t *testing.T) {
	tt := routertest.NewTester(call, http.NotFoundHandler(), methodNotAllowedBuilder, optionsHandlerBuilder)

	t.Run("params", func(t *testing.T) {
		a := assert.New(t, false)
		tt.Params(a, func(ctx *types.Context) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				p := GetParams(r)
				if p != nil {
					p.Params().Range(func(k, v string) {
						ctx.Set(k, v)
					})
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
	a.NotEqual(WithValue(r, &types.Context{}), r)

	r = rest.Get(a, "/to/path").Request()
	pp := types.NewContext()
	pp.Set("k1", "v1")
	r = WithValue(r, pp)

	pp = types.NewContext()
	pp.Set("k2", "v2")
	r = WithValue(r, pp)
	ps := GetParams(r)
	a.NotNil(ps).
		Equal(ps.Params().MustString("k2", "def"), "v2").
		Equal(ps.Params().MustString("k1", "def"), "v1")
}

func TestGetParams(t *testing.T) {
	a := assert.New(t, false)

	r := rest.Get(a, "/to/path").Request()
	ps := GetParams(r)
	a.Nil(ps)

	c := types.NewContext()
	c.Set("key1", "1")
	r = rest.Get(a, "/to/path").Request()
	ctx := context.WithValue(r.Context(), contextKeyParams, c)
	r = r.WithContext(ctx)
	a.Equal(GetParams(r).Params().MustString("key1", "def"), "1")
}
