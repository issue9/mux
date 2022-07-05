// SPDX-License-Identifier: MIT

package std

import (
	"context"
	"net/http"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"

	"github.com/issue9/mux/v7"
	"github.com/issue9/mux/v7/internal/params"
	"github.com/issue9/mux/v7/routertest"
	"github.com/issue9/mux/v7/types"
)

var (
	_ http.Handler = &Router{}
	_ Middleware   = MiddlewareFunc(func(http.Handler) http.Handler { return nil })
)

func TestRouter(t *testing.T) {
	a := assert.New(t, false)
	tt := routertest.NewTester(call, http.NotFoundHandler(), methodNotAllowedBuilder, optionsHandlerBuilder)

	a.Run("params", func(a *assert.Assertion) {
		tt.Params(a, func(ps *types.Params) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				p := GetParams(r)
				if p != nil {
					p.Range(func(k, v string) {
						(*ps).Set(k, v)
					})
				}
			})
		})
	})

	a.Run("serve", func(a *assert.Assertion) {
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
	a.Equal(WithValue(r, &params.Params{}), r)

	r = rest.Get(a, "/to/path").Request()
	pp := mux.NewParams()
	pp.Set("k1", "v1")
	r = WithValue(r, pp)

	pp = mux.NewParams()
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

	kvs := []params.Param{{K: "key1", V: "1"}}
	r = rest.Get(a, "/to/path").Request()
	ctx := context.WithValue(r.Context(), contextKeyParams, &params.Params{Params: kvs})
	r = r.WithContext(ctx)
	a.Equal(GetParams(r).MustString("key1", "def"), "1")
}
