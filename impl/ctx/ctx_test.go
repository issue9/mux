// SPDX-License-Identifier: MIT

package ctx

import (
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/routertest"
)

func TestContextRouter_Params(t *testing.T) {
	a := assert.New(t, false)
	tt := routertest.NewTester(call, optionsHandlerBuilder)

	a.Run("params", func(a *assert.Assertion) {
		tt.Params(a, func(ps *mux.Params) Handler {
			return HandlerFunc(func(c *CTX) {
				if c.P != nil {
					c.P.Range(func(k, v string) {
						(*ps).Set(k, v)
					})
				}
			})
		})
	})

	a.Run("serve", func(a *assert.Assertion) {
		tt.Serve(a, func(status int) Handler {
			return HandlerFunc(func(c *CTX) {
				c.W.WriteHeader(status)
			})
		})
	})
}
