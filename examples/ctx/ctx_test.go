// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

package ctx

import (
	"testing"

	"github.com/issue9/assert/v5"

	"github.com/issue9/mux/v10/routertest"
	"github.com/issue9/mux/v10/types"
)

func TestContextRouter_Params(t *testing.T) {
	tt := routertest.NewTester[Handler](call, HandlerFunc(notFound), HandlerFunc(trace), methodNotAllowedBuilder, optionsHandlerBuilder)

	t.Run("params", func(t *testing.T) {
		a := assert.New(t, false)
		tt.Params(a, func(ctx *types.Context) Handler {
			return HandlerFunc(func(c *CTX) {
				if c.P != nil {
					for k, v := range c.P.Params().IterSeq() {
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
