// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

package ctx

import (
	"testing"

	"github.com/issue9/assert/v5"

	"github.com/issue9/mux/v10"
	"github.com/issue9/mux/v10/routertest"
	"github.com/issue9/mux/v10/types"
)

func TestContextRouter_Params(t *testing.T) {
	tt := routertest.NewTester(func(o ...mux.Option) *routertest.TestRouter[Handler] {
		router := mux.NewRouter[Handler]("test", call, HandlerFunc(notFound), HandlerFunc(trace), methodNotAllowedBuilder, optionsHandlerBuilder, o...)
		return &routertest.TestRouter[Handler]{
			ServeHTTP: router.ServeHTTP,
			Handle:    func(pattern string, h Handler, methods ...string) { router.Handle(pattern, h, nil, methods...) },
			Clean:     router.Clean,
			Remove:    router.Remove,
			URL:       router.URL,
		}
	})

	t.Run("params", func(t *testing.T) {
		a := assert.New(t, false)
		tt.Params(a, func(ctx *types.Route) Handler {
			return HandlerFunc(func(c *CTX) {
				if c.P != nil {
					for k, v := range c.P.Params() {
						ctx.Set(k, v)
					}
				}
			})
		})
	})

	t.Run("serve", func(t *testing.T) {
		a := assert.New(t, false)
		tt.Serve(a, func(status int) Handler {
			return HandlerFunc(func(c *CTX) {
				c.W.WriteHeader(status)
			})
		})
	})
}
