// SPDX-License-Identifier: MIT

package mux_test

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/routertest"
)

type (
	ctx struct {
		R *http.Request
		W http.ResponseWriter
		P mux.Params
	}
	ctxHandlerFunc func(ctx *ctx)
)

func contextCall(w http.ResponseWriter, r *http.Request, ps mux.Params, h ctxHandlerFunc) {
	h(&ctx{R: r, W: w, P: ps})
}

func TestDefaultRouter(t *testing.T) {
	a := assert.New(t, false)
	tt := routertest.NewTester[http.Handler](mux.DefaultCall)

	a.Run("params", func(a *assert.Assertion) {
		tt.Params(a, func(ps *mux.Params) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				p := mux.GetParams(r)
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

func TestContextRouter_Params(t *testing.T) {
	a := assert.New(t, false)
	tt := routertest.NewTester[ctxHandlerFunc](contextCall)

	a.Run("params", func(a *assert.Assertion) {
		tt.Params(a, func(ps *mux.Params) ctxHandlerFunc {
			return func(c *ctx) {
				if c.P != nil {
					c.P.Range(func(k, v string) {
						(*ps).Set(k, v)
					})
				}
			}
		})
	})

	a.Run("serve", func(a *assert.Assertion) {
		tt.Serve(a, func(status int) ctxHandlerFunc {
			return func(c *ctx) {
				c.W.WriteHeader(status)
			}
		})
	})
}
